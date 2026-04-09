# xdxtools build and test matrix

## Root repo

### `xdxtools`

- Toolchain: `conda activate go-env`
- Build: `go build -o xdxtools .`
- Test: `go test -v ./...`
- Static checks: `go vet ./...`
- Release-style build: `./scripts/build.sh vX.Y.Z`

## Go submodules

### `Paireads`

- Toolchain: `conda activate go-env`
- Build: `go build -o paireads ./cmd/paireads`
- Test: `go test ./...`

### `bamdriver-go`

- Toolchain: `conda activate go-env`
- Build: prefer consumer-driven validation; if needed use `go test ./...`
- Test: `go test ./...`

### `gomats`

- Toolchain: `conda activate go-env`
- Build: `go build -o gomats ./cmd/gomats`
- Test: `go test ./...`

### `htseq2matrix-go`

- Toolchain: `conda activate go-env`
- Build: `go build -o htseq2matrix cmd/htseq2matrix/main.go`
- Test: `go test ./...`
- Notes: Gene mapping assets may be embedded; keep data-path assumptions aligned with the repo README.

### `xenofilter-go`

- Toolchain: `conda activate go-env`
- Build: `go build -o xenofilter ./cmd/xenofilter`
- Test: `go test ./...`

## Rust submodules

### `enva`

- Toolchain: `conda activate rust_build`
- Build: `cargo build --release`
- Test: `cargo test`

### `fastqc-rs`

- Toolchain: `conda activate rust_build`
- Build: `cargo build --release`
- Test: `cargo test`

### `methrix-cli`

- Toolchain: `conda activate rust_build`
- Build: `cargo build --release`
- Test: `cargo test`
- Notes: HDF5 development libraries may be required for a full build.

### `qctb`

- Toolchain: `conda activate rust_build`
- Build: `cargo build --release`
- Test: `cargo test`

## Validation strategy

- Prefer repo-local formatter and tests first.
- For root `xdxtools` changes, run `go test -v ./...` and `go vet ./...` when the change can affect shared CLI or config behavior.
- For shared library changes in `bamdriver-go`, validate at least one downstream consumer when feasible.
- If a command cannot run because a system dependency is missing, report the missing dependency instead of guessing success.
