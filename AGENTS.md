# Repository Guidelines

## Project Structure & Module Organization
- Root Go CLI: `main.go`, `cmd/`, `internal/`, `pkg/`.
- Embedded workflow/runtime assets: `inst/` (Snakemake files, rules, env YAMLs, R helper scripts).
- Test fixtures: `testdata/` (FASTQ, pdata CSVs, e2e placeholders).
- Docs: `docs/` (architecture, build/install, manual pages).
- Helper scripts: `scripts/` (`setup.sh`, `build.sh`, `release.sh`, `build-all-submodules.sh`).
- Git submodules/tools live in their own directories (for example `enva/`, `fastqc-rs/`, `methrix-cli-local/`, `qctb/`, `xenofilter-go/`, `gomats/`). Treat each as an independent module when changing code.

## Build, Test, and Development Commands
- Activate the correct toolchain first:
  - Go work: `conda activate go-env`
  - Rust/submodule work: `conda activate rust_build`
- `go build -o xdxtools .` builds the main CLI locally.
- `go test -v ./...` runs all Go unit tests.
- `go vet ./...` runs static checks used by CI.
- `./scripts/build.sh vX.Y.Z` produces release-style binaries in `dist/`.
- `./scripts/setup.sh --dry-run` previews local install/runtime initialization.
- `./scripts/build-all-submodules.sh` builds companion binaries from submodules into `$HOME/.cargo/bin`.

## Development Environment
- Default Go environment: `go-env` (Conda).
- Default Rust environment: `rust_build` (Conda).
- Example session:
  - `conda activate go-env` for `go build`, `go test`, `go vet`.
  - `conda activate rust_build` for `cargo build` in Rust submodules (for example `enva/`, `fastqc-rs/`, `methrix-cli-local/`, `qctb/`).

## Coding Style & Naming Conventions
- Go style follows standard tooling: run `gofmt` (or `go fmt ./...`) before commit.
- Keep packages focused by domain (`internal/input`, `internal/engine`, `internal/workflow`, etc.).
- Use `snake_case` for workflow/config file names and `CamelCase` only where an external format requires it.
- Keep CLI flags and user-facing options explicit and stable (`--mode`, `--engine`, `--dry-run`, `--resume`).

## Testing Guidelines
- Go tests are in `*_test.go` files (mostly under `internal/`).
- Prefer table-driven tests for parser/validator logic.
- Use fixtures from `testdata/` instead of ad-hoc local files.
- For changes affecting execution modes, test both local and SLURM code paths where possible.
- Before opening a PR: run `go test -v ./... && go vet ./...`.

## Commit & Pull Request Guidelines
- Follow Conventional Commit style seen in history: `feat: ...`, `fix: ...`, `chore: ...`, `refactor: ...`.
- Keep commits scoped (one concern per commit), with imperative summaries.
- PRs should include:
  - What changed and why.
  - Affected modules/submodules.
  - Validation commands and key output.
  - Screenshots or terminal captures for TUI/output-format changes.
