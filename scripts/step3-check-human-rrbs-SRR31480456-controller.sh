#!/usr/bin/env bash
#SBATCH --job-name=otter-step3-check-human-rrbs-SRR31480456
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=12:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/step3-check-human-rrbs-SRR31480456-controller.log
set -euo pipefail
export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"
otter_bin="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/otter"
craftmake_bin="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/craftmake"
run_yaml="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/human-rrbs-SRR31480456/runs/run-20260815T093050Z-ykunpt/run.yaml"
printf 'controller-start %s\n' "$(date -u +%FT%TZ)"
"${otter_bin}" run --foreground --config "${run_yaml}" --executor craftmake --phase step3-check --craftmake-binary "${craftmake_bin}" --parallel-jobs 4 --catalog "${CRAFTMAKE_WORKFLOW_CATALOG}"
printf 'controller-exit %d %s\n' "$?" "$(date -u +%FT%TZ)"
