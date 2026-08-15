#!/usr/bin/env bash
#SBATCH --job-name=otter-g6-canary-inputs
#SBATCH --partition=amd_512
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=12:00:00
#SBATCH --array=0-3
#SBATCH --output=/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801/slurm-%A_%a.out
#SBATCH --error=/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801/slurm-%A_%a.err

set -euo pipefail

readonly evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801"
readonly output_root="/public3/home/scg9946/otter-gate6/canaries/gate6-20260801"
readonly publisher="${evidence_root}/scripts/gate6_publish_canary_inputs.py"
readonly configurations=(
  "${evidence_root}/configs/rrbs.json"
  "${evidence_root}/configs/rnaseq.json"
  "${evidence_root}/configs/bs-pdx.json"
  "${evidence_root}/configs/rna-pdx.json"
)
readonly scenario_names=(rrbs rnaseq bs-pdx rna-pdx)

readonly configuration_path="${configurations[${SLURM_ARRAY_TASK_ID}]}"
readonly scenario_name="${scenario_names[${SLURM_ARRAY_TASK_ID}]}"
readonly output_directory="${output_root}/${scenario_name}"

[[ -f "${publisher}" ]]
[[ -f "${configuration_path}" ]]
[[ ! -e "${output_directory}" ]]

python3 "${publisher}" --config "${configuration_path}" --output "${output_directory}"
(
  cd "${output_directory}"
  sha256sum -c checksums.sha256
  gzip -t R1.fastq.gz R2.fastq.gz
)
python3 - "${output_directory}" <<'PY'
import gzip
import json
import sys
from pathlib import Path

output_directory = Path(sys.argv[1])
manifest = json.loads((output_directory / "canary-inputs.json").read_text(encoding="utf-8"))
expected_pairs = manifest["selection"]["paired_record_count"]
for mate_name in ("R1.fastq.gz", "R2.fastq.gz"):
    line_count = 0
    with gzip.open(output_directory / mate_name, "rb") as fastq_file:
        for line_count, _ in enumerate(fastq_file, start=1):
            pass
    if line_count == 0 or line_count % 4 != 0 or line_count // 4 != expected_pairs:
        raise SystemExit(f"{mate_name} has {line_count} lines; expected {expected_pairs} paired records")
PY
chmod 0555 "${output_directory}"
chmod 0444 "${output_directory}"/*
printf 'scenario=%s\noutput=%s\n' "${scenario_name}" "${output_directory}" > "${evidence_root}/${scenario_name}-publication-${SLURM_JOB_ID}.txt"
