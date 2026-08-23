#!/usr/bin/env bash
set -euo pipefail

readonly root='/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z'
readonly output_root="${root}/decoded"
readonly expected_archive_sha256='713d1eb153b3db279e8aff6df7139c48687458e79a79008f2ceed5248b2569bb'
readonly expected_r1_sha256='aa8a2225ee4f68b81348e709632a7567e7cd2ada7ed9a4e9e2475c8299810706'
readonly expected_r2_sha256='2d564420b68119087863eeb4b5f8ae0120a9ccaf7cf20708c6b0678d8eac3c05'
readonly expected_record_count='71242412'

[[ -d "${output_root}" && ! -L "${output_root}" ]]
[[ "$(stat -c '%a' "${output_root}")" == '555' ]]
for artifact_name in R1.fastq.gz R2.fastq.gz checksums.sha256 sra-acquisition-decode.json; do
  artifact_path="${output_root}/${artifact_name}"
  [[ -f "${artifact_path}" && ! -L "${artifact_path}" ]]
  [[ "$(stat -c '%a' "${artifact_path}")" == '444' ]]
done
(cd "${output_root}" && sha256sum -c checksums.sha256)
gzip -t "${output_root}/R1.fastq.gz" "${output_root}/R2.fastq.gz"

python3 - "${output_root}" "${expected_archive_sha256}" "${expected_r1_sha256}" "${expected_r2_sha256}" "${expected_record_count}" <<'PY'
import gzip
import hashlib
import json
import sys
from pathlib import Path

output_root = Path(sys.argv[1])
expected_archive_sha256, expected_r1_sha256, expected_r2_sha256 = sys.argv[2:5]
expected_record_count = int(sys.argv[5])


def calculate_sha256(path: Path) -> str:
    checksum = hashlib.sha256()
    with path.open("rb") as input_handle:
        for byte_chunk in iter(lambda: input_handle.read(1024 * 1024), b""):
            checksum.update(byte_chunk)
    return checksum.hexdigest()


def count_fastq_records(path: Path) -> int:
    total_records = 0
    with gzip.open(path, "rb") as input_handle:
        while header_line := input_handle.readline():
            sequence_line = input_handle.readline()
            separator_line = input_handle.readline()
            quality_line = input_handle.readline()
            if not sequence_line or not separator_line or not quality_line:
                raise RuntimeError(f"truncated FASTQ record in {path}")
            if not header_line.startswith(b"@") or not separator_line.startswith(b"+"):
                raise RuntimeError(f"invalid FASTQ record in {path}")
            if len(sequence_line.rstrip()) != len(quality_line.rstrip()):
                raise RuntimeError(f"sequence/quality length mismatch in {path}")
            total_records += 1
    return total_records


r1_path = output_root / "R1.fastq.gz"
r2_path = output_root / "R2.fastq.gz"
manifest = json.loads((output_root / "sra-acquisition-decode.json").read_text())
assert manifest["archive"]["sha256"] == "sha256:" + expected_archive_sha256
assert manifest["outputs"]["r1"]["path"] == str(r1_path)
assert manifest["outputs"]["r2"]["path"] == str(r2_path)
r1_sha256, r2_sha256 = calculate_sha256(r1_path), calculate_sha256(r2_path)
assert r1_sha256 == expected_r1_sha256, f"R1 SHA256 mismatch: {r1_sha256}"
assert r2_sha256 == expected_r2_sha256, f"R2 SHA256 mismatch: {r2_sha256}"
assert manifest["outputs"]["r1"]["sha256"] == "sha256:" + r1_sha256
assert manifest["outputs"]["r2"]["sha256"] == "sha256:" + r2_sha256
r1_records, r2_records = count_fastq_records(r1_path), count_fastq_records(r2_path)
assert r1_records == r2_records == expected_record_count
assert manifest["outputs"]["r1"]["paired_record_count"] == expected_record_count
assert manifest["outputs"]["r2"]["paired_record_count"] == expected_record_count
print(json.dumps({"status": "verified", "paired_record_count": expected_record_count, "r1_sha256": r1_sha256, "r2_sha256": r2_sha256}, sort_keys=True))
PY
