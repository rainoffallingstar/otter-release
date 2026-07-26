package reference

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
)

type Resolver struct {
	RegistryRoot string
	Lock         configv1.ReferencesLock
}

func LoadLock(path string) (configv1.ReferencesLock, error) {
	return configv1.LoadReferencesLock(path)
}

func (resolver Resolver) Resolve(role configv1.ReferenceRole, selection configv1.ReferenceSelection, scenario configv1.Scenario) (configv1.ResolvedReference, error) {
	lockedReference, exists := resolver.Lock.References[string(role)]
	if !exists {
		return configv1.ResolvedReference{}, fmt.Errorf("references lock has no %q entry", role)
	}
	id, release, err := configv1.ParseReferenceSelection(selection)
	if err != nil {
		return configv1.ResolvedReference{}, err
	}
	if lockedReference.ID != id || lockedReference.Release != release {
		return configv1.ResolvedReference{}, fmt.Errorf(
			"reference %q selection %s does not match lock %s@%s",
			role,
			selection,
			lockedReference.ID,
			lockedReference.Release,
		)
	}
	return resolver.resolve(role, selection, scenario, lockedReference.ManifestDigest)
}

func (resolver Resolver) ResolveOverride(role configv1.ReferenceRole, selection configv1.ReferenceSelection, scenario configv1.Scenario) (configv1.ResolvedReference, error) {
	id, release, err := configv1.ParseReferenceSelection(selection)
	if err != nil {
		return configv1.ResolvedReference{}, err
	}
	manifestPath := filepath.Join(resolver.RegistryRoot, "genomes", id, release, "manifest.json")
	manifestDigest, err := digestFile(manifestPath)
	if err != nil {
		return configv1.ResolvedReference{}, fmt.Errorf("resolve override manifest: %w", err)
	}
	return resolver.resolve(role, selection, scenario, manifestDigest)
}

func (resolver Resolver) resolve(role configv1.ReferenceRole, selection configv1.ReferenceSelection, scenario configv1.Scenario, manifestDigest string) (configv1.ResolvedReference, error) {
	if !filepath.IsAbs(resolver.RegistryRoot) {
		return configv1.ResolvedReference{}, fmt.Errorf("reference registry root must be absolute")
	}
	id, release, err := configv1.ParseReferenceSelection(selection)
	if err != nil {
		return configv1.ResolvedReference{}, err
	}
	releaseRoot := filepath.Join(resolver.RegistryRoot, "genomes", id, release)
	definitionPath := filepath.Join(releaseRoot, "reference.yaml")
	definition, err := configv1.LoadReferenceDefinition(definitionPath)
	if err != nil {
		return configv1.ResolvedReference{}, err
	}
	if definition.Reference.ID != id || definition.Reference.Release != release {
		return configv1.ResolvedReference{}, fmt.Errorf("reference definition identity does not match %s", selection)
	}
	if !supportsScenario(definition.Compatibility.Scenarios, scenario) {
		return configv1.ResolvedReference{}, fmt.Errorf("reference %s does not support scenario %s", selection, scenario)
	}

	resolved := configv1.ResolvedReference{
		Role:           role,
		ID:             id,
		Release:        release,
		RegistryRoot:   releaseRoot,
		ManifestDigest: manifestDigest,
		Fasta: configv1.ResolvedAsset{
			Type:   "fasta",
			Path:   filepath.Join(releaseRoot, definition.Assets.Fasta.Path),
			SHA256: definition.Assets.Fasta.SHA256,
		},
	}
	for _, annotation := range definition.Assets.Annotations {
		resolved.Annotations = append(resolved.Annotations, configv1.ResolvedAsset{
			Type:   annotation.Type,
			Path:   filepath.Join(releaseRoot, annotation.Path),
			SHA256: annotation.SHA256,
		})
	}
	for _, index := range definition.Assets.Indexes {
		resolved.Indexes = append(resolved.Indexes, configv1.ResolvedAsset{
			Type:   index.Type,
			Path:   filepath.Join(releaseRoot, index.Path),
			SHA256: index.ManifestSHA256,
		})
	}
	return resolved, nil
}

func digestFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func supportsScenario(scenarios []configv1.Scenario, expected configv1.Scenario) bool {
	for _, scenario := range scenarios {
		if scenario == expected {
			return true
		}
	}
	return false
}
