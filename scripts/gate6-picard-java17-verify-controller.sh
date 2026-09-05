#!/usr/bin/env bash
#SBATCH --job-name=gate6-picard-verify
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=1
#SBATCH --mem=4G
#SBATCH --time=00:15:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-picard-verify-%j.log
set -euo pipefail

prefix="/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17"
java_binary="${prefix}/bin/java"
picard_jar="${prefix}/share/picard-3.4.0-0/picard.jar"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/picard-java17-verify-${SLURM_JOB_ID}"

for required_path in "${java_binary}" "${picard_jar}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable Picard runtime component: %s\n' "${required_path}" >&2
        exit 1
    fi
done
if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite existing evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"

"${java_binary}" -version > "${evidence_directory}/java-version.txt" 2>&1

run_picard_information_command() {
    local output_path="$1"
    shift
    set +e
    "${java_binary}" -jar "${picard_jar}" "$@" > "${output_path}" 2>&1
    local command_status=$?
    set -e
    if [[ ${command_status} -ne 0 && ${command_status} -ne 1 ]]; then
        printf 'Picard information command exited %d: %s\n' "${command_status}" "$*" >&2
        return "${command_status}"
    fi
}

run_picard_information_command "${evidence_directory}/picard-version.txt" MarkDuplicates --version
run_picard_information_command "${evidence_directory}/picard-set-nm-help.txt" SetNmMdAndUqTags --help

grep -F 'Version:3.4.0' "${evidence_directory}/picard-version.txt" > /dev/null
grep -F 'USAGE: SetNmMdAndUqTags' "${evidence_directory}/picard-set-nm-help.txt" > /dev/null

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'environment_prefix=%s\n' "${prefix}"
    printf 'java_sha256=%s\n' "$(sha256sum "${java_binary}" | awk '{print $1}')"
    printf 'picard_jar_sha256=%s\n' "$(sha256sum "${picard_jar}" | awk '{print $1}')"
    printf 'picard_invocation=%s -jar %s SetNmMdAndUqTags I=<coordinate-sorted-bam> O=<patched-bam> R=<reference-fasta>\n' "${java_binary}" "${picard_jar}"
} > "${evidence_directory}/manifest.properties"

printf '%s\n' "${evidence_directory}"
