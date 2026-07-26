package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	craftmakeclient "github.com/rainoffallingstar/otter/internal/craftmake"
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
				exitCode = 1
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
	configPath, _, snapshot, err := loadCraftmakeSnapshot()
	if err != nil {
		return err
	}
	binaryPath, err := craftmakeclient.ResolveBinary(runCraftmakeBinary)
	if err != nil {
		return err
	}
	craftmakeCommand, arguments, err := craftmakeRunArguments(configPath, snapshot)
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

func submitBackgroundCraftmakeRun() error {
	configPath, projectDir, snapshot, err := loadCraftmakeSnapshot()
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
	record.CraftmakeRunID = snapshot.Run.ID
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

func loadCraftmakeSnapshot() (string, string, configv1.RunSnapshot, error) {
	configPath, projectDir, err := resolveRunPaths(runConfigFile, runProjectDir)
	if err != nil {
		return "", "", configv1.RunSnapshot{}, err
	}
	snapshot, err := configv1.LoadRunSnapshot(configPath)
	if err != nil {
		return "", "", configv1.RunSnapshot{}, fmt.Errorf("Craftmake executor requires an immutable otter.run/v1 run.yaml: %w", err)
	}
	if snapshot.Execution.Executor.Value != configv1.ExecutorCraftmake {
		return "", "", configv1.RunSnapshot{}, fmt.Errorf("run snapshot selects executor %q, not craftmake", snapshot.Execution.Executor.Value)
	}
	resolvedBackend := string(snapshot.Execution.Backend.Value)
	if runBackend != "auto" && runBackend != resolvedBackend {
		return "", "", configv1.RunSnapshot{}, fmt.Errorf("--backend %s conflicts with immutable run backend %s", runBackend, resolvedBackend)
	}
	return configPath, projectDir, snapshot, nil
}

func craftmakeRunArguments(configPath string, snapshot configv1.RunSnapshot) (craftmakeclient.Command, []string, error) {
	statePath := filepath.Join(snapshot.Paths.State, "state.sqlite")
	if resumeFlag {
		return craftmakeclient.CommandResume, []string{
			"--state", statePath,
			"--run", snapshot.Run.ID,
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
		"--run-id", snapshot.Run.ID,
		"--max-parallel", fmt.Sprintf("%d", parallelJobs),
	)
	if slurmPartition != "" {
		arguments = append(arguments, "--partition", slurmPartition)
	}
	return craftmakeclient.CommandRun, arguments, nil
}
