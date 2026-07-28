package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rainoffallingstar/otter/internal/assets"
	"github.com/rainoffallingstar/otter/internal/config"
	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	"github.com/rainoffallingstar/otter/internal/engine"
	execution "github.com/rainoffallingstar/otter/internal/execution"
	"github.com/rainoffallingstar/otter/internal/logger"
	taskruntime "github.com/rainoffallingstar/otter/internal/task"
	"github.com/rainoffallingstar/otter/internal/workflow"
	"github.com/spf13/cobra"
)

type snakemakeSnapshotInvocation struct {
	SnapshotPath        string
	ProjectDirectory    string
	Snapshot            configv1.RunSnapshot
	Configuration       *config.OtterConfig
	CompatibilityConfig string
}

func executeSnakemakeSnapshotRun(command *cobra.Command) (runErr error) {
	if internalWorker {
		defer finishCurrentSnakemakeTask(&runErr)
	}
	invocation, err := loadSnakemakeSnapshotInvocation(command)
	if err != nil {
		return err
	}
	if err := verifyProjectAssets(invocation.ProjectDirectory); err != nil {
		return err
	}

	sampleNames := make([]string, 0, len(invocation.Snapshot.Samples))
	for _, sample := range invocation.Snapshot.Samples {
		sampleNames = append(sampleNames, sample.ID)
	}
	workflowInstance := workflow.NewWorkflow(invocation.Configuration, sampleNames)
	workflowInstance.Options = &workflow.WorkflowOptions{DryRun: dryRun, Resume: resumeFlag}
	manager := workflow.NewManager(workflowInstance)
	manager.SetSamples(sampleNames)
	manager.SetStepResources(invocation.Configuration.StepResources)
	manager.SetParallelJobs(parallelJobs)
	if loadRatio > 0 {
		manager.SetLoadRatio(loadRatio)
	}
	if err := manager.SetConfigFile(invocation.CompatibilityConfig); err != nil {
		return err
	}

	executionEngine, err := engine.CreateEngineFromConfig(invocation.Configuration)
	if err != nil {
		return fmt.Errorf("create Snakemake execution engine: %w", err)
	}
	workflowInstance.SetEngine(executionEngine)
	if err := validateResources(executionEngine.GetName().String(), invocation.Configuration, invocation.Configuration.StepResources, parallelJobs, dryRun); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}
	if dryRun {
		manager.SetDryRun(true)
	}
	condaEnvironment := runCondaEnv
	if condaEnvironment == "" {
		condaEnvironment = invocation.Configuration.Engine.CondaEnv
	}
	if condaEnvironment != "" {
		manager.SetCondaEnv(condaEnvironment)
	}
	fallbackEnvironment := invocation.Configuration.Engine.FallbackEnv
	if fallbackEnvironment == "" {
		fallbackEnvironment = workflow.DefaultFallbackEnv
	}
	manager.SetFallbackConfig(fallbackEnvironment, invocation.Configuration.Engine.NoFallback)
	if err := validateRNAsplicingDependencies(invocation.Configuration); err != nil {
		return err
	}

	originalDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}
	if err := os.Chdir(invocation.ProjectDirectory); err != nil {
		return fmt.Errorf("switch to project directory %q: %w", invocation.ProjectDirectory, err)
	}
	defer func() {
		if restoreErr := os.Chdir(originalDirectory); restoreErr != nil {
			logger.Warnf("Failed to restore working directory %s: %v", originalDirectory, restoreErr)
		}
	}()
	if err := manager.ExecuteAll(); err != nil {
		return fmt.Errorf("Snakemake workflow execution failed: %w", err)
	}
	if dryRun {
		return nil
	}
	publicationResult, err := workflow.PublishSnakemakeArtifacts(workflow.SnakemakeArtifactPublicationRequest{
		Context:      command.Context(),
		SnapshotPath: invocation.SnapshotPath,
		Snapshot:     invocation.Snapshot,
	})
	if err != nil {
		return fmt.Errorf("publish Snakemake compatibility artifacts: %w", err)
	}
	if publicationResult.AlreadyPublished {
		fmt.Fprintf(command.OutOrStdout(), "Verified existing immutable artifact manifest: %s\n", publicationResult.ManifestPath)
		return nil
	}
	fmt.Fprintf(command.OutOrStdout(), "Published %d immutable artifacts: %s\n", publicationResult.ArtifactCount, publicationResult.ManifestPath)
	return nil
}

func submitBackgroundSnakemakeSnapshotRun(command *cobra.Command) error {
	invocation, err := loadSnakemakeSnapshotInvocation(command)
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
	workerArguments := buildBackgroundWorkerArguments(os.Args[1:], invocation.SnapshotPath, invocation.ProjectDirectory)
	record := taskruntime.NewRecord(taskID, invocation.ProjectDirectory, invocation.SnapshotPath, runExecutorSnakemake, append([]string{executablePath}, workerArguments...))
	record.LogPath = store.LogPath(taskID)
	record.StatePath = workflow.NewState(invocation.Snapshot.Paths.RunRoot, invocation.Snapshot.Run.ID).GetFilePath()
	if err := store.Create(record); err != nil {
		return err
	}
	logFile, err := os.OpenFile(record.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open background task log: %w", err)
	}
	defer logFile.Close()
	workerCommand := exec.Command(executablePath, workerArguments...)
	workerCommand.Dir = invocation.ProjectDirectory
	workerCommand.Stdout = logFile
	workerCommand.Stderr = logFile
	workerCommand.Env = append(os.Environ(),
		taskruntime.EnvironmentTaskID+"="+taskID,
		taskruntime.EnvironmentTaskStateDir+"="+store.RootDir(),
	)
	workerCommand.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := workerCommand.Start(); err != nil {
		return fmt.Errorf("start background Snakemake task: %w", err)
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
		return fmt.Errorf("release background Snakemake worker: %w", err)
	}
	fmt.Printf("Task submitted: %s\n", taskID)
	fmt.Printf("Run:            %s\n", invocation.Snapshot.Run.ID)
	fmt.Printf("Executor:       snakemake\n")
	fmt.Printf("Status:         otter task status %s\n", taskID)
	fmt.Printf("Logs:           otter task logs %s --follow\n", taskID)
	return nil
}

func loadSnakemakeSnapshotInvocation(command *cobra.Command) (snakemakeSnapshotInvocation, error) {
	configPath, _, err := resolveRunPaths(runConfigFile, runProjectDir)
	if err != nil {
		return snakemakeSnapshotInvocation{}, err
	}
	invocation, err := execution.LoadRunInvocation(configPath, configv1.ExecutorSnakemake)
	if err != nil {
		return snakemakeSnapshotInvocation{}, fmt.Errorf("Snakemake compatibility executor requires an immutable otter.run/v1 run.yaml: %w", err)
	}
	if err := rejectSnakemakeRuntimeOverrides(command, invocation.Snapshot); err != nil {
		return snakemakeSnapshotInvocation{}, err
	}
	configuration, err := workflow.SnapshotToLegacyConfig(invocation.Snapshot)
	if err != nil {
		return snakemakeSnapshotInvocation{}, err
	}
	compatibilityConfigPath, err := writeSnakemakeCompatibilityConfig(invocation.Snapshot, configuration)
	if err != nil {
		return snakemakeSnapshotInvocation{}, err
	}
	return snakemakeSnapshotInvocation{
		SnapshotPath:        invocation.SnapshotPath,
		ProjectDirectory:    invocation.ProjectDirectory,
		Snapshot:            invocation.Snapshot,
		Configuration:       configuration,
		CompatibilityConfig: compatibilityConfigPath,
	}, nil
}

func rejectSnakemakeRuntimeOverrides(command *cobra.Command, snapshot configv1.RunSnapshot) error {
	resolvedBackend := string(snapshot.Execution.Backend.Value)
	if command.Flags().Changed("backend") && runBackend != "auto" && runBackend != resolvedBackend {
		return fmt.Errorf("--backend %s conflicts with immutable run backend %s", runBackend, resolvedBackend)
	}
	if command.Flags().Changed("engine") && runEngine != "auto" && runEngine != resolvedBackend {
		return fmt.Errorf("--engine %s conflicts with immutable run backend %s", runEngine, resolvedBackend)
	}
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

func writeSnakemakeCompatibilityConfig(snapshot configv1.RunSnapshot, configuration *config.OtterConfig) (string, error) {
	encoded, err := config.MarshalWithMapstructureTags(configuration)
	if err != nil {
		return "", err
	}
	path := filepath.Join(snapshot.Paths.State, "snakemake-config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create Snakemake compatibility config directory: %w", err)
	}
	if existingContent, err := os.ReadFile(path); err == nil {
		if string(existingContent) != string(encoded) {
			return "", fmt.Errorf("Snakemake compatibility config already exists with different content: %s", path)
		}
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect Snakemake compatibility config: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o444); err != nil {
		return "", fmt.Errorf("write Snakemake compatibility config: %w", err)
	}
	return path, nil
}

func verifyProjectAssets(projectDirectory string) error {
	if !verifyAssets {
		return nil
	}
	manifest, err := assets.LoadManifest(projectDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Assets manifest not found: %s", assets.ManifestPath(projectDirectory))
			return nil
		}
		return fmt.Errorf("load assets manifest: %w", err)
	}
	difference, err := assets.VerifyWorkflowAssets(projectDirectory, manifest)
	if err != nil {
		return fmt.Errorf("verify workflow assets: %w", err)
	}
	if difference.IsClean() {
		return nil
	}
	message := fmt.Sprintf("assets differ from manifest (missing=%d extra=%d modified=%d)", len(difference.Missing), len(difference.Extra), len(difference.Modified))
	if strictAssets {
		return fmt.Errorf("%s", message)
	}
	logger.Warn(message)
	return nil
}

func finishCurrentSnakemakeTask(runError *error) func() {
	return func() {
		exitCode := 0
		status := taskruntime.StatusCompleted
		errorMessage := ""
		if *runError != nil {
			exitCode = 1
			status = taskruntime.StatusFailed
			errorMessage = (*runError).Error()
		}
		_ = taskruntime.UpdateCurrent(func(record *taskruntime.Record) error {
			if record.Status == taskruntime.StatusStopping {
				status = taskruntime.StatusStopped
			}
			record.Status = status
			record.ExitCode = &exitCode
			record.Error = errorMessage
			record.FinishedAt = time.Now()
			return nil
		})
	}
}
