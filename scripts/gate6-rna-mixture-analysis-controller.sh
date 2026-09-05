#!/usr/bin/env bash
#SBATCH --job-name=gate6-rna-mixture-analysis
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=40
#SBATCH --mem=160G
#SBATCH --time=12:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-rna-mixture-analysis-%j.log
set -euo pipefail

runtime_directory="${GATE6_RUNTIME_DIRECTORY:-/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime}"
mixture_directory="${GATE6_MIXTURE_DIRECTORY:?GATE6_MIXTURE_DIRECTORY is required}"
evidence_root="${GATE6_EVIDENCE_ROOT:-/public3/home/scg9946/otter-gate6/evidence/gate6-human-mouse-rna-mixtures-20260903}"
evidence_directory="${evidence_root}/analysis-${SLURM_JOB_ID}"
mixture_evaluator="${GATE6_MIXTURE_EVALUATOR:-${runtime_directory}/gate6-mixture-evaluate.py}"
samtools_binary="${GATE6_SAMTOOLS_BINARY:-/public3/home/scg9946/TTest/breg/soft/bin/samtools}"
star_binary="${GATE6_STAR_BINARY:-/public3/home/scg9946/TTest/soft/MyMiniconda/bin/STAR}"
java_binary="${GATE6_JAVA_BINARY:-/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17/bin/java}"
picard_jar="${GATE6_PICARD_JAR:-/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17/share/picard-3.4.0-0/picard.jar}"
rscript_binary="${GATE6_RSCRIPT_BINARY:-/public3/home/scg9946/TTest/soft/MyMiniconda/bin/Rscript}"
xenofilx_binary="${GATE6_XENOFILX_BINARY:-${runtime_directory}/gate6-xenofilx-fixed-strand-aware-static}"
legacy_script="${GATE6_LEGACY_SCRIPT:-${runtime_directory}/gate6-xenofilter-legacy.R}"
graft_source="${GATE6_GRAFT_SOURCE:-mouse}"
mm_threshold="${GATE6_MM_THRESHOLD:-6}"
unmapped_penalty="${GATE6_UNMAPPED_PENALTY:-8}"
mixture_r1="${mixture_directory}/mixture_R1.fastq.gz"
mixture_r2="${mixture_directory}/mixture_R2.fastq.gz"
truth_file="${mixture_directory}/truth.tsv"

case "${graft_source}" in
    human)
        graft_label=human
        host_label=mouse
        graft_reference="${GATE6_GRAFT_REFERENCE:-/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa}"
        graft_index="${GATE6_GRAFT_INDEX:-/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/indexes/star}"
        host_reference="${GATE6_HOST_REFERENCE:-/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa}"
        host_index="${GATE6_HOST_INDEX:-/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/indexes/star}"
        ;;
    mouse)
        graft_label=mouse
        host_label=human
        graft_reference="${GATE6_GRAFT_REFERENCE:-/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa}"
        graft_index="${GATE6_GRAFT_INDEX:-/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/indexes/star}"
        host_reference="${GATE6_HOST_REFERENCE:-/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa}"
        host_index="${GATE6_HOST_INDEX:-/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/indexes/star}"
        ;;
    *)
        printf 'GATE6_GRAFT_SOURCE must be human or mouse: %s\n' "${graft_source}" >&2
        exit 2
        ;;
esac

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite analysis directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${mixture_r1}" "${mixture_r2}" "${truth_file}" "${mixture_evaluator}" "${samtools_binary}" "${star_binary}" "${java_binary}" "${picard_jar}" "${rscript_binary}" "${xenofilx_binary}" "${legacy_script}" "${graft_reference}" "${host_reference}" "${graft_index}" "${host_index}"; do
    if [[ ! -e "${required_path}" || ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}" "${evidence_directory}"
mkdir "${evidence_directory}/star-graft" "${evidence_directory}/star-host" "${evidence_directory}/legacy-xenofilter"
{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'assay=RNA-seq\n'
    printf 'mixture_directory=%s\n' "${mixture_directory}"
    printf 'mixture_manifest_sha256=%s\n' "$(sha256sum "${mixture_directory}/manifest.json" | awk '{print $1}')"
    printf 'truth_sha256=%s\n' "$(sha256sum "${truth_file}" | awk '{print $1}')"
    printf 'graft_source=%s\n' "${graft_source}"
    printf 'host_source=%s\n' "${host_label}"
    printf 'graft_reference=%s\n' "${graft_reference}"
    printf 'host_reference=%s\n' "${host_reference}"
    printf 'graft_star_index=%s\n' "${graft_index}"
    printf 'host_star_index=%s\n' "${host_index}"
    printf 'star_version=%s\n' "$(${star_binary} --version 2>&1 | head -n 1)"
    printf 'samtools_version=%s\n' "$(${samtools_binary} --version | head -n 1)"
    printf 'xenofilx_sha256=%s\n' "$(sha256sum "${xenofilx_binary}" | awk '{print $1}')"
    printf 'legacy_wrapper_sha256=%s\n' "$(sha256sum "${legacy_script}" | awk '{print $1}')"
    printf 'mm_threshold=%s\n' "${mm_threshold}"
    printf 'unmapped_penalty=%s\n' "${unmapped_penalty}"
    printf 'modern_nm_semantics=conventional_reference_nm\n'
    printf 'legacy_nm_semantics=picard_conventional_reference_nm\n'
    printf 'input_preprocessing=raw_mixed_fastq_no_additional_trimming\n'
} > "${evidence_directory}/manifest.properties"

gzip -t "${mixture_r1}" "${mixture_r2}"

run_star_alignment() {
    local reference_index="$1"
    local output_directory="$2"
    local log_path="$3"
    local aligned_bam="$4"
    mkdir -p "${output_directory}"
    "${star_binary}" \
        --genomeDir "${reference_index}" \
        --readFilesIn "${mixture_r1}" "${mixture_r2}" \
        --readFilesCommand zcat \
        --runThreadN "${SLURM_CPUS_PER_TASK}" \
        --outFileNamePrefix "${output_directory}/" \
        --outSAMtype BAM SortedByCoordinate \
        --outSAMattributes All \
        --limitBAMsortRAM 120000000000 \
        > "${log_path}" 2>&1
    mv "${output_directory}/Aligned.sortedByCoord.out.bam" "${aligned_bam}"
    "${samtools_binary}" index -@ "${SLURM_CPUS_PER_TASK}" -b "${aligned_bam}"
    "${samtools_binary}" quickcheck -v "${aligned_bam}" > "${aligned_bam}.quickcheck.txt" 2>&1
}

graft_bam="${evidence_directory}/mixture_${graft_label}.bam"
host_bam="${evidence_directory}/mixture_${host_label}.bam"
run_star_alignment "${graft_index}" "${evidence_directory}/star-graft" "${evidence_directory}/star-graft.log" "${graft_bam}"
run_star_alignment "${host_index}" "${evidence_directory}/star-host" "${evidence_directory}/star-host.log" "${host_bam}"

modern_output_directory="${evidence_directory}/modern-xenofilx"
mkdir -p "${modern_output_directory}"
modern_output_name="mixture_${graft_label}_Filtered.bam"
"${xenofilx_binary}" run \
    --graft "${graft_bam}" \
    --host "${host_bam}" \
    --output "${modern_output_directory}" \
    --output-names "${modern_output_name}" \
    --graft-ref "${graft_reference}" \
    --host-ref "${host_reference}" \
    --mm-threshold "${mm_threshold}" \
    --unmapped-penalty "${unmapped_penalty}" \
    --threads "${SLURM_CPUS_PER_TASK}" \
    --sort-memory 8G \
    --recalculate-nm \
    > "${evidence_directory}/modern-xenofilx.stdout.txt" \
    2> "${evidence_directory}/modern-xenofilx.stderr.txt"
modern_filtered_bam="${modern_output_directory}/${modern_output_name}"
"${samtools_binary}" quickcheck -v "${modern_filtered_bam}" > "${evidence_directory}/modern-filtered.quickcheck.txt" 2>&1

legacy_graft_bam="${evidence_directory}/legacy-xenofilter/mixture_${graft_label}.picard-nm.bam"
legacy_host_bam="${evidence_directory}/legacy-xenofilter/mixture_${host_label}.picard-nm.bam"
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags I="${graft_bam}" O="${legacy_graft_bam}" R="${graft_reference}" > "${evidence_directory}/picard-graft.log" 2>&1
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags I="${host_bam}" O="${legacy_host_bam}" R="${host_reference}" > "${evidence_directory}/picard-host.log" 2>&1
legacy_output_directory="${evidence_directory}/legacy-xenofilter/destination"
mkdir -p "${legacy_output_directory}"
"${rscript_binary}" "${legacy_script}" \
    "${legacy_graft_bam}" \
    "${legacy_host_bam}" \
    "${legacy_output_directory}" \
    "${unmapped_penalty}" \
    "${mm_threshold}" \
    "${unmapped_penalty}" \
    NM \
    "${evidence_directory}/xenofilter-run.json" \
    > "${evidence_directory}/legacy-xenofilter.stdout.txt" \
    2> "${evidence_directory}/legacy-xenofilter.stderr.txt"
mapfile -t legacy_filtered_candidates < <(printf '%s\n' "${legacy_output_directory}/Filtered_bams/"*_Filtered.bam)
if (( ${#legacy_filtered_candidates[@]} != 1 )); then
    printf 'expected exactly one XenofilteR filtered BAM in %s, found %d\n' "${legacy_output_directory}/Filtered_bams" "${#legacy_filtered_candidates[@]}" >&2
    exit 1
fi
legacy_filtered_bam="${legacy_filtered_candidates[0]}"
"${samtools_binary}" quickcheck -v "${legacy_filtered_bam}" > "${evidence_directory}/legacy-filtered.quickcheck.txt" 2>&1

python3 "${mixture_evaluator}" \
    --truth "${truth_file}" \
    --samtools "${samtools_binary}" \
    --tool-output "modern_xenofilx=${modern_filtered_bam}" \
    --tool-output "legacy_xenofilter=${legacy_filtered_bam}" \
    --require-known-selection \
    --graft-source "${graft_source}" \
    --report "${evidence_directory}/source-performance.json"

{
    printf 'graft_alignment_records\t'; "${samtools_binary}" view -c "${graft_bam}"
    printf 'host_alignment_records\t'; "${samtools_binary}" view -c "${host_bam}"
    printf 'modern_filtered_records\t'; "${samtools_binary}" view -c "${modern_filtered_bam}"
    printf 'legacy_filtered_records\t'; "${samtools_binary}" view -c "${legacy_filtered_bam}"
} > "${evidence_directory}/summary.tsv"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
