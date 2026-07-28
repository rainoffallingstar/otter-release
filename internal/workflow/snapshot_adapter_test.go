package workflow

import (
	"testing"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
)

func TestSnapshotToLegacyConfigPreservesCanonicalInputs(t *testing.T) {
	snapshot := configv1.RunSnapshot{
		Workflow: configv1.ResolvedWorkflow{Scenario: configv1.ScenarioWGBS},
		Execution: configv1.ResolvedExecution{
			Executor: configv1.ResolvedExecutor{Value: configv1.ExecutorSnakemake},
			Backend:  configv1.ResolvedBackend{Value: configv1.BackendLocal},
			Resources: configv1.ProjectResources{
				Defaults: configv1.ResourceSpec{Cores: 4, Memory: "8GiB", Partition: "cpu"},
				Phases: map[string]configv1.ResourceSpec{
					"align": {Cores: 8, Memory: "16GiB"},
				},
			},
		},
		Run:     configv1.RunMetadata{ID: "run-20260726T013245Z-abcdef"},
		Project: configv1.ResolvedProject{ID: "cohort-a", Root: "/project"},
		Samples: []configv1.SampleRecord{
			{ID: "S01", R1: "/inputs/S01_R1.fastq.gz", R2: "/inputs/S01_R2.fastq.gz"},
		},
		References: configv1.ResolvedReferences{Resolved: []configv1.ResolvedReference{
			{
				Role:  configv1.ReferenceRolePrimary,
				ID:    "hg38",
				Fasta: configv1.ResolvedAsset{Path: "/refs/hg38/genome.fa.gz"},
				Indexes: []configv1.ResolvedAsset{
					{Type: "bismark", Path: "/refs/hg38/bismark"},
				},
			},
		}},
		Paths: configv1.RunPaths{
			RunRoot: "/project/runs/run-20260726T013245Z-abcdef",
			Work:    "/project/runs/run-20260726T013245Z-abcdef/work",
			Results: "/project/runs/run-20260726T013245Z-abcdef/results",
			Logs:    "/project/runs/run-20260726T013245Z-abcdef/logs",
		},
	}
	configuration, err := SnapshotToLegacyConfig(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Workflow.Mode != "WGBS" || configuration.Workflow.JobID != snapshot.Run.ID {
		t.Fatalf("workflow identity was not preserved: %+v", configuration.Workflow)
	}
	if len(configuration.Workflow.Samples) != 1 || configuration.Workflow.Samples[0].R1 != snapshot.Samples[0].R1 {
		t.Fatalf("sample paths were not preserved: %+v", configuration.Workflow.Samples)
	}
	if configuration.Reference.Files.Fasta[0] != "/refs/hg38/genome.fa.gz" {
		t.Fatalf("reference FASTA was not preserved: %+v", configuration.Reference.Files.Fasta)
	}
	if configuration.StepResources[2].Cores != 8 {
		t.Fatalf("canonical align resources were not mapped: %+v", configuration.StepResources)
	}
}
