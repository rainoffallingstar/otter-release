#!/usr/bin/env bash
set -u
# Bind accepted parent step2 work outputs into the new immutable snapshot work/.
# Parent outputs are reused only through symlinks; the snapshot itself is untouched.

base="/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects"

# modern: child run-20260821T002101Z-hycjez, parent run-20260817T001228Z-vclyug
modern_child="${base}/bs-pdx-SRR23802966/runs/run-20260821T002101Z-hycjez"
modern_parent="${base}/bs-pdx-SRR23802966/runs/run-20260817T001228Z-vclyug"
mkdir -p "${modern_child}/work"
for d in trim QC fastqc_raw fastqc_clean bsmap; do
  if [ -e "${modern_child}/work/${d}" ]; then
    printf 'skip-existing modern %s\n' "${d}"
  else
    ln -s "${modern_parent}/work/${d}" "${modern_child}/work/${d}"
    printf 'linked modern %s\n' "${d}"
  fi
done

# legacy: child run-20260821T002226Z-nfioxh, parent run-20260817T102532Z-emvbei
legacy_child="${base}/bs-pdx-SRR23802966-legacy-equivalent/runs/run-20260821T002226Z-nfioxh"
legacy_parent="${base}/bs-pdx-SRR23802966-legacy-equivalent/runs/run-20260817T102532Z-emvbei"
mkdir -p "${legacy_child}/work"
for d in trim QC fastqc_raw fastqc_clean bsmap; do
  if [ -e "${legacy_child}/work/${d}" ]; then
    printf 'skip-existing legacy %s\n' "${d}"
  else
    ln -s "${legacy_parent}/work/${d}" "${legacy_child}/work/${d}"
    printf 'linked legacy %s\n' "${d}"
  fi
done

echo 'bind-complete'
