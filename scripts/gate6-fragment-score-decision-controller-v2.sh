#!/usr/bin/env bash
#SBATCH --job-name=gate6-fragment-score-decision
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-fragment-score-decision-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/fragment-score-decision-${SLURM_JOB_ID}"
stratification_job_id="${GATE6_STRATIFICATION_JOB_ID:?GATE6_STRATIFICATION_JOB_ID is required}"
stratification_directory="${evidence_root}/fragment-disagreement-stratification-${stratification_job_id}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
scoreaudit_binary="${GATE6_SCOREAUDIT_BINARY:-${runtime_directory}/gate6-xenofilx-scoreaudit}"
comparison_script="${runtime_directory}/gate6-fragment-score-decision-compare.py"
samples="${stratification_directory}/sampled-fragments.tsv"
modern_graft_input="${project_root}/work/bsmap/SRR36187610_hg38.bam"
modern_host_input="${project_root}/work/bsmap/SRR36187610_mm10.bam"
graft_reference="/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"
host_reference="/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa"
legacy_evidence_prefix="${GATE6_LEGACY_EVIDENCE_PREFIX:-legacy-picard-strand-aware-xenofilter}"
legacy_job_id="${GATE6_LEGACY_JOB_ID:-41967065}"
legacy_directory="${evidence_root}/${legacy_evidence_prefix}-${legacy_job_id}"
legacy_graft_input="${legacy_directory}/SRR36187610_hg38.picard-strand-aware-nm.bam"
legacy_host_input="${legacy_directory}/SRR36187610_mm10.picard-strand-aware-nm.bam"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in \
    "${samtools_binary}" "${scoreaudit_binary}" "${comparison_script}" "${samples}" \
    "${modern_graft_input}" "${modern_host_input}" "${graft_reference}" "${host_reference}" \
    "${legacy_graft_input}" "${legacy_host_input}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
names_file="${evidence_directory}/sampled-qnames.txt"
python3 - "${samples}" "${names_file}" <<'PY'
import csv
import sys

samples_path, names_path = sys.argv[1:]
with open(samples_path, encoding="utf-8", newline="") as samples_file:
    reader = csv.DictReader(samples_file, delimiter="\t")
    names = sorted({row["fragment_id"] for row in reader})
if not names:
    raise SystemExit("sampled fragments file contains no QNAMEs")
with open(names_path, "w", encoding="utf-8", newline="") as names_file:
    names_file.write("\n".join(names) + "\n")
PY

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'stratification_job_id=%s\n' "${stratification_job_id}"
    printf 'sample_count=%s\n' "$(wc -l < "${names_file}")"
    printf 'legacy_extraction=streaming_qname_filter_without_samtools_view_N\n'
    printf 'scoreaudit_sha256=%s\n' "$(sha256sum "${scoreaudit_binary}" | awk '{print $1}')"
    printf 'comparison_script_sha256=%s\n' "$(sha256sum "${comparison_script}" | awk '{print $1}')"
    printf 'sample_qnames_sha256=%s\n' "$(sha256sum "${names_file}" | awk '{print $1}')"
    printf 'modern_graft_sha256=%s\n' "$(sha256sum "${modern_graft_input}" | awk '{print $1}')"
    printf 'modern_host_sha256=%s\n' "$(sha256sum "${modern_host_input}" | awk '{print $1}')"
    printf 'legacy_graft_sha256=%s\n' "$(sha256sum "${legacy_graft_input}" | awk '{print $1}')"
    printf 'legacy_host_sha256=%s\n' "$(sha256sum "${legacy_host_input}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

extract_legacy_sample() {
    local input_bam="$1"
    local output_sam="$2"
    python3 - "${samtools_binary}" "${input_bam}" "${names_file}" "${output_sam}" <<'PY'
import subprocess
import sys

samtools_path, bam_path, names_path, output_path = sys.argv[1:]
with open(names_path, encoding="utf-8") as names_file:
    selected_names = {line.rstrip("\n") for line in names_file}
if not selected_names or "" in selected_names:
    raise SystemExit("invalid QNAME allowlist")
with subprocess.Popen(
    [samtools_path, "view", bam_path],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True,
    encoding="utf-8",
) as process, open(output_path, "w", encoding="utf-8", newline="") as output_file:
    assert process.stdout is not None
    for line in process.stdout:
        qname, separator, _ = line.partition("\t")
        if separator and qname in selected_names:
            output_file.write(line)
    stderr = process.stderr.read() if process.stderr is not None else ""
    if process.wait() != 0:
        raise SystemExit(f"samtools view failed for {bam_path}: {stderr.strip()}")
PY
}

"${samtools_binary}" quickcheck -v \
    "${modern_graft_input}" "${modern_host_input}" "${legacy_graft_input}" "${legacy_host_input}" \
    > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${scoreaudit_binary}" \
    --input "${modern_graft_input}" \
    --reference "${graft_reference}" \
    --names "${names_file}" \
    --bisulfite \
    --report "${evidence_directory}/modern-graft-scores.tsv"
"${scoreaudit_binary}" \
    --input "${modern_host_input}" \
    --reference "${host_reference}" \
    --names "${names_file}" \
    --bisulfite \
    --report "${evidence_directory}/modern-host-scores.tsv"
extract_legacy_sample "${legacy_graft_input}" "${evidence_directory}/legacy-graft.sam"
extract_legacy_sample "${legacy_host_input}" "${evidence_directory}/legacy-host.sam"
python3 "${comparison_script}" \
    --modern-graft "${evidence_directory}/modern-graft-scores.tsv" \
    --modern-host "${evidence_directory}/modern-host-scores.tsv" \
    --legacy-graft-sam "${evidence_directory}/legacy-graft.sam" \
    --legacy-host-sam "${evidence_directory}/legacy-host.sam" \
    --samples "${samples}" \
    --names "${names_file}" \
    --report "${evidence_directory}/score-decision-comparison.json"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
