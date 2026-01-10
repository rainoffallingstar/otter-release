package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateReferencePaths(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		species1  string
		species2  string
		wantError bool
		checkFunc func(fasta, index, gtf, ref []string) bool
	}{
		{
			name:     "单物种RRBS模式",
			mode:     "RRBS",
			species1: "human",
			species2: "",
			checkFunc: func(fasta, index, gtf, ref []string) bool {
				return len(fasta) == 1 &&
					len(index) == 1 &&
					len(gtf) == 0 &&
					len(ref) == 0 &&
					strings.Contains(fasta[0], "human.fasta") &&
					strings.Contains(index[0], "homo_sapiens")
			},
		},
		{
			name:     "PDX RRBS模式",
			mode:     "RRBS",
			species1: "human",
			species2: "mouse",
			checkFunc: func(fasta, index, gtf, ref []string) bool {
				return len(fasta) == 2 &&
					len(index) == 2 &&
					len(gtf) == 0 &&
					len(ref) == 0 &&
					strings.Contains(fasta[0], "human.fasta") &&
					strings.Contains(fasta[1], "mouse.fasta")
			},
		},
		{
			name:     "单物种RNAseq模式",
			mode:     "RNASEQ",
			species1: "human",
			species2: "",
			checkFunc: func(fasta, index, gtf, ref []string) bool {
				return len(fasta) == 1 &&
					len(index) == 1 &&
					len(gtf) == 1 &&
					len(ref) == 1 &&
					strings.Contains(gtf[0], "human.ensGene_sorted.gtf") &&
					strings.Contains(ref[0], "homo_sapiens")
			},
		},
		{
			name:     "PDX RNAseq模式",
			mode:     "RNASEQ",
			species1: "human",
			species2: "mouse",
			checkFunc: func(fasta, index, gtf, ref []string) bool {
				return len(fasta) == 2 &&
					len(index) == 2 &&
					len(gtf) == 2 &&
					len(ref) == 2 &&
					strings.Contains(gtf[0], "human.ensGene_sorted.gtf") &&
					strings.Contains(gtf[1], "mouse.ensGene_sorted.gtf")
			},
		},
		{
			name:      "不支持的物种",
			mode:      "RRBS",
			species1:  "unsupported_species",
			species2:  "",
			wantError: true,
			checkFunc: func(fasta, index, gtf, ref []string) bool { return true },
		},
		{
			name:     "小写物种名",
			mode:     "RRBS",
			species1: "mouse",
			species2: "",
			checkFunc: func(fasta, index, gtf, ref []string) bool {
				return len(fasta) == 1 &&
					strings.Contains(fasta[0], "mouse.fasta") &&
					strings.Contains(index[0], "mouse")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fasta, index, gtf, ref, err := generateReferencePaths(tt.mode, tt.species1, tt.species2)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported")
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.checkFunc(fasta, index, gtf, ref),
					"Reference paths check failed for mode=%s, species1=%s, species2=%s",
					tt.mode, tt.species1, tt.species2)
			}
		})
	}
}
