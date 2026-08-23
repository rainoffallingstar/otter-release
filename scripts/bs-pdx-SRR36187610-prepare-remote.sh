#!/usr/bin/env bash
set -euo pipefail
readonly root='/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z'
[[ ! -e "${root}" ]]
mkdir -p "${root}/archive" "${root}/runtime" "${root}/workflows/SRAArchiveDecode"
echo "${root}"
