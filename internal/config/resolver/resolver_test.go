package resolver

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	runstate "github.com/rainoffallingstar/otter/internal/run"
)

const fixtureDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"

func TestResolveIsDeterministicAndTracksOverrides(t *testing.T) {
	projectRoot := t.TempDir()
	referenceRoot := filepath.Join(projectRoot, "reference-registry")
	writeProjectFixture(t, projectRoot)
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	writeReferenceFixture(t, referenceRoot, "hg19", "GRCh37.p13", true)
	runDirectory := runstate.Directory{
		ID:        "run-20260726T013245Z-abcdef",
		CreatedAt: time.Date(2026, 7, 26, 1, 32, 45, 0, time.UTC),
		Root:      filepath.Join(projectRoot, "runs", "run-20260726T013245Z-abcdef"),
	}
	options := Options{
		ProjectPath:      filepath.Join(projectRoot, "project.yaml"),
		RunDirectory:     runDirectory,
		ReferenceRoot:    referenceRoot,
		BackendOverride:  configv1.BackendSlurm,
		ExecutorOverride: configv1.ExecutorSnakemake,
		ReferenceOverride: map[configv1.ReferenceRole]configv1.ReferenceSelection{
			configv1.ReferenceRolePrimary: "hg19@GRCh37.p13",
		},
	}
	first, err := (Resolver{}).Resolve(options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := (Resolver{}).Resolve(options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digests != second.Digests || first.Run != second.Run {
		t.Fatal("same inputs did not resolve deterministically")
	}
	if first.Execution.Executor.Source != configv1.SourceCLI || first.Execution.Backend.Source != configv1.SourceCLI {
		t.Fatalf("CLI source tracking failed: %+v", first.Execution)
	}
	if !first.References.Override || first.References.EffectiveSelection.Primary != "hg19@GRCh37.p13" {
		t.Fatalf("reference override was not captured: %+v", first.References)
	}
	if first.References.ProjectSelection.Primary != "hg38@GRCh38.p14" {
		t.Fatalf("project reference selection was mutated: %+v", first.References.ProjectSelection)
	}
}

func TestResolveRequiresExplicitBackendBeforeGateThree(t *testing.T) {
	projectRoot := t.TempDir()
	referenceRoot := filepath.Join(projectRoot, "reference-registry")
	writeProjectFixture(t, projectRoot)
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	runDirectory := runstate.Directory{ID: "run-20260726T013245Z-abcdef", CreatedAt: time.Now().UTC(), Root: filepath.Join(projectRoot, "runs", "run-20260726T013245Z-abcdef")}
	_, err := (Resolver{}).Resolve(Options{ProjectPath: filepath.Join(projectRoot, "project.yaml"), RunDirectory: runDirectory, ReferenceRoot: referenceRoot})
	if err == nil {
		t.Fatal("expected unresolved backend auto to fail closed")
	}
}

func writeProjectFixture(t *testing.T, projectRoot string) {
	t.Helper()
	project := `schema_version: otter.project/v1
project:
  id: cohort-a
workflow:
  scenario: rrbs
execution:
  executor: craftmake
  backend: auto
  site: auto
samples:
  manifest: samples.tsv
references:
  primary: hg38@GRCh38.p14
observability:
  metrics: true
  retain_logs: true
`
	writeFile(t, filepath.Join(projectRoot, "project.yaml"), project)
	writeFile(t, filepath.Join(projectRoot, "samples.tsv"), "sample_id\tr1\tr2\nS01\tdata/S01_R1.fastq.gz\tdata/S01_R2.fastq.gz\n")
	lock := `schema_version: otter.references.lock/v1
references:
  primary:
    id: hg38
    release: GRCh38.p14
    manifest_digest: ` + fixtureDigest + "\n"
	writeFile(t, filepath.Join(projectRoot, "references.lock.yaml"), lock)
	if err := os.MkdirAll(filepath.Join(projectRoot, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(projectRoot, "workflows", "rrbs.yaml"), "name: rrbs\n")
}

func writeReferenceFixture(t *testing.T, referenceRoot string, id string, release string, includeManifest bool) {
	t.Helper()
	releaseRoot := filepath.Join(referenceRoot, "genomes", id, release)
	definition := `schema_version: otter.reference/v1
reference:
  id: ` + id + `
  release: ` + release + `
  organism: Homo sapiens
  assembly: test
assets:
  fasta:
    path: fasta/genome.fa.gz
    sha256: ` + fixtureDigest + `
    size_bytes: 1
    fai: fasta/genome.fa.gz.fai
  indexes:
    - type: bismark
      path: indexes/bismark
      reference_fasta_sha256: ` + fixtureDigest + `
      tool: bismark
      tool_version: test
      manifest_sha256: ` + fixtureDigest + `
compatibility:
  scenarios: [rrbs]
`
	writeFile(t, filepath.Join(releaseRoot, "reference.yaml"), definition)
	if includeManifest {
		writeFile(t, filepath.Join(releaseRoot, "manifest.json"), "{}\n")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
