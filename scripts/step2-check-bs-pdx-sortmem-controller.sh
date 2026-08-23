#!/usr/bin/env bash
#SBATCH --job-name=otter-bs-pdx-step2-check-sortmem
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=24:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/step2-check-bs-pdx-sortmem-controller.log
set -uo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"

runtime_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
projects_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects"
otter_binary="${runtime_root}/otter"
craftmake_binary="${runtime_root}/craftmake"

# Independent immutable snapshots for modern and legacy. Each has its own
# isolated Craftmake state; no cross-snapshot cache/state reuse.
run_specifications=(
  "bs-pdx-SRR23802966:run-20260820T103531Z-tjhxpo"
  "bs-pdx-SRR23802966-legacy-equivalent:run-20260820T103656Z-gnahig"
)

controller_exit=0
for run_specification in "${run_specifications[@]}"; do
  project_name="${run_specification%%:*}"
  run_identifier="${run_specification##*:}"
  run_configuration="${projects_root}/${project_name}/runs/${run_identifier}/run.yaml"

  printf 'step2-check-start %s %s\n' "${project_name}" "$(date -u +%FT%TZ)"
  "${otter_binary}" run --foreground --config "${run_configuration}" --executor craftmake --phase step2-check --craftmake-binary "${craftmake_binary}" --parallel-jobs 2 --catalog "${CRAFTMAKE_WORKFLOW_CATALOG}"
  project_exit=$?
  printf 'step2-check-exit %s %d %s\n' "${project_name}" "${project_exit}" "$(date -u +%FT%TZ)"

  if [ "${project_exit}" -ne 0 ]; then
    controller_exit="${project_exit}"
  fi
done

exit "${controller_exit}"
