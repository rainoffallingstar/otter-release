#!/usr/bin/env bash
#SBATCH --job-name=otter-bs-pdx-step2-check-plan
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/step2-check-bs-pdx-sortmem-plan.log
set -uo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"

runtime_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
otter_binary="${runtime_root}/otter"
craftmake_binary="${runtime_root}/craftmake"

modern_config="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR23802966/runs/run-20260820T103531Z-tjhxpo/run.yaml"
legacy_config="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR23802966-legacy-equivalent/runs/run-20260820T103656Z-gnahig/run.yaml"

for label_config in "modern:${modern_config}" "legacy:${legacy_config}"; do
  label="${label_config%%:*}"
  config="${label_config##*:}"
  printf '=== plan %s ===\n' "${label}"
  "${otter_binary}" run --dry-run --config "${config}" --executor craftmake --phase step2-check --craftmake-binary "${craftmake_binary}" --parallel-jobs 2 --catalog "${CRAFTMAKE_WORKFLOW_CATALOG}" 2>&1
  printf 'plan-exit %s %d\n' "${label}" "$?"
done
printf 'plan-bs-pdx-sortmem-complete\n'
