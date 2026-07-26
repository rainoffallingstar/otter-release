package cmd

import (
	"path/filepath"
	"testing"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	craftmakeclient "github.com/rainoffallingstar/otter/internal/craftmake"
)

func TestSelectedRunExecutorDefaultsToCraftmake(t *testing.T) {
	originalExecutor := runExecutor
	t.Cleanup(func() { runExecutor = originalExecutor })
	runExecutor = runExecutorCraftmake
	selected, err := selectedRunExecutor()
	if err != nil {
		t.Fatal(err)
	}
	if selected != runExecutorCraftmake {
		t.Fatalf("unexpected default executor %q", selected)
	}
}

func TestCraftmakeRunArgumentsPreserveSnapshotIdentity(t *testing.T) {
	originalPhase := runPhase
	originalWorkflowPath := runWorkflowPath
	originalCatalog := runWorkflowCatalog
	originalDryRun := dryRun
	originalResume := resumeFlag
	originalParallelJobs := parallelJobs
	t.Cleanup(func() {
		runPhase = originalPhase
		runWorkflowPath = originalWorkflowPath
		runWorkflowCatalog = originalCatalog
		dryRun = originalDryRun
		resumeFlag = originalResume
		parallelJobs = originalParallelJobs
	})
	runPhase = "step1"
	runWorkflowPath = ""
	runWorkflowCatalog = "/opt/craftmake/workflows"
	dryRun = false
	resumeFlag = false
	parallelJobs = 4
	snapshot := configv1.RunSnapshot{
		Run: configv1.RunMetadata{ID: "run-20260726T013245Z-kxqjrm"},
		Execution: configv1.ResolvedExecution{
			Backend: configv1.ResolvedBackend{Value: configv1.BackendSlurm},
		},
		Paths: configv1.RunPaths{
			RunRoot: "/shared/project/runs/run-20260726T013245Z-kxqjrm",
			State:   "/shared/project/runs/run-20260726T013245Z-kxqjrm/state",
		},
	}
	command, arguments, err := craftmakeRunArguments("/shared/project/runs/run-20260726T013245Z-kxqjrm/run.yaml", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if command != craftmakeclient.CommandRun {
		t.Fatalf("unexpected command %q", command)
	}
	expectedStateDirectory := filepath.Join(snapshot.Paths.RunRoot, "state")
	assertArgumentPair(t, arguments, "--run-id", snapshot.Run.ID)
	assertArgumentPair(t, arguments, "--backend", "slurm")
	assertArgumentPair(t, arguments, "--state-dir", expectedStateDirectory)
	assertArgumentPair(t, arguments, "--phase", "step1")
}

func TestCraftmakeRunArgumentsRequireExplicitCompilationEntry(t *testing.T) {
	originalPhase := runPhase
	originalWorkflowPath := runWorkflowPath
	originalDryRun := dryRun
	originalResume := resumeFlag
	t.Cleanup(func() {
		runPhase = originalPhase
		runWorkflowPath = originalWorkflowPath
		dryRun = originalDryRun
		resumeFlag = originalResume
	})
	runPhase = ""
	runWorkflowPath = ""
	dryRun = false
	resumeFlag = false
	_, _, err := craftmakeRunArguments("/tmp/run.yaml", configv1.RunSnapshot{})
	if err == nil {
		t.Fatal("expected missing phase or workflow to fail")
	}
}

func assertArgumentPair(t *testing.T, arguments []string, name string, expectedValue string) {
	t.Helper()
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == name {
			if arguments[index+1] != expectedValue {
				t.Fatalf("argument %s: got %q, expected %q", name, arguments[index+1], expectedValue)
			}
			return
		}
	}
	t.Fatalf("argument %s was not present in %#v", name, arguments)
}
