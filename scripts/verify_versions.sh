#!/usr/bin/env bash

set -euo pipefail

BIN_DIR="${1:-$HOME/.cargo/bin}"
required_bins=(enva fqc xenofilter paireads htseq2matrix methrix-cli qctb gomats)

missing_bins=()
invalid_versions=()

for binary in "${required_bins[@]}"; do
  path="${BIN_DIR}/${binary}"
  if [ ! -x "${path}" ] && [ -x "${BIN_DIR}/${binary}-linux-amd64" ]; then
    path="${BIN_DIR}/${binary}-linux-amd64"
  fi
  if [ ! -x "${path}" ]; then
    echo "✗ ${binary}: missing executable at ${path}"
    missing_bins+=("${binary}")
    continue
  fi

  output="$("${path}" --version 2>&1 | head -1 || true)"
  expected_re="^${binary} [0-9]+\\.[0-9]+\\.[0-9]+([.-][0-9A-Za-z.-]+)?$"
  if printf '%s\n' "${output}" | grep -Eq "${expected_re}"; then
    echo "✓ ${binary}: ${output}"
  else
    echo "✗ ${binary}: invalid --version output: ${output}"
    invalid_versions+=("${binary}")
  fi
done

if [ "${#missing_bins[@]}" -gt 0 ] || [ "${#invalid_versions[@]}" -gt 0 ]; then
  [ "${#missing_bins[@]}" -gt 0 ] && echo "Missing binaries: ${missing_bins[*]}"
  [ "${#invalid_versions[@]}" -gt 0 ] && echo "Invalid version outputs: ${invalid_versions[*]}"
  exit 1
fi

echo "All tool versions are valid."
