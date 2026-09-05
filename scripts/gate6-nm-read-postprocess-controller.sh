#!/usr/bin/env bash
#SBATCH --job-name=gate6-nm-read-postprocess
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=64G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-nm-read-postprocess-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
dataset_label="${GATE6_NM_DATASET_LABEL:?GATE6_NM_DATASET_LABEL is required}"
source_job_id="${GATE6_NM_SOURCE_JOB_ID:?GATE6_NM_SOURCE_JOB_ID is required}"
evidence_directory="${evidence_root}/nm-read-parity-${dataset_label}-${source_job_id}"
input_bam="${GATE6_NM_INPUT_BAM:?GATE6_NM_INPUT_BAM is required}"
record_limit="${GATE6_NM_RECORD_LIMIT:-200000}"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
read_compare="${runtime_directory}/gate6-nm-read-compare.py"
score_compare="${runtime_directory}/gate6-nm-score-compare.py"
nm_oracle="${runtime_directory}/gate6-nmoracle"
xenofilx_audit="${runtime_directory}/gate6-xenofilx-scoreaudit"

if [[ ! -d "${evidence_directory}" ]]; then
    printf 'missing source evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${input_bam}" "${samtools_binary}" "${read_compare}" "${score_compare}" "${evidence_directory}/oracle-records.tsv" "${evidence_directory}/xenofilx-records.tsv" "${evidence_directory}/input.picard-conventional-nm.bam"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

picard_arguments=(--picard-bam "conventional=${evidence_directory}/input.picard-conventional-nm.bam")
if [[ -r "${evidence_directory}/input.picard-bisulfite-nm.bam" ]]; then
    picard_arguments+=(--picard-bam "picard_bisulfite=${evidence_directory}/input.picard-bisulfite-nm.bam")
fi
if [[ -r "${evidence_directory}/input.forward.picard-ct-nm.bam" ]]; then
    picard_arguments+=(--picard-bam "forward_ct=${evidence_directory}/input.forward.picard-ct-nm.bam")
fi
if [[ -r "${evidence_directory}/input.reverse.picard-ga-nm.bam" ]]; then
    picard_arguments+=(--picard-bam "reverse_ga=${evidence_directory}/input.reverse.picard-ga-nm.bam")
fi
mode="${GATE6_NM_MODE:?GATE6_NM_MODE is required}"
if [[ "${mode}" == "bsseq" ]]; then
    xenofilx_mode=bisulfite
else
    xenofilx_mode=conventional
fi
python3 "${score_compare}" \
    --oracle "${evidence_directory}/oracle-records.tsv" \
    --xenofilx "${evidence_directory}/xenofilx-records.tsv" \
    --xenofilx-mode "${xenofilx_mode}" \
    --report "${evidence_directory}/xenofilx-oracle-comparison.json"
python3 "${read_compare}" \
    --samtools "${samtools_binary}" \
    --original-bam "${input_bam}" \
    "${picard_arguments[@]}" \
    --oracle "${evidence_directory}/oracle-records.tsv" \
    --xenofilx "${evidence_directory}/xenofilx-records.tsv" \
    --records-read "${record_limit}" \
    --xenofilx-mode "${xenofilx_mode}" \
    --report "${evidence_directory}/picard-xenofilx-comparison.json" \
    --differences "${evidence_directory}/picard-xenofilx-differences.jsonl"
if [[ "${mode}" == "bsseq" ]]; then
    for required_path in \
        "${nm_oracle}" "${xenofilx_audit}" \
        "${evidence_directory}/input.forward.bam" \
        "${evidence_directory}/input.reverse.bam" \
        "${evidence_directory}/input.forward.picard-ct-nm.bam" \
        "${evidence_directory}/input.reverse.picard-ga-nm.bam" \
        "${evidence_directory}/reference.ct.fa" \
        "${evidence_directory}/reference.ga.fa"; do
        if [[ ! -r "${required_path}" ]]; then
            printf 'missing strand-specific required path: %s\n' "${required_path}" >&2
            exit 1
        fi
    done
    forward_bam="${evidence_directory}/input.forward.bam"
    reverse_bam="${evidence_directory}/input.reverse.bam"
    forward_reference="${evidence_directory}/reference.ct.fa"
    reverse_reference="${evidence_directory}/reference.ga.fa"
    forward_oracle_records="${evidence_directory}/oracle-forward-ct-records.tsv"
    reverse_oracle_records="${evidence_directory}/oracle-reverse-ga-records.tsv"
    forward_xenofilx_records="${evidence_directory}/xenofilx-forward-ct-records.tsv"
    reverse_xenofilx_records="${evidence_directory}/xenofilx-reverse-ga-records.tsv"
    if [[ ! -r "${forward_oracle_records}" ]]; then
        "${nm_oracle}" --input "${forward_bam}" --reference "${forward_reference}" \
            --limit "${record_limit}" \
            --report "${evidence_directory}/oracle-forward-ct-summary.json" \
            --records "${forward_oracle_records}"
    fi
    if [[ ! -r "${reverse_oracle_records}" ]]; then
        "${nm_oracle}" --input "${reverse_bam}" --reference "${reverse_reference}" \
            --limit "${record_limit}" \
            --report "${evidence_directory}/oracle-reverse-ga-summary.json" \
            --records "${reverse_oracle_records}"
    fi
    if [[ ! -r "${forward_xenofilx_records}" ]]; then
        "${xenofilx_audit}" --input "${forward_bam}" --reference "${forward_reference}" \
            --limit "${record_limit}" --report "${forward_xenofilx_records}"
    fi
    if [[ ! -r "${reverse_xenofilx_records}" ]]; then
        "${xenofilx_audit}" --input "${reverse_bam}" --reference "${reverse_reference}" \
            --limit "${record_limit}" --report "${reverse_xenofilx_records}"
    fi
    python3 "${read_compare}" \
        --samtools "${samtools_binary}" \
        --original-bam "${forward_bam}" \
        --picard-bam "forward_ct=${evidence_directory}/input.forward.picard-ct-nm.bam" \
        --oracle "${forward_oracle_records}" \
        --xenofilx "${forward_xenofilx_records}" \
        --records-read "${record_limit}" \
        --xenofilx-mode conventional \
        --report "${evidence_directory}/forward-ct-picard-xenofilx-comparison.json" \
        --differences "${evidence_directory}/forward-ct-picard-xenofilx-differences.jsonl"
    python3 "${read_compare}" \
        --samtools "${samtools_binary}" \
        --original-bam "${reverse_bam}" \
        --picard-bam "reverse_ga=${evidence_directory}/input.reverse.picard-ga-nm.bam" \
        --oracle "${reverse_oracle_records}" \
        --xenofilx "${reverse_xenofilx_records}" \
        --records-read "${record_limit}" \
        --xenofilx-mode conventional \
        --report "${evidence_directory}/reverse-ga-picard-xenofilx-comparison.json" \
        --differences "${evidence_directory}/reverse-ga-picard-xenofilx-differences.jsonl"
fi
printf 'postprocessed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
