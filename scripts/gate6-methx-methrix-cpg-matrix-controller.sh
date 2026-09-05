#!/usr/bin/env bash
#SBATCH --job-name=gate6-methx-methrix-matrix
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=48G
#SBATCH --time=24:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/evidence/gate6-methx-methrix-cpg-matrix-20260904/controller-%j.out
#SBATCH --error=/public3/home/scg9946/otter-gate6/evidence/gate6-methx-methrix-cpg-matrix-20260904/controller-%j.err
set -euo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/public3/home/scg9946/.cargo/bin:/opt/slurm/slurm/bin:${PATH}"

cov_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260821T112353Z-tunvxb/work/mCall"
methx_h5="${cov_directory}/methrixh5/assays.h5"
output_directory="/public3/home/scg9946/otter-gate6/evidence/gate6-methx-methrix-cpg-matrix-20260904/run-${SLURM_JOB_ID}"
comparison_script="${output_directory}/gate6-methx-methrix-cpg-matrix-compare.R"
local_script="${GATE6_METHX_METHRIX_SCRIPT:-/public3/home/scg9946/otter-gate6/evidence/gate6-methx-methrix-cpg-matrix-20260904/gate6-methx-methrix-cpg-matrix-compare.R}"

mkdir -p "${output_directory}"
cp "${local_script}" "${comparison_script}"
chmod 0555 "${comparison_script}"

printf 'job_id=%s\n' "${SLURM_JOB_ID}"
printf 'started_at_utc=%s\n' "$(date -u +%FT%TZ)"
printf 'cov_directory=%s\n' "${cov_directory}"
printf 'methx_h5=%s\n' "${methx_h5}"
printf 'output_directory=%s\n' "${output_directory}"
sha256sum "${comparison_script}" "${methx_h5}" "${cov_directory}/SRR36187610_nsort.bismark.cov.gz"
Rscript "${comparison_script}" "${cov_directory}" "${methx_h5}" hg38 "${output_directory}"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)"
