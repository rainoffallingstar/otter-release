#!/usr/bin/env bash
#SBATCH --job-name=gate6-xenofilter-fixture
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=1
#SBATCH --mem=8G
#SBATCH --time=00:30:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-xenofilter-fixture-%j.log
set -euo pipefail

runtime_directory="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime"
evidence_root="/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx"
evidence_directory="${evidence_root}/xenofilter-fixture-${SLURM_JOB_ID}"
legacy_script="${runtime_directory}/gate6-xenofilter-legacy.R"
package_root="/public3/home/scg9946/TTest/soft/MyMiniconda/lib/R/library/XenofilteR"

if [[ -e "${evidence_directory}" ]]; then
    printf 'refusing to overwrite evidence directory: %s\n' "${evidence_directory}" >&2
    exit 1
fi
if [[ ! -r "${legacy_script}" ]]; then
    printf 'missing legacy wrapper: %s\n' "${legacy_script}" >&2
    exit 1
fi

mkdir -p "${evidence_root}"
mkdir "${evidence_directory}"
mkdir "${evidence_directory}/destination"

Rscript "${legacy_script}" \
    "${package_root}/extdata/Test_hg19_NRAS.bam" \
    "${package_root}/extdata/Test_mm10_NRAS.bam" \
    "${evidence_directory}/destination" \
    1 \
    6 \
    8 \
    NM \
    "${evidence_directory}/legacy-run.json" \
    > "${evidence_directory}/stdout.txt" \
    2> "${evidence_directory}/stderr.txt"

python3 - "${evidence_directory}" > "${evidence_directory}/files.tsv" <<'PY'
import pathlib
import sys
root = pathlib.Path(sys.argv[1])
for path in sorted(item for item in root.rglob("*") if item.is_file() and item.name != "files.tsv"):
    print(f"{path.relative_to(root)}\t{path.stat().st_size}")
PY
printf '%s\n' "${evidence_directory}"
