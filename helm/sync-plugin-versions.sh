#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
source "${ROOT_DIR}/scripts/dev-shell-lib.sh"

HIGRESS_DIR=""
CONSOLE_PLUGIN_PROPERTIES_FILE=""
PLUGIN_SERVER_PROPERTIES_FILE=""
CONSOLE_PLUGIN_RESOURCE_DIR=""
FORCE_SOURCE_VERSION_PLUGINS=""
DRY_RUN=false

usage() {
  cat <<'EOF'
Usage:
  ./higress/helm/sync-plugin-versions.sh --higress-dir <dir> --console-plugin-properties-file <file> \
    --plugin-server-properties-file <file> --console-plugin-resource-dir <dir> [options]

Options:
  --force-source-version-plugins <csv>   Plugins that must follow local source VERSION
  --dry-run                              Print pending updates without writing files
  -h, --help                             Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --higress-dir)
      HIGRESS_DIR="$2"
      shift 2
      ;;
    --console-plugin-properties-file)
      CONSOLE_PLUGIN_PROPERTIES_FILE="$2"
      shift 2
      ;;
    --plugin-server-properties-file)
      PLUGIN_SERVER_PROPERTIES_FILE="$2"
      shift 2
      ;;
    --console-plugin-resource-dir)
      CONSOLE_PLUGIN_RESOURCE_DIR="$2"
      shift 2
      ;;
    --force-source-version-plugins)
      FORCE_SOURCE_VERSION_PLUGINS="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
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

[[ -n "${HIGRESS_DIR}" ]] || dev_die "--higress-dir is required"
[[ -n "${CONSOLE_PLUGIN_PROPERTIES_FILE}" ]] || dev_die "--console-plugin-properties-file is required"
[[ -n "${PLUGIN_SERVER_PROPERTIES_FILE}" ]] || dev_die "--plugin-server-properties-file is required"
[[ -n "${CONSOLE_PLUGIN_RESOURCE_DIR}" ]] || dev_die "--console-plugin-resource-dir is required"

declare -A PLUGIN_VERSIONS=()
declare -A FORCE_SOURCE=()

IFS=',' read -r -a force_items <<< "${FORCE_SOURCE_VERSION_PLUGINS}"
for item in "${force_items[@]:-}"; do
  item="$(dev_trim "${item}")"
  [[ -z "${item}" ]] && continue
  FORCE_SOURCE["${item}"]=1
done

extract_plugin_version() {
  local raw_value="$1"
  printf '%s\n' "${raw_value##*:}"
}

replace_plugin_version() {
  local raw_value="$1"
  local version="$2"
  printf '%s:%s\n' "${raw_value%:*}" "${version}"
}

major_version() {
  local version="$1"
  [[ "${version}" =~ ^([0-9]+) ]] || return 0
  printf '%s\n' "${BASH_REMATCH[1]}"
}

resolve_source_dir() {
  local plugin_name="$1"
  local alias_name=""
  local candidate

  if [[ "${plugin_name}" == "json-converter" ]]; then
    alias_name="jsonrpc-converter"
  fi

  for candidate in \
    "${HIGRESS_DIR}/plugins/wasm-go/extensions/${plugin_name}" \
    "${HIGRESS_DIR}/plugins/wasm-rust/extensions/${plugin_name}" \
    "${HIGRESS_DIR}/plugins/wasm-cpp/extensions/${plugin_name//-/_}"
  do
    if [[ -d "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done

  if [[ -n "${alias_name}" ]]; then
    for candidate in \
      "${HIGRESS_DIR}/plugins/wasm-go/extensions/${alias_name}" \
      "${HIGRESS_DIR}/plugins/wasm-rust/extensions/${alias_name}" \
      "${HIGRESS_DIR}/plugins/wasm-cpp/extensions/${alias_name//-/_}"
    do
      if [[ -d "${candidate}" ]]; then
        printf '%s\n' "${candidate}"
        return 0
      fi
    done
  fi

  return 1
}

choose_effective_version() {
  local plugin_name="$1"
  local current_version="$2"
  local source_version="$3"
  local current_major source_major

  if [[ "${current_version}" == "${source_version}" ]]; then
    printf '%s|source-version\n' "${source_version}"
    return 0
  fi

  if [[ -n "${FORCE_SOURCE[${plugin_name}]:-}" ]]; then
    printf '%s|forced-source-version\n' "${source_version}"
    return 0
  fi

  current_major="$(major_version "${current_version}")"
  source_major="$(major_version "${source_version}")"
  if [[ -n "${current_major}" && -n "${source_major}" && "${current_major}" == "${source_major}" ]]; then
    printf '%s|same-major-source-version\n' "${source_version}"
    return 0
  fi

  printf '%s|kept-current-version (source=%s)\n' "${current_version}" "${source_version}"
}

collect_versions() {
  local file="$1"
  local line plugin_name raw_value version

  while IFS= read -r line; do
    [[ -z "${line}" || "${line}" =~ ^[[:space:]]*# ]] && continue
    [[ "${line}" == *"="* ]] || continue
    plugin_name="$(dev_trim "${line%%=*}")"
    raw_value="$(dev_trim "${line#*=}")"
    version="$(extract_plugin_version "${raw_value}")"
    if [[ -n "${PLUGIN_VERSIONS[${plugin_name}]:-}" && "${PLUGIN_VERSIONS[${plugin_name}]}" != "${version}" ]]; then
      dev_die "Version mismatch across properties files for ${plugin_name}: ${PLUGIN_VERSIONS[${plugin_name}]} vs ${version}"
    fi
    PLUGIN_VERSIONS["${plugin_name}"]="${version}"
  done < "${file}"
}

update_property_file() {
  local file="$1"
  local plugin_name="$2"
  local version="$3"
  local tmp

  tmp="$(mktemp "${TMPDIR:-/tmp}/plugin-props.XXXXXX")"
  awk -F= -v plugin_name="${plugin_name}" -v version="${version}" '
    {
      if ($0 ~ /^[[:space:]]*#/ || $0 !~ /=/) {
        print
        next
      }
      key = $1
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", key)
      if (key != plugin_name) {
        print
        next
      }
      value = substr($0, index($0, "=") + 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      prefix = value
      sub(/:[^:]*$/, "", prefix)
      print key "=" prefix ":" version
      next
    }
  ' "${file}" > "${tmp}"

  if [[ "${DRY_RUN}" == "true" ]]; then
    rm -f "${tmp}"
  else
    mv "${tmp}" "${file}"
  fi
}

update_spec_version() {
  local spec_path="$1"
  local version="$2"
  local tmp

  [[ -f "${spec_path}" ]] || return 1

  tmp="$(mktemp "${TMPDIR:-/tmp}/plugin-spec.XXXXXX")"
  awk -v version="${version}" '
    BEGIN {
      in_info = 0
      updated = 0
    }
    {
      line = $0
      stripped = line
      sub(/^[[:space:]]+/, "", stripped)
      indent = match(line, /[^ ]/) - 1
      if (indent < 0) {
        indent = length(line)
      }

      if (!in_info) {
        print line
        if (stripped == "info:") {
          in_info = 1
        }
        next
      }

      if (indent == 0 && stripped != "") {
        in_info = 0
      }

      if (in_info && indent == 2 && stripped ~ /^version:[[:space:]]*/) {
        print "  version: " version
        updated = 1
        in_info = 2
        next
      }

      print line
    }
    END {
      if (!updated) {
        exit 3
      }
    }
  ' "${spec_path}" > "${tmp}" || {
    local rc=$?
    rm -f "${tmp}"
    [[ ${rc} -eq 3 ]] && return 1
    dev_die "Failed to update spec file: ${spec_path}"
  }

  if [[ "${DRY_RUN}" == "true" ]]; then
    rm -f "${tmp}"
  else
    mv "${tmp}" "${spec_path}"
  fi
  return 0
}

collect_versions "${CONSOLE_PLUGIN_PROPERTIES_FILE}"
collect_versions "${PLUGIN_SERVER_PROPERTIES_FILE}"

property_updates=0
spec_updates=0

while IFS= read -r plugin_name; do
  current_version="${PLUGIN_VERSIONS[${plugin_name}]}"
  if ! source_dir="$(resolve_source_dir "${plugin_name}")"; then
    echo "[skip] ${plugin_name}: no local source directory, keep ${current_version}"
    continue
  fi

  version_file="${source_dir}/VERSION"
  [[ -f "${version_file}" ]] || dev_die "VERSION file not found: ${version_file}"
  source_version="$(dev_trim "$(cat "${version_file}")")"

  effective="$(choose_effective_version "${plugin_name}" "${current_version}" "${source_version}")"
  effective_version="${effective%%|*}"
  reason="${effective#*|}"

  if [[ "${effective_version}" != "${current_version}" ]]; then
    update_property_file "${CONSOLE_PLUGIN_PROPERTIES_FILE}" "${plugin_name}" "${effective_version}"
    update_property_file "${PLUGIN_SERVER_PROPERTIES_FILE}" "${plugin_name}" "${effective_version}"
    property_updates=$((property_updates + 2))
    action="updated"
    [[ "${DRY_RUN}" == "true" ]] && action="would update"
    echo "[sync] ${plugin_name}: ${current_version} -> ${effective_version} (${reason}, ${action} properties)"
  else
    echo "[keep] ${plugin_name}: ${current_version} (${reason})"
  fi

  spec_path="${CONSOLE_PLUGIN_RESOURCE_DIR}/${plugin_name}/spec.yaml"
  if update_spec_version "${spec_path}" "${effective_version}"; then
    spec_updates=$((spec_updates + 1))
    action="updated"
    [[ "${DRY_RUN}" == "true" ]] && action="would update"
    echo "[sync] ${plugin_name}: ${action} spec version -> ${effective_version}"
  fi
done < <(printf '%s\n' "${!PLUGIN_VERSIONS[@]}" | sort)

action="Updated"
[[ "${DRY_RUN}" == "true" ]] && action="Would update"
echo "${action} ${property_updates} properties entries and ${spec_updates} spec files."
