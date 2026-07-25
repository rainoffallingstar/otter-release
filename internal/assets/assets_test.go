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

func TestWorkflowCommandsQuotePathArguments(t *testing.T) {
	testCases := []struct {
		name              string
		rulePath          string
		requiredSnippets  []string
		forbiddenSnippets []string
	}{
		{
			name:     "RNA splicing",
			rulePath: filepath.Join("..", "..", "inst", "rules", "rnaseq_splicing.smk"),
			requiredSnippets: []string{
				"--root {params.run_dir:q}",
				"--pdata {input.pdata:q}",
				"--seqlengthQC {params.seqlengthQC:q}",
				"--gtf {input.gtf:q}",
				"> {output.marker:q}",
			},
			forbiddenSnippets: []string{
				"--root {params.run_dir} ",
				"--pdata {params.pdata}",
				"--gtf {params.gtf}",
			},
		},
		{
			name:     "bisulfite QC summary",
			rulePath: filepath.Join("..", "..", "inst", "rules", "bs_qc_summary.smk"),
			requiredSnippets: []string{
				"--config {input.config_file:q}",
				"--output {output.summary:q}",
			},
			forbiddenSnippets: []string{"--config {params.self_config}"},
		},
		{
			name:     "RNA QC summary",
			rulePath: filepath.Join("..", "..", "inst", "rules", "rna_qc_summary.smk"),
			requiredSnippets: []string{
				"--config {input.config_file:q}",
				"--output {output.summary:q}",
			},
			forbiddenSnippets: []string{"--config {params.self_config}"},
		},
		{
			name:     "raw FastQC",
			rulePath: filepath.Join("..", "..", "inst", "rules", "01fqcAtfirst.smk"),
			requiredSnippets: []string{
				"fqc -q {input.R1:q} -s {params.R1_dir:q} --no-html",
				"fqc -q {input.R2:q} -s {params.R2_dir:q} --no-html",
				"{sample}_R1_fqc\", \"fastqc_data.txt",
				"{sample}_R2_fqc\", \"fastqc_data.txt",
			},
			forbiddenSnippets: []string{
				"fqc -q {input.R1} ",
				"fqc -q {input.R2} ",
			},
		},
		{
			name:     "clean FastQC",
			rulePath: filepath.Join("..", "..", "inst", "rules", "03-0-fqcAtclean.smk"),
			requiredSnippets: []string{
				"fqc -q {input.R1:q} -s {params.R1_dir:q} --no-html",
				"fqc -q {input.R2:q} -s {params.R2_dir:q} --no-html",
				"{sample}_val_1_fqc\", \"fastqc_data.txt",
				"{sample}_val_2_fqc\", \"fastqc_data.txt",
			},
			forbiddenSnippets: []string{
				"fqc -q {input.R1} ",
				"fqc -q {input.R2} ",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ruleContent, err := os.ReadFile(testCase.rulePath)
			if err != nil {
				t.Fatalf("read workflow rule %s: %v", testCase.rulePath, err)
			}
			ruleText := string(ruleContent)
			for _, requiredSnippet := range testCase.requiredSnippets {
				if !strings.Contains(ruleText, requiredSnippet) {
					t.Errorf("workflow rule %s is missing safely quoted command fragment %q", testCase.rulePath, requiredSnippet)
				}
			}
			for _, forbiddenSnippet := range testCase.forbiddenSnippets {
				if strings.Contains(ruleText, forbiddenSnippet) {
					t.Errorf("workflow rule %s still contains unsafe command fragment %q", testCase.rulePath, forbiddenSnippet)
				}
			}
		})
	}
}

func TestQCWorkflowDeclaresQCTBConsumedArtifacts(t *testing.T) {
	testCases := []struct {
		name             string
		rulePath         string
		requiredSnippets []string
	}{
		{
			name:     "bisulfite summary inputs",
			rulePath: filepath.Join("..", "..", "inst", "rules", "bs_qc_summary.smk"),
			requiredSnippets: []string{
				"{sample}_val_1_bismark_bt2_PE_report.txt",
				"genome_results.txt",
				"CpG_coverage.xlsx",
				"CpG_annotation_report.xlsx",
			},
		},
		{
			name:     "methrix producer outputs",
			rulePath: filepath.Join("..", "..", "inst", "rules", "methrix_object.smk"),
			requiredSnippets: []string{
				"assays.h5",
				"methrix_data.h5",
				"CpG_coverage.xlsx",
				"CpG_annotation_report.xlsx",
				"CpG_annotation_details.tsv.gz",
			},
		},
		{
			name:             "RNA summary inputs",
			rulePath:         filepath.Join("..", "..", "inst", "rules", "rna_qc_summary.smk"),
			requiredSnippets: []string{"{sample}Log.final.out"},
		},
		{
			name:             "Bismark producer outputs",
			rulePath:         filepath.Join("..", "..", "inst", "rules", "04bsmap2sort_bismark.smk"),
			requiredSnippets: []string{"alignment_report=", "{sample}_val_1_bismark_bt2_PE_report.txt"},
		},
		{
			name:             "Qualimap producer outputs",
			rulePath:         filepath.Join("..", "..", "inst", "rules", "05-3-qualimap.smk"),
			requiredSnippets: []string{"genome_results=", "genome_results.txt"},
		},
		{
			name:             "STAR producer outputs",
			rulePath:         filepath.Join("..", "..", "inst", "rules", "rnaseq_mapping.smk"),
			requiredSnippets: []string{"star_log=", "{sample}Log.final.out"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ruleContent, err := os.ReadFile(testCase.rulePath)
			if err != nil {
				t.Fatalf("read workflow rule %s: %v", testCase.rulePath, err)
			}
			for _, snippet := range testCase.requiredSnippets {
				if !strings.Contains(string(ruleContent), snippet) {
					t.Errorf("workflow rule %s does not declare qctb artifact %q", testCase.rulePath, snippet)
				}
			}
		})
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
