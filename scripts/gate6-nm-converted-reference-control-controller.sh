#!/usr/bin/env bash
#SBATCH --job-name=gate6-nm-converted-control
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-nm-converted-control-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
dataset_label="${GATE6_NM_DATASET_LABEL:?GATE6_NM_DATASET_LABEL is required}"
source_job_id="${GATE6_NM_SOURCE_JOB_ID:?GATE6_NM_SOURCE_JOB_ID is required}"
input_bam="${GATE6_NM_INPUT_BAM:?GATE6_NM_INPUT_BAM is required}"
record_limit="${GATE6_NM_RECORD_LIMIT:-200000}"
evidence_directory="${evidence_root}/nm-read-parity-${dataset_label}-${source_job_id}"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
control_compare="${runtime_directory}/gate6-nm-converted-reference-compare.py"

for required_path in \
    "${input_bam}" "${samtools_binary}" "${control_compare}" \
    "${evidence_directory}/oracle-records.tsv" \
    "${evidence_directory}/xenofilx-records.tsv" \
    "${evidence_directory}/input.xg-ct.bam" \
    "${evidence_directory}/input.xg-ga.bam" \
    "${evidence_directory}/oracle-xg-ct-records.tsv" \
    "${evidence_directory}/oracle-xg-ga-records.tsv" \
    "${evidence_directory}/xenofilx-xg-ct-records.tsv" \
    "${evidence_directory}/xenofilx-xg-ga-records.tsv" \
    "${evidence_directory}/input.xg-ct.picard-nm.bam" \
    "${evidence_directory}/input.xg-ga.picard-nm.bam"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

python3 "${control_compare}" \
    --samtools "${samtools_binary}" \
    --original-bam "${input_bam}" \
    --original-oracle "${evidence_directory}/oracle-records.tsv" \
    --original-xenofilx "${evidence_directory}/xenofilx-records.tsv" \
    --control "xg_ct=${evidence_directory}/input.xg-ct.bam,${evidence_directory}/oracle-xg-ct-records.tsv,${evidence_directory}/xenofilx-xg-ct-records.tsv,${evidence_directory}/input.xg-ct.picard-nm.bam" \
    --control "xg_ga=${evidence_directory}/input.xg-ga.bam,${evidence_directory}/oracle-xg-ga-records.tsv,${evidence_directory}/xenofilx-xg-ga-records.tsv,${evidence_directory}/input.xg-ga.picard-nm.bam" \
    --records-read "${record_limit}" \
    --report "${evidence_directory}/converted-reference-control-comparison-v2.json"
printf 'converted_reference_control_job_id=%s\n' "${SLURM_JOB_ID}" >> "${evidence_directory}/manifest.properties"
printf 'converted_reference_control_completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
