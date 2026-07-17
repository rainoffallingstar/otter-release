package config

import (
	"testing"
)

func TestDeriveSuffix2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"_R1.fastq.gz", "_R2.fastq.gz"},
		{"_1.fastq.gz", "_2.fastq.gz"},
		{"sample_R1.fq", "sample_R2.fq"},
		{"_R1", "_R2"},
		{"no_suffix_here", "no_suffix_here"},
		{".fastq.gz", ".fastq.gz"},
	}

	for _, tc := range tests {
		result := deriveSuffix2(tc.input)
		if result != tc.expected {
			t.Errorf("deriveSuffix2(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestMergeReferenceFields(t *testing.T) {
	loader := NewLoader("/tmp/test.yaml")

	cfg := LoadDefaults()
	cfg.Reference.GenomeFasta = []string{"/path/to/hg19.fasta"}
	cfg.Reference.Files.Fasta = nil

	loader.mergeReferenceFields(cfg)

	if len(cfg.Reference.Files.Fasta) != 1 {
		t.Errorf("expected 1 fasta file, got %d", len(cfg.Reference.Files.Fasta))
	}
	if cfg.Reference.Files.Fasta[0] != "/path/to/hg19.fasta" {
		t.Errorf("expected /path/to/hg19.fasta, got %s", cfg.Reference.Files.Fasta[0])
	}
}

func TestMergeReferenceFields_NoOverride(t *testing.T) {
	loader := NewLoader("/tmp/test.yaml")

	cfg := LoadDefaults()
	cfg.Reference.Files.Fasta = []string{"/existing.fasta"}
	cfg.Reference.GenomeFasta = []string{"/new.fasta"}

	loader.mergeReferenceFields(cfg)

	// Should not override if Files.Fasta already has values
	if len(cfg.Reference.Files.Fasta) != 1 {
		t.Errorf("expected 1 fasta file, got %d", len(cfg.Reference.Files.Fasta))
	}
	if cfg.Reference.Files.Fasta[0] != "/existing.fasta" {
		t.Errorf("expected /existing.fasta, got %s", cfg.Reference.Files.Fasta[0])
	}
}

func TestMergeFlatCompatFields_EmptyMode(t *testing.T) {
	loader := NewLoader("/tmp/test.yaml")

	cfg := LoadDefaults()
	cfg.Workflow.Mode = ""

	// Without viper data, should not change
	loader.mergeFlatCompatFields(cfg)
	if cfg.Workflow.Mode != "" {
		t.Errorf("expected empty mode, got %s", cfg.Workflow.Mode)
	}
}

func TestMergeFlatCompatFields_ExistingMode(t *testing.T) {
	loader := NewLoader("/tmp/test.yaml")

	cfg := LoadDefaults()
	cfg.Workflow.Mode = "WGBS"

	// Should not override existing mode
	loader.mergeFlatCompatFields(cfg)
	if cfg.Workflow.Mode != "WGBS" {
		t.Errorf("expected WGBS, got %s", cfg.Workflow.Mode)
	}
}
