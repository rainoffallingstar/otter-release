package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

func init() {
	logger.Init(false)
}

func setupTestFS(t *testing.T) {
	t.Helper()
	EmbeddedAssets = fstest.MapFS{
		"inst/Rscripts/test.R": &fstest.MapFile{
			Data: []byte("# Test R script\nprint('hello')\n"),
		},
		"inst/snakefiles/BeaverBS_step1.snakemake": &fstest.MapFile{
			Data: []byte("rule step1:\n  shell: 'echo hello'\n"),
		},
		"inst/rules/01fqc.smk": &fstest.MapFile{
			Data: []byte("rule fqc:\n  shell: 'fqc {input}'\n"),
		},
		"inst/envs/xdxtools-snakemake.yaml": &fstest.MapFile{
			Data: []byte("name: xdxtools-snakemake\n"),
		},
		"inst/data/gene_mapping.csv": &fstest.MapFile{
			Data: []byte("ensembl_id,symbol\nENSG001,BRCA1\n"),
		},
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	copier := NewAssetCopier(tmpDir, "rootless")

	err := copier.CreateDirectoryStructure()
	if err != nil {
		t.Fatalf("CreateDirectoryStructure failed: %v", err)
	}

	expectedDirs := []string{
		"config", "data", "envs", "inst", "R", "rules",
		"saveRDS", "temp", "userspace", "workflows", "www",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if info, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		} else if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}
}

func TestCopyDir_FileContent(t *testing.T) {
	setupTestFS(t)
	tmpDir := t.TempDir()
	copier := NewAssetCopier(tmpDir, "rootless")

	// Copy Rscripts
	err := copier.copyDir("inst/Rscripts", "R")
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify file was copied
	destPath := filepath.Join(tmpDir, "R", "test.R")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist after copy", destPath)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	expected := "# Test R script\nprint('hello')\n"
	if string(content) != expected {
		t.Errorf("file content mismatch.\nExpected: %q\nGot: %q", expected, string(content))
	}
}

func TestCopyDir_CreatesParentDirs(t *testing.T) {
	setupTestFS(t)
	tmpDir := t.TempDir()
	copier := NewAssetCopier(tmpDir, "rootless")

	// Copy into a nested subdirectory that doesn't exist
	err := copier.copyDir("inst/rules", "rules")
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify nested file exists
	destPath := filepath.Join(tmpDir, "rules", "01fqc.smk")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist after copy", destPath)
	}
}

func TestCopyAll(t *testing.T) {
	setupTestFS(t)
	tmpDir := t.TempDir()
	copier := NewAssetCopier(tmpDir, "rootless")

	err := copier.CopyAll()
	if err != nil {
		t.Fatalf("CopyAll failed: %v", err)
	}

	// Verify key files were copied
	checks := []struct {
		path string
		desc string
	}{
		{filepath.Join(tmpDir, "R", "test.R"), "R script"},
		{filepath.Join(tmpDir, "BeaverBS_step1.snakemake"), "snakefile"},
		{filepath.Join(tmpDir, "rules", "01fqc.smk"), "rule"},
		{filepath.Join(tmpDir, "envs", "xdxtools-snakemake.yaml"), "env"},
		{filepath.Join(tmpDir, "data", "gene_mapping.csv"), "data"},
	}

	for _, check := range checks {
		if _, err := os.Stat(check.path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist (%s)", check.path, check.desc)
		}
	}
}

func TestListEmbeddedFiles(t *testing.T) {
	setupTestFS(t)

	files, err := ListEmbeddedFiles()
	if err != nil {
		t.Fatalf("ListEmbeddedFiles failed: %v", err)
	}

	if len(files) != 5 {
		t.Errorf("expected 5 embedded files, got %d: %v", len(files), files)
	}
}

func TestSetEmbeddedAssets(t *testing.T) {
	testFS := fstest.MapFS{
		"custom/path/file.txt": &fstest.MapFile{
			Data: []byte("custom content"),
		},
	}

	SetEmbeddedAssets(testFS)

	err := fs.WalkDir(EmbeddedAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
}

func TestNewAssetCopier_Defaults(t *testing.T) {
	// Test with empty rulesType (should default to "rootless")
	copier := NewAssetCopier("/tmp/project", "")
	if copier.RulesType != "rootless" {
		t.Errorf("expected RulesType 'rootless', got '%s'", copier.RulesType)
	}

	// Test with explicit legacy type
	copier2 := NewAssetCopier("/tmp/project", "legacy")
	if copier2.RulesType != "legacy" {
		t.Errorf("expected RulesType 'legacy', got '%s'", copier2.RulesType)
	}
}

func TestXenofilterSuccessMarkersRequireSuccessfulCommands(t *testing.T) {
	rulePaths := []string{
		filepath.Join("..", "..", "inst", "rules", "XenofilteR.smk"),
		filepath.Join("..", "..", "inst", "rules_legacy", "XenofilteR.smk"),
	}
	for _, rulePath := range rulePaths {
		ruleContent, err := os.ReadFile(rulePath)
		if err != nil {
			t.Fatalf("read xenofilter rule %s: %v", rulePath, err)
		}
		if !strings.Contains(string(ruleContent), "&& touch {params.filter_root}/Filtered_bams/filtered_success.txt") {
			t.Fatalf("xenofilter rule %s can create its success marker without an explicit successful-command guard", rulePath)
		}
	}
}

func TestCopyAll_EmptyEmbeddedAssets(t *testing.T) {
	// Reset to empty
	EmbeddedAssets = fstest.MapFS{}
	tmpDir := t.TempDir()
	copier := NewAssetCopier(tmpDir, "rootless")

	// CopyAll should fail because embedded assets are empty
	err := copier.CopyAll()
	if err == nil {
		t.Error("expected error with empty embedded assets")
	}
}
