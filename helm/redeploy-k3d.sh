#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
# Prefer the chart colocated with this script (higress/helm/higress).
# Fall back to the legacy monorepo-level helm path when needed.
LOCAL_CHART_DIR="${SCRIPT_DIR}/higress"
LEGACY_CHART_DIR="${ROOT_DIR}/helm/higress"
if [[ -d "${LOCAL_CHART_DIR}" ]]; then
  DEFAULT_CHART_DIR="${LOCAL_CHART_DIR}"
else
  DEFAULT_CHART_DIR="${LEGACY_CHART_DIR}"
fi
CHART_DIR="${CHART_DIR:-${DEFAULT_CHART_DIR}}"
if [[ -d "${CHART_DIR}" ]]; then
  CHART_DIR="$(cd -- "${CHART_DIR}" && pwd -P)"
fi

DEFAULT_VALUES_FILE="${CHART_DIR}/values-production-k3d.yaml"
if [[ ! -f "${DEFAULT_VALUES_FILE}" ]]; then
  DEFAULT_VALUES_FILE="${CHART_DIR}/values-production-gray.yaml"
fi

VALUES_FILE="${VALUES_FILE:-${DEFAULT_VALUES_FILE}}"
NAMESPACE="${NAMESPACE:-aigateway-system}"
RELEASE_NAME="${RELEASE_NAME:-aigateway}"
BUILD_COMPONENTS="${BUILD_COMPONENTS:-aigateway,controller,gateway,pilot,console,portal,plugin-server}"
HELM_TIMEOUT="${HELM_TIMEOUT:-15m}"
K3D_CLUSTER="${K3D_CLUSTER:-}"

SKIP_BUILD=false
SKIP_LOAD=false
SKIP_DEPLOY=false

usage() {
  cat <<'USAGE'
Usage:
  redeploy-k3d.sh [options]

Options:
  --values <path>         Helm values file (default: higress/helm/higress/values-production-k3d.yaml)
  --namespace <name>      Kubernetes namespace (default: aigateway-system)
  --release <name>        Helm release name (default: aigateway)
  --components <list>     Components passed to build-local-images.sh
  --timeout <duration>    Helm/kubectl rollout timeout (default: 15m)
  --cluster <name>        k3d cluster name (default: inferred from current kubectl context)
  --skip-build            Skip local image build
  --skip-load             Skip k3d image import
  --skip-deploy           Skip helm upgrade + rollout restart
  -h, --help              Show this help

Environment overrides:
  CHART_DIR, VALUES_FILE, NAMESPACE, RELEASE_NAME, BUILD_COMPONENTS, HELM_TIMEOUT, K3D_CLUSTER
USAGE
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
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

run() {
  echo "+ $*"
  "$@"
}

resolve_k3d_cluster() {
  local context inferred

  if [[ -n "${K3D_CLUSTER}" ]]; then
    return
  fi

  context="$(kubectl config current-context 2>/dev/null || true)"
  if [[ "${context}" == k3d-* ]]; then
    K3D_CLUSTER="${context#k3d-}"
    return
  fi

  inferred="$(
    k3d cluster list -o json 2>/dev/null | python3 -c '
import json
import sys

try:
    clusters = json.load(sys.stdin)
except Exception:
    print("")
    raise SystemExit(0)

names = [item.get("name", "").strip() for item in clusters if isinstance(item, dict)]
names = [name for name in names if name]

print(names[0] if len(names) == 1 else "")
'
  )"

  if [[ -n "${inferred}" ]]; then
    K3D_CLUSTER="${inferred}"
    return
  fi

  echo "Unable to infer k3d cluster. Please provide --cluster <name>." >&2
  exit 1
}

verify_context_matches_cluster() {
  local current expected
  current="$(kubectl config current-context 2>/dev/null || true)"
  expected="k3d-${K3D_CLUSTER}"

  if [[ "${current}" != "${expected}" ]]; then
    echo "kubectl context mismatch: current='${current}', expected='${expected}'." >&2
    echo "Please run: kubectl config use-context ${expected}" >&2
    exit 1
  fi
}

declare -A FALLBACK_IMAGE_REPO=(
  ["aigateway/grafana"]="grafana/grafana"
  ["aigateway/prometheus"]="prom/prometheus"
  ["aigateway/loki"]="grafana/loki"
  ["aigateway/promtail"]="grafana/promtail"
)

resolve_images() {
  python3 - "${VALUES_FILE}" <<'PY'
import sys
import yaml

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    values = yaml.safe_load(f) or {}


def get(data, *keys, default=None):
    cur = data
    for key in keys:
        if not isinstance(cur, dict) or key not in cur:
            return default
        cur = cur[key]
    return cur


def add(images, repository, tag):
    if repository and tag:
        images.add(f"{repository}:{tag}")


images = set()

add(images, get(values, "higress-core", "gateway", "repository"), get(values, "higress-core", "gateway", "tag"))
add(images, get(values, "higress-core", "controller", "repository"), get(values, "higress-core", "controller", "tag"))
add(images, get(values, "higress-core", "pilot", "repository"), get(values, "higress-core", "pilot", "tag"))
add(images, get(values, "higress-core", "pluginServer", "repository"), get(values, "higress-core", "pluginServer", "tag"))
add(images, get(values, "higress-core", "redis", "redis", "repository"), get(values, "higress-core", "redis", "redis", "tag"))

add(images, get(values, "aigateway-console", "image", "repository"), get(values, "aigateway-console", "image", "tag"))

add(images, get(values, "global", "o11y", "grafana", "image", "repository"), get(values, "global", "o11y", "grafana", "image", "tag"))
add(images, get(values, "global", "o11y", "prometheus", "image", "repository"), get(values, "global", "o11y", "prometheus", "image", "tag"))
add(images, get(values, "global", "o11y", "loki", "image", "repository"), get(values, "global", "o11y", "loki", "image", "tag"))
add(images, get(values, "global", "o11y", "promtail", "image", "repository"), get(values, "global", "o11y", "promtail", "image", "tag"))

portal_repository = get(values, "aigateway-portal", "backend", "image", "repository")
portal_tag = get(values, "aigateway-portal", "backend", "image", "tag")
if not portal_repository:
    portal_repository = get(values, "aigateway-portal", "image", "repository")
if not portal_tag:
    portal_tag = get(values, "aigateway-portal", "image", "tag")
add(images, portal_repository, portal_tag)
add(images,
    get(values, "aigateway-portal", "mysql", "image", "repository", default="mariadb"),
    get(values, "aigateway-portal", "mysql", "image", "tag", default="11.4"))

for image in sorted(images):
    print(image)
PY
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

  if [[ -z "${fallback_repo}" ]]; then
    echo "Local image not found and no fallback mapping configured: ${image}" >&2
    return 1
  fi

  fallback_image="${fallback_repo}:${tag}"
  if ! docker image inspect "${fallback_image}" >/dev/null 2>&1; then
    run docker pull "${fallback_image}"
  fi
  run docker tag "${fallback_image}" "${image}"
}

need_cmd docker
need_cmd helm
need_cmd kubectl
need_cmd k3d
need_cmd python3

if [[ ! -f "${VALUES_FILE}" ]]; then
  echo "Values file not found: ${VALUES_FILE}" >&2
  exit 1
fi

if [[ ! -d "${CHART_DIR}" ]]; then
  echo "Chart directory not found: ${CHART_DIR}" >&2
  exit 1
fi

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
  if [[ ${#IMAGES[@]} -eq 0 ]]; then
    echo "No images resolved from ${VALUES_FILE}" >&2
    exit 1
  fi

  echo "Importing images into k3d (${#IMAGES[@]} images)..."
  for image in "${IMAGES[@]}"; do
    ensure_local_image "${image}"
    run k3d image import "${image}" -c "${K3D_CLUSTER}"
  done
fi

if [[ "${SKIP_DEPLOY}" != "true" ]]; then
  run helm dependency update "${CHART_DIR}"
  run helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
    -n "${NAMESPACE}" \
    --create-namespace \
    --render-subchart-notes \
    -f "${VALUES_FILE}" \
    --wait \
    --timeout "${HELM_TIMEOUT}"

  DEPLOYMENTS="$(kubectl -n "${NAMESPACE}" get deploy -l "app.kubernetes.io/instance=${RELEASE_NAME}" -o name || true)"
  STATEFULSETS="$(kubectl -n "${NAMESPACE}" get statefulset -l "app.kubernetes.io/instance=${RELEASE_NAME}" -o name || true)"

  if [[ -n "${DEPLOYMENTS}" ]]; then
    run kubectl -n "${NAMESPACE}" rollout restart deploy -l "app.kubernetes.io/instance=${RELEASE_NAME}"
    while IFS= read -r item; do
      [[ -z "${item}" ]] && continue
      run kubectl -n "${NAMESPACE}" rollout status "${item}" --timeout "${HELM_TIMEOUT}"
    done <<< "${DEPLOYMENTS}"
  fi

  if [[ -n "${STATEFULSETS}" ]]; then
    run kubectl -n "${NAMESPACE}" rollout restart statefulset -l "app.kubernetes.io/instance=${RELEASE_NAME}"
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
