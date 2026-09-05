#!/usr/bin/env bash
#SBATCH --job-name=gate6-rna-mixture-build
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-rna-mixture-build-%j.log
set -euo pipefail

runtime_directory="${GATE6_RUNTIME_DIRECTORY:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime}"
mixture_root="${GATE6_MIXTURE_ROOT:-/public3/home/scg9946/otter-gate6/evidence/gate6-human-mouse-rna-mixtures-20260903}"
mixture_builder="${GATE6_MIXTURE_BUILDER:-${runtime_directory}/gate6-mixture-build.py}"
human_r1="${GATE6_HUMAN_R1:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/human-rnaseq-SRR1039508/data/SRR1039508_R1.fastq.gz}"
human_r2="${GATE6_HUMAN_R2:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/human-rnaseq-SRR1039508/data/SRR1039508_R2.fastq.gz}"
mouse_r1="${GATE6_MOUSE_R1:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/mouse-rnaseq-SRR037954/data/SRR037954_R1.fastq.gz}"
mouse_r2="${GATE6_MOUSE_R2:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/mouse-rnaseq-SRR037954/data/SRR037954_R2.fastq.gz}"
human_fraction="${GATE6_HUMAN_FRACTION:?GATE6_HUMAN_FRACTION is required}"
total_fragments="${GATE6_TOTAL_FRAGMENTS:-1000000}"
replicate="${GATE6_REPLICATE:?GATE6_REPLICATE is required}"
seed="${GATE6_MIXTURE_SEED:-gate6-human-mouse-rna-v1}"
label_suffix="${GATE6_LABEL_SUFFIX:-}"
formatted_fraction="$(python3 -c 'import sys; print(f"{float(sys.argv[1]):.6f}".rstrip("0").rstrip("."))' "${human_fraction}")"
mixture_label="human${formatted_fraction}-replicate${replicate}-n${total_fragments}${label_suffix}"
output_directory="${mixture_root}/${mixture_label}"

if [[ -e "${output_directory}" ]]; then
    printf 'refusing to overwrite mixture directory: %s\n' "${output_directory}" >&2
    exit 1
fi
for required_path in "${mixture_builder}" "${human_r1}" "${human_r2}" "${mouse_r1}" "${mouse_r2}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${mixture_root}"
properties_path="${mixture_root}/build-${SLURM_JOB_ID}.properties"
{
    printf 'job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'mixture_label=%s\n' "${mixture_label}"
    printf 'assay=RNA-seq\n'
    printf 'builder_sha256=%s\n' "$(sha256sum "${mixture_builder}" | awk '{print $1}')"
    printf 'human_r1_sha256=%s\n' "$(sha256sum "${human_r1}" | awk '{print $1}')"
    printf 'human_r2_sha256=%s\n' "$(sha256sum "${human_r2}" | awk '{print $1}')"
    printf 'mouse_r1_sha256=%s\n' "$(sha256sum "${mouse_r1}" | awk '{print $1}')"
    printf 'mouse_r2_sha256=%s\n' "$(sha256sum "${mouse_r2}" | awk '{print $1}')"
    printf 'human_fraction=%s\n' "${human_fraction}"
    printf 'total_fragments=%s\n' "${total_fragments}"
    printf 'replicate=%s\n' "${replicate}"
    printf 'seed=%s\n' "${seed}"
} > "${properties_path}"

python3 "${mixture_builder}" \
    --human-r1 "${human_r1}" \
    --human-r2 "${human_r2}" \
    --mouse-r1 "${mouse_r1}" \
    --mouse-r2 "${mouse_r2}" \
    --output-directory "${output_directory}" \
    --human-fraction "${human_fraction}" \
    --total-fragments "${total_fragments}" \
    --seed "${seed}" \
    --replicate "${replicate}" \
    > "${output_directory}.manifest.stdout.json"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${properties_path}"
printf '%s\n' "${output_directory}"
