# Full-featured xdxtools container
# Includes:
# - All submodule tools (enva, fqc, xenofilter, paireads, htseq2matrix, methrix-cli, qctb, gomats)
# - Main xdxtools binary
# - Three runtime conda environments: xdxtools-core, xdxtools-snakemake, xdxtools-extra
# - Build helper environments: go-env, rust_build
#
# Notes:
# - Runtime environment initialization uses enva (rattler-first).
# - Environment and HDF5 runtime variables are configured via ENV.

FROM continuumio/miniconda3:latest

SHELL ["/bin/bash", "-lc"]

ENV DEBIAN_FRONTEND=noninteractive
ENV XDXTOOLS_HOME=/opt/xdxtools
ENV XDXTOOLS_BIN=/opt/xdxtools/bin
ENV MAMBA_ROOT_PREFIX=/opt/mamba

# Deterministic prefixes for the 3 runtime envs created by enva.
ENV XDXTOOLS_CORE_PREFIX=${MAMBA_ROOT_PREFIX}/envs/xdxtools-core
ENV XDXTOOLS_SNAKEMAKE_PREFIX=${MAMBA_ROOT_PREFIX}/envs/xdxtools-snakemake
ENV XDXTOOLS_EXTRA_PREFIX=${MAMBA_ROOT_PREFIX}/envs/xdxtools-extra

# HDF5 runtime configuration (for methrix-cli and related tooling).
ENV HDF5_DIR=${XDXTOOLS_CORE_PREFIX}
ENV HDF5_INCLUDE_DIR=${HDF5_DIR}/include
ENV HDF5_LIB_DIR=${HDF5_DIR}/lib
ENV LD_LIBRARY_PATH=${HDF5_LIB_DIR}:${LD_LIBRARY_PATH}
ENV PKG_CONFIG_PATH=${HDF5_LIB_DIR}/pkgconfig:${PKG_CONFIG_PATH}

# Global PATH with all runtime env bins available out-of-box.
ENV PATH=/opt/xdxtools/bin:/root/.cargo/bin:/opt/conda/bin:${XDXTOOLS_CORE_PREFIX}/bin:${XDXTOOLS_SNAKEMAKE_PREFIX}/bin:${XDXTOOLS_EXTRA_PREFIX}/bin:${PATH}

# System packages required for building Go/Rust submodule tools.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       bash \
       ca-certificates \
       clang \
       curl \
       git \
       pkg-config \
       build-essential \
       libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR ${XDXTOOLS_HOME}
COPY . ${XDXTOOLS_HOME}

# Configure conda channels.
RUN conda config --system --set channel_priority strict \
    && conda config --system --add channels conda-forge \
    && conda config --system --add channels bioconda

# Toolchain environments used by existing build scripts.
RUN conda create -y -n go-env go \
    && conda create -y -n rust_build rust hdf5

# Prepare deterministic mamba root directory.
RUN mkdir -p ${MAMBA_ROOT_PREFIX} ${XDXTOOLS_BIN}

# Build bootstrap enva binary first, then use enva for the 3 runtime envs.
RUN conda run -n rust_build cargo build --manifest-path enva/Cargo.toml --release \
    && cp enva/target/release/enva ${XDXTOOLS_BIN}/enva

# Three primary runtime environments (out-of-box), created via enva.
RUN ${XDXTOOLS_BIN}/enva create --yaml inst/envs/xdxtools-core.yaml --name xdxtools-core --force --clean-cache \
    && ${XDXTOOLS_BIN}/enva create --yaml inst/envs/xdxtools-snakemake.yaml --name xdxtools-snakemake --force \
    && ${XDXTOOLS_BIN}/enva create --yaml inst/envs/xdxtools-extra.yaml --name xdxtools-extra --force

# Build main binary + all submodule tools using project-provided script.
RUN conda run -n go-env go build -o ${XDXTOOLS_BIN}/xdxtools . \
    && STRICT_MODE=1 bash scripts/build-all-submodules.sh \
    && cp /root/.cargo/bin/enva ${XDXTOOLS_BIN}/enva \
    && cp /root/.cargo/bin/fqc ${XDXTOOLS_BIN}/fqc \
    && cp /root/.cargo/bin/xenofilter ${XDXTOOLS_BIN}/xenofilter \
    && cp /root/.cargo/bin/paireads ${XDXTOOLS_BIN}/paireads \
    && cp /root/.cargo/bin/htseq2matrix ${XDXTOOLS_BIN}/htseq2matrix \
    && cp /root/.cargo/bin/methrix-cli ${XDXTOOLS_BIN}/methrix-cli \
    && cp /root/.cargo/bin/qctb ${XDXTOOLS_BIN}/qctb \
    && cp /root/.cargo/bin/gomats ${XDXTOOLS_BIN}/gomats

# Keep image ready for immediate interactive use.
WORKDIR /workspace
CMD ["bash"]
