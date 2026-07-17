#!/usr/bin/env bash

set -euo pipefail

DIST_DIR="${1:-dist}"

required=(
  xdxtools-linux-amd64-static
  enva-linux-amd64-static
  xenofilter-linux-amd64-static
  paireads-linux-amd64-static
  htseq2matrix-linux-amd64-static
  gomats-linux-amd64-static
  qctb-linux-amd64-static
  fqc-linux-amd64-static
  methrix-linux-amd64-static
)

missing=()
not_exec=()

for artifact in "${required[@]}"; do
  path="${DIST_DIR}/${artifact}"
  if [ ! -f "$path" ]; then
    echo "✗ missing: $artifact"
    missing+=("$artifact")
    continue
  fi
  if [ ! -x "$path" ]; then
    echo "✗ not executable: $artifact"
    not_exec+=("$artifact")
    continue
  fi
  echo "✓ $artifact"
done

if [ ${#missing[@]} -gt 0 ] || [ ${#not_exec[@]} -gt 0 ]; then
  [ ${#missing[@]} -gt 0 ] && echo "Missing artifacts: ${missing[*]}"
  [ ${#not_exec[@]} -gt 0 ] && echo "Non-executable artifacts: ${not_exec[*]}"
  exit 1
fi

echo "All required release artifacts are present and executable."
