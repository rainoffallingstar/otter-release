#!/usr/bin/env bash
#SBATCH --job-name=gate6-fragment-membership-fixed
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-fragment-membership-fixed-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/fragment-membership-fixed-strand-aware-${SLURM_JOB_ID}"
legacy_job_id="${GATE6_LEGACY_JOB_ID:?GATE6_LEGACY_JOB_ID is required}"
modern_job_id="${GATE6_MODERN_JOB_ID:?GATE6_MODERN_JOB_ID is required}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
comparison_script="${runtime_directory}/gate6-fragment-membership-compare.py"
original_graft="${project_root}/work/bsmap/SRR36187610_hg38.bam"
modern_filtered="${evidence_root}/modern-xenofilx-fixed-strand-aware-${modern_job_id}/output/SRR36187610_fixed_hg38_Filtered.bam"
legacy_evidence_directory="${evidence_root}/legacy-picard-xenofilter-${legacy_job_id}"
legacy_filtered="${legacy_evidence_directory}/xenofilter-destination/Filtered_bams/SRR36187610_hg38.picard-nm_Filtered.bam"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${samtools_binary}" "${comparison_script}" "${original_graft}" "${modern_filtered}" "${legacy_filtered}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
mkdir "${evidence_directory}/temporary"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'legacy_job_id=%s\n' "${legacy_job_id}"
    printf 'modern_job_id=%s\n' "${modern_job_id}"
    printf 'comparison_script_sha256=%s\n' "$(sha256sum "${comparison_script}" | awk '{print $1}')"
    printf 'original_graft_sha256=%s\n' "$(sha256sum "${original_graft}" | awk '{print $1}')"
    printf 'modern_filtered_sha256=%s\n' "$(sha256sum "${modern_filtered}" | awk '{print $1}')"
    printf 'legacy_filtered_sha256=%s\n' "$(sha256sum "${legacy_filtered}" | awk '{print $1}')"
    printf 'modern_nm_semantics=strand_aware_bisulfite\n'
    printf 'classification_scope=graft-selected-versus-discarded\n'
    printf 'classification_limitation=XenofilteR_filtered_BAM_only_exposes_selected_graft_membership\n'
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v "${original_graft}" "${modern_filtered}" "${legacy_filtered}" > "${evidence_directory}/quickcheck.txt" 2>&1
python3 "${comparison_script}" \
    --samtools "${samtools_binary}" \
    --original-graft "${original_graft}" \
    --modern-filtered "${modern_filtered}" \
    --legacy-filtered "${legacy_filtered}" \
    --report "${evidence_directory}/comparison.json" \
    --disagreements "${evidence_directory}/disagreements.tsv.gz" \
    --temporary-directory "${evidence_directory}/temporary"

rm -rf "${evidence_directory}/temporary"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
