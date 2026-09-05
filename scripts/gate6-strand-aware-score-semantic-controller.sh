#!/usr/bin/env bash
#SBATCH --job-name=gate6-strand-score-semantic
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=32G
#SBATCH --time=04:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-strand-score-semantic-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/strand-score-semantic-${SLURM_JOB_ID}"
stratification_job_id="${GATE6_STRATIFICATION_JOB_ID:?GATE6_STRATIFICATION_JOB_ID is required}"
stratification_directory="${evidence_root}/fragment-disagreement-stratification-${stratification_job_id}"
project_root="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/runs/run-20260823T034624Z-ndcwfa"
old_scoreaudit_binary="${runtime_directory}/gate6-xenofilx-scoreaudit-pinned-r35"
fixed_scoreaudit_binary="${runtime_directory}/gate6-xenofilx-scoreaudit-names"
samples="${stratification_directory}/sampled-fragments.tsv"
graft_input="${project_root}/work/bsmap/SRR36187610_hg38.bam"
graft_reference="/public3/home/scg9946/otter-gate6/references/genomes/hg38/GRCh38-gencode-v44/fasta/hg38.fa"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
for required_path in \
    "${old_scoreaudit_binary}" "${fixed_scoreaudit_binary}" "${samples}" \
    "${graft_input}" "${graft_reference}"; do
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
    printf 'old_scoreaudit_sha256=%s\n' "$(sha256sum "${old_scoreaudit_binary}" | awk '{print $1}')"
    printf 'fixed_scoreaudit_sha256=%s\n' "$(sha256sum "${fixed_scoreaudit_binary}" | awk '{print $1}')"
    printf 'sample_qnames_sha256=%s\n' "$(sha256sum "${names_file}" | awk '{print $1}')"
    printf 'graft_input_sha256=%s\n' "$(sha256sum "${graft_input}" | awk '{print $1}')"
    printf 'graft_reference_sha256=%s\n' "$(sha256sum "${graft_reference}" | awk '{print $1}')"
} > "${evidence_directory}/manifest.properties"

"${old_scoreaudit_binary}" \
    --input "${graft_input}" \
    --reference "${graft_reference}" \
    --names "${names_file}" \
    --bisulfite \
    --report "${evidence_directory}/old-r35-scores.tsv"
"${fixed_scoreaudit_binary}" \
    --input "${graft_input}" \
    --reference "${graft_reference}" \
    --names "${names_file}" \
    --bisulfite \
    --report "${evidence_directory}/fixed-strand-aware-scores.tsv"

python3 - \
    "${evidence_directory}/old-r35-scores.tsv" \
    "${evidence_directory}/fixed-strand-aware-scores.tsv" \
    "${names_file}" \
    "${evidence_directory}/score-semantic-comparison.json" <<'PY'
import csv
import json
import sys
from collections import Counter, defaultdict
from pathlib import Path

old_path, fixed_path, names_path, report_path = map(Path, sys.argv[1:])
selected_names = {line.rstrip("\n") for line in names_path.read_text(encoding="utf-8").splitlines()}
if not selected_names or "" in selected_names:
    raise SystemExit("invalid QNAME allowlist")


def load_scores(path: Path) -> dict[str, dict[str, int]]:
    scores: dict[str, dict[str, int]] = defaultdict(dict)
    with path.open(encoding="utf-8", newline="") as score_file:
        reader = csv.DictReader(score_file, delimiter="\t")
        required = {"qname", "flag", "classification_score"}
        if reader.fieldnames is None or not required.issubset(reader.fieldnames):
            raise ValueError(f"unexpected score header in {path}: {reader.fieldnames}")
        for row in reader:
            qname = row["qname"]
            if qname not in selected_names:
                continue
            flag = int(row["flag"])
            if flag & (0x4 | 0x100 | 0x800):
                continue
            first_mate = bool(flag & 0x40)
            second_mate = bool(flag & 0x80)
            if first_mate == second_mate:
                raise ValueError(f"invalid mate flags for {qname}: {flag}")
            mate = "first" if first_mate else "second"
            if mate in scores[qname]:
                raise ValueError(f"duplicate primary {mate} mate for {qname}")
            scores[qname][mate] = int(row["classification_score"])
    return scores


def paired_decision(scores: dict[str, int]) -> str:
    if set(scores) != {"first", "second"}:
        return "incomplete_pair"
    if scores["first"] >= 6 or scores["second"] >= 6:
        return "discarded_threshold"
    return "graft_host_absent"

old_scores = load_scores(old_path)
fixed_scores = load_scores(fixed_path)
missing_scores = Counter()
decision_pairs = Counter()
score_transitions = Counter()
for qname in selected_names:
    if qname not in old_scores:
        missing_scores["old_r35"] += 1
        continue
    if qname not in fixed_scores:
        missing_scores["fixed_strand_aware"] += 1
        continue
    old_pair = old_scores[qname]
    fixed_pair = fixed_scores[qname]
    old_decision = paired_decision(old_pair)
    fixed_decision = paired_decision(fixed_pair)
    decision_pairs[f"old={old_decision};fixed={fixed_decision}"] += 1
    score_transitions[
        f"old_scores={old_pair.get('first')},{old_pair.get('second')};"
        f"fixed_scores={fixed_pair.get('first')},{fixed_pair.get('second')}"
    ] += 1

report = {
    "schema_version": "gate6.strand-score-semantic/v1",
    "sample_size": len(selected_names),
    "threshold": 6,
    "decision_pairs": dict(sorted(decision_pairs.items())),
    "missing_score_records": dict(sorted(missing_scores.items())),
    "score_transitions": dict(sorted(score_transitions.items(), key=lambda item: (-item[1], item[0]))),
}
report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
expected = decision_pairs["old=graft_host_absent;fixed=discarded_threshold"]
raise SystemExit(0 if expected > 0 and not missing_scores else 1)
PY

printf 'completed_at_utc=%s\n' "$(date -u +%FT%TZ)" >> "${evidence_directory}/manifest.properties"
printf '%s\n' "${evidence_directory}"
