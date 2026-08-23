#!/usr/bin/env bash
#SBATCH --job-name=otter-publish-rna3
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=16G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/publish-rna-all-controller.log
set -uo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/opt/slurm/slurm/bin:${PATH}"
export CRAFTMAKE_WORKFLOW_CATALOG="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog"

runtime_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
projects_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects"
otter_binary="${runtime_root}/otter"
craftmake_binary="${runtime_root}/craftmake"

run_specifications=(
  "human-rnaseq-SRR1039508:run-20260815T134115Z-fvhpxa"
  "human-rnaseq-SRR018258:run-20260815T134119Z-vdwxod"
  "mouse-rnaseq-SRR037954:run-20260815T134123Z-iusjfz"
)

controller_exit=0
for run_specification in "${run_specifications[@]}"; do
  project_name="${run_specification%%:*}"
  run_identifier="${run_specification##*:}"
  run_configuration="${projects_root}/${project_name}/runs/${run_identifier}/run.yaml"

  printf 'publish-start %s %s\n' "${project_name}" "$(date -u +%FT%TZ)"
  "${otter_binary}" run --foreground --config "${run_configuration}" --executor craftmake --phase publish --craftmake-binary "${craftmake_binary}" --parallel-jobs 2
  project_exit=$?
  printf 'publish-exit %s %d %s\n' "${project_name}" "${project_exit}" "$(date -u +%FT%TZ)"

  if [ "${project_exit}" -ne 0 ]; then
    controller_exit="${project_exit}"
  fi
done

exit "${controller_exit}"
