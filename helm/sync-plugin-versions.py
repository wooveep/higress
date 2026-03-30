#!/usr/bin/env python3

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path


ALIAS_MAP = {
    "json-converter": "jsonrpc-converter",
}


@dataclass(frozen=True)
class PropertyEntry:
    file_path: Path
    line_index: int
    name: str
    raw_value: str
    version: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Synchronize built-in plugin versions used by local build scripts."
    )
    parser.add_argument("--higress-dir", required=True, help="Path to the higress repo root.")
    parser.add_argument(
        "--console-plugin-properties-file",
        required=True,
        help="Path to higress-console plugins.properties.",
    )
    parser.add_argument(
        "--plugin-server-properties-file",
        required=True,
        help="Path to plugin-server plugins.properties.",
    )
    parser.add_argument(
        "--console-plugin-resource-dir",
        required=True,
        help="Path to higress-console built-in plugin resource directory.",
    )
    parser.add_argument(
        "--force-source-version-plugins",
        default="",
        help="Comma-separated plugin names that must always follow local source VERSION.",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print pending updates without writing files.",
    )
    return parser.parse_args()


def extract_version(raw_value: str) -> str:
    if ":" not in raw_value:
        raise ValueError(f"Unsupported plugin image reference: {raw_value}")
    return raw_value.rsplit(":", 1)[-1].strip()


def replace_version(raw_value: str, version: str) -> str:
    prefix, _ = raw_value.rsplit(":", 1)
    return f"{prefix}:{version}"


def load_properties(path: Path) -> tuple[list[str], list[PropertyEntry]]:
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    entries: list[PropertyEntry] = []

    for index, line in enumerate(lines):
        stripped = line.strip()
        if not stripped or stripped.startswith("#") or "=" not in line:
            continue
        name, raw_value = line.split("=", 1)
        value = raw_value.rstrip("\n")
        entries.append(
            PropertyEntry(
                file_path=path,
                line_index=index,
                name=name.strip(),
                raw_value=value,
                version=extract_version(value.strip()),
            )
        )

    return lines, entries


def resolve_source_dir(higress_dir: Path, plugin_name: str) -> Path | None:
    candidates = [
        higress_dir / "plugins/wasm-go/extensions" / plugin_name,
        higress_dir / "plugins/wasm-rust/extensions" / plugin_name,
        higress_dir / "plugins/wasm-cpp/extensions" / plugin_name.replace("-", "_"),
    ]
    alias = ALIAS_MAP.get(plugin_name)
    if alias:
        candidates.extend(
            [
                higress_dir / "plugins/wasm-go/extensions" / alias,
                higress_dir / "plugins/wasm-rust/extensions" / alias,
                higress_dir / "plugins/wasm-cpp/extensions" / alias.replace("-", "_"),
            ]
        )

    for candidate in candidates:
        if candidate.is_dir():
            return candidate
    return None


def read_source_version(source_dir: Path) -> str:
    version_file = source_dir / "VERSION"
    if not version_file.is_file():
        raise FileNotFoundError(f"VERSION file not found: {version_file}")
    return version_file.read_text(encoding="utf-8").strip()


def major_version(version: str) -> str:
    match = re.match(r"^(\d+)", version)
    return match.group(1) if match else ""


def choose_effective_version(
    plugin_name: str,
    current_version: str,
    source_version: str,
    force_source_plugins: set[str],
) -> tuple[str, str]:
    if current_version == source_version:
        return source_version, "source-version"
    if plugin_name in force_source_plugins:
        return source_version, "forced-source-version"

    current_major = major_version(current_version)
    source_major = major_version(source_version)
    if current_major and source_major and current_major == source_major:
        return source_version, "same-major-source-version"

    return current_version, f"kept-current-version (source={source_version})"


def update_spec_version(spec_path: Path, version: str, dry_run: bool) -> bool:
    if not spec_path.is_file():
        return False

    lines = spec_path.read_text(encoding="utf-8").splitlines(keepends=True)
    in_info_block = False
    changed = False

    for index, line in enumerate(lines):
        stripped = line.strip()
        indent = len(line) - len(line.lstrip(" "))

        if not in_info_block:
            if stripped == "info:":
                in_info_block = True
            continue

        if indent == 0 and stripped:
            break

        if indent == 2 and stripped.startswith("version:"):
            prefix = line.split("version:", 1)[0] + "version: "
            newline = "\n" if line.endswith("\n") else ""
            current_version = stripped.split(":", 1)[1].strip()
            if current_version != version:
                lines[index] = f"{prefix}{version}{newline}"
                changed = True
            break

    if changed and not dry_run:
        spec_path.write_text("".join(lines), encoding="utf-8")
    return changed


def main() -> int:
    args = parse_args()

    higress_dir = Path(args.higress_dir).resolve()
    console_properties_path = Path(args.console_plugin_properties_file).resolve()
    plugin_server_properties_path = Path(args.plugin_server_properties_file).resolve()
    console_resource_root = Path(args.console_plugin_resource_dir).resolve()
    force_source_plugins = {
        item.strip()
        for item in args.force_source_version_plugins.split(",")
        if item.strip()
    }

    property_files = [console_properties_path, plugin_server_properties_path]
    property_contents: dict[Path, list[str]] = {}
    entries_by_plugin: dict[str, list[PropertyEntry]] = {}

    for path in property_files:
        lines, entries = load_properties(path)
        property_contents[path] = lines
        for entry in entries:
            entries_by_plugin.setdefault(entry.name, []).append(entry)

    mismatched_versions = {
        plugin_name: sorted({entry.version for entry in entries})
        for plugin_name, entries in entries_by_plugin.items()
        if len({entry.version for entry in entries}) > 1
    }
    if mismatched_versions:
        for plugin_name, versions in sorted(mismatched_versions.items()):
            print(
                f"[ERROR] Version mismatch across properties files for {plugin_name}: {versions}",
                file=sys.stderr,
            )
        return 1

    property_updates = 0
    spec_updates = 0

    for plugin_name in sorted(entries_by_plugin):
        current_version = entries_by_plugin[plugin_name][0].version
        source_dir = resolve_source_dir(higress_dir, plugin_name)
        if source_dir is None:
            print(f"[skip] {plugin_name}: no local source directory, keep {current_version}")
            continue

        try:
            source_version = read_source_version(source_dir)
        except FileNotFoundError as exc:
            print(f"[ERROR] {exc}", file=sys.stderr)
            return 1

        effective_version, reason = choose_effective_version(
            plugin_name,
            current_version,
            source_version,
            force_source_plugins,
        )

        if effective_version != current_version:
            for entry in entries_by_plugin[plugin_name]:
                new_value = replace_version(entry.raw_value.strip(), effective_version)
                newline = "\n" if property_contents[entry.file_path][entry.line_index].endswith("\n") else ""
                property_contents[entry.file_path][entry.line_index] = (
                    f"{entry.name}={new_value}{newline}"
                )
                property_updates += 1

            action = "would update" if args.dry_run else "updated"
            print(
                f"[sync] {plugin_name}: {current_version} -> {effective_version} ({reason}, {action} properties)"
            )
        else:
            print(f"[keep] {plugin_name}: {current_version} ({reason})")

        spec_path = console_resource_root / plugin_name / "spec.yaml"
        if update_spec_version(spec_path, effective_version, args.dry_run):
            spec_updates += 1
            action = "would update" if args.dry_run else "updated"
            print(f"[sync] {plugin_name}: {action} spec version -> {effective_version}")

    if not args.dry_run:
        for path, lines in property_contents.items():
            path.write_text("".join(lines), encoding="utf-8")

    action = "Would update" if args.dry_run else "Updated"
    print(f"{action} {property_updates} properties entries and {spec_updates} spec files.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
