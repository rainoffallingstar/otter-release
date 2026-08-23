#!/usr/bin/env bash
#SBATCH --job-name=otter-verify-srr36187610-decode
#SBATCH --partition=amd_512
#SBATCH --account=zc-m6
#SBATCH --qos=normal
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=02:00:00
#SBATCH --output=/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/verify-SRR36187610-decode.out
#SBATCH --error=/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/verify-SRR36187610-decode.err
set -uo pipefail
bash /public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/verify-decoded.sh
