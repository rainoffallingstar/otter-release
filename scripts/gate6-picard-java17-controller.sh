#!/usr/bin/env bash
#SBATCH --job-name=gate6-picard-java17
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=2
#SBATCH --mem=8G
#SBATCH --time=01:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-picard-java17-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
environment_name="gate6-picard-java17"
environment_yaml="${runtime_directory}/${environment_name}.yaml"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/picard-java17-${SLURM_JOB_ID}"
enva_binary="/public3/home/scg9946/.cargo/bin/enva"

if [[ ! -r "${environment_yaml}" ]]; then
    printf 'missing environment manifest: %s\n' "${environment_yaml}" >&2
    exit 1
fi
if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite existing evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"

"${enva_binary}" create --yaml "${environment_yaml}" --name "${environment_name}" --output stream 2>&1 | tee "${evidence_directory}/enva-create.log"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'environment_name=%s\n' "${environment_name}"
    printf 'environment_yaml_sha256=%s\n' "$(sha256sum "${environment_yaml}" | awk '{print $1}')"
    printf 'enva_path=%s\n' "${enva_binary}"
    printf 'enva_version=%s\n' "$("${enva_binary}" --version)"
} > "${evidence_directory}/manifest.properties"

"${enva_binary}" run "${environment_name}" -- java -version > "${evidence_directory}/java-version.txt" 2>&1
"${enva_binary}" run "${environment_name}" -- picard --version > "${evidence_directory}/picard-version.txt" 2>&1
"${enva_binary}" run "${environment_name}" -- picard SetNmMdAndUqTags --help > "${evidence_directory}/picard-set-nm-help.txt" 2>&1

printf '%s\n' "${evidence_directory}"
