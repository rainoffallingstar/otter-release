#!/usr/bin/env bash
#SBATCH --job-name=otter-resolve-bs-pdx-sortmem
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=01:30:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/resolve-bs-pdx-sortmem-slurm.log
set -u

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"

otter_bin="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/otter"
projects_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects"

printf '=== modern bs-pdx resolve ===\n'
"${otter_bin}" config resolve \
  --project "${projects_root}/bs-pdx-SRR23802966/project.yaml" \
  --site paracloud-gate6 \
  --parent-run-id run-20260817T001228Z-vclyug 2>&1
printf 'modern-exit=%d\n' "$?"

printf '=== legacy bs-pdx resolve ===\n'
"${otter_bin}" config resolve \
  --project "${projects_root}/bs-pdx-SRR23802966-legacy-equivalent/project.yaml" \
  --site paracloud-gate6 \
  --parent-run-id run-20260817T102532Z-emvbei 2>&1
printf 'legacy-exit=%d\n' "$?"

printf 'resolve-bs-pdx-sortmem-complete\n'
