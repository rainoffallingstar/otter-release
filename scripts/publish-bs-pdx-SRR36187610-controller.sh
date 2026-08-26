#!/usr/bin/env bash
#SBATCH --job-name=otter-publish-bs-pdx
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/publish-bs-pdx-SRR36187610-controller.log
set -euo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"

runtime_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
otter_binary="${runtime_root}/otter"
craftmake_binary="${runtime_root}/craftmake"
run_yaml="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa/run.yaml"

printf 'publish-start bs-pdx-SRR36187610 %s\n' "$(date -u +%FT%TZ)"
"${otter_binary}" run --foreground --config "${run_yaml}" --executor craftmake --phase publish --craftmake-binary "${craftmake_binary}" --parallel-jobs 2 --catalog "${CRAFTMAKE_WORKFLOW_CATALOG}"
project_exit=$?
printf 'publish-exit bs-pdx-SRR36187610 %d %s\n' "${project_exit}" "$(date -u +%FT%TZ)"
exit "${project_exit}"
