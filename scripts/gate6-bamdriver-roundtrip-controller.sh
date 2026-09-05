#!/usr/bin/env bash
#SBATCH --job-name=gate6-bamdriver-roundtrip
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-bamdriver-roundtrip-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
roundtrip_binary="${runtime_directory}/gate6-bamroundtrip"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/bamdriver-roundtrip-${SLURM_JOB_ID}"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"

if [[ ! -x "${roundtrip_binary}" ]]; then
    printf 'missing executable round-trip binary: %s\n' "${roundtrip_binary}" >&2
    exit 1
fi
if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'roundtrip_binary=%s\n' "${roundtrip_binary}"
    printf 'roundtrip_binary_sha256=%s\n' "$(sha256sum "${roundtrip_binary}" | awk '{print $1}')"
    printf 'samtools_path=%s\n' "${samtools_binary}"
    printf 'samtools_version=%s\n' "$("${samtools_binary}" --version 2>&1 | head -1)"
} > "${evidence_directory}/manifest.properties"

run_roundtrip() {
    local label="$1"
    local input_bam="$2"
    local output_bam="${evidence_directory}/${label}.roundtrip.bam"

    if [[ ! -r "${input_bam}" ]]; then
        printf 'missing readable input BAM: %s\n' "${input_bam}" >&2
        return 1
    fi

    "${samtools_binary}" quickcheck -v "${input_bam}" > "${evidence_directory}/${label}.input.quickcheck.txt" 2>&1
    "${roundtrip_binary}" roundtrip \
        --input "${input_bam}" \
        --output "${output_bam}" \
        --report "${evidence_directory}/${label}.roundtrip.json"
    "${samtools_binary}" quickcheck -v "${output_bam}" > "${evidence_directory}/${label}.output.quickcheck.txt" 2>&1
    "${roundtrip_binary}" compare \
        --left "${input_bam}" \
        --right "${output_bam}" \
        --report "${evidence_directory}/${label}.comparison.json"

    {
        printf 'input_records\t'
        "${samtools_binary}" view -c "${input_bam}"
        printf 'output_records\t'
        "${samtools_binary}" view -c "${output_bam}"
        printf 'input_sha256\t'
        sha256sum "${input_bam}" | awk '{print $1}'
        printf 'output_sha256\t'
        sha256sum "${output_bam}" | awk '{print $1}'
    } > "${evidence_directory}/${label}.summary.tsv"
}

run_roundtrip graft-original "${project_root}/work/bsmap/SRR36187610_hg38.bam"
run_roundtrip host-original "${project_root}/work/bsmap/SRR36187610_mm10.bam"

printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
