package reference

import (
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
)

const fixtureDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"

func TestResolveMatchingLockSucceeds(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p14", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	resolved, err := resolver.Resolve(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioRRBS)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ID != "hg38" || resolved.Release != "GRCh38.p14" || resolved.Role != configv1.ReferenceRolePrimary {
		t.Fatalf("unexpected resolved reference: %+v", resolved)
	}
	if resolved.Fasta.SHA256 != fixtureDigest {
		t.Fatalf("unexpected fasta digest: %q", resolved.Fasta.SHA256)
	}
	if resolved.ManifestDigest != fixtureDigest {
		t.Fatalf("unexpected manifest digest: %q", resolved.ManifestDigest)
	}
	if resolved.RegistryRoot == "" {
		t.Fatal("resolved reference must include registry root")
	}
}

func TestResolveRejectsLockIDMismatch(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "mm10", Release: "GRCm39", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.Resolve(configv1.ReferenceRolePrimary, "mm10@GRCm39", configv1.ScenarioRRBS)
	if err == nil {
		t.Fatal("expected lock mismatch to fail when lock has mm10 but selection is hg38")
	}
}

func TestResolveRejectsLockReleaseMismatch(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p13", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.Resolve(configv1.ReferenceRolePrimary, "hg38@GRCh38.p13", configv1.ScenarioRRBS)
	if err == nil {
		t.Fatal("expected lock release mismatch to fail")
	}
}

func TestResolveRejectsLockMissingRole(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References:    map[string]configv1.LockedReference{},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.Resolve(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioRRBS)
	if err == nil {
		t.Fatal("expected missing lock entry to fail")
	}
}

func TestResolveRejectsScenarioIncompatibility(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p14", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.Resolve(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioRNASeq)
	if err == nil {
		t.Fatal("expected scenario incompatibility to fail (fixture only supports rrbs)")
	}
}

func TestResolveRejectsReferenceDefinitionIdentityMismatch(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p14", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.Resolve(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioBSPDX)
	if err == nil {
		t.Fatal("expected reference definition scenario mismatch for bs-pdx to fail")
	}
}

func TestResolveOverrideComputesFreshManifestDigest(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", true)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p14", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	resolved, err := resolver.ResolveOverride(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioRRBS)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ManifestDigest == fixtureDigest {
		t.Fatalf("ResolveOverride should compute a fresh manifest digest from the actual manifest.json file, but got the lock fixture digest %q", resolved.ManifestDigest)
	}
	if len(resolved.ManifestDigest) != len(fixtureDigest) || resolved.ManifestDigest[:7] != "sha256:" {
		t.Fatalf("unexpected manifest digest format: %q", resolved.ManifestDigest)
	}
}

func TestResolveOverrideRejectsLockMismatch(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", true)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "mm10", Release: "GRCm39", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.ResolveOverride(configv1.ReferenceRolePrimary, "mm10@GRCm39", configv1.ScenarioRRBS)
	if err == nil {
		t.Fatal("expected ResolveOverride with lock mismatch to fail")
	}
}

func TestResolveOverrideRejectsMissingManifest(t *testing.T) {
	referenceRoot := t.TempDir()
	writeReferenceFixture(t, referenceRoot, "hg38", "GRCh38.p14", false)
	lock := configv1.ReferencesLock{
		SchemaVersion: configv1.ReferencesLockSchemaVersion,
		References: map[string]configv1.LockedReference{
			"primary": {ID: "hg38", Release: "GRCh38.p14", ManifestDigest: fixtureDigest},
		},
	}
	resolver := Resolver{RegistryRoot: referenceRoot, Lock: lock}
	_, err := resolver.ResolveOverride(configv1.ReferenceRolePrimary, "hg38@GRCh38.p14", configv1.ScenarioRRBS)
	if err == nil {
		t.Fatal("expected ResolveOverride to fail when manifest.json is missing")
	}
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
