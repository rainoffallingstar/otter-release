#!/usr/bin/env bash
#SBATCH --job-name=gate6-nm-score-audit
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-nm-score-audit-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/nm-score-audit-${SLURM_JOB_ID}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
record_limit=200000

nm_oracle="${runtime_directory}/gate6-nmoracle"
xenofilx_audit="${runtime_directory}/gate6-xenofilx-scoreaudit"
comparison_script="${runtime_directory}/gate6-nm-score-compare.py"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for executable_path in "${nm_oracle}" "${xenofilx_audit}" "${comparison_script}"; do
    if [[ ! -x "${executable_path}" ]]; then
        printf 'missing executable audit dependency: %s\n' "${executable_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'record_limit=%s\n' "${record_limit}"
    printf 'nm_oracle_sha256=%s\n' "$(sha256sum "${nm_oracle}" | awk '{print $1}')"
    printf 'xenofilx_audit_sha256=%s\n' "$(sha256sum "${xenofilx_audit}" | awk '{print $1}')"
    printf 'comparison_script_sha256=%s\n' "$(sha256sum "${comparison_script}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

run_score_audit() {
    local label="$1"
    local input_bam="$2"
    local reference_fasta="$3"

    "${nm_oracle}" \
        --input "${input_bam}" \
        --reference "${reference_fasta}" \
        --limit "${record_limit}" \
        --report "${evidence_directory}/${label}.oracle-summary.json" \
        --records "${evidence_directory}/${label}.oracle-records.tsv"
    "${xenofilx_audit}" \
        --input "${input_bam}" \
        --reference "${reference_fasta}" \
        --bisulfite \
        --limit "${record_limit}" \
        --report "${evidence_directory}/${label}.xenofilx-records.tsv"
    "${comparison_script}" \
        --oracle "${evidence_directory}/${label}.oracle-records.tsv" \
        --xenofilx "${evidence_directory}/${label}.xenofilx-records.tsv" \
        --report "${evidence_directory}/${label}.comparison.json"
}

run_score_audit \
    graft \
    "${project_root}/work/bsmap/SRR36187610_hg38.bam" \
    "/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"
run_score_audit \
    host \
    "${project_root}/work/bsmap/SRR36187610_mm10.bam" \
    "/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa"

printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
