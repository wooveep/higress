#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
HIGRESS_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
ROOT_DIR="$(cd -- "${HIGRESS_DIR}/.." && pwd)"
CONSOLE_DIR="${CONSOLE_DIR:-$(cd -- "${HIGRESS_DIR}/../aigateway-console" && pwd)}"
PLUGIN_SERVER_DIR="${PLUGIN_SERVER_DIR:-$(cd -- "${HIGRESS_DIR}/../plugin-server" && pwd)}"
PORTAL_DIR="${PORTAL_DIR:-${HIGRESS_DIR}/../aigateway-portal}"

# Prefer the wrapper values colocated with this script (higress/helm/higress).
# Fall back to the legacy monorepo-level helm path when needed.
if [[ -f "${SCRIPT_DIR}/higress/values-production-gray.yaml" ]]; then
  DEFAULT_WRAPPER_VALUES_FILE="${SCRIPT_DIR}/higress/values-production-gray.yaml"
elif [[ -f "${ROOT_DIR}/helm/higress/values-production-gray.yaml" ]]; then
  DEFAULT_WRAPPER_VALUES_FILE="${ROOT_DIR}/helm/higress/values-production-gray.yaml"
else
  DEFAULT_WRAPPER_VALUES_FILE="${SCRIPT_DIR}/higress/values-production-gray.yaml"
fi
WRAPPER_VALUES_FILE="${WRAPPER_VALUES_FILE:-${DEFAULT_WRAPPER_VALUES_FILE}}"
CORE_VALUES_FILE="${CORE_VALUES_FILE:-${SCRIPT_DIR}/core/values-production-gray.yaml}"
CONSOLE_VALUES_FILE="${CONSOLE_VALUES_FILE:-${CONSOLE_DIR}/helm/values-production-gray.yaml}"

DEFAULT_CONSOLE_PLUGIN_RESOURCE_DIR=""
for candidate in \
  "${CONSOLE_DIR}/backend/resource/public/plugin" \
  "${CONSOLE_DIR}/backend-java-legacy/sdk/src/main/resources/plugins" \
  "${CONSOLE_DIR}/backend/sdk/src/main/resources/plugins"
do
  if [[ -f "${candidate}/plugins.properties" ]]; then
    DEFAULT_CONSOLE_PLUGIN_RESOURCE_DIR="${candidate}"
    break
  fi
done
if [[ -z "${DEFAULT_CONSOLE_PLUGIN_RESOURCE_DIR}" ]]; then
  DEFAULT_CONSOLE_PLUGIN_RESOURCE_DIR="${CONSOLE_DIR}/backend/resource/public/plugin"
fi

CONSOLE_PLUGIN_RESOURCE_DIR="${CONSOLE_PLUGIN_RESOURCE_DIR:-${DEFAULT_CONSOLE_PLUGIN_RESOURCE_DIR}}"
CONSOLE_PLUGIN_PROPERTIES_FILE="${CONSOLE_PLUGIN_PROPERTIES_FILE:-${CONSOLE_PLUGIN_RESOURCE_DIR}/plugins.properties}"
PLUGIN_SERVER_PROPERTIES_FILE="${PLUGIN_SERVER_PROPERTIES_FILE:-${PLUGIN_SERVER_DIR}/plugins.properties}"
SYNC_PLUGIN_VERSIONS_SCRIPT="${SYNC_PLUGIN_VERSIONS_SCRIPT:-${SCRIPT_DIR}/sync-plugin-versions.py}"
FORCE_SOURCE_VERSION_PLUGINS="${FORCE_SOURCE_VERSION_PLUGINS:-ai-quota}"

LOCAL_PLUGIN_OUTPUT_DIR="${LOCAL_PLUGIN_OUTPUT_DIR:-${HIGRESS_DIR}/out/local-wasm-plugins}"
LOCAL_PLUGIN_LAYOUT_ROOT="${LOCAL_PLUGIN_LAYOUT_ROOT:-${LOCAL_PLUGIN_OUTPUT_DIR}/oci-layouts}"
PLUGIN_SERVER_LOCAL_PLUGINS_DIR="${PLUGIN_SERVER_LOCAL_PLUGINS_DIR:-${PLUGIN_SERVER_DIR}/local-plugins}"

ARCH="${ARCH:-amd64}"
BASE_HUB="${BASE_HUB:-higress-registry.cn-hangzhou.cr.aliyuncs.com/higress}"
ORAS_IMAGE="${ORAS_IMAGE:-ghcr.io/oras-project/oras:v1.2.3}"
COMPONENTS="${COMPONENTS:-aigateway,controller,gateway,pilot,console,portal,plugins,plugin-server}"
DRY_RUN="${DRY_RUN:-false}"

PLUGIN_BUILD_PLAN_FILE=""
PLUGIN_LAYOUT_REF_FILE=""
PLUGINS_BUILT=false
ISTIO_CACHE_VOLUMES_PREPARED=false
PLUGIN_VERSIONS_SYNCED=false

cleanup() {
  if [[ -n "${PLUGIN_BUILD_PLAN_FILE}" && -f "${PLUGIN_BUILD_PLAN_FILE}" ]]; then
    rm -f "${PLUGIN_BUILD_PLAN_FILE}"
  fi
}
trap cleanup EXIT

usage() {
  cat <<'EOF'
Usage:
  build-local-images.sh [--dry-run] [--arch amd64|arm64] [--components list] [--base-hub repo]

Examples:
  ./helm/build-local-images.sh
  ./helm/build-local-images.sh --dry-run
  ./helm/build-local-images.sh --components aigateway,controller,gateway,console,portal,plugins,plugin-server

Environment overrides:
  WRAPPER_VALUES_FILE           Parent chart values file. Default: higress/helm/higress/values-production-gray.yaml
  CORE_VALUES_FILE              higress-core values file. Default: helm/core/values-production-gray.yaml
  CONSOLE_VALUES_FILE           aigateway-console values file. Default: ../aigateway-console/helm/values-production-gray.yaml
  CONSOLE_DIR                   Path to the aigateway-console repo. Default: ../aigateway-console
  PORTAL_DIR                    Path to the aigateway-portal repo. Default: ../aigateway-portal
  PLUGIN_SERVER_DIR             Path to the plugin-server repo. Default: ../plugin-server
  ARCH                          Target architecture. Default: amd64
  BASE_HUB                      Base image hub used by the existing higress build scripts.
  COMPONENTS                    Comma-separated list:
                                aigateway,controller,gateway,pilot,console,portal,plugins,plugin-server
  LOCAL_PLUGIN_OUTPUT_DIR       Root directory for generated Wasm artifacts and OCI layouts.
  PLUGIN_SERVER_LOCAL_PLUGINS_DIR
                                Generated plugin-server local plugin directory.
  ORAS_IMAGE                    ORAS container image used to write OCI layouts.
  DRY_RUN                       Set to true to print commands only.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --arch)
      ARCH="$2"
      shift 2
      ;;
    --components)
      COMPONENTS="$2"
      shift 2
      ;;
    --base-hub)
      BASE_HUB="$2"
      shift 2
      ;;
    --help|-h)
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

resolve_makefile_export() {
  local var_name="$1"

  python3 - "${HIGRESS_DIR}/Makefile.core.mk" "${var_name}" <<'PY'
import re
import sys

makefile_path, var_name = sys.argv[1:3]
pattern = re.compile(rf'^\s*export\s+{re.escape(var_name)}\s*\??=\s*(.*?)\s*$')

with open(makefile_path, "r", encoding="utf-8") as f:
    for line in f:
        match = pattern.match(line)
        if match:
            print(match.group(1))
            sys.exit(0)

sys.exit(1)
PY
}

run() {
  echo "+ $*"
  if [[ "${DRY_RUN}" != "true" ]]; then
    "$@"
  fi
}

run_in_dir() {
  local dir="$1"
  shift
  echo "+ (cd ${dir} && $*)"
  if [[ "${DRY_RUN}" != "true" ]]; then
    (
      cd "${dir}"
      "$@"
    )
  fi
}

remove_dir_with_docker() {
  local dir="$1"
  local parent_dir base_name

  parent_dir="$(dirname "${dir}")"
  base_name="$(basename "${dir}")"

  echo "+ docker run --rm -v ${parent_dir}:/workspace alpine:3.20 sh -c rm -rf /workspace/${base_name}"
  if [[ "${DRY_RUN}" != "true" ]]; then
    docker run --rm \
      -v "${parent_dir}:/workspace" \
      alpine:3.20 \
      sh -c "rm -rf /workspace/${base_name}"
  fi
}

prepare_istio_cache_volumes() {
  local uid gid

  if [[ "${ISTIO_CACHE_VOLUMES_PREPARED}" == "true" ]]; then
    return
  fi

  uid="$(id -u)"
  gid="$(id -g)"

  echo "+ docker run --rm -v go:/workspace/go -v gocache:/workspace/gocache -v cache:/workspace/cache alpine:3.20 sh -c chown -R ${uid}:${gid} /workspace/go /workspace/gocache /workspace/cache"
  if [[ "${DRY_RUN}" != "true" ]]; then
    docker run --rm \
      -v go:/workspace/go \
      -v gocache:/workspace/gocache \
      -v cache:/workspace/cache \
      alpine:3.20 \
      sh -c "chown -R ${uid}:${gid} /workspace/go /workspace/gocache /workspace/cache"
  fi

  ISTIO_CACHE_VOLUMES_PREPARED=true
}

reset_dir() {
  local dir="$1"
  echo "+ rm -rf ${dir} && mkdir -p ${dir}"
  if [[ "${DRY_RUN}" != "true" ]]; then
    if ! rm -rf "${dir}" 2>/dev/null; then
      remove_dir_with_docker "${dir}"
    fi
    mkdir -p "${dir}"
  fi
}

tag_if_needed() {
  local source_image="$1"
  local target_image="$2"

  if [[ "${source_image}" == "${target_image}" ]]; then
    return
  fi

  run docker tag "${source_image}" "${target_image}"
}

has_component() {
  local wanted="$1"
  if [[ "${COMPONENTS}" == "all" ]]; then
    return 0
  fi
  local padded=",${COMPONENTS},"
  [[ "${padded}" == *",${wanted},"* ]]
}

for file in "${WRAPPER_VALUES_FILE}" "${CORE_VALUES_FILE}" "${CONSOLE_VALUES_FILE}"; do
  if [[ ! -f "${file}" ]]; then
    echo "Required file not found: ${file}" >&2
    exit 1
  fi
done

if (has_component "plugins" || has_component "plugin-server") && [[ ! -f "${CONSOLE_PLUGIN_PROPERTIES_FILE}" ]]; then
  echo "Required file not found: ${CONSOLE_PLUGIN_PROPERTIES_FILE}" >&2
  exit 1
fi

if (has_component "plugins" || has_component "plugin-server") && [[ ! -f "${PLUGIN_SERVER_PROPERTIES_FILE}" ]]; then
  echo "Required file not found: ${PLUGIN_SERVER_PROPERTIES_FILE}" >&2
  exit 1
fi

if (has_component "plugins" || has_component "plugin-server") && [[ ! -f "${SYNC_PLUGIN_VERSIONS_SCRIPT}" ]]; then
  echo "Required file not found: ${SYNC_PLUGIN_VERSIONS_SCRIPT}" >&2
  exit 1
fi

for dir in "${CONSOLE_DIR}"; do
  if [[ ! -d "${dir}" ]]; then
    echo "Required directory not found: ${dir}" >&2
    exit 1
  fi
done

if (has_component "plugins" || has_component "plugin-server") && [[ ! -d "${PLUGIN_SERVER_DIR}" ]]; then
  echo "Required directory not found: ${PLUGIN_SERVER_DIR}" >&2
  exit 1
fi

if (has_component "plugins" || has_component "plugin-server") && [[ ! -d "${CONSOLE_PLUGIN_RESOURCE_DIR}" ]]; then
  echo "Required directory not found: ${CONSOLE_PLUGIN_RESOURCE_DIR}" >&2
  exit 1
fi

if has_component "portal" && [[ ! -d "${PORTAL_DIR}" ]]; then
  echo "Required directory not found: ${PORTAL_DIR}" >&2
  exit 1
fi

if has_component "console" && [[ ! -d "${PORTAL_DIR}/backend" ]]; then
  echo "Required directory not found: ${PORTAL_DIR}/backend" >&2
  exit 1
fi

need_cmd docker
need_cmd make
need_cmd python3
need_cmd tar
need_cmd file
need_cmd go

sync_plugin_versions() {
  local cmd=(
    python3
    "${SYNC_PLUGIN_VERSIONS_SCRIPT}"
    --higress-dir "${HIGRESS_DIR}"
    --console-plugin-properties-file "${CONSOLE_PLUGIN_PROPERTIES_FILE}"
    --plugin-server-properties-file "${PLUGIN_SERVER_PROPERTIES_FILE}"
    --console-plugin-resource-dir "${CONSOLE_PLUGIN_RESOURCE_DIR}"
    --force-source-version-plugins "${FORCE_SOURCE_VERSION_PLUGINS}"
  )

  if [[ "${PLUGIN_VERSIONS_SYNCED}" == "true" ]]; then
    return
  fi

  if [[ "${DRY_RUN}" == "true" ]]; then
    cmd+=(--dry-run)
  fi

  echo "+ ${cmd[*]}"
  "${cmd[@]}"
  PLUGIN_VERSIONS_SYNCED=true
}

if has_component "plugins" || has_component "plugin-server"; then
  sync_plugin_versions
fi

if ! PARSED_IMAGE_VALUES="$(
python3 - "${WRAPPER_VALUES_FILE}" "${CORE_VALUES_FILE}" "${CONSOLE_VALUES_FILE}" <<'PY'
import shlex
import sys
import yaml

wrapper_path, core_path, console_path = sys.argv[1:4]

with open(wrapper_path, "r", encoding="utf-8") as f:
    wrapper = yaml.safe_load(f) or {}
with open(core_path, "r", encoding="utf-8") as f:
    core = yaml.safe_load(f) or {}
with open(console_path, "r", encoding="utf-8") as f:
    console = yaml.safe_load(f) or {}

def get(data, *path):
    cur = data
    for key in path:
        if not isinstance(cur, dict):
            return ""
        cur = cur.get(key)
    return "" if cur is None else cur

def coalesce(*values):
    for value in values:
        if value != "":
            return value
    return ""

values = {
    "AIGATEWAY_REPOSITORY": get(console, "certmanager", "image", "repository"),
    "AIGATEWAY_TAG": get(console, "certmanager", "image", "tag"),
    "CONTROLLER_REPOSITORY": get(wrapper, "higress-core", "controller", "repository"),
    "CONTROLLER_TAG": get(wrapper, "higress-core", "controller", "tag"),
    "GATEWAY_REPOSITORY": get(wrapper, "higress-core", "gateway", "repository"),
    "GATEWAY_TAG": get(wrapper, "higress-core", "gateway", "tag"),
    "PILOT_REPOSITORY": get(wrapper, "higress-core", "pilot", "repository"),
    "PILOT_TAG": get(wrapper, "higress-core", "pilot", "tag"),
    "PLUGIN_SERVER_REPOSITORY": get(wrapper, "higress-core", "pluginServer", "repository"),
    "PLUGIN_SERVER_TAG": get(wrapper, "higress-core", "pluginServer", "tag"),
    "CONSOLE_REPOSITORY": get(wrapper, "aigateway-console", "image", "repository"),
    "CONSOLE_TAG": get(wrapper, "aigateway-console", "image", "tag"),
    "PORTAL_BACKEND_REPOSITORY": coalesce(
        get(wrapper, "aigateway-portal", "backend", "image", "repository"),
        get(wrapper, "aigateway-portal", "image", "repository"),
    ),
    "PORTAL_BACKEND_TAG": coalesce(
        get(wrapper, "aigateway-portal", "backend", "image", "tag"),
        get(wrapper, "aigateway-portal", "image", "tag"),
    ),
}

checks = [
    ("certmanager.image.repository", values["AIGATEWAY_REPOSITORY"], get(wrapper, "aigateway-console", "certmanager", "image", "repository")),
    ("certmanager.image.tag", values["AIGATEWAY_TAG"], get(wrapper, "aigateway-console", "certmanager", "image", "tag")),
    ("controller.repository", values["CONTROLLER_REPOSITORY"], get(core, "controller", "repository")),
    ("controller.tag", values["CONTROLLER_TAG"], get(core, "controller", "tag")),
    ("gateway.repository", values["GATEWAY_REPOSITORY"], get(core, "gateway", "repository")),
    ("gateway.tag", values["GATEWAY_TAG"], get(core, "gateway", "tag")),
    ("pilot.repository", values["PILOT_REPOSITORY"], get(core, "pilot", "repository")),
    ("pilot.tag", values["PILOT_TAG"], get(core, "pilot", "tag")),
    ("pluginServer.repository", values["PLUGIN_SERVER_REPOSITORY"], get(core, "pluginServer", "repository")),
    ("pluginServer.tag", values["PLUGIN_SERVER_TAG"], get(core, "pluginServer", "tag")),
    ("console.image.repository", values["CONSOLE_REPOSITORY"], get(console, "image", "repository")),
    ("console.image.tag", values["CONSOLE_TAG"], get(console, "image", "tag")),
]

errors = []
for name, wrapper_value, standalone_value in checks:
    if str(wrapper_value) != str(standalone_value):
        errors.append(
            f"Values mismatch for {name}: wrapper={wrapper_value!r}, standalone={standalone_value!r}"
        )

required = [key for key, value in values.items() if value == ""]
if required:
    errors.append("Missing required image values: " + ", ".join(required))

if errors:
    for error in errors:
        print(error, file=sys.stderr)
    sys.exit(1)

for key, value in values.items():
    print(f"{key}={shlex.quote(str(value))}")
PY
)"; then
  echo "Failed to resolve image tags from values files. Please fix the mismatch above and retry." >&2
  exit 1
fi

eval "${PARSED_IMAGE_VALUES}"

HIGRESS_BASE_VERSION="${HIGRESS_BASE_VERSION:-$(resolve_makefile_export HIGRESS_BASE_VERSION || true)}"
if [[ -z "${HIGRESS_BASE_VERSION}" ]]; then
  echo "Unable to resolve HIGRESS_BASE_VERSION from ${HIGRESS_DIR}/Makefile.core.mk" >&2
  exit 1
fi

ENVOY_PACKAGE_URL_PATTERN="${ENVOY_PACKAGE_URL_PATTERN:-$(resolve_makefile_export ENVOY_PACKAGE_URL_PATTERN || true)}"
if [[ -z "${ENVOY_PACKAGE_URL_PATTERN}" ]]; then
  echo "Unable to resolve ENVOY_PACKAGE_URL_PATTERN from ${HIGRESS_DIR}/Makefile.core.mk" >&2
  exit 1
fi

export HIGRESS_BASE_VERSION
export ENVOY_PACKAGE_URL_PATTERN

PLUGIN_LAYOUT_REF_FILE="${LOCAL_PLUGIN_OUTPUT_DIR}/plugin-layout-refs.properties"

echo "Using values:"
echo "  wrapper          : ${WRAPPER_VALUES_FILE}"
echo "  core             : ${CORE_VALUES_FILE}"
echo "  console          : ${CONSOLE_VALUES_FILE}"
echo "  arch             : ${ARCH}"
echo "  base hub         : ${BASE_HUB}"
echo "  base version     : ${HIGRESS_BASE_VERSION}"
echo "  envoy package    : ${ENVOY_PACKAGE_URL_PATTERN}"
echo "  plugin output    : ${LOCAL_PLUGIN_OUTPUT_DIR}"
echo "  plugin layouts   : ${LOCAL_PLUGIN_LAYOUT_ROOT}"
echo "  plugin-server dir: ${PLUGIN_SERVER_DIR}"
echo "  console dir      : ${CONSOLE_DIR}"
echo "  portal dir       : ${PORTAL_DIR}"

echo "Resolved image tags:"
echo "  aigateway    : ${AIGATEWAY_REPOSITORY}:${AIGATEWAY_TAG}"
echo "  controller   : ${CONTROLLER_REPOSITORY}:${CONTROLLER_TAG}"
echo "  gateway      : ${GATEWAY_REPOSITORY}:${GATEWAY_TAG}"
echo "  pilot        : ${PILOT_REPOSITORY}:${PILOT_TAG}"
echo "  console      : ${CONSOLE_REPOSITORY}:${CONSOLE_TAG}"
echo "  portal-image : ${PORTAL_BACKEND_REPOSITORY}:${PORTAL_BACKEND_TAG}"
echo "  plugin-server: ${PLUGIN_SERVER_REPOSITORY}:${PLUGIN_SERVER_TAG}"

prepare_higress_main_context() {
  local out_linux="${HIGRESS_DIR}/out/linux_${ARCH}"
  local binary="${out_linux}/higress"
  local context_dir="${out_linux}/docker_build/docker.higress"

  if [[ ! -f "${binary}" ]]; then
    run_in_dir "${HIGRESS_DIR}" \
      env TARGET_ARCH="${ARCH}" make build-linux
  fi

  run mkdir -p "${context_dir}"
  run_in_dir "${HIGRESS_DIR}" \
    env TARGET_ARCH="${ARCH}" ./docker/docker-copy.sh "${binary}" docker/Dockerfile.higress "${context_dir}"
}

build_aigateway_main() {
  local target_image="${AIGATEWAY_REPOSITORY}:${AIGATEWAY_TAG}"
  local context_dir="${HIGRESS_DIR}/out/linux_${ARCH}/docker_build/docker.higress"

  prepare_higress_main_context

  run_in_dir "${context_dir}" \
    docker build \
      --build-arg BASE_VERSION="${HIGRESS_BASE_VERSION}" \
      --build-arg HUB="${BASE_HUB}" \
      --build-arg TARGETARCH="${ARCH}" \
      -t "${target_image}" \
      -f Dockerfile.higress \
      .
}

build_controller() {
  local source_image="${BASE_HUB}/controller:${CONTROLLER_TAG}"
  local target_image="${CONTROLLER_REPOSITORY}:${CONTROLLER_TAG}"

  run_in_dir "${HIGRESS_DIR}" \
    env HUB="${BASE_HUB}" IMG="controller" TAG="${CONTROLLER_TAG}" TARGET_ARCH="${ARCH}" \
    make docker-build

  tag_if_needed "${source_image}" "${target_image}"
}

build_gateway() {
  local target_image="${GATEWAY_REPOSITORY}:${GATEWAY_TAG}"
  local source_hub="${GATEWAY_REPOSITORY%/*}"
  local source_image="${source_hub}/proxyv2:${GATEWAY_TAG}"

  prepare_istio_cache_volumes

  run_in_dir "${HIGRESS_DIR}" \
    env TARGET_ARCH="${ARCH}" ./tools/hack/build-golang-filters.sh

  run_in_dir "${HIGRESS_DIR}" \
    env USE_REAL_USER=1 HUB="${BASE_HUB}" IMG_URL="${target_image}" TARGET_ARCH="${ARCH}" DOCKER_TARGETS="docker.proxyv2" \
    ./tools/hack/build-istio-image.sh docker

  tag_if_needed "${source_image}" "${target_image}"
}

build_pilot() {
  local target_image="${PILOT_REPOSITORY}:${PILOT_TAG}"

  prepare_istio_cache_volumes

  run_in_dir "${HIGRESS_DIR}" \
    env USE_REAL_USER=1 HUB="${BASE_HUB}" IMG_URL="${target_image}" TARGET_ARCH="${ARCH}" DOCKER_TARGETS="docker.pilot" \
    ./tools/hack/build-istio-image.sh docker
}

build_console() {
  local target_image="${CONSOLE_REPOSITORY}:${CONSOLE_TAG}"

  # Rebuild the frontend bundle and copy it into the Go backend's static resource directory
  # before building the final console image.
  run rm -rf "${CONSOLE_DIR}/frontend/.ice" "${CONSOLE_DIR}/frontend/build"
  run rm -rf "${CONSOLE_DIR}/backend/resource/public/html"
  run mkdir -p "${CONSOLE_DIR}/backend/resource/public/html"

  run_in_dir "${CONSOLE_DIR}/frontend" \
    npm install
  run_in_dir "${CONSOLE_DIR}/frontend" \
    npm run build

  run cp -R "${CONSOLE_DIR}/frontend/build/." "${CONSOLE_DIR}/backend/resource/public/html/"

  run_in_dir "${CONSOLE_DIR}/backend" \
    docker build \
      --platform "linux/${ARCH}" \
      --build-context "portal_backend=${PORTAL_DIR}/backend" \
      --build-arg TARGETARCH="${ARCH}" \
      -f Dockerfile.monorepo \
      -t "${target_image}" \
      .
}

build_portal() {
  local portal_image="${PORTAL_BACKEND_REPOSITORY}:${PORTAL_BACKEND_TAG}"

  run_in_dir "${PORTAL_DIR}" \
    docker build \
      --platform "linux/${ARCH}" \
      --build-arg TARGETARCH="${ARCH}" \
      -f backend/Dockerfile \
      -t "${portal_image}" \
      .
}

prepare_plugin_build_plan() {
  if [[ -n "${PLUGIN_BUILD_PLAN_FILE}" && -f "${PLUGIN_BUILD_PLAN_FILE}" ]]; then
    return
  fi

  PLUGIN_BUILD_PLAN_FILE="$(mktemp)"

  python3 - "${CONSOLE_PLUGIN_PROPERTIES_FILE}" "${PLUGIN_SERVER_PROPERTIES_FILE}" "${HIGRESS_DIR}" "${CONSOLE_PLUGIN_RESOURCE_DIR}" > "${PLUGIN_BUILD_PLAN_FILE}" <<'PY'
from pathlib import Path
import sys

console_props = Path(sys.argv[1])
plugin_server_props = Path(sys.argv[2])
higress_dir = Path(sys.argv[3])
console_resource_root = Path(sys.argv[4])

alias_map = {
    "json-converter": "jsonrpc-converter",
}

def load_properties(path):
    result = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        name, value = line.split("=", 1)
        name = name.strip()
        value = value.strip()
        image_ref = value.replace("oci://", "", 1)
        version = image_ref.rsplit(":", 1)[-1]
        result[name] = version
    return result

versions = {}
for path in (console_props, plugin_server_props):
    for name, version in load_properties(path).items():
        previous = versions.get(name)
        if previous and previous != version:
            raise SystemExit(f"Version mismatch for plugin {name}: {previous} vs {version}")
        versions[name] = version

def resolve_source(name):
    candidates = [
        ("go", higress_dir / "plugins/wasm-go/extensions" / name),
        ("rust", higress_dir / "plugins/wasm-rust/extensions" / name),
        ("cpp", higress_dir / "plugins/wasm-cpp/extensions" / name.replace("-", "_")),
    ]
    alias = alias_map.get(name)
    if alias:
        candidates.extend([
            ("go", higress_dir / "plugins/wasm-go/extensions" / alias),
            ("rust", higress_dir / "plugins/wasm-rust/extensions" / alias),
            ("cpp", higress_dir / "plugins/wasm-cpp/extensions" / alias.replace("-", "_")),
        ])

    for plugin_type, source_dir in candidates:
        if source_dir.is_dir():
            return plugin_type, source_dir
    raise SystemExit(f"No local source found for plugin {name}")

for name in sorted(versions):
    plugin_type, source_dir = resolve_source(name)
    resource_dir = console_resource_root / name
    if not resource_dir.is_dir():
        resource_dir = source_dir
    print("\t".join([name, versions[name], plugin_type, str(source_dir), str(resource_dir)]))
PY
}

write_plugin_metadata() {
  local plugin_dir="$1"
  local plugin_name="$2"

  run python3 - "${plugin_dir}" "${plugin_name}" <<'PY'
import hashlib
import os
import sys
from datetime import datetime

plugin_dir = sys.argv[1]
plugin_name = sys.argv[2]
wasm_path = os.path.join(plugin_dir, "plugin.wasm")

with open(wasm_path, "rb") as f:
    md5 = hashlib.md5(f.read()).hexdigest()

stat_info = os.stat(wasm_path)
metadata_path = os.path.join(plugin_dir, "metadata.txt")

with open(metadata_path, "w", encoding="utf-8") as f:
    f.write(f"Plugin Name: {plugin_name}\n")
    f.write(f"Size: {stat_info.st_size} bytes\n")
    f.write(f"Last Modified: {datetime.fromtimestamp(stat_info.st_mtime).isoformat()}\n")
    f.write(f"Created: {datetime.fromtimestamp(stat_info.st_ctime).isoformat()}\n")
    f.write(f"MD5: {md5}\n")
PY
}

copy_plugin_doc_file() {
  local stage_dir="$1"
  local filename="$2"
  shift 2

  local source_dir
  for source_dir in "$@"; do
    if [[ -f "${source_dir}/${filename}" ]]; then
      run cp "${source_dir}/${filename}" "${stage_dir}/${filename}"
      return 0
    fi
  done

  return 1
}

copy_plugin_docs() {
  local stage_dir="$1"
  local source_dir="$2"
  local resource_dir="$3"
  local filename

  copy_plugin_doc_file "${stage_dir}" "spec.yaml" "${resource_dir}" "${source_dir}" || true

  for filename in README.md README_EN.md README_CN.md README_ZH.md; do
    copy_plugin_doc_file "${stage_dir}" "${filename}" "${resource_dir}" "${source_dir}" || true
  done
}

build_go_plugin() {
  local source_dir="$1"
  local stage_dir="$2"

  run_in_dir "${source_dir}" \
    env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o "${stage_dir}/plugin.wasm" .
}

build_rust_plugin() {
  local source_dir="$1"
  local stage_dir="$2"
  local plugin_dir_name
  local uid gid

  plugin_dir_name="$(basename "${source_dir}")"
  uid="$(id -u)"
  gid="$(id -g)"

  run_in_dir "${HIGRESS_DIR}/plugins/wasm-rust" \
    env DOCKER_BUILDKIT=1 docker build \
      --build-arg PLUGIN_NAME="${plugin_dir_name}" \
      --output "type=local,dest=${stage_dir},uid=${uid},gid=${gid}" \
      -f Dockerfile \
      .
}

build_cpp_plugin() {
  local source_dir="$1"
  local stage_dir="$2"
  local plugin_dir_name
  local uid gid

  plugin_dir_name="$(basename "${source_dir}")"
  uid="$(id -u)"
  gid="$(id -g)"

  run_in_dir "${HIGRESS_DIR}/plugins/wasm-cpp" \
    env DOCKER_BUILDKIT=1 docker build \
      --build-arg PLUGIN_NAME="${plugin_dir_name}" \
      --output "type=local,dest=${stage_dir},uid=${uid},gid=${gid}" \
      -f Dockerfile \
      .
}

write_plugin_oci_layout() {
  local plugin_name="$1"
  local plugin_version="$2"
  local stage_dir="$3"
  local target_ref="/workspace/layouts/plugin/${plugin_name}:${plugin_version}"
  local uid gid
  local oras_args=(
    push
    --oci-layout
    "${target_ref}"
  )
  uid="$(id -u)"
  gid="$(id -g)"

  if [[ -f "${stage_dir}/spec.yaml" ]]; then
    oras_args+=("./spec.yaml:application/vnd.module.wasm.spec.v1+yaml")
  fi
  if [[ -f "${stage_dir}/README.md" ]]; then
    oras_args+=("./README.md:application/vnd.module.wasm.doc.v1+markdown")
  fi
  if [[ -f "${stage_dir}/README_EN.md" ]]; then
    oras_args+=("./README_EN.md:application/vnd.module.wasm.doc.v1.EN+markdown")
  fi
  if [[ -f "${stage_dir}/README_CN.md" ]]; then
    oras_args+=("./README_CN.md:application/vnd.module.wasm.doc.v1.CN+markdown")
  fi
  if [[ -f "${stage_dir}/README_ZH.md" ]]; then
    oras_args+=("./README_ZH.md:application/vnd.module.wasm.doc.v1.ZH+markdown")
  fi
  oras_args+=("./plugin.tar.gz:application/vnd.oci.image.layer.v1.tar+gzip")

  run mkdir -p "${LOCAL_PLUGIN_LAYOUT_ROOT}/plugin"
  echo "+ docker run --rm --user ${uid}:${gid} -v ${stage_dir}:/workspace/plugin -v ${LOCAL_PLUGIN_LAYOUT_ROOT}:/workspace/layouts -w /workspace/plugin ${ORAS_IMAGE} ${oras_args[*]}"
  if [[ "${DRY_RUN}" != "true" ]]; then
    docker run --rm \
      --user "${uid}:${gid}" \
      -v "${stage_dir}:/workspace/plugin" \
      -v "${LOCAL_PLUGIN_LAYOUT_ROOT}:/workspace/layouts" \
      -w /workspace/plugin \
      "${ORAS_IMAGE}" \
      "${oras_args[@]}"
  fi
}

build_single_plugin() {
  local plugin_name="$1"
  local plugin_version="$2"
  local plugin_type="$3"
  local source_dir="$4"
  local resource_dir="$5"
  local stage_dir="${LOCAL_PLUGIN_OUTPUT_DIR}/plugins/${plugin_name}/${plugin_version}"

  echo "Building plugin ${plugin_name}:${plugin_version} (${plugin_type})"
  reset_dir "${stage_dir}"

  case "${plugin_type}" in
    go)
      build_go_plugin "${source_dir}" "${stage_dir}"
      ;;
    rust)
      build_rust_plugin "${source_dir}" "${stage_dir}"
      ;;
    cpp)
      build_cpp_plugin "${source_dir}" "${stage_dir}"
      ;;
    *)
      echo "Unsupported plugin type ${plugin_type} for ${plugin_name}" >&2
      exit 1
      ;;
  esac

  write_plugin_metadata "${stage_dir}" "${plugin_name}"
  copy_plugin_docs "${stage_dir}" "${source_dir}" "${resource_dir}"
  run_in_dir "${stage_dir}" tar czf plugin.tar.gz plugin.wasm
  write_plugin_oci_layout "${plugin_name}" "${plugin_version}" "${stage_dir}"

  if [[ "${DRY_RUN}" != "true" ]]; then
    printf '%s=%s/plugin/%s:%s\n' "${plugin_name}" "${LOCAL_PLUGIN_LAYOUT_ROOT}" "${plugin_name}" "${plugin_version}" >> "${PLUGIN_LAYOUT_REF_FILE}"
  fi
}

build_plugins() {
  local line
  local plugin_name
  local plugin_version
  local plugin_type
  local source_dir
  local resource_dir

  if [[ "${PLUGINS_BUILT}" == "true" ]]; then
    return
  fi

  prepare_plugin_build_plan
  reset_dir "${LOCAL_PLUGIN_OUTPUT_DIR}/plugins"
  reset_dir "${LOCAL_PLUGIN_LAYOUT_ROOT}"
  if [[ "${DRY_RUN}" != "true" ]]; then
    mkdir -p "$(dirname "${PLUGIN_LAYOUT_REF_FILE}")"
    : > "${PLUGIN_LAYOUT_REF_FILE}"
  else
    echo "+ : > ${PLUGIN_LAYOUT_REF_FILE}"
  fi

  while IFS=$'\t' read -r plugin_name plugin_version plugin_type source_dir resource_dir; do
    build_single_plugin "${plugin_name}" "${plugin_version}" "${plugin_type}" "${source_dir}" "${resource_dir}"
  done < "${PLUGIN_BUILD_PLAN_FILE}"

  PLUGINS_BUILT=true
}

sync_plugin_server_local_plugins() {
  reset_dir "${PLUGIN_SERVER_LOCAL_PLUGINS_DIR}"
  echo "+ cp -a ${LOCAL_PLUGIN_OUTPUT_DIR}/plugins/. ${PLUGIN_SERVER_LOCAL_PLUGINS_DIR}/"
  if [[ "${DRY_RUN}" != "true" ]]; then
    cp -a "${LOCAL_PLUGIN_OUTPUT_DIR}/plugins/." "${PLUGIN_SERVER_LOCAL_PLUGINS_DIR}/"
  fi
}

build_plugin_server() {
  local target_image="${PLUGIN_SERVER_REPOSITORY}:${PLUGIN_SERVER_TAG}"

  build_plugins
  sync_plugin_server_local_plugins

  run_in_dir "${PLUGIN_SERVER_DIR}" \
    env DOCKER_BUILDKIT=1 docker build \
      --build-arg USE_LOCAL_PLUGINS=true \
      -t "${target_image}" \
      -f Dockerfile \
      .
}

if has_component "aigateway"; then
  build_aigateway_main
fi

if has_component "controller"; then
  build_controller
fi

if has_component "gateway"; then
  build_gateway
fi

if has_component "pilot"; then
  build_pilot
fi

if has_component "console"; then
  build_console
fi

if has_component "portal"; then
  build_portal
fi

if has_component "plugins"; then
  build_plugins
fi

if has_component "plugin-server"; then
  build_plugin_server
fi

echo "Done."
if [[ "${PLUGINS_BUILT}" == "true" ]]; then
  echo "Local Wasm plugin refs: ${PLUGIN_LAYOUT_REF_FILE}"
fi
echo "Skipped by design: redis/grafana/prometheus/loki/promtail/cert-manager."
