#!/usr/bin/env bash
#SBATCH --job-name=gate6-bam-nm-baseline
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-bam-nm-baseline-%j.log
set -euo pipefail

export PATH="/public3/home/scg9946/otter-gate6/binaries/current/bin:/public3/home/scg9946/.cargo/bin:/opt/slurm/slurm/bin:${PATH}"

project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/baseline-${SLURM_JOB_ID}"
run_snapshot="${project_root}/run.yaml"
graft_bam="${project_root}/work/bsmap/SRR36187610_hg38.bam"
host_bam="${project_root}/work/bsmap/SRR36187610_mm10.bam"
filtered_bam="${project_root}/work/bsmap/Filtered_bams/SRR36187610_fixed_hg38_Filtered.bam"

if [[ ! -r "${run_snapshot}" ]]; then
    printf 'missing immutable run snapshot: %s\n' "${run_snapshot}" >&2
    exit 1
fi

for input_path in "${graft_bam}" "${host_bam}" "${filtered_bam}"; do
    if [[ ! -r "${input_path}" ]]; then
        printf 'missing readable BAM: %s\n' "${input_path}" >&2
        exit 1
    fi
    if [[ ! -r "${input_path}.bai" ]]; then
        printf 'missing readable BAM index: %s\n' "${input_path}.bai" >&2
        exit 1
    fi
done

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite existing evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"

write_tool_version() {
    local program_name="$1"
    local output_path="${evidence_directory}/tool-${program_name}.txt"
    {
        printf 'path=' 
        command -v "${program_name}" || true
        printf '\nversion:\n'
        "${program_name}" --version 2>&1 || "${program_name}" version 2>&1 || true
    } > "${output_path}"
}

write_tool_version samtools
write_tool_version xenofilx
write_tool_version pairbam

{
    printf 'collection_started_utc=%s\n' "$(date -u +%FT%TZ)"
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'project_root=%s\n' "${project_root}"
    printf 'run_snapshot=%s\n' "${run_snapshot}"
    printf 'graft_bam=%s\n' "${graft_bam}"
    printf 'host_bam=%s\n' "${host_bam}"
    printf 'filtered_bam=%s\n' "${filtered_bam}"
    printf 'xenofilx_contract=--recalculate-nm --bisulfite --mm-threshold 6 --unmapped-penalty 8\n'
} > "${evidence_directory}/manifest.properties"

sha256sum "${run_snapshot}" "${graft_bam}" "${graft_bam}.bai" "${host_bam}" "${host_bam}.bai" "${filtered_bam}" "${filtered_bam}.bai" > "${evidence_directory}/input-sha256.txt"

collect_bam_baseline() {
    local label="$1"
    local bam_path="$2"
    local prefix="${evidence_directory}/${label}"

    samtools quickcheck -v "${bam_path}" > "${prefix}.quickcheck.txt" 2>&1
    samtools view -H "${bam_path}" > "${prefix}.header.sam"
    samtools idxstats "${bam_path}" > "${prefix}.idxstats.tsv"
    samtools flagstat "${bam_path}" > "${prefix}.flagstat.txt"
    samtools stats "${bam_path}" > "${prefix}.stats.txt"

    {
        printf 'total_records\t'
        samtools view -c "${bam_path}"
        printf 'mapped_records\t'
        samtools view -c -F 4 "${bam_path}"
        printf 'unmapped_records\t'
        samtools view -c -f 4 "${bam_path}"
        printf 'primary_records\t'
        samtools view -c -F 2308 "${bam_path}"
        printf 'secondary_records\t'
        samtools view -c -f 256 "${bam_path}"
        printf 'supplementary_records\t'
        samtools view -c -f 2048 "${bam_path}"
        printf 'read1_records\t'
        samtools view -c -f 64 "${bam_path}"
        printf 'read2_records\t'
        samtools view -c -f 128 "${bam_path}"
    } > "${prefix}.record-counts.tsv"

    samtools view -h "${bam_path}" | sha256sum > "${prefix}.ordered-sam-sha256.txt"

    samtools view "${bam_path}" | awk -v nm_output_path="${prefix}.nm-summary.tsv" -v tag_output_path="${prefix}.tag-types.tsv" '
        BEGIN { OFS="\t" }
        {
            record_count++
            has_nm=0
            for (field_index = 12; field_index <= NF; field_index++) {
                split($field_index, components, ":")
                tag_type[components[1] ":" components[2]]++
                if (components[1] == "NM" && components[2] == "i") {
                    has_nm=1
                    nm_histogram[components[3]]++
                }
            }
            if (has_nm) { nm_present++ } else { nm_missing++ }
        }
        END {
            print "metric", "value" > nm_output_path
            print "records", record_count >> nm_output_path
            print "nm_present", nm_present + 0 >> nm_output_path
            print "nm_missing", nm_missing + 0 >> nm_output_path
            for (nm_value in nm_histogram) {
                print "nm_value:" nm_value, nm_histogram[nm_value] >> nm_output_path
            }
            for (tag in tag_type) {
                print tag, tag_type[tag] > tag_output_path
            }
        }
    '

    sort -t $'\t' -k1,1n "${prefix}.nm-summary.tsv" -o "${prefix}.nm-summary.tsv"
    sort -t $'\t' -k1,1 "${prefix}.tag-types.tsv" -o "${prefix}.tag-types.tsv"
}

collect_bam_baseline graft-original "${graft_bam}"
collect_bam_baseline host-original "${host_bam}"
collect_bam_baseline graft-filtered-modern "${filtered_bam}"

printf 'collection_completed_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
