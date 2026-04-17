#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
source "${ROOT_DIR}/scripts/dev-shell-lib.sh"

LOCAL_CHART_DIR="${SCRIPT_DIR}/higress"
LEGACY_CHART_DIR="${ROOT_DIR}/helm/higress"
if [[ -d "${LOCAL_CHART_DIR}" ]]; then
  DEFAULT_CHART_DIR="${LOCAL_CHART_DIR}"
else
  DEFAULT_CHART_DIR="${LEGACY_CHART_DIR}"
fi

CHART_DIR="${CHART_DIR:-${DEFAULT_CHART_DIR}}"
DEFAULT_VALUES_FILE="${CHART_DIR}/values-production-k3d.yaml"
if [[ ! -f "${DEFAULT_VALUES_FILE}" ]]; then
  DEFAULT_VALUES_FILE="${CHART_DIR}/values-production-gray.yaml"
fi

VALUES_FILE="${VALUES_FILE:-${DEFAULT_VALUES_FILE}}"
NAMESPACE="${NAMESPACE:-aigateway-system}"
RELEASE_NAME="${RELEASE_NAME:-aigateway}"
BUILD_COMPONENTS="${BUILD_COMPONENTS:-aigateway,controller,gateway,pilot,console,portal,plugins,plugin-server}"
HELM_TIMEOUT="${HELM_TIMEOUT:-15m}"
K3D_CLUSTER="${K3D_CLUSTER:-}"

SKIP_BUILD=false
SKIP_LOAD=false
SKIP_DEPLOY=false

usage() {
  cat <<'EOF'
Usage:
  redeploy-k3d.sh [options]

Options:
  --values <path>         Helm values file
  --namespace <name>      Kubernetes namespace (default: aigateway-system)
  --release <name>        Helm release name (default: aigateway)
  --components <list>     Components passed to build-local-images.sh
  --timeout <duration>    Helm/kubectl rollout timeout (default: 15m)
  --cluster <name>        k3d cluster name (default: inferred from current context)
  --skip-build            Skip local image build
  --skip-load             Skip k3d image import
  --skip-deploy           Skip helm upgrade + rollout restart
  -h, --help              Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --values)
      VALUES_FILE="$2"
      shift 2
      ;;
    --namespace)
      NAMESPACE="$2"
      shift 2
      ;;
    --release)
      RELEASE_NAME="$2"
      shift 2
      ;;
    --components)
      BUILD_COMPONENTS="$2"
      shift 2
      ;;
    --timeout)
      HELM_TIMEOUT="$2"
      shift 2
      ;;
    --cluster)
      K3D_CLUSTER="$2"
      shift 2
      ;;
    --skip-build)
      SKIP_BUILD=true
      shift
      ;;
    --skip-load)
      SKIP_LOAD=true
      shift
      ;;
    --skip-deploy)
      SKIP_DEPLOY=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      dev_die "Unknown argument: $1"
      ;;
  esac
done

[[ -d "${CHART_DIR}" ]] || dev_die "Chart directory not found: ${CHART_DIR}"
CHART_DIR="$(cd -- "${CHART_DIR}" && pwd -P)"
VALUES_FILE="$(cd -- "$(dirname -- "${VALUES_FILE}")" 2>/dev/null && pwd -P)/$(basename "${VALUES_FILE}")"
[[ -f "${VALUES_FILE}" ]] || dev_die "Values file not found: ${VALUES_FILE}"

dev_need_cmd docker
dev_need_cmd helm
dev_need_cmd kubectl
dev_need_cmd k3d
dev_need_cmd jq

run() {
  echo "+ $*"
  "$@"
}

resolve_k3d_cluster() {
  local context inferred

  if [[ -n "${K3D_CLUSTER}" ]]; then
    return 0
  fi

  context="$(kubectl config current-context 2>/dev/null || true)"
  if [[ "${context}" == k3d-* ]]; then
    K3D_CLUSTER="${context#k3d-}"
    return 0
  fi

  inferred="$(k3d cluster list -o json 2>/dev/null | jq -r 'if length == 1 then .[0].name else "" end')"
  [[ -n "${inferred}" ]] || dev_die "Unable to infer k3d cluster. Please provide --cluster <name>."
  K3D_CLUSTER="${inferred}"
}

verify_context_matches_cluster() {
  local current expected
  current="$(kubectl config current-context 2>/dev/null || true)"
  expected="k3d-${K3D_CLUSTER}"
  [[ "${current}" == "${expected}" ]] || dev_die "kubectl context mismatch: current='${current}', expected='${expected}'. Run: kubectl config use-context ${expected}"
}

declare -A FALLBACK_IMAGE_REPO=(
  ["aigateway/grafana"]="grafana/grafana"
  ["aigateway/prometheus"]="prom/prometheus"
  ["aigateway/loki"]="grafana/loki"
  ["aigateway/promtail"]="grafana/promtail"
)

resolve_value() {
  local path="$1"
  yaml_get_scalar "${VALUES_FILE}" "${path}"
}

append_image() {
  local repository="$1"
  local tag="$2"
  [[ -n "${repository}" && -n "${tag}" ]] || return 0
  printf '%s:%s\n' "${repository}" "${tag}"
}

resolve_images() {
  {
    append_image "$(resolve_value "higress-core.gateway.repository")" "$(resolve_value "higress-core.gateway.tag")"
    append_image "$(resolve_value "higress-core.controller.repository")" "$(resolve_value "higress-core.controller.tag")"
    append_image "$(resolve_value "higress-core.pilot.repository")" "$(resolve_value "higress-core.pilot.tag")"
    append_image "$(resolve_value "higress-core.pluginServer.repository")" "$(resolve_value "higress-core.pluginServer.tag")"
    append_image "$(resolve_value "higress-core.redis.redis.repository")" "$(resolve_value "higress-core.redis.redis.tag")"
    append_image "$(resolve_value "aigateway-console.image.repository")" "$(resolve_value "aigateway-console.image.tag")"
    append_image "$(resolve_value "global.o11y.grafana.image.repository")" "$(resolve_value "global.o11y.grafana.image.tag")"
    append_image "$(resolve_value "global.o11y.prometheus.image.repository")" "$(resolve_value "global.o11y.prometheus.image.tag")"
    append_image "$(resolve_value "global.o11y.loki.image.repository")" "$(resolve_value "global.o11y.loki.image.tag")"
    append_image "$(resolve_value "global.o11y.promtail.image.repository")" "$(resolve_value "global.o11y.promtail.image.tag")"
    append_image \
      "$(yaml_get_scalar_from_files "aigateway-portal.backend.image.repository" "${VALUES_FILE}")$( [[ -z "$(resolve_value "aigateway-portal.backend.image.repository")" ]] && printf '%s' "$(resolve_value "aigateway-portal.image.repository")" )" \
      "$(yaml_get_scalar_from_files "aigateway-portal.backend.image.tag" "${VALUES_FILE}")$( [[ -z "$(resolve_value "aigateway-portal.backend.image.tag")" ]] && printf '%s' "$(resolve_value "aigateway-portal.image.tag")" )"
    append_image \
      "$(resolve_value "aigateway-portal.mysql.image.repository")$( [[ -z "$(resolve_value "aigateway-portal.mysql.image.repository")" ]] && printf 'mariadb' )" \
      "$(resolve_value "aigateway-portal.mysql.image.tag")$( [[ -z "$(resolve_value "aigateway-portal.mysql.image.tag")" ]] && printf '11.4' )"
  } | awk 'NF { images[$0] = 1 } END { for (image in images) print image }' | sort
}

ensure_local_image() {
  local image="$1"
  local repository tag fallback_repo fallback_image

  if docker image inspect "${image}" >/dev/null 2>&1; then
    return 0
  fi

  repository="${image%:*}"
  tag="${image##*:}"
  fallback_repo="${FALLBACK_IMAGE_REPO[${repository}]:-}"
  [[ -n "${fallback_repo}" ]] || dev_die "Local image not found and no fallback mapping configured: ${image}"

  fallback_image="${fallback_repo}:${tag}"
  if ! docker image inspect "${fallback_image}" >/dev/null 2>&1; then
    run docker pull "${fallback_image}"
  fi
  run docker tag "${fallback_image}" "${image}"
}

resolve_k3d_cluster
verify_context_matches_cluster

echo "Using chart dir  : ${CHART_DIR}"
echo "Using values file: ${VALUES_FILE}"
echo "Using k3d cluster: ${K3D_CLUSTER}"

if [[ "${SKIP_BUILD}" != "true" ]]; then
  run env WRAPPER_VALUES_FILE="${VALUES_FILE}" "${SCRIPT_DIR}/build-local-images.sh" --components "${BUILD_COMPONENTS}"
fi

if [[ "${SKIP_LOAD}" != "true" ]]; then
  mapfile -t IMAGES < <(resolve_images)
  [[ ${#IMAGES[@]} -gt 0 ]] || dev_die "No images resolved from ${VALUES_FILE}"

  echo "Importing images into k3d (${#IMAGES[@]} images)..."
  for image in "${IMAGES[@]}"; do
    ensure_local_image "${image}"
    run k3d image import "${image}" -c "${K3D_CLUSTER}"
  done
fi

if [[ "${SKIP_DEPLOY}" != "true" ]]; then
  run helm dependency build "${CHART_DIR}"
  run helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
    -n "${NAMESPACE}" \
    --create-namespace \
    --render-subchart-notes \
    -f "${VALUES_FILE}" \
    --wait \
    --timeout "${HELM_TIMEOUT}"

  DEPLOYMENTS="$(kubectl -n "${NAMESPACE}" get deploy -l "app.kubernetes.io/managed-by=Helm" -o name || true)"
  STATEFULSETS="$(kubectl -n "${NAMESPACE}" get statefulset -l "app.kubernetes.io/managed-by=Helm" -o name || true)"

  if [[ -n "${DEPLOYMENTS}" ]]; then
    run kubectl -n "${NAMESPACE}" rollout restart deploy -l "app.kubernetes.io/managed-by=Helm"
    while IFS= read -r item; do
      [[ -z "${item}" ]] && continue
      run kubectl -n "${NAMESPACE}" rollout status "${item}" --timeout "${HELM_TIMEOUT}"
    done <<< "${DEPLOYMENTS}"
  fi

  if [[ -n "${STATEFULSETS}" ]]; then
    run kubectl -n "${NAMESPACE}" rollout restart statefulset -l "app.kubernetes.io/managed-by=Helm"
    while IFS= read -r item; do
      [[ -z "${item}" ]] && continue
      run kubectl -n "${NAMESPACE}" rollout status "${item}" --timeout "${HELM_TIMEOUT}"
    done <<< "${STATEFULSETS}"
  fi
fi

echo "k3d redeploy complete."
echo "  release   : ${RELEASE_NAME}"
echo "  namespace : ${NAMESPACE}"
echo "  values    : ${VALUES_FILE}"
echo "  cluster   : ${K3D_CLUSTER}"
