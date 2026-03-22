package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xdxtools/xdxtools-go/internal/config"
)

func TestDiscoverProjectDirFromConfigUsesAncestorManifest(t *testing.T) {
	projectDir := t.TempDir()
	manifestPath := filepath.Join(projectDir, ".xdxtools", "assets.manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	configDir := filepath.Join(projectDir, "userspace", "demo_rrbs", "config")
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("workflow: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got := discoverProjectDirFromConfig(configPath)
	if got != projectDir {
		t.Fatalf("discoverProjectDirFromConfig() = %q, want %q", got, projectDir)
	}
}

func TestDiscoverProjectDirFromConfigFallsBackToConfigDir(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "standalone", "config")
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("workflow: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got := discoverProjectDirFromConfig(configPath)
	if got != configDir {
		t.Fatalf("discoverProjectDirFromConfig() = %q, want %q", got, configDir)
	}
}

func TestResolveRunPathsWithExplicitProjectDir(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("workflow: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	projectDir := filepath.Join(t.TempDir(), "project-root")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	gotConfig, gotProject, err := resolveRunPaths(configPath, projectDir)
	if err != nil {
		t.Fatalf("resolveRunPaths() error = %v", err)
	}
	if gotConfig != configPath {
		t.Fatalf("resolveRunPaths() config = %q, want %q", gotConfig, configPath)
	}
	if gotProject != projectDir {
		t.Fatalf("resolveRunPaths() project = %q, want %q", gotProject, projectDir)
	}
}

func TestShouldRelaxLocalDryRunResourceValidation(t *testing.T) {
	tests := []struct {
		name          string
		stepResources map[int]*config.StepResource
		dryRun        bool
		want          bool
	}{
		{
			name:          "dry run with defaults",
			stepResources: map[int]*config.StepResource{},
			dryRun:        true,
			want:          true,
		},
		{
			name: "dry run with zero-value override",
			stepResources: map[int]*config.StepResource{
				1: {},
			},
			dryRun: true,
			want:   true,
		},
		{
			name: "dry run with explicit core override",
			stepResources: map[int]*config.StepResource{
				1: {Cores: 4},
			},
			dryRun: true,
			want:   false,
		},
		{
			name: "dry run with explicit memory override",
			stepResources: map[int]*config.StepResource{
				1: {Memory: "1G"},
			},
			dryRun: true,
			want:   false,
		},
		{
			name:          "real run never relaxes",
			stepResources: map[int]*config.StepResource{},
			dryRun:        false,
			want:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldRelaxLocalDryRunResourceValidation(1, tc.stepResources, tc.dryRun)
			if got != tc.want {
				t.Fatalf("shouldRelaxLocalDryRunResourceValidation() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateCreateOutputDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "directory path", path: "my_project/userspace", wantErr: false},
		{name: "yaml path", path: "config/config.yaml", wantErr: true},
		{name: "json path", path: "out/config.json", wantErr: true},
		{name: "empty path", path: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateOutputDir(tc.path)
			if tc.wantErr && err == nil {
				t.Fatalf("validateCreateOutputDir(%q) error = nil, want error", tc.path)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateCreateOutputDir(%q) error = %v, want nil", tc.path, err)
			}
		})
	}
}
