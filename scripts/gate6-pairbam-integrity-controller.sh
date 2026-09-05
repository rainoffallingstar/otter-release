#!/usr/bin/env bash
#SBATCH --job-name=gate6-pairbam-integrity
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-pairbam-integrity-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/pairbam-integrity-${SLURM_JOB_ID}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
pairbam_binary="/public3/home/scg9946/.cargo/bin/pairbam"
paircompare_binary="${runtime_directory}/gate6-paircompare"
input_bam="${project_root}/work/bsmap/SRR36187610_mm10.bam"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for executable_path in "${samtools_binary}" "${pairbam_binary}" "${paircompare_binary}"; do
    if [[ ! -x "${executable_path}" ]]; then
        printf 'missing executable: %s\n' "${executable_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
temporary_directory="${evidence_directory}/temporary"
mkdir "${temporary_directory}"
mkdir "${temporary_directory}/paircompare-sort"

name_sorted_input="${temporary_directory}/input.queryname.bam"
pairbam_output="${evidence_directory}/host.paired.bam"
filtered_names="${evidence_directory}/host.paired_filtered_readnames.txt"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'samtools_path=%s\n' "${samtools_binary}"
    printf 'pairbam_path=%s\n' "${pairbam_binary}"
    printf 'paircompare_path=%s\n' "${paircompare_binary}"
    printf 'pairbam_version=%s\n' "$("${pairbam_binary}" --version)"
    printf 'paircompare_sha256=%s\n' "$(sha256sum "${paircompare_binary}" | awk '{print $1}')"
    printf 'input_sha256=%s\n' "$(sha256sum "${input_bam}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v "${input_bam}" > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${paircompare_binary}" sort-name \
    --input "${input_bam}" \
    --output "${name_sorted_input}" \
    --temporary "${temporary_directory}/paircompare-sort"
"${pairbam_binary}" "${input_bam}" "${pairbam_output}" > "${evidence_directory}/pairbam.stdout.txt" 2> "${evidence_directory}/pairbam.stderr.txt"
"${samtools_binary}" quickcheck -v "${pairbam_output}" > "${evidence_directory}/output.quickcheck.txt" 2>&1
"${paircompare_binary}" \
    --input "${name_sorted_input}" \
    --output "${pairbam_output}" \
    --filtered-names "${filtered_names}" \
    --report "${evidence_directory}/comparison.json"

{
    printf 'input_records\t'
    "${samtools_binary}" view -c "${input_bam}"
    printf 'name_sorted_records\t'
    "${samtools_binary}" view -c "${name_sorted_input}"
    printf 'output_records\t'
    "${samtools_binary}" view -c "${pairbam_output}"
    printf 'output_sha256\t'
    sha256sum "${pairbam_output}" | awk '{print $1}'
    printf 'filtered_names_sha256\t'
    sha256sum "${filtered_names}" | awk '{print $1}'
} > "${evidence_directory}/summary.tsv"

# Retain the native name-sorted input in evidence because the comparator report
# references it and it is part of the reproducible subset audit.
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
