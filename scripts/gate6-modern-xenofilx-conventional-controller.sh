#!/usr/bin/env bash
#SBATCH --job-name=gate6-modern-xenofilx-conventional
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=8
#SBATCH --mem=160G
#SBATCH --time=08:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-modern-xenofilx-conventional-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/modern-xenofilx-conventional-ndcwfa-${SLURM_JOB_ID}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
graft_input="${project_root}/work/bsmap/SRR36187610_hg38.bam"
host_input="${project_root}/work/bsmap/SRR36187610_mm10.bam"
graft_reference="/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"
host_reference="/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
xenofilx_binary="/public3/home/scg9946/.cargo/bin/xenofilx"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${graft_input}" "${host_input}" "${graft_reference}" "${host_reference}" "${samtools_binary}" "${xenofilx_binary}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
mkdir "${evidence_directory}/output"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'run_id=run-20260823T034624Z-ndcwfa\n'
    printf 'graft_input_sha256=%s\n' "$(sha256sum "${graft_input}" | awk '{print $1}')"
    printf 'host_input_sha256=%s\n' "$(sha256sum "${host_input}" | awk '{print $1}')"
    printf 'graft_reference_sha256=%s\n' "$(sha256sum "${graft_reference}" | awk '{print $1}')"
    printf 'host_reference_sha256=%s\n' "$(sha256sum "${host_reference}" | awk '{print $1}')"
    printf 'xenofilx_path=%s\n' "${xenofilx_binary}"
    printf 'xenofilx_sha256=%s\n' "$(sha256sum "${xenofilx_binary}" | awk '{print $1}')"
    printf 'xenofilx_version=%s\n' "$("${xenofilx_binary}" --version)"
    printf 'mm_threshold=6\n'
    printf 'unmapped_penalty=8\n'
    printf 'threads=%s\n' "${SLURM_CPUS_PER_TASK}"
    printf 'recalculate_nm=true\n'
    printf 'bisulfite=false\n'
    printf 'purpose=conventional_nm_control_against_picard_xenofilter\n'
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v "${graft_input}" "${host_input}" > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${xenofilx_binary}" run \
    --graft "${graft_input}" \
    --host "${host_input}" \
    --output "${evidence_directory}/output" \
    --output-names SRR36187610_fixed_hg38_Filtered.bam \
    --graft-ref "${graft_reference}" \
    --host-ref "${host_reference}" \
    --mm-threshold 6 \
    --unmapped-penalty 8 \
    --threads "${SLURM_CPUS_PER_TASK}" \
    --sort-memory 8G \
    --recalculate-nm \
    > "${evidence_directory}/xenofilx.stdout.txt" \
    2> "${evidence_directory}/xenofilx.stderr.txt"

modern_filtered_bam="${evidence_directory}/output/SRR36187610_fixed_hg38_Filtered.bam"
"${samtools_binary}" quickcheck -v "${modern_filtered_bam}" > "${evidence_directory}/output.quickcheck.txt" 2>&1

{
    printf 'graft_input_records\t'
    "${samtools_binary}" view -c "${graft_input}"
    printf 'host_input_records\t'
    "${samtools_binary}" view -c "${host_input}"
    printf 'modern_conventional_filtered_records\t'
    "${samtools_binary}" view -c "${modern_filtered_bam}"
    printf 'modern_conventional_filtered_bam_sha256\t'
    sha256sum "${modern_filtered_bam}" | awk '{print $1}'
} > "${evidence_directory}/summary.tsv"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
