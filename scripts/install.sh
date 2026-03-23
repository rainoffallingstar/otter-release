#!/usr/bin/env bash
# =============================================================================
#  xdxtools Installer
#  Downloads pre-built binaries from GitHub Releases and sets up conda envs.
#
#  Usage:
#    bash install.sh [OPTIONS]
#
#  Options:
#    --install-dir PATH   Override binary installation directory
#    --skip-envs          Skip conda environment creation
#    --skip-hdf5          Skip HDF5 configuration for methrix-cli
#    --non-interactive    Use all defaults without prompting
#    --dry-run            Print all actions without executing
#    --version VER        Specify release version (e.g. v0.3.0); default: latest
#    --releases-repo REPO  Override GitHub release repo (owner/name)
#    --help               Show this help message
#
#  Environment:
#    GITHUB_TOKEN / GH_TOKEN        Optional GitHub token for private release downloads
#    GITHUB_RELEASES_REPO           Optional release repo override (owner/name)
#
#  Interactive behavior:
#    If GitHub access fails and no token is configured, interactive mode can
#    prompt for a hidden token input and retry once for the current session.
# =============================================================================

set -euo pipefail

# ── Top-level configuration ───────────────────────────────────────────────────
RELEASES_REPO="${GITHUB_RELEASES_REPO:-rainoffallingstar/xdxtools-go}"
XDXTOOLS_VERSION="latest"
DEFAULT_INSTALL_DIR="$HOME/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENVS_DIR="$SCRIPT_DIR/../inst/envs"
GITHUB_AUTH_TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"

# All 9 tools: binary_name:release_asset_stem:linkage(static|dynamic)
TOOLS=(
  "xdxtools:xdxtools:static"
  "enva:enva:static"
  "xenofilter:xenofilter:static"
  "paireads:paireads:static"
  "htseq2matrix:htseq2matrix:static"
  "methrix-cli:methrix:dynamic"
  "qctb:qctb:static"
  "fqc:fqc:static"
  "gomats:gomats:static"
)

# Conda environment yaml files
ENV_FILES=(
  "xdxtools-core.yaml"
  "xdxtools-snakemake.yaml"
  "xdxtools-extra.yaml"
)

# ── Parse arguments ───────────────────────────────────────────────────────────
INSTALL_DIR=""
SKIP_ENVS=false
SKIP_HDF5=false
NON_INTERACTIVE=false
DRY_RUN=false
SHOW_HELP=false
INSTALL_ENVS_CHOICE="all"   # all | core | snakemake | extra
GITHUB_TOKEN_PROMPT_ATTEMPTED=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --install-dir)    INSTALL_DIR="$2"; shift 2 ;;
    --skip-envs)      SKIP_ENVS=true;   shift ;;
    --skip-hdf5)      SKIP_HDF5=true;   shift ;;
    --non-interactive) NON_INTERACTIVE=true; shift ;;
    --dry-run)        DRY_RUN=true;     shift ;;
    --version)        XDXTOOLS_VERSION="$2"; shift 2 ;;
    --releases-repo)  RELEASES_REPO="$2"; shift 2 ;;
    --help)           SHOW_HELP=true; shift ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information."
      exit 1
      ;;
  esac
done

# ── Helpers ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'; YELLOW='\033[1;33m'; GREEN='\033[0;32m'
BOLD='\033[1m'; RESET='\033[0m'

log_info()    { echo -e "  ${BOLD}[INFO]${RESET}  $*"; }
log_success() { echo -e "  ${GREEN}✓${RESET} $*"; }
log_warn()    { echo -e "  ${YELLOW}⚠${RESET}  $*"; }
log_error()   { echo -e "  ${RED}✗${RESET}  $*" >&2; }


print_help() {
  cat <<'EOF'
# =============================================================================
#  xdxtools Installer
#  Downloads pre-built binaries from GitHub Releases and sets up conda envs.
#
#  Usage:
#    bash install.sh [OPTIONS]
#
#  Options:
#    --install-dir PATH   Override binary installation directory
#    --skip-envs          Skip conda environment creation
#    --skip-hdf5          Skip HDF5 configuration for methrix-cli
#    --non-interactive    Use all defaults without prompting
#    --dry-run            Print all actions without executing
#    --version VER        Specify release version (e.g. v0.3.0); default: latest
#    --releases-repo REPO  Override GitHub release repo (owner/name)
#    --help               Show this help message
#
#  Environment:
#    GITHUB_TOKEN / GH_TOKEN        Optional GitHub token for private release downloads
#    GITHUB_RELEASES_REPO           Optional release repo override (owner/name)
#
#  Interactive behavior:
#    If GitHub access fails and no token is configured, interactive mode can
#    prompt for a hidden token input and retry once for the current session.
EOF
}

if [ "$SHOW_HELP" = true ]; then
  print_help
  exit 0
fi

github_api_get() {
  local url="$1"
  if [ -n "$GITHUB_AUTH_TOKEN" ]; then
    curl -sf       -H "Accept: application/vnd.github+json"       -H "Authorization: Bearer $GITHUB_AUTH_TOKEN"       -H "X-GitHub-Api-Version: 2022-11-28"       "$url"
  else
    curl -sf "$url"
  fi
}

github_release_download() {
  local url="$1" dest="$2"
  if [ -n "$GITHUB_AUTH_TOKEN" ]; then
    curl -fL --progress-bar       -H "Authorization: Bearer $GITHUB_AUTH_TOKEN"       "$url" -o "$dest"
  else
    curl -fL --progress-bar "$url" -o "$dest"
  fi
}

resolve_latest_release_tag() {
  github_api_get "https://api.github.com/repos/${RELEASES_REPO}/releases/latest" \
    | grep '"tag_name"' | cut -d'"' -f4
}

run() {
  if [ "$DRY_RUN" = true ]; then
    echo -e "  ${YELLOW}[DRY-RUN]${RESET} $*"
  else
    eval "$@"
  fi
}

ask() {
  # ask <prompt> <default>
  local prompt="$1" default="$2"
  if [ "$NON_INTERACTIVE" = true ]; then
    echo "$default"
    return
  fi
  read -rp "  $prompt [$default]: " answer
  echo "${answer:-$default}"
}

ask_yn() {
  # ask_yn <prompt> <Y|n>
  local prompt="$1" default="$2"
  if [ "$NON_INTERACTIVE" = true ]; then
    [[ "$default" =~ ^[Yy] ]] && return 0 || return 1
  fi
  read -rp "  $prompt (${default}): " answer
  answer="${answer:-$default}"
  [[ "$answer" =~ ^[Yy] ]] && return 0 || return 1
}


ask_secret() {
  # ask_secret <prompt>
  local prompt="$1"
  if [ "$NON_INTERACTIVE" = true ]; then
    echo ""
    return
  fi
  local answer
  read -rsp "  $prompt: " answer
  echo >&2
  echo "$answer"
}

maybe_prompt_github_token_on_failure() {
  local reason="${1:-GitHub access failed.}"

  if [ -n "$GITHUB_AUTH_TOKEN" ] || [ "$NON_INTERACTIVE" = true ] || [ "$GITHUB_TOKEN_PROMPT_ATTEMPTED" = true ]; then
    return 1
  fi

  GITHUB_TOKEN_PROMPT_ATTEMPTED=true
  log_warn "$reason"

  if ask_yn "Enter a GitHub token now and retry once" "Y"; then
    GITHUB_AUTH_TOKEN=$(ask_secret "GitHub token (input hidden, used only for this run)")
    if [ -n "$GITHUB_AUTH_TOKEN" ]; then
      log_success "GitHub token captured for this session"
      return 0
    fi
    log_warn "Empty token entered; continuing without authenticated release access"
  fi

  return 1
}

divider() { echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; }

# ── Step 0: Banner ────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}╔══════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║        xdxtools Installer                ║${RESET}"
echo -e "${BOLD}║  Bioinformatics Workflow Manager         ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════════╝${RESET}"
echo ""
[ "$DRY_RUN" = true ] && echo -e "${YELLOW}  DRY-RUN MODE – no changes will be made${RESET}\n"

# ── Step 1: System dependency check ──────────────────────────────────────────
echo -e "${BOLD}Step 1: Checking system dependencies${RESET}"

# Required: curl
if ! command -v curl &>/dev/null; then
  log_error "curl is required but not found. Please install curl and re-run."
  exit 1
fi
log_success "curl found"
if [ -n "$GITHUB_AUTH_TOKEN" ]; then
  log_success "GitHub token detected – authenticated release access enabled"
fi

# Architecture detection
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    log_error "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac
log_success "Architecture: $ARCH"

# Shell detection
DETECTED_SHELL="$(basename "${SHELL:-bash}")"
case "$DETECTED_SHELL" in
  zsh)  DEFAULT_SHELL_CONFIG="$HOME/.zshrc"  ;;
  bash) DEFAULT_SHELL_CONFIG="$HOME/.bashrc" ;;
  *)    DEFAULT_SHELL_CONFIG="$HOME/.profile" ;;
esac
log_success "Shell: $DETECTED_SHELL → $DEFAULT_SHELL_CONFIG"

# conda / mamba / micromamba detection
PM=""
HAS_CONDA=false
for pm in mamba micromamba conda; do
  if command -v "$pm" &>/dev/null; then
    PM="$pm"
    HAS_CONDA=true
    log_success "Package manager: $pm"
    break
  fi
done
if [ "$HAS_CONDA" = false ]; then
  log_warn "No conda/mamba/micromamba found – conda environment steps will be skipped"
  SKIP_ENVS=true
fi

echo ""

# ── Step 2: Interactive configuration ────────────────────────────────────────
echo -e "${BOLD}Step 2: Configuration${RESET}"

if [ -z "$INSTALL_DIR" ]; then
  INSTALL_DIR=$(ask "Installation directory" "$DEFAULT_INSTALL_DIR")
fi
log_info "Binaries will be installed to: $INSTALL_DIR"
log_info "GitHub releases repo: $RELEASES_REPO"
SHELL_CONFIG=$(ask "Shell config file" "$DEFAULT_SHELL_CONFIG")
log_info "Shell config: $SHELL_CONFIG"

if [ "$SKIP_ENVS" = false ]; then
  if ask_yn "Install conda environments?" "Y"; then
    echo -e "  Which environments? [a]ll / [c]ore / [s]nakemake / [e]xtra"
    if [ "$NON_INTERACTIVE" = false ]; then
      read -rp "  Choice [a]: " env_choice
    else
      env_choice="a"
    fi
    case "${env_choice:-a}" in
      c|core)       INSTALL_ENVS_CHOICE="core"       ;;
      s|snakemake)  INSTALL_ENVS_CHOICE="snakemake"  ;;
      e|extra)      INSTALL_ENVS_CHOICE="extra"       ;;
      *)            INSTALL_ENVS_CHOICE="all"         ;;
    esac
    log_info "Environments to install: $INSTALL_ENVS_CHOICE"
  else
    SKIP_ENVS=true
    log_info "Skipping conda environments"
  fi
fi

echo ""

# ── Step 3: Resolve version and download binaries ────────────────────────────
echo -e "${BOLD}Step 3: Downloading binaries${RESET}"

# Resolve "latest" tag via GitHub API
if [ "$XDXTOOLS_VERSION" = "latest" ]; then
  log_info "Querying GitHub API for latest release..."
  latest_release_tag=""
  if ! latest_release_tag=$(resolve_latest_release_tag 2>/dev/null); then
    latest_release_tag=""
  fi
  if [ -z "$latest_release_tag" ] && maybe_prompt_github_token_on_failure "Latest release query failed. Private release repos usually require a GitHub token."; then
    if ! latest_release_tag=$(resolve_latest_release_tag 2>/dev/null); then
      latest_release_tag=""
    fi
  fi
  if [ -z "$latest_release_tag" ]; then
    log_error "Could not determine latest version. If the release repo is private, export GITHUB_TOKEN or GH_TOKEN and retry, or use --version to specify."
    exit 1
  fi
  XDXTOOLS_VERSION="$latest_release_tag"
fi
log_info "Version: $XDXTOOLS_VERSION"

run "mkdir -p \"$INSTALL_DIR\""

BASE_URL="https://github.com/${RELEASES_REPO}/releases/download/${XDXTOOLS_VERSION}"

for entry in "${TOOLS[@]}"; do
  IFS=':' read -r bin_name asset_stem linkage <<< "$entry"
  asset="${asset_stem}-linux-${ARCH}"
  if [ "${linkage}" = "static" ]; then
    asset="${asset}-static"
  fi
  dest="${INSTALL_DIR}/${bin_name}"

  if [ -f "$dest" ] && [ "$DRY_RUN" = false ]; then
    log_info "[已安装] $bin_name – skipping (use --install-dir to reinstall)"
    continue
  fi

  url="${BASE_URL}/${asset}"
  log_info "Downloading $bin_name ..."

  if [ "$DRY_RUN" = true ]; then
    if [ -n "$GITHUB_AUTH_TOKEN" ]; then
      printf '  %b[DRY-RUN]%b curl -fL --progress-bar -H "Authorization: Bearer $GITHUB_TOKEN" "%s" -o "%s"\n' "$YELLOW" "$RESET" "$url" "$dest"
    else
      printf '  %b[DRY-RUN]%b curl -fL --progress-bar "%s" -o "%s"\n' "$YELLOW" "$RESET" "$url" "$dest"
    fi
  else
    if github_release_download "$url" "$dest"; then
      chmod +x "$dest"
      log_success "$bin_name installed"
    elif maybe_prompt_github_token_on_failure "Download failed for $bin_name. Private release assets usually require a GitHub token."; then
      if github_release_download "$url" "$dest"; then
        chmod +x "$dest"
        log_success "$bin_name installed"
      else
        log_warn "$bin_name download failed (HTTP error – asset may not exist for this release, or authentication is required)"
      fi
    else
      log_warn "$bin_name download failed (HTTP error – asset may not exist for this release, or authentication is required)"
    fi
  fi
done
# Keep backward compatibility for legacy scripts that call `methrix`.
if [ "$DRY_RUN" = false ]; then
  if [ -f "${INSTALL_DIR}/methrix-cli" ] && [ ! -e "${INSTALL_DIR}/methrix" ]; then
    ln -s "${INSTALL_DIR}/methrix-cli" "${INSTALL_DIR}/methrix"
    log_info "Created compatibility symlink: methrix -> methrix-cli"
  fi
else
  echo -e "  ${YELLOW}[DRY-RUN]${RESET} ln -s \"${INSTALL_DIR}/methrix-cli\" \"${INSTALL_DIR}/methrix\""
fi

echo ""

# ── Step 4: Verify tools ──────────────────────────────────────────────────────
echo -e "${BOLD}Step 4: Verifying installed tools${RESET}"

# Temporarily add INSTALL_DIR to PATH for verification
export PATH="${INSTALL_DIR}:${PATH}"

for entry in "${TOOLS[@]}"; do
  bin_name="${entry%%:*}"
  dest="${INSTALL_DIR}/${bin_name}"

  if [ "$bin_name" = "methrix-cli" ]; then
    echo -e "  ${YELLOW}⚠${RESET}  methrix-cli  (needs HDF5 config – will verify in Step 8)"
    continue
  fi

  if [ "$DRY_RUN" = true ]; then
    echo -e "  ${YELLOW}[DRY-RUN]${RESET} would verify: $bin_name --version"
    continue
  fi

  if [ -f "$dest" ]; then
    ver=$("$dest" --version 2>&1 | head -1 || echo "(version unknown)")
    log_success "$bin_name  $ver"
  else
    log_warn "$bin_name  not found at $dest"
  fi
done

echo ""

# ── Step 5: Configure PATH ────────────────────────────────────────────────────
echo -e "${BOLD}Step 5: Configuring PATH${RESET}"

if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  log_info "Adding $INSTALL_DIR to PATH in $SHELL_CONFIG"
  run "echo '' >> \"$SHELL_CONFIG\""
  run "echo '# xdxtools: binary install directory' >> \"$SHELL_CONFIG\""
  run "echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> \"$SHELL_CONFIG\""
  log_success "PATH updated in $SHELL_CONFIG"
else
  log_success "$INSTALL_DIR is already in PATH"
fi

echo ""

# ── Step 6: Create conda environments ────────────────────────────────────────
if [ "$SKIP_ENVS" = false ]; then
  echo -e "${BOLD}Step 6: Creating conda environments${RESET}"

  # Check that inst/envs/ is accessible
  if [ ! -d "$ENVS_DIR" ]; then
    log_warn "inst/envs/ not found at $ENVS_DIR"
    log_warn "Clone the full repository to access environment files:"
    echo ""
    echo "    git clone --recurse-submodules https://github.com/rainoffallingstar/xdxtools-go.git"
    echo "    bash xdxtools-go/scripts/install.sh"
    echo ""
    SKIP_ENVS=true
    SKIP_HDF5=true
  fi
fi

if [ "$SKIP_ENVS" = false ]; then
  # Map choice → env files to install
  declare -A ENV_SELECT
  case "$INSTALL_ENVS_CHOICE" in
    all)
      for f in "${ENV_FILES[@]}"; do ENV_SELECT["$f"]=1; done
      ;;
    core)
      ENV_SELECT["xdxtools-core.yaml"]=1
      ;;
    snakemake)
      ENV_SELECT["xdxtools-core.yaml"]=1
      ENV_SELECT["xdxtools-snakemake.yaml"]=1
      ;;
    extra)
      for f in "${ENV_FILES[@]}"; do ENV_SELECT["$f"]=1; done
      ;;
  esac

  for yaml_file in "${ENV_FILES[@]}"; do
    [ -z "${ENV_SELECT[$yaml_file]+x}" ] && continue
    env_name="${yaml_file%.yaml}"
    yaml_path="${ENVS_DIR}/${yaml_file}"

    if [ ! -f "$yaml_path" ]; then
      log_warn "$yaml_file not found, skipping $env_name"
      continue
    fi

    if [ "$DRY_RUN" = false ]; then
      if $PM env list 2>/dev/null | grep -q "^${env_name}\b"; then
        log_info "[跳过] $env_name 已存在"
      else
        log_info "Creating environment: $env_name ..."
        $PM env create -f "$yaml_path"
        log_success "$env_name created"
      fi
    else
      echo -e "  ${YELLOW}[DRY-RUN]${RESET} $PM env create -f \"$yaml_path\""
    fi
  done
fi
[ "$SKIP_ENVS" = true ] || echo ""

# ── Step 7: Install HDF5 into xdxtools-core ───────────────────────────────────
if [ "$SKIP_HDF5" = false ] && [ "$SKIP_ENVS" = false ] && [ "$HAS_CONDA" = true ]; then
  echo -e "${BOLD}Step 7: Installing HDF5 into xdxtools-core${RESET}"

  if [ "$DRY_RUN" = false ]; then
    if $PM env list 2>/dev/null | grep -q "^xdxtools-core\b"; then
      log_info "Installing hdf5 into xdxtools-core ..."
      $PM install -n xdxtools-core -c conda-forge hdf5 -y
      log_success "HDF5 installed"
    else
      log_warn "xdxtools-core environment not found, skipping HDF5 install"
      SKIP_HDF5=true
    fi
  else
    echo -e "  ${YELLOW}[DRY-RUN]${RESET} $PM install -n xdxtools-core -c conda-forge hdf5 -y"
  fi
  echo ""
elif [ "$SKIP_HDF5" = false ]; then
  echo -e "${BOLD}Step 7: HDF5 configuration${RESET}"
  log_warn "Skipped (conda environments not set up)"
  echo ""
fi

# ── Step 8: Configure HDF5 environment variables ─────────────────────────────
HDF5_ENV_PATH=""
if [ "$SKIP_HDF5" = false ] && [ "$HAS_CONDA" = true ]; then
  echo -e "${BOLD}Step 8: Configuring HDF5 environment variables${RESET}"

  if [ "$DRY_RUN" = false ]; then
    HDF5_ENV_PATH=$($PM env list 2>/dev/null \
      | grep "^xdxtools-core" | awk '{print $NF}' | head -1)

    if [ -n "$HDF5_ENV_PATH" ]; then
      cat >> "$SHELL_CONFIG" << 'HEREDOC'

# xdxtools: HDF5 configuration (required by methrix-cli)
HEREDOC
      echo "export HDF5_DIR=\"${HDF5_ENV_PATH}\"" >> "$SHELL_CONFIG"
      cat >> "$SHELL_CONFIG" << 'HEREDOC'
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:${LD_LIBRARY_PATH:-}"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
HEREDOC
      log_success "HDF5 environment variables written to $SHELL_CONFIG"

      # Load for immediate verification
      export HDF5_DIR="$HDF5_ENV_PATH"
      export LD_LIBRARY_PATH="${HDF5_DIR}/lib:${LD_LIBRARY_PATH:-}"

      METHRIX_BIN="${INSTALL_DIR}/methrix-cli"
      if [ -f "$METHRIX_BIN" ] && "$METHRIX_BIN" --version &>/dev/null; then
        ver=$("$METHRIX_BIN" --version 2>&1 | head -1)
        log_success "methrix-cli verified: $ver"
      else
        log_warn "methrix-cli could not be verified (may need to source $SHELL_CONFIG first)"
      fi
    else
      log_warn "xdxtools-core path not found, HDF5 variables not configured"
    fi
  else
    echo -e "  ${YELLOW}[DRY-RUN]${RESET} would append HDF5_DIR / LD_LIBRARY_PATH to $SHELL_CONFIG"
    echo -e "  ${YELLOW}[DRY-RUN]${RESET} would verify: methrix-cli --version"
  fi
  echo ""
fi

# ── Step 9: Reference genome download instructions ───────────────────────────
echo ""
divider
echo -e "${BOLD}参考基因组下载指引${RESET}"
divider
cat << 'EOF'

参考基因组托管于 HuggingFace：
  https://huggingface.co/datasets/Genomiclab/xdxtools-genomes

步骤：

  1. 安装 huggingface-cli
       pip install huggingface_hub

  2. 下载全部基因组到项目目录
       cd <your-project>
       huggingface-cli download Genomiclab/xdxtools-genomes \
         --local-dir ./inst/ --repo-type dataset

  3. 仅下载人类基因组（RRBS/WGBS 分析）
       huggingface-cli download Genomiclab/xdxtools-genomes \
         --include "pdx/homo_sapiens/*" \
         --local-dir ./inst/ --repo-type dataset

  或直接访问浏览器下载后解压至项目 inst/ 目录

EOF
divider

# ── Final summary ─────────────────────────────────────────────────────────────
echo ""
divider
echo -e "${GREEN}${BOLD}安装完成！${RESET}"
divider
echo ""
echo "请重新打开终端，或执行："
echo ""
echo "    source ${SHELL_CONFIG}"
echo ""
echo "快速开始："
echo ""
echo "    xdxtools init my_project"
echo "    xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv --output my_project/userspace --jobid demo_rrbs"
echo "    xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml"
echo ""
divider
echo ""
