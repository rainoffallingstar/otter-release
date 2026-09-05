#!/usr/bin/env bash
#SBATCH --job-name=gate6-paired-decision-audit
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-paired-decision-audit-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/paired-decision-audit-${SLURM_JOB_ID}"
membership_job_id="${GATE6_MEMBERSHIP_JOB_ID:?GATE6_MEMBERSHIP_JOB_ID is required}"
membership_directory="${evidence_root}/fragment-membership-strand-aware-legacy-${membership_job_id}"
legacy_job_id="${GATE6_LEGACY_JOB_ID:?GATE6_LEGACY_JOB_ID is required}"
legacy_directory="${evidence_root}/legacy-picard-strand-aware-xenofilter-${legacy_job_id}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
samtools_binary="/public3/home/scg9946/TTest/breg/soft/bin/samtools"
scoreaudit_binary="${runtime_directory}/gate6-xenofilx-scoreaudit-names"
audit_script="${runtime_directory}/gate6-paired-decision-audit.py"
modern_graft="${project_root}/work/bsmap/SRR36187610_hg38.bam"
modern_host="${project_root}/work/bsmap/SRR36187610_mm10.bam"
graft_reference="/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"
host_reference="/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25/fasta/mm10.fa"
legacy_graft="${legacy_directory}/SRR36187610_hg38.picard-strand-aware-nm.bam"
legacy_host="${legacy_directory}/SRR36187610_mm10.picard-strand-aware-nm.bam"
disagreements="${membership_directory}/disagreements.tsv.gz"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in "${samtools_binary}" "${scoreaudit_binary}" "${audit_script}" "${modern_graft}" "${modern_host}" "${graft_reference}" "${host_reference}" "${legacy_graft}" "${legacy_host}" "${disagreements}"; do
    if [[ ! -r "${required_path}" ]]; then
        printf 'missing readable required path: %s\n' "${required_path}" >&2
        exit 1
    fi
done

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
sample_path="${evidence_directory}/membership-sample.tsv"
names_path="${evidence_directory}/sampled-qnames.txt"
python3 - "${disagreements}" "${sample_path}" "${names_path}" <<'PY'
import csv
import gzip
import sys

source_path, sample_path, names_path = sys.argv[1:]
limits = {("discarded", "graft"): 2000, ("graft", "discarded"): 2000}
selected = []
counts = {key: 0 for key in limits}
with gzip.open(source_path, "rt", encoding="utf-8", newline="") as source_file:
    reader = csv.DictReader(source_file, delimiter="\t")
    for row in reader:
        key = (row["modern_classification"], row["legacy_classification"])
        if key not in limits or counts[key] >= limits[key]:
            continue
        selected.append(row)
        counts[key] += 1
        if all(counts[key] == limits[key] for key in limits):
            break
if not selected:
    raise SystemExit("no disagreement sample was selected")
with open(sample_path, "w", encoding="utf-8", newline="") as sample_file:
    writer = csv.DictWriter(sample_file, fieldnames=selected[0].keys(), delimiter="\t")
    writer.writeheader()
    writer.writerows(selected)
with open(names_path, "w", encoding="utf-8") as names_file:
    names_file.write("\n".join(sorted(row["fragment_id"] for row in selected)) + "\n")
print(counts)
PY

{
    printf 'slurm_job_id=%s\n' "${SLURM_JOB_ID}"
    printf 'membership_job_id=%s\n' "${membership_job_id}"
    printf 'legacy_job_id=%s\n' "${legacy_job_id}"
    printf 'sample_count=%s\n' "$(wc -l < "${names_path}")"
    printf 'scoreaudit_sha256=%s\n' "$(sha256sum "${scoreaudit_binary}" | awk '{print $1}')"
    printf 'audit_script_sha256=%s\n' "$(sha256sum "${audit_script}" | awk '{print $1}')"
    printf 'sample_sha256=%s\n' "$(sha256sum "${sample_path}" | awk '{print $1}')"
    printf 'modern_graft_sha256=%s\n' "$(sha256sum "${modern_graft}" | awk '{print $1}')"
    printf 'modern_host_sha256=%s\n' "$(sha256sum "${modern_host}" | awk '{print $1}')"
    printf 'legacy_graft_sha256=%s\n' "$(sha256sum "${legacy_graft}" | awk '{print $1}')"
    printf 'legacy_host_sha256=%s\n' "$(sha256sum "${legacy_host}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

extract_sample() {
    local input_bam="$1"
    local output_sam="$2"
    python3 - "${samtools_binary}" "${input_bam}" "${names_path}" "${output_sam}" <<'PY'
import subprocess
import sys

samtools_path, bam_path, names_path, output_path = sys.argv[1:]
with open(names_path, encoding="utf-8") as names_file:
    selected_names = {line.rstrip("\n") for line in names_file}
with subprocess.Popen([samtools_path, "view", bam_path], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, encoding="utf-8") as process, open(output_path, "w", encoding="utf-8") as output_file:
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

"${samtools_binary}" quickcheck -v "${modern_graft}" "${modern_host}" "${legacy_graft}" "${legacy_host}" > "${evidence_directory}/input.quickcheck.txt" 2>&1
"${scoreaudit_binary}" --input "${modern_graft}" --reference "${graft_reference}" --names "${names_path}" --bisulfite --report "${evidence_directory}/modern-graft-scores.tsv"
"${scoreaudit_binary}" --input "${modern_host}" --reference "${host_reference}" --names "${names_path}" --bisulfite --report "${evidence_directory}/modern-host-scores.tsv"
extract_sample "${legacy_graft}" "${evidence_directory}/legacy-graft.sam"
extract_sample "${legacy_host}" "${evidence_directory}/legacy-host.sam"
python3 "${audit_script}" \
    --sample "${sample_path}" \
    --modern-graft "${evidence_directory}/modern-graft-scores.tsv" \
    --modern-host "${evidence_directory}/modern-host-scores.tsv" \
    --legacy-graft "${evidence_directory}/legacy-graft.sam" \
    --legacy-host "${evidence_directory}/legacy-host.sam" \
    --report "${evidence_directory}/paired-decision-report.json"
printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
