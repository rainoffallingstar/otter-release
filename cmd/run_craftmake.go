package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	craftmakeclient "github.com/rainoffallingstar/otter/internal/craftmake"
	execution "github.com/rainoffallingstar/otter/internal/execution"
	taskruntime "github.com/rainoffallingstar/otter/internal/task"
	"github.com/spf13/cobra"
)

const (
	runExecutorCraftmake = "craftmake"
	runExecutorSnakemake = "snakemake"
)

var (
	runExecutor        string
	runBackend         string
	runPhase           string
	runWorkflowPath    string
	runWorkflowCatalog string
	runCraftmakeBinary string
)

func init() {
	runCmd.Flags().StringVar(&runExecutor, "executor", runExecutorCraftmake, "Workflow executor (craftmake/snakemake)")
	runCmd.Flags().StringVar(&runBackend, "backend", "auto", "Execution backend (auto/local/slurm)")
	runCmd.Flags().StringVar(&runPhase, "phase", "", "Craftmake workflow phase, for example step1")
	runCmd.Flags().StringVar(&runWorkflowPath, "workflow", "", "Explicit Craftmake workflow YAML path")
	runCmd.Flags().StringVar(&runWorkflowCatalog, "catalog", "", "Craftmake workflow catalog root")
	runCmd.Flags().StringVar(&runCraftmakeBinary, "craftmake-binary", "", "Craftmake executable path")
}

func selectedRunExecutor() (string, error) {
	selected := strings.ToLower(strings.TrimSpace(runExecutor))
	switch selected {
	case runExecutorCraftmake, runExecutorSnakemake:
		return selected, nil
	default:
		return "", fmt.Errorf("unsupported executor %q; expected craftmake or snakemake", runExecutor)
	}
}

func executeCraftmakeRun(command *cobra.Command) (runErr error) {
	if internalWorker {
		defer func() {
			exitCode := 0
			finalStatus := taskruntime.StatusCompleted
			errorMessage := ""
			if runErr != nil {
				var craftmakeExitError *craftmakeclient.ExitCodeError
				if errors.As(runErr, &craftmakeExitError) {
					exitCode = craftmakeExitError.ExitCode()
				} else {
					exitCode = 1
				}
				finalStatus = taskruntime.StatusFailed
				errorMessage = runErr.Error()
			}
			_ = taskruntime.UpdateCurrent(func(record *taskruntime.Record) error {
				if record.Status == taskruntime.StatusStopping {
					finalStatus = taskruntime.StatusStopped
				}
				record.Status = finalStatus
				record.ExitCode = &exitCode
				record.Error = errorMessage
				record.FinishedAt = time.Now()
				return nil
			})
		}()
	}
	snapshotPath, _, snapshot, err := loadCraftmakeSnapshot(command)
	if err != nil {
		return err
	}
	binaryPath, err := craftmakeclient.ResolveBinary(runCraftmakeBinary)
	if err != nil {
		return err
	}
	craftmakeCommand, arguments, err := craftmakeRunArguments(snapshotPath, snapshot)
	if err != nil {
		return err
	}
	result, executeErr := craftmakeclient.Execute(command.Context(), craftmakeclient.Request{
		Binary: binaryPath, Command: craftmakeCommand, Arguments: arguments,
	})
	if result.Stderr != "" {
		fmt.Fprint(command.ErrOrStderr(), result.Stderr)
	}
	if executeErr != nil {
		return executeErr
	}
	if result.Failed() {
		return fmt.Errorf("Craftmake %s failed with exit code %d", craftmakeCommand, result.ExitCode)
	}
	if _, err := fmt.Fprint(command.OutOrStdout(), result.Stdout); err != nil {
		return fmt.Errorf("write Craftmake response: %w", err)
	}
	return taskruntime.UpdateCurrent(func(record *taskruntime.Record) error {
		record.CraftmakeRunID = result.Envelope.RunID
		if result.Envelope.StatePath != "" {
			record.StatePath = result.Envelope.StatePath
		}
		record.Message = fmt.Sprintf("Craftmake %s completed", craftmakeCommand)
		return nil
	})
}

func submitBackgroundCraftmakeRun(command *cobra.Command) error {
	configPath, projectDir, snapshot, err := loadCraftmakeSnapshot(command)
	if err != nil {
		return err
	}
	binaryPath, err := craftmakeclient.ResolveBinary(runCraftmakeBinary)
	if err != nil {
		return err
	}
	store, err := taskruntime.DefaultStore()
	if err != nil {
		return err
	}
	taskID, err := taskruntime.GenerateID()
	if err != nil {
		return err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve otter executable: %w", err)
	}
	workerArguments := buildBackgroundWorkerArguments(os.Args[1:], configPath, projectDir)
	record := taskruntime.NewRecord(taskID, projectDir, configPath, runExecutorCraftmake, append([]string{executablePath}, workerArguments...))
	record.CraftmakeRunID = craftmakePhaseRunID(snapshot.Run.ID, runPhase)
	record.CraftmakeBinary = binaryPath
	record.LogPath = store.LogPath(taskID)
	record.StatePath = filepath.Join(snapshot.Paths.State, "state.sqlite")
	if err := store.Create(record); err != nil {
		return err
	}

	logFile, err := os.OpenFile(record.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open background task log: %w", err)
	}
	defer logFile.Close()

	workerCommand := exec.Command(executablePath, workerArguments...)
	workerCommand.Dir = projectDir
	workerCommand.Stdout = logFile
	workerCommand.Stderr = logFile
	workerCommand.Env = append(os.Environ(),
		taskruntime.EnvironmentTaskID+"="+taskID,
		taskruntime.EnvironmentTaskStateDir+"="+store.RootDir(),
	)
	workerCommand.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := workerCommand.Start(); err != nil {
		_ = store.Update(taskID, func(current *taskruntime.Record) error {
			current.Status = taskruntime.StatusFailed
			current.Error = err.Error()
			current.FinishedAt = time.Now()
			return nil
		})
		return fmt.Errorf("start background Craftmake task: %w", err)
	}
	processID := workerCommand.Process.Pid
	if err := store.Update(taskID, func(current *taskruntime.Record) error {
		current.PID = processID
		current.ProcessGroupID = processID
		current.Status = taskruntime.StatusRunning
		current.StartedAt = time.Now()
		return nil
	}); err != nil {
		_ = syscall.Kill(-processID, syscall.SIGTERM)
		return err
	}
	if err := workerCommand.Process.Release(); err != nil {
		return fmt.Errorf("release background Craftmake worker: %w", err)
	}

	fmt.Printf("Task submitted: %s\n", taskID)
	fmt.Printf("Run:            %s\n", snapshot.Run.ID)
	fmt.Printf("Executor:       craftmake\n")
	fmt.Printf("Status:         otter task status %s\n", taskID)
	fmt.Printf("Logs:           otter task logs %s --follow\n", taskID)
	return nil
}

func loadCraftmakeSnapshot(command *cobra.Command) (string, string, configv1.RunSnapshot, error) {
	if err := rejectCraftmakeRuntimeOverrides(command); err != nil {
		return "", "", configv1.RunSnapshot{}, err
	}
	configPath, _, err := resolveRunPaths(runConfigFile, runProjectDir)
	if err != nil {
		return "", "", configv1.RunSnapshot{}, err
	}
	invocation, err := execution.LoadRunInvocation(configPath, configv1.ExecutorCraftmake)
	if err != nil {
		return "", "", configv1.RunSnapshot{}, fmt.Errorf("Craftmake executor requires an immutable otter.run/v1 run.yaml: %w", err)
	}
	resolvedBackend := string(invocation.Snapshot.Execution.Backend.Value)
	if runBackend != "auto" && runBackend != resolvedBackend {
		return "", "", configv1.RunSnapshot{}, fmt.Errorf("--backend %s conflicts with immutable run backend %s", runBackend, resolvedBackend)
	}
	return invocation.SnapshotPath, invocation.ProjectDirectory, invocation.Snapshot, nil
}

func craftmakePhaseRunID(snapshotRunID string, phase string) string {
	if strings.TrimSpace(phase) == "" {
		return snapshotRunID
	}
	return snapshotRunID + "--" + strings.TrimSpace(phase)
}

func craftmakeRunArguments(configPath string, snapshot configv1.RunSnapshot) (craftmakeclient.Command, []string, error) {
	statePath := filepath.Join(snapshot.Paths.State, "state.sqlite")
	phaseRunID := craftmakePhaseRunID(snapshot.Run.ID, runPhase)
	if resumeFlag {
		return craftmakeclient.CommandResume, []string{
			"--state", statePath,
			"--run", phaseRunID,
			"--format", "json",
		}, nil
	}
	arguments := []string{
		"--config", configPath,
		"--project-dir", snapshot.Paths.RunRoot,
		"--state-dir", snapshot.Paths.State,
		"--format", "json",
	}
	if runWorkflowPath != "" {
		arguments = append(arguments, "--workflow", runWorkflowPath)
	} else {
		if runPhase == "" {
			return "", nil, fmt.Errorf("--phase is required for Craftmake catalog routing when --workflow is not provided")
		}
		arguments = append(arguments, "--phase", runPhase)
		if runWorkflowCatalog != "" {
			arguments = append(arguments, "--catalog", runWorkflowCatalog)
		}
	}
	if dryRun {
		return craftmakeclient.CommandPlan, arguments, nil
	}
	arguments = append(arguments,
		"--backend", string(snapshot.Execution.Backend.Value),
		"--max-parallel", fmt.Sprintf("%d", parallelJobs),
	)
	return craftmakeclient.CommandRun, arguments, nil
}

func rejectCraftmakeRuntimeOverrides(command *cobra.Command) error {
	immutableFlags := []string{
		"slurm-partition", "slurm-cores", "slurm-memory", "slurm-unified-partition",
		"step1-cores", "step1-memory", "step1-partition",
		"step2-cores", "step2-memory", "step2-partition",
		"step3-cores", "step3-memory", "step3-partition",
		"step2-checker-cores", "step2-checker-memory",
		"step3-checker-cores", "step3-checker-memory",
		"copy-fastq", "move-fastq", "compress-fastq",
	}
	for _, flagName := range immutableFlags {
		if command.Flags().Changed(flagName) {
			return fmt.Errorf("--%s cannot override an immutable run snapshot; resolve a new run.yaml instead", flagName)
		}
	}
	return nil
}
