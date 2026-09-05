#!/usr/bin/env bash
#SBATCH --job-name=gate6-fragment-stratify
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-fragment-stratify-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/fragment-disagreement-stratification-${SLURM_JOB_ID}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
modern_job_id="${GATE6_MODERN_JOB_ID:?GATE6_MODERN_JOB_ID is required}"
legacy_job_id="${GATE6_LEGACY_JOB_ID:?GATE6_LEGACY_JOB_ID is required}"
modern_evidence_prefix="${GATE6_MODERN_EVIDENCE_PREFIX:-modern-xenofilx-fixed-strand-aware}"
legacy_evidence_prefix="${GATE6_LEGACY_EVIDENCE_PREFIX:-legacy-picard-strand-aware-xenofilter}"
source_membership_job_id="${GATE6_MEMBERSHIP_JOB_ID:?GATE6_MEMBERSHIP_JOB_ID is required}"
membership_evidence_prefix="${GATE6_MEMBERSHIP_EVIDENCE_PREFIX:-fragment-membership-strand-aware-legacy}"
sample_size="${GATE6_FRAGMENT_SAMPLE_SIZE:-10000}"
stratification_script="${runtime_directory}/gate6-fragment-disagreement-stratify.py"
disagreements="${evidence_root}/${membership_evidence_prefix}-${source_membership_job_id}/disagreements.tsv.gz"
original_graft="${project_root}/work/bsmap/SRR36187610_hg38.bam"
original_host="${project_root}/work/bsmap/SRR36187610_mm10.bam"
modern_filtered="${evidence_root}/${modern_evidence_prefix}-${modern_job_id}/output/SRR36187610_fixed_hg38_Filtered.bam"
legacy_filtered="${evidence_root}/${legacy_evidence_prefix}-${legacy_job_id}/xenofilter-destination/Filtered_bams/SRR36187610_hg38.picard-strand-aware-nm_Filtered.bam"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
if [[ ! "${sample_size}" =~ ^[1-9][0-9]*$ ]]; then
    printf 'sample size must be a positive integer: %s\n' "${sample_size}" >&2
    exit 2
fi
for required_path in \
    "${samtools_binary}" "${stratification_script}" "${disagreements}" \
    "${original_graft}" "${original_host}" "${modern_filtered}" "${legacy_filtered}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'membership_job_id=%s\n' "${source_membership_job_id}"
    printf 'sample_size=%s\n' "${sample_size}"
    printf 'sample_seed=gate6-strand-aware-v1\n'
    printf 'selection_classification=graft/discarded\n'
    printf 'stratification_script_sha256=%s\n' "$(sha256sum "${stratification_script}" | awk '{print $1}')"
    printf 'disagreements_sha256=%s\n' "$(sha256sum "${disagreements}" | awk '{print $1}')"
    printf 'original_graft_sha256=%s\n' "$(sha256sum "${original_graft}" | awk '{print $1}')"
    printf 'original_host_sha256=%s\n' "$(sha256sum "${original_host}" | awk '{print $1}')"
    printf 'modern_filtered_sha256=%s\n' "$(sha256sum "${modern_filtered}" | awk '{print $1}')"
    printf 'legacy_filtered_sha256=%s\n' "$(sha256sum "${legacy_filtered}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v \
    "${original_graft}" "${original_host}" "${modern_filtered}" "${legacy_filtered}" \
    > "${evidence_directory}/input.quickcheck.txt" 2>&1
python3 "${stratification_script}" \
    --samtools "${samtools_binary}" \
    --disagreements "${disagreements}" \
    --original-graft "${original_graft}" \
    --original-host "${original_host}" \
    --modern-filtered "${modern_filtered}" \
    --legacy-filtered "${legacy_filtered}" \
    --sample-size "${sample_size}" \
    --report "${evidence_directory}/stratification.json" \
    --samples "${evidence_directory}/sampled-fragments.tsv"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
