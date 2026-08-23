#!/usr/bin/env bash
#SBATCH --job-name=otter-bs-pdx-legacy-step2-check
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=24:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/step2-check-bs-pdx-legacy-sortmem-controller.log
set -uo pipefail

export PATH="/public3/home/scg9946/.cargo/bin:/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"

runtime_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
otter_binary="${runtime_root}/otter"
craftmake_binary="${runtime_root}/craftmake"

run_configuration="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR23802966-legacy-equivalent/runs/run-20260821T002226Z-nfioxh/run.yaml"

printf 'step2-check-start legacy %s\n' "$(date -u +%FT%TZ)"
"${otter_binary}" run --foreground --config "${run_configuration}" --executor craftmake --phase step2-check --craftmake-binary "${craftmake_binary}" --parallel-jobs 2 --catalog "${CRAFTMAKE_WORKFLOW_CATALOG}"
project_exit=$?
printf 'step2-check-exit legacy %d %s\n' "${project_exit}" "$(date -u +%FT%TZ)"
exit "${project_exit}"
