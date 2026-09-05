#!/usr/bin/env bash
#SBATCH --job-name=gate6-legacy-picard-strand-aware
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=8
#SBATCH --mem=160G
#SBATCH --time=08:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-legacy-picard-strand-aware-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/legacy-picard-strand-aware-xenofilter-${SLURM_JOB_ID}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
graft_input="${project_root}/work/bsmap/SRR36187610_hg38.bam"
host_input="${project_root}/work/bsmap/SRR36187610_mm10.bam"
graft_reference="/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"
host_reference="/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
picard_prefix="/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17"
java_binary="${picard_prefix}/bin/java"
picard_jar="${picard_prefix}/share/picard-3.4.0-0/picard.jar"
legacy_script="${runtime_directory}/gate6-xenofilter-legacy.R"
reference_converter="${runtime_directory}/gate6-bisulfite-reference.py"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${graft_input}" "${host_input}" "${graft_reference}" "${host_reference}" "${samtools_binary}" "${java_binary}" "${picard_jar}" "${legacy_script}" "${reference_converter}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
mkdir "${evidence_directory}/xenofilter-destination"

graft_forward="${evidence_directory}/graft.forward.bam"
graft_reverse="${evidence_directory}/graft.reverse.bam"
host_forward="${evidence_directory}/host.forward.bam"
host_reverse="${evidence_directory}/host.reverse.bam"
graft_forward_patched="${evidence_directory}/graft.forward.picard-ct-nm.bam"
graft_reverse_patched="${evidence_directory}/graft.reverse.picard-ga-nm.bam"
host_forward_patched="${evidence_directory}/host.forward.picard-ct-nm.bam"
host_reverse_patched="${evidence_directory}/host.reverse.picard-ga-nm.bam"
graft_merged="${evidence_directory}/SRR36187610_hg38.picard-strand-aware-nm.bam"
host_merged="${evidence_directory}/SRR36187610_mm10.picard-strand-aware-nm.bam"
ct_reference="${evidence_directory}/reference.hg38.ct.fa"
ga_reference="${evidence_directory}/reference.hg38.ga.fa"
host_ct_reference="${evidence_directory}/reference.mm10.ct.fa"
host_ga_reference="${evidence_directory}/reference.mm10.ga.fa"

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'graft_input_sha256=%s\n' "$(sha256sum "${graft_input}" | awk '{print $1}')"
    printf 'host_input_sha256=%s\n' "$(sha256sum "${host_input}" | awk '{print $1}')"
    printf 'graft_reference_sha256=%s\n' "$(sha256sum "${graft_reference}" | awk '{print $1}')"
    printf 'host_reference_sha256=%s\n' "$(sha256sum "${host_reference}" | awk '{print $1}')"
    printf 'java_sha256=%s\n' "$(sha256sum "${java_binary}" | awk '{print $1}')"
    printf 'picard_jar_sha256=%s\n' "$(sha256sum "${picard_jar}" | awk '{print $1}')"
    printf 'xenofilter_wrapper_sha256=%s\n' "$(sha256sum "${legacy_script}" | awk '{print $1}')"
    printf 'reference_converter_sha256=%s\n' "$(sha256sum "${reference_converter}" | awk '{print $1}')"
    printf 'legacy_nm_semantics=strand_aware_converted_reference_picard\n'
    printf 'mm_threshold=6\n'
    printf 'unmapped_penalty=8\n'
    printf 'nm_tag=NM\n'
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v "${graft_input}" "${host_input}" > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${samtools_binary}" view -bh -F 16 -o "${graft_forward}" "${graft_input}"
"${samtools_binary}" view -bh -f 16 -o "${graft_reverse}" "${graft_input}"
"${samtools_binary}" view -bh -F 16 -o "${host_forward}" "${host_input}"
"${samtools_binary}" view -bh -f 16 -o "${host_reverse}" "${host_input}"

python3 "${reference_converter}" --input "${graft_reference}" --output "${ct_reference}" --conversion ct
python3 "${reference_converter}" --input "${graft_reference}" --output "${ga_reference}" --conversion ga
python3 "${reference_converter}" --input "${host_reference}" --output "${host_ct_reference}" --conversion ct
python3 "${reference_converter}" --input "${host_reference}" --output "${host_ga_reference}" --conversion ga
"${samtools_binary}" faidx "${ct_reference}"
"${samtools_binary}" faidx "${ga_reference}"
"${samtools_binary}" faidx "${host_ct_reference}"
"${samtools_binary}" faidx "${host_ga_reference}"

"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
    I="${graft_forward}" O="${graft_forward_patched}" R="${ct_reference}" \
    > "${evidence_directory}/picard-graft-forward.log" 2>&1
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
    I="${graft_reverse}" O="${graft_reverse_patched}" R="${ga_reference}" \
    > "${evidence_directory}/picard-graft-reverse.log" 2>&1
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
    I="${host_forward}" O="${host_forward_patched}" R="${host_ct_reference}" \
    > "${evidence_directory}/picard-host-forward.log" 2>&1
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
    I="${host_reverse}" O="${host_reverse_patched}" R="${host_ga_reference}" \
    > "${evidence_directory}/picard-host-reverse.log" 2>&1

"${samtools_binary}" merge -f "${graft_merged}" "${graft_forward_patched}" "${graft_reverse_patched}"
"${samtools_binary}" merge -f "${host_merged}" "${host_forward_patched}" "${host_reverse_patched}"
"${samtools_binary}" quickcheck -v "${graft_forward_patched}" "${graft_reverse_patched}" "${host_forward_patched}" "${host_reverse_patched}" "${graft_merged}" "${host_merged}" > "${evidence_directory}/patched-output.quickcheck.txt" 2>&1

Rscript "${legacy_script}" \
    "${graft_merged}" \
    "${host_merged}" \
    "${evidence_directory}/xenofilter-destination" \
    "${SLURM_CPUS_PER_TASK}" \
    6 \
    8 \
    NM \
    "${evidence_directory}/xenofilter-run.json" \
    > "${evidence_directory}/xenofilteR.stdout.txt" \
    2> "${evidence_directory}/xenofilteR.stderr.txt"

legacy_filtered_bam="${evidence_directory}/xenofilter-destination/Filtered_bams/SRR36187610_hg38.picard-strand-aware-nm_Filtered.bam"
"${samtools_binary}" quickcheck -v "${legacy_filtered_bam}" > "${evidence_directory}/legacy-filtered.quickcheck.txt" 2>&1

{
    printf 'graft_input_records\t'
    "${samtools_binary}" view -c "${graft_input}"
    printf 'graft_merged_records\t'
    "${samtools_binary}" view -c "${graft_merged}"
    printf 'host_input_records\t'
    "${samtools_binary}" view -c "${host_input}"
    printf 'host_merged_records\t'
    "${samtools_binary}" view -c "${host_merged}"
    printf 'legacy_filtered_records\t'
    "${samtools_binary}" view -c "${legacy_filtered_bam}"
    printf 'legacy_filtered_bam_sha256\t'
    sha256sum "${legacy_filtered_bam}" | awk '{print $1}'
} > "${evidence_directory}/summary.tsv"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
