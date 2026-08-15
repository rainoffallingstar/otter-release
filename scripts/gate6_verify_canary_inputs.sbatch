#!/usr/bin/env bash
#SBATCH --job-name=otter-g6-canary-verify
#SBATCH --partition=amd_512
#SBATCH --cpus-per-task=2
#SBATCH --mem=8G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801/verify-%j.out
#SBATCH --error=/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801/verify-%j.err

set -euo pipefail

readonly evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801"
readonly output_root="/public3/home/scg9946/otter-gate6/canaries/gate6-20260801"

for scenario_name in rrbs rnaseq bs-pdx rna-pdx; do
  output_directory="${output_root}/${scenario_name}"
  test -d "${output_directory}"
  (
    cd "${output_directory}"
    sha256sum -c checksums.sha256
    gzip -t R1.fastq.gz R2.fastq.gz
  )
  python3 - "${output_directory}" "${scenario_name}" <<'PY'
import gzip
import json
import sys
from pathlib import Path

output_directory = Path(sys.argv[1])
scenario_name = sys.argv[2]
manifest = json.loads((output_directory / "canary-inputs.json").read_text(encoding="utf-8"))
expected_pairs = manifest["selection"]["paired_record_count"]
if manifest["schema_version"] != "otter.canary-inputs/v1":
    raise SystemExit("unexpected input manifest schema")
if manifest["scenario"] != scenario_name:
    raise SystemExit("scenario mismatch")
for mate_name in ("R1.fastq.gz", "R2.fastq.gz"):
    record_count = 0
    with gzip.open(output_directory / mate_name, "rb") as fastq_file:
        while True:
            header = fastq_file.readline()
            if not header:
                break
            sequence = fastq_file.readline()
            separator = fastq_file.readline()
            quality = fastq_file.readline()
            if not sequence or not separator or not quality or not header.startswith(b"@") or not separator.startswith(b"+"):
                raise SystemExit(f"invalid FASTQ record in {mate_name}")
            record_count += 1
    if record_count != expected_pairs:
        raise SystemExit(f"{mate_name} has {record_count} records; expected {expected_pairs}")
print(json.dumps({"scenario": scenario_name, "paired_record_count": expected_pairs, "verified_on": "compute_node"}))
PY
done

sha256sum "${output_root}"/*/canary-inputs.json "${output_root}"/*/checksums.sha256 > "${evidence_root}/compute-node-input-manifests-${SLURM_JOB_ID}.sha256"
printf 'verified_at=%s\njob_id=%s\n' "$(date -u +%FT%TZ)" "${SLURM_JOB_ID}" > "${evidence_root}/compute-node-input-verify-${SLURM_JOB_ID}.txt"
