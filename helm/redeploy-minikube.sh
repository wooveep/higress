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
VALUES_FILE="${VALUES_FILE:-${CHART_DIR}/values-local-minikube.yaml}"
NAMESPACE="${NAMESPACE:-aigateway-system}"
RELEASE_NAME="${RELEASE_NAME:-aigateway}"
BUILD_COMPONENTS="${BUILD_COMPONENTS:-aigateway,controller,gateway,pilot,console,portal,plugins,plugin-server}"
HELM_TIMEOUT="${HELM_TIMEOUT:-15m}"
MINIKUBE_PROFILE="${MINIKUBE_PROFILE:-}"

SKIP_BUILD=false
SKIP_LOAD=false
SKIP_DEPLOY=false

declare -a VALUES_FILES=()

usage() {
  cat <<'EOF'
Usage:
  redeploy-minikube.sh [options]

Options:
  --values <path>         Primary Helm values file
  --extra-values <path>   Additional Helm values file, applied in order
  --namespace <name>      Kubernetes namespace (default: aigateway-system)
  --release <name>        Helm release name (default: aigateway)
  --components <list>     Components passed to build-local-images.sh
  --timeout <duration>    Helm/kubectl rollout timeout (default: 15m)
  --profile <name>        Minikube profile name (default: current profile)
  --skip-build            Skip local image build
  --skip-load             Skip minikube image load
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
    --extra-values)
      VALUES_FILES+=("$2")
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
    --profile)
      MINIKUBE_PROFILE="$2"
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

FINAL_VALUES_FILES=("${VALUES_FILE}")
for extra_file in "${VALUES_FILES[@]}"; do
  extra_file="$(cd -- "$(dirname -- "${extra_file}")" 2>/dev/null && pwd -P)/$(basename "${extra_file}")"
  [[ -f "${extra_file}" ]] || dev_die "Values file not found: ${extra_file}"
  FINAL_VALUES_FILES+=("${extra_file}")
done

dev_need_cmd docker
dev_need_cmd helm
dev_need_cmd kubectl
dev_need_cmd minikube

MINIKUBE_ARGS=()
if [[ -n "${MINIKUBE_PROFILE}" ]]; then
  MINIKUBE_ARGS=(-p "${MINIKUBE_PROFILE}")
fi

declare -A FALLBACK_IMAGE_REPO=(
  ["aigateway/grafana"]="grafana/grafana"
  ["aigateway/prometheus"]="prom/prometheus"
  ["aigateway/loki"]="grafana/loki"
  ["aigateway/promtail"]="grafana/promtail"
  ["bitnamilegacy/redis"]="bitnamilegacy/redis"
  ["bitnamilegacy/redis-sentinel"]="bitnamilegacy/redis-sentinel"
  ["bitnamilegacy/postgresql-repmgr"]="bitnamilegacy/postgresql-repmgr"
  ["bitnamilegacy/pgpool"]="bitnamilegacy/pgpool"
  ["bitnami/redis"]="bitnami/redis"
  ["bitnami/redis-sentinel"]="bitnami/redis-sentinel"
  ["bitnami/postgresql-repmgr"]="bitnami/postgresql-repmgr"
  ["bitnami/pgpool"]="bitnami/pgpool"
)

run() {
  echo "+ $*"
  "$@"
}

force_remove_minikube_image() {
  local image="$1"
  local quoted cleanup_cmd

  quoted="$(printf '%q' "${image}")"
  cleanup_cmd="if command -v docker >/dev/null 2>&1; then docker image rm -f ${quoted} >/dev/null 2>&1 || true; elif command -v nerdctl >/dev/null 2>&1; then nerdctl --namespace k8s.io image rm -f ${quoted} >/dev/null 2>&1 || true; elif command -v crictl >/dev/null 2>&1; then crictl rmi ${quoted} >/dev/null 2>&1 || true; fi"
  run minikube "${MINIKUBE_ARGS[@]}" ssh -- "${cleanup_cmd}"
}

resolve_value() {
  local path="$1"
  yaml_get_scalar_from_files "${path}" "${FINAL_VALUES_FILES[@]}"
}

value_is_true() {
  local value
  value="$(printf '%s' "${1:-}" | tr '[:upper:]' '[:lower:]')"
  [[ "${value}" == "true" || "${value}" == "yes" || "${value}" == "on" || "${value}" == "1" ]]
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
    if value_is_true "$(resolve_value "higress-core.redis.enabled")"; then
      append_image \
        "$(resolve_value "higress-core.redis.image.repository")$( [[ -z "$(resolve_value "higress-core.redis.image.repository")" ]] && printf '%s' "$(resolve_value "higress-core.redis.redis.repository")" )" \
        "$(resolve_value "higress-core.redis.image.tag")$( [[ -z "$(resolve_value "higress-core.redis.image.tag")" ]] && printf '%s' "$(resolve_value "higress-core.redis.redis.tag")" )"
      if value_is_true "$(resolve_value "higress-core.redis.sentinel.enabled")"; then
        append_image \
          "$(resolve_value "higress-core.redis.sentinel.image.repository")" \
          "$(resolve_value "higress-core.redis.sentinel.image.tag")"
      fi
    fi
    if value_is_true "$(resolve_value "higress-core.postgresql.enabled")"; then
      append_image \
        "$(resolve_value "higress-core.postgresql.postgresql.image.repository")" \
        "$(resolve_value "higress-core.postgresql.postgresql.image.tag")"
      append_image \
        "$(resolve_value "higress-core.postgresql.pgpool.image.repository")" \
        "$(resolve_value "higress-core.postgresql.pgpool.image.tag")"
    fi
    append_image "$(resolve_value "aigateway-console.image.repository")" "$(resolve_value "aigateway-console.image.tag")"
    append_image "$(resolve_value "global.o11y.grafana.image.repository")" "$(resolve_value "global.o11y.grafana.image.tag")"
    append_image "$(resolve_value "global.o11y.prometheus.image.repository")" "$(resolve_value "global.o11y.prometheus.image.tag")"
    append_image "$(resolve_value "global.o11y.loki.image.repository")" "$(resolve_value "global.o11y.loki.image.tag")"
    append_image "$(resolve_value "global.o11y.promtail.image.repository")" "$(resolve_value "global.o11y.promtail.image.tag")"
    append_image \
      "$(resolve_value "aigateway-portal.backend.image.repository")$( [[ -z "$(resolve_value "aigateway-portal.backend.image.repository")" ]] && printf '%s' "$(resolve_value "aigateway-portal.image.repository")" )" \
      "$(resolve_value "aigateway-portal.backend.image.tag")$( [[ -z "$(resolve_value "aigateway-portal.backend.image.tag")" ]] && printf '%s' "$(resolve_value "aigateway-portal.image.tag")" )"
    if value_is_true "$(resolve_value "aigateway-portal.mysql.enabled")"; then
      append_image \
        "$(resolve_value "aigateway-portal.mysql.image.repository")$( [[ -z "$(resolve_value "aigateway-portal.mysql.image.repository")" ]] && printf 'mariadb' )" \
        "$(resolve_value "aigateway-portal.mysql.image.tag")$( [[ -z "$(resolve_value "aigateway-portal.mysql.image.tag")" ]] && printf '11.4' )"
    fi
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

echo "Using chart dir    : ${CHART_DIR}"
printf 'Using values files :\n'
for file in "${FINAL_VALUES_FILES[@]}"; do
  echo "  - ${file}"
done

if [[ "${SKIP_BUILD}" != "true" ]]; then
  run "${SCRIPT_DIR}/build-local-images.sh" --components "${BUILD_COMPONENTS}"
fi

if [[ "${SKIP_LOAD}" != "true" ]]; then
  mapfile -t IMAGES < <(resolve_images)
  [[ ${#IMAGES[@]} -gt 0 ]] || dev_die "No images resolved from values files"

  echo "Loading images into minikube (${#IMAGES[@]} images)..."
  for image in "${IMAGES[@]}"; do
    ensure_local_image "${image}"
    force_remove_minikube_image "${image}"
    run minikube "${MINIKUBE_ARGS[@]}" image load --overwrite=true "${image}"
  done
fi

if [[ "${SKIP_DEPLOY}" != "true" ]]; then
  helm_args=()
  for file in "${FINAL_VALUES_FILES[@]}"; do
    helm_args+=(-f "${file}")
  done

  run helm dependency build "${CHART_DIR}"
  run helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
    -n "${NAMESPACE}" \
    --create-namespace \
    --render-subchart-notes \
    "${helm_args[@]}" \
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

echo "Redeploy complete."
echo "  release   : ${RELEASE_NAME}"
echo "  namespace : ${NAMESPACE}"
