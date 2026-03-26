#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "${SCRIPT_DIR}/.." && pwd)
BIN=${BIN:-"${REPO_ROOT}/enva/target/debug/enva"}
WORKDIR=$(mktemp -d /tmp/enva-e2e-gha-XXXXXX)
ENV_NAME="enva-e2e-gha"
ROOT_PREFIX="${WORKDIR}/root"
YAML_PATH="${WORKDIR}/minimal.yaml"

cleanup() {
  set +e
  if [[ -x "${BIN}" ]]; then
    ENVA_RATTLER_ROOT_PREFIX="${ROOT_PREFIX}" "${BIN}" remove "${ENV_NAME}" >/dev/null 2>&1 || true
  fi
  rm -rf "${WORKDIR}"
}
trap cleanup EXIT

mkdir -p "${ROOT_PREFIX}"
cat > "${YAML_PATH}" <<'YAML'
name: ignored-by-cli
channels:
  - conda-forge
dependencies:
  - xz
YAML

export ENVA_RATTLER_ROOT_PREFIX="${ROOT_PREFIX}"
PREFIX="${ROOT_PREFIX}/envs/${ENV_NAME}"

"${BIN}" --version
"${BIN}" create --yaml "${YAML_PATH}" --name "${ENV_NAME}" --with jq

test -d "${PREFIX}/conda-meta"

XZ_OUTPUT=$("${BIN}" run "${ENV_NAME}" -- xz --version 2>&1)
printf '%s\n' "${XZ_OUTPUT}"
grep -q "XZ Utils" <<<"${XZ_OUTPUT}"

JQ_OUTPUT=$("${BIN}" run "${ENV_NAME}" -- jq --version 2>&1)
printf '%s\n' "${JQ_OUTPUT}"
grep -q "jq-" <<<"${JQ_OUTPUT}"

BIN="${BIN}" PREFIX="${PREFIX}" ENV_NAME="${ENV_NAME}" bash -lc '
  set -euo pipefail
  PATH_BEFORE="$PATH"
  eval "$($BIN shell hook bash)"
  type enva >/dev/null
  enva activate "$ENV_NAME"
  test "$CONDA_PREFIX" = "$PREFIX"
  test "$ENVA_ACTIVE_NAME" = "$ENV_NAME"
  test "${PATH%%:*}" = "$PREFIX/bin"
  enva deactivate
  test -z "${CONDA_PREFIX-}"
  test -z "${ENVA_ACTIVE_NAME-}"
  test "$PATH" = "$PATH_BEFORE"
'

echo "enva e2e passed"
