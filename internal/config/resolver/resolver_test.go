package resolver

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	runstate "github.com/rainoffallingstar/otter/internal/run"
	"github.com/rainoffallingstar/otter/internal/site"
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
		ProjectPath:     filepath.Join(projectRoot, "project.yaml"),
		RunDirectory:    runDirectory,
		ReferenceRoot:   referenceRoot,
		BackendOverride: configv1.BackendLocal,
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

func TestResolveAutoDetectsLocalBackend(t *testing.T) {
	projectRoot := t.TempDir()
	referenceRoot := filepath.Join(projectRoot, "reference-registry")
	writeProjectFixture(t, projectRoot)
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	runDirectory := runstate.Directory{
		ID:        "run-20260726T013245Z-abcdef",
		CreatedAt: time.Date(2026, 7, 26, 1, 32, 45, 0, time.UTC),
		Root:      filepath.Join(projectRoot, "runs", "run-20260726T013245Z-abcdef"),
	}
	mockDetector := site.NewDetector(
		site.WithToolChecker(func(tool string) bool { return false }),
	)
	options := Options{
		ProjectPath:   filepath.Join(projectRoot, "project.yaml"),
		RunDirectory:  runDirectory,
		ReferenceRoot: referenceRoot,
		Detector:      mockDetector,
	}
	snapshot, err := (Resolver{}).Resolve(options)
	if err != nil {
		t.Fatalf("auto-detection should succeed with no SLURM tools: %v", err)
	}
	if snapshot.Execution.Backend.Value != configv1.BackendLocal {
		t.Fatalf("expected backend local from auto-detection, got %q", snapshot.Execution.Backend.Value)
	}
	if snapshot.Execution.Backend.Source != configv1.SourceDetection {
		t.Fatalf("expected source detection, got %q", snapshot.Execution.Backend.Source)
	}
}

func TestResolveAutoDetectionFailsOnPartialSlurm(t *testing.T) {
	projectRoot := t.TempDir()
	referenceRoot := filepath.Join(projectRoot, "reference-registry")
	writeProjectFixture(t, projectRoot)
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	runDirectory := runstate.Directory{
		ID:        "run-20260726T013245Z-abcdef",
		CreatedAt: time.Now().UTC(),
		Root:      filepath.Join(projectRoot, "runs", "run-20260726T013245Z-abcdef"),
	}
	mockDetector := site.NewDetector(
		site.WithToolChecker(func(tool string) bool { return tool == "sbatch" || tool == "squeue" }),
	)
	options := Options{
		ProjectPath:   filepath.Join(projectRoot, "project.yaml"),
		RunDirectory:  runDirectory,
		ReferenceRoot: referenceRoot,
		Detector:      mockDetector,
	}
	_, err := (Resolver{}).Resolve(options)
	if err == nil {
		t.Fatal("expected partial SLURM toolchain to fail closed")
	}
}

func TestResolveAutoBackendWithNamedSiteProfile(t *testing.T) {
	projectRoot := t.TempDir()
	referenceRoot := filepath.Join(projectRoot, "reference-registry")

	writeProjectFixture(t, projectRoot)
	sitePath := filepath.Join(projectRoot, "sites", "hpc-cluster-1.yaml")
	siteYAML := "schema_version: otter.site/v1\nsite:\n  id: hpc-cluster-1\n  backend: slurm\nslurm:\n  partition: cpu\n  account: genomics\n  qos: normal\n  default_time: 24:00:00\npaths:\n  reference_root: /shared/reference\n"
	writeFile(t, sitePath, siteYAML)
	t.Setenv("OTTER_SITE_PROFILE", sitePath)

	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	runDirectory := runstate.Directory{
		ID:        "run-20260726T013245Z-abcdef",
		CreatedAt: time.Date(2026, 7, 26, 1, 32, 45, 0, time.UTC),
		Root:      filepath.Join(projectRoot, "runs", "run-20260726T013245Z-abcdef"),
	}
	mockDetector := site.NewDetector(
		site.WithToolChecker(func(tool string) bool {
			return tool == "sbatch" || tool == "squeue" || tool == "sacct" || tool == "scancel"
		}),
		site.WithClusterNameGetter(func() (string, bool) {
			return "test-cluster", true
		}),
		site.WithSlurmValidator(func(profile *site.SiteProfile) error { return nil }),
	)
	options := Options{
		ProjectPath:   filepath.Join(projectRoot, "project.yaml"),
		RunDirectory:  runDirectory,
		ReferenceRoot: referenceRoot,
		SiteOverride:  "hpc-cluster-1",
		Detector:      mockDetector,
	}
	snapshot, err := (Resolver{}).Resolve(options)
	if err != nil {
		t.Fatalf("named site profile resolve should succeed with mocked SLURM: %v", err)
	}
	if snapshot.Execution.Backend.Value != configv1.BackendSlurm {
		t.Fatalf("expected backend slurm from named site profile, got %q", snapshot.Execution.Backend.Value)
	}
	if snapshot.Execution.Site.Value != "hpc-cluster-1" {
		t.Fatalf("expected site hpc-cluster-1, got %q", snapshot.Execution.Site.Value)
	}
}

func writeProjectFixture(t *testing.T, projectRoot string) {
	t.Helper()
	content := "schema_version: otter.project/v1\nproject:\n  id: cohort-a\nworkflow:\n  scenario: rrbs\nexecution:\n  executor: craftmake\n  backend: auto\n  site: auto\nsamples:\n  manifest: samples.tsv\nreferences:\n  primary: hg38@GRCh38.p14\nobservability:\n  metrics: true\n  retain_logs: true\n"
	writeFile(t, filepath.Join(projectRoot, "project.yaml"), content)
	writeFile(t, filepath.Join(projectRoot, "samples.tsv"), "sample_id\tr1\tr2\nS01\tdata/S01_R1.fastq.gz\tdata/S01_R2.fastq.gz\n")
	lock := "schema_version: otter.references.lock/v1\nreferences:\n  primary:\n    id: hg38\n    release: GRCh38.p14\n    manifest_digest: " + fixtureDigest + "\n"
	writeFile(t, filepath.Join(projectRoot, "references.lock.yaml"), lock)
	if err := os.MkdirAll(filepath.Join(projectRoot, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(projectRoot, "workflows", "rrbs.yaml"), "name: rrbs\n")
}

func writeReferenceFixture(t *testing.T, referenceRoot string, id string, release string, includeManifest bool) {
	t.Helper()
	releaseRoot := filepath.Join(referenceRoot, "genomes", id, release)
	definition := "schema_version: otter.reference/v1\nreference:\n  id: " + id + "\n  release: " + release + "\n  organism: Homo sapiens\n  assembly: test\nassets:\n  fasta:\n    path: fasta/genome.fa.gz\n    sha256: " + fixtureDigest + "\n    size_bytes: 1\n    fai: fasta/genome.fa.gz.fai\n  indexes:\n    - type: bismark\n      path: indexes/bismark\n      reference_fasta_sha256: " + fixtureDigest + "\n      tool: bismark\n      tool_version: test\n      manifest_sha256: " + fixtureDigest + "\ncompatibility:\n  scenarios: [rrbs]\n"
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
