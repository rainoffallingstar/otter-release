#!/usr/bin/env bash
#SBATCH --job-name=gate6-nm-read-parity
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=8
#SBATCH --mem=64G
#SBATCH --time=08:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-nm-read-parity-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
record_limit="${GATE6_NM_RECORD_LIMIT:-200000}"
dataset_label="${GATE6_NM_DATASET_LABEL:?GATE6_NM_DATASET_LABEL is required}"
input_bam="${GATE6_NM_INPUT_BAM:?GATE6_NM_INPUT_BAM is required}"
reference_fasta="${GATE6_NM_REFERENCE_FASTA:?GATE6_NM_REFERENCE_FASTA is required}"
mode="${GATE6_NM_MODE:?GATE6_NM_MODE must be rna or bsseq}"
evidence_directory="${evidence_root}/nm-read-parity-${dataset_label}-${SLURM_JOB_ID}"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
picard_prefix="/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17"
java_binary="${picard_prefix}/bin/java"
picard_jar="${picard_prefix}/share/picard-3.4.0-0/picard.jar"
nm_oracle="${runtime_directory}/gate6-nmoracle"
xenofilx_audit="${runtime_directory}/gate6-xenofilx-scoreaudit"
score_compare="${runtime_directory}/gate6-nm-score-compare.py"
read_compare="${runtime_directory}/gate6-nm-read-compare.py"
bisulfite_reference_converter="${runtime_directory}/gate6-bisulfite-reference.py"

if [[ "${mode}" != "rna" && "${mode}" != "bsseq" ]]; then
    printf 'unsupported mode: %s\n' "${mode}" >&2
    exit 2
fi
if [[ "${record_limit}" -le 0 ]]; then
    printf 'record limit must be positive: %s\n' "${record_limit}" >&2
    exit 2
fi
if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${input_bam}" "${reference_fasta}" "${samtools_binary}" "${java_binary}" "${picard_jar}" "${nm_oracle}" "${xenofilx_audit}" "${score_compare}" "${read_compare}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done
if [[ "${mode}" == "bsseq" && ! -r "${bisulfite_reference_converter}" ]]; then
    printf 'missing readable required path: %s\n' "${bisulfite_reference_converter}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
picard_conventional_bam="${evidence_directory}/input.picard-conventional-nm.bam"
picard_bisulfite_bam="${evidence_directory}/input.picard-bisulfite-nm.bam"
{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'dataset_label=%s\n' "${dataset_label}"
    printf 'mode=%s\n' "${mode}"
    printf 'record_limit=%s\n' "${record_limit}"
    printf 'input_bam=%s\n' "${input_bam}"
    printf 'reference_fasta=%s\n' "${reference_fasta}"
    printf 'input_bam_sha256=%s\n' "$(sha256sum "${input_bam}" | awk '{print $1}')"
    printf 'reference_fasta_sha256=%s\n' "$(sha256sum "${reference_fasta}" | awk '{print $1}')"
    printf 'picard_version=%s\n' "$("${java_binary}" -jar "${picard_jar}" MarkDuplicates --version 2>&1 | tr '\n' ' ')"
    printf 'nm_oracle_sha256=%s\n' "$(sha256sum "${nm_oracle}" | awk '{print $1}')"
    printf 'xenofilx_audit_sha256=%s\n' "$(sha256sum "${xenofilx_audit}" | awk '{print $1}')"
    printf 'score_compare_sha256=%s\n' "$(sha256sum "${score_compare}" | awk '{print $1}')"
    printf 'read_compare_sha256=%s\n' "$(sha256sum "${read_compare}" | awk '{print $1}')"
    printf 'xenofilx_bisulfite=%s\n' "$( [[ "${mode}" == "bsseq" ]] && printf true || printf false )"
} > "${evidence_directory}/manifest.properties"

"${samtools_binary}" quickcheck -v "${input_bam}" > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
    I="${input_bam}" O="${picard_conventional_bam}" R="${reference_fasta}" \
    > "${evidence_directory}/picard-conventional.log" 2>&1
"${samtools_binary}" quickcheck -v "${picard_conventional_bam}" > "${evidence_directory}/picard-conventional.quickcheck.txt" 2>&1

picard_comparison_arguments=(--picard-bam "conventional=${picard_conventional_bam}")
if [[ "${mode}" == "bsseq" ]]; then
    "${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
        I="${input_bam}" O="${picard_bisulfite_bam}" R="${reference_fasta}" IS_BISULFITE_SEQUENCE=true \
        > "${evidence_directory}/picard-bisulfite.log" 2>&1
    "${samtools_binary}" quickcheck -v "${picard_bisulfite_bam}" > "${evidence_directory}/picard-bisulfite.quickcheck.txt" 2>&1

    ct_bam="${evidence_directory}/input.xg-ct.bam"
    ga_bam="${evidence_directory}/input.xg-ga.bam"
    ct_picard_bam="${evidence_directory}/input.xg-ct.picard-nm.bam"
    ga_picard_bam="${evidence_directory}/input.xg-ga.picard-nm.bam"
    ct_reference="${evidence_directory}/reference.ct.fa"
    ga_reference="${evidence_directory}/reference.ga.fa"
    ct_oracle_summary="${evidence_directory}/oracle-xg-ct-summary.json"
    ct_oracle_records="${evidence_directory}/oracle-xg-ct-records.tsv"
    ga_oracle_summary="${evidence_directory}/oracle-xg-ga-summary.json"
    ga_oracle_records="${evidence_directory}/oracle-xg-ga-records.tsv"
    ct_xenofilx_records="${evidence_directory}/xenofilx-xg-ct-records.tsv"
    ga_xenofilx_records="${evidence_directory}/xenofilx-xg-ga-records.tsv"

    filter_by_xg_context() {
        local context="$1"
        local output_bam="$2"
        python3 - "${samtools_binary}" "${input_bam}" "${context}" "${output_bam}" <<'PY'
import subprocess
import sys

samtools_path, input_bam, context, output_bam = sys.argv[1:]
with subprocess.Popen(
    [samtools_path, "view", "-h", input_bam],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
) as reader, subprocess.Popen(
    [samtools_path, "view", "-bh", "-o", output_bam, "-"],
    stdin=subprocess.PIPE,
    stderr=subprocess.PIPE,
) as writer:
    assert reader.stdout is not None
    assert reader.stderr is not None
    assert writer.stdin is not None
    for raw_line in reader.stdout:
        if raw_line.startswith(b"@") or (b"XG:Z:" + context.encode("ascii")) in raw_line.splitlines()[0].split(b"\t"):
            writer.stdin.write(raw_line)
    writer.stdin.close()
    reader_stderr = reader.stderr.read().decode("utf-8", errors="replace")
    writer_stderr = writer.stderr.read().decode("utf-8", errors="replace")
    reader_code = reader.wait()
    writer_code = writer.wait()
    if reader_code != 0:
        raise SystemExit(f"samtools view input failed: {reader_stderr.strip()}")
    if writer_code != 0:
        raise SystemExit(f"samtools view output failed: {writer_stderr.strip()}")
PY
    }

    filter_by_xg_context CT "${ct_bam}"
    filter_by_xg_context GA "${ga_bam}"
    python3 "${bisulfite_reference_converter}" --input "${reference_fasta}" --output "${ct_reference}" --conversion ct
    python3 "${bisulfite_reference_converter}" --input "${reference_fasta}" --output "${ga_reference}" --conversion ga
    "${samtools_binary}" faidx "${ct_reference}"
    "${samtools_binary}" faidx "${ga_reference}"
    "${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
        I="${ct_bam}" O="${ct_picard_bam}" R="${ct_reference}" \
        > "${evidence_directory}/picard-xg-ct.log" 2>&1
    "${java_binary}" -jar "${picard_jar}" SetNmMdAndUqTags \
        I="${ga_bam}" O="${ga_picard_bam}" R="${ga_reference}" \
        > "${evidence_directory}/picard-xg-ga.log" 2>&1
    "${samtools_binary}" quickcheck -v "${ct_picard_bam}" "${ga_picard_bam}" \
        > "${evidence_directory}/picard-converted-reference.quickcheck.txt" 2>&1

    "${nm_oracle}" \
        --input "${ct_bam}" \
        --reference "${ct_reference}" \
        --limit "${record_limit}" \
        --report "${ct_oracle_summary}" \
        --records "${ct_oracle_records}"
    "${nm_oracle}" \
        --input "${ga_bam}" \
        --reference "${ga_reference}" \
        --limit "${record_limit}" \
        --report "${ga_oracle_summary}" \
        --records "${ga_oracle_records}"
    "${xenofilx_audit}" \
        --input "${ct_bam}" \
        --reference "${ct_reference}" \
        --limit "${record_limit}" \
        --report "${ct_xenofilx_records}"
    "${xenofilx_audit}" \
        --input "${ga_bam}" \
        --reference "${ga_reference}" \
        --limit "${record_limit}" \
        --report "${ga_xenofilx_records}"

    picard_comparison_arguments+=(
        --picard-bam "picard_bisulfite=${picard_bisulfite_bam}"
    )
fi

"${nm_oracle}" \
    --input "${input_bam}" \
    --reference "${reference_fasta}" \
    --limit "${record_limit}" \
    --report "${evidence_directory}/oracle-summary.json" \
    --records "${evidence_directory}/oracle-records.tsv"

xenofilx_arguments=(
    --input "${input_bam}"
    --reference "${reference_fasta}"
    --limit "${record_limit}"
    --report "${evidence_directory}/xenofilx-records.tsv"
)
if [[ "${mode}" == "bsseq" ]]; then
    xenofilx_arguments+=(--bisulfite)
fi
"${xenofilx_audit}" "${xenofilx_arguments[@]}"

set +e
python3 "${score_compare}" \
    --oracle "${evidence_directory}/oracle-records.tsv" \
    --xenofilx "${evidence_directory}/xenofilx-records.tsv" \
    --xenofilx-mode "$( [[ "${mode}" == "bsseq" ]] && printf bisulfite || printf conventional )" \
    --report "${evidence_directory}/xenofilx-oracle-comparison.json"
score_compare_exit_code=$?
python3 "${read_compare}" \
    --samtools "${samtools_binary}" \
    --original-bam "${input_bam}" \
    "${picard_comparison_arguments[@]}" \
    --oracle "${evidence_directory}/oracle-records.tsv" \
    --xenofilx "${evidence_directory}/xenofilx-records.tsv" \
    --records-read "${record_limit}" \
    --xenofilx-mode "$( [[ "${mode}" == "bsseq" ]] && printf bisulfite || printf conventional )" \
    --report "${evidence_directory}/picard-xenofilx-comparison.json" \
    --differences "${evidence_directory}/picard-xenofilx-differences.jsonl"
read_compare_exit_code=$?
forward_strand_compare_exit_code=0
reverse_strand_compare_exit_code=0
if [[ "${mode}" == "bsseq" ]]; then
    python3 "${read_compare}" \
        --samtools "${samtools_binary}" \
        --original-bam "${ct_bam}" \
        --picard-bam "xg_ct=${ct_picard_bam}" \
        --oracle "${ct_oracle_records}" \
        --xenofilx "${ct_xenofilx_records}" \
        --records-read "${record_limit}" \
        --xenofilx-mode conventional \
        --report "${evidence_directory}/xg-ct-picard-xenofilx-comparison.json" \
        --differences "${evidence_directory}/xg-ct-picard-xenofilx-differences.jsonl"
    forward_strand_compare_exit_code=$?
    python3 "${read_compare}" \
        --samtools "${samtools_binary}" \
        --original-bam "${ga_bam}" \
        --picard-bam "xg_ga=${ga_picard_bam}" \
        --oracle "${ga_oracle_records}" \
        --xenofilx "${ga_xenofilx_records}" \
        --records-read "${record_limit}" \
        --xenofilx-mode conventional \
        --report "${evidence_directory}/xg-ga-picard-xenofilx-comparison.json" \
        --differences "${evidence_directory}/xg-ga-picard-xenofilx-differences.jsonl"
    reverse_strand_compare_exit_code=$?
fi
set -e
printf 'xenofilx_oracle_compare_exit_code=%s\n' "${score_compare_exit_code}" >> "${evidence_directory}/manifest.properties"
printf 'read_compare_exit_code=%s\n' "${read_compare_exit_code}" >> "${evidence_directory}/manifest.properties"
printf 'forward_strand_compare_exit_code=%s\n' "${forward_strand_compare_exit_code}" >> "${evidence_directory}/manifest.properties"
printf 'reverse_strand_compare_exit_code=%s\n' "${reverse_strand_compare_exit_code}" >> "${evidence_directory}/manifest.properties"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
