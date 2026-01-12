package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// NodeInfo 存储节点信息
type NodeInfo struct {
	Name            string
	State           string  // idle/alloc/mixed/down
	TotalCores      int     // 总CPU核心数
	AvailableCores  int     // 可用CPU核心数
	TotalMemory     int     // 总内存 (MB)
	AvailableMemory int     // 可用内存 (MB)
}

// ValidateSlurmPartition checks if a SLURM partition exists and is accessible
func ValidateSlurmPartition(partition string) error {
	if partition == "" {
		return fmt.Errorf("partition name cannot be empty")
	}
	cmd := exec.Command("sinfo", "-h", "-p", partition)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("partition '%s' does not exist or is not accessible: %w", partition, err)
	}
	// sinfo returns empty output for non-existent partitions (but exit code 0)
	if len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("partition '%s' does not exist", partition)
	}
	return nil
}

// ValidateSlurmNodeResources 检查分区中是否有节点满足资源需求
// 注意：sbatch是按空闲节点投递的，只需要检查是否有任意一个节点满足需求
func ValidateSlurmNodeResources(partition string, requestedCores int, requestedMemory string) error {
	// 1. 验证输入参数
	if requestedCores <= 0 {
		return fmt.Errorf("requested cores must be greater than 0, got %d", requestedCores)
	}

	// 2. 获取分区节点信息
	nodes, err := getPartitionNodes(partition)
	if err != nil {
		return fmt.Errorf("failed to get partition nodes: %w", err)
	}

	// 3. 解析请求的内存
	requestedMemMB, err := ParseMemory(requestedMemory)
	if err != nil {
		return fmt.Errorf("invalid memory format: %w", err)
	}

	// 4. 查找满足条件的节点
	suitableNodes := 0
	for _, node := range nodes {
		// 只考虑空闲或部分空闲的节点
		if node.State == "idle" || node.State == "mixed" {
			// 检查节点是否有足够资源
			if node.AvailableCores >= requestedCores && int64(node.AvailableMemory) >= requestedMemMB {
				suitableNodes++
				logger.Debugf("Found suitable node: %s - CPU: %d/%d, Memory: %dMB",
					node.Name, node.AvailableCores, node.TotalCores, node.AvailableMemory)
			}
		}
	}

	// 5. 验证结果
	if suitableNodes == 0 {
		return fmt.Errorf("no suitable nodes found in partition '%s' with %d cores and %dMB memory",
			partition, requestedCores, requestedMemMB)
	}

	logger.Infof("Found %d suitable nodes in partition '%s' for %d cores and %dMB",
		suitableNodes, partition, requestedCores, requestedMemMB)

	return nil
}

// getPartitionNodes 获取分区节点信息
func getPartitionNodes(partition string) ([]NodeInfo, error) {
	// 使用 -N 标志按节点列出
	// 格式: NODELIST,STATE,CPUS(A/I/O/T),MEMORY
	cmd := exec.Command("sinfo", "-N", "-h", "-p", partition,
		"-o", "%n,%T,%C,%m")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute sinfo: %w", err)
	}

	var nodes []NodeInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// sinfo output is comma-separated: NODELIST,STATE,CPUS(A/I/O/T),MEMORY
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}

		// 解析 CPU 信息 (A/I/O/T = Allocated/Idle/Other/Total)
		cpuParts := strings.Split(parts[2], "/")
		totalCores := 0
		availableCores := 0
		if len(cpuParts) >= 4 {
			totalCores, _ = strconv.Atoi(cpuParts[3])  // Total (第4个值)
			idle, _ := strconv.Atoi(cpuParts[1])         // Idle (第2个值，available)
			availableCores = idle
		}

		// 解析内存 (MB) - 跳过CPU parts，使用索引 3
		// 因为 parts[2] 是CPU，parts[3] 是内存
		totalMem := 0
		if len(parts) >= 4 {
			totalMem, _ = strconv.Atoi(parts[3])
		}

		// 检查节点状态
		state := parts[1]
		if state == "down" {
			continue // 跳过故障节点
		}

		node := NodeInfo{
			Name:            parts[0],
			State:           state,
			TotalCores:      totalCores,
			AvailableCores:  availableCores,
			TotalMemory:     totalMem,
			AvailableMemory: totalMem, // 简化：假设全部内存可用
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// SlurmEngine represents a Slurm cluster execution engine
type SlurmEngine struct {
	partition  string
	cores      int
	memory     string
	jobName    string
	maxRetries int
	scriptPath string
	status     *Status
	logDir     string
}

// SlurmConfig represents Slurm-specific configuration
type SlurmConfig struct {
	Partition  string `json:"partition"`
	Cores      int    `json:"cores"`
	Memory     string `json:"memory"`
	JobName    string `json:"job_name"`
	MaxRetries int    `json:"max_retries"`
}

// NewSlurmEngine creates a new Slurm engine
func NewSlurmEngine(config *SlurmConfig) *SlurmEngine {
	// Set default values if not provided
	partition := config.Partition
	if partition == "" {
		partition = "cpu112c"
	}
	cores := config.Cores
	if cores <= 0 {
		cores = 4
	}
	memory := config.Memory
	if memory == "" {
		memory = "8G"
	}
	jobName := config.JobName
	if jobName == "" {
		jobName = "xdxtools_job"
	}
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &SlurmEngine{
		partition:  partition,
		cores:      cores,
		memory:     memory,
		jobName:    jobName,
		maxRetries: maxRetries,
		status: &Status{
			State:     StatusPending,
			StartTime: time.Now(),
		},
	}
}

// Execute executes a command via Slurm
func (e *SlurmEngine) Execute(cmd []string) error {
	e.status.State = StatusRunning
	e.status.Message = "Executing via Slurm"

	// Generate Slurm script
	scriptPath, err := e.generateSlurmScript(cmd)
	if err != nil {
		return fmt.Errorf("failed to generate Slurm script: %w", err)
	}
	e.scriptPath = scriptPath

	// Submit job
	jobID, err := e.submitJob(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to submit Slurm job: %w", err)
	}

	e.status.JobID = jobID
	logger.Infof("Slurm job submitted: %s", jobID)

	// Wait for completion
	if err := e.waitForCompletion(); err != nil {
		return fmt.Errorf("job execution failed: %w", err)
	}

	return nil
}

// ExecuteWithOutput executes a command and returns output
func (e *SlurmEngine) ExecuteWithOutput(cmd []string) (string, error) {
	// For Slurm engine, ExecuteWithOutput first submits the job and then
	// waits for completion while collecting output

	// Generate and submit the job
	scriptPath, err := e.generateSlurmScript(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to generate Slurm script: %w", err)
	}
	e.scriptPath = scriptPath

	// Submit job
	jobID, err := e.submitJob(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to submit Slurm job: %w", err)
	}

	e.status.JobID = jobID
	logger.Infof("Slurm job submitted: %s", jobID)

	// Wait for completion and collect output
	if err := e.waitForCompletion(); err != nil {
		return "", fmt.Errorf("job execution failed: %w", err)
	}

	// Collect job output using sacct
	return e.collectJobOutput(jobID)
}

// GetName returns the engine name
func (e *SlurmEngine) GetName() EngineType {
	return EngineSlurm
}

// GetStatus returns the current status
func (e *SlurmEngine) GetStatus() *Status {
	return e.status
}

// Wait waits for the job to complete
func (e *SlurmEngine) Wait() error {
	return e.waitForCompletion()
}

// getWorkDir returns the working directory for the job
func (e *SlurmEngine) getWorkDir() string {
	// Get current working directory
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	// Fallback to home directory
	return os.Getenv("HOME")
}

// Kill terminates the Slurm job
func (e *SlurmEngine) Kill() error {
	if e.status.JobID == "" {
		return fmt.Errorf("no job ID available")
	}

	cmd := exec.Command("scancel", e.status.JobID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	e.status.State = StatusKilled
	e.status.EndTime = time.Now()
	logger.Infof("Slurm job cancelled: %s", e.status.JobID)

	return nil
}

// SetLogDir sets the directory for SLURM log files
func (e *SlurmEngine) SetLogDir(dir string) error {
	e.logDir = dir
	return nil
}

// generateSlurmScript generates a Slurm batch script
func (e *SlurmEngine) generateSlurmScript(cmd []string) (string, error) {
	// Use log directory if set, otherwise fall back to temp directory
	tmpDir := e.logDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("xdxtools_%s.sh", e.jobName))

	// Slurm script template
	const scriptTemplate = `#!/bin/bash
#SBATCH --job-name={{.JobName}}
#SBATCH --partition={{.Partition}}
#SBATCH --cpus-per-task={{.Cores}}
#SBATCH --mem={{.Memory}}
#SBATCH --output={{.OutputFile}}
#SBATCH --error={{.ErrorFile}}

set -e

# Change to working directory
cd {{.WorkDir}}

# Execute command
{{range .Commands}}{{.}} {{end}}

# Touch success file
touch {{.SuccessFile}}
`

	data := struct {
		JobName     string
		Partition   string
		Cores       int
		Memory      string
		WorkDir     string
		OutputFile  string
		ErrorFile   string
		SuccessFile string
		Commands    []string
	}{
		JobName:     e.jobName,
		Partition:   e.partition,
		Cores:       e.cores,
		Memory:      e.memory,
		WorkDir:     e.getWorkDir(),
		OutputFile:  filepath.Join(tmpDir, fmt.Sprintf("%s.out", e.jobName)),
		ErrorFile:   filepath.Join(tmpDir, fmt.Sprintf("%s.err", e.jobName)),
		SuccessFile: filepath.Join(tmpDir, fmt.Sprintf("%s.success", e.jobName)),
		Commands:    cmd,
	}

	tmpl, err := template.New("slurm").Parse(scriptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	// Write script to file
	if err := os.WriteFile(scriptPath, buf.Bytes(), 0755); err != nil {
		return "", err
	}

	return scriptPath, nil
}

// submitJob submits a Slurm batch job
func (e *SlurmEngine) submitJob(scriptPath string) (string, error) {
	cmd := exec.Command("sbatch", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		stdoutStr := stdout.String()
		if stderrStr != "" {
			return "", fmt.Errorf("sbatch command failed: %w\nstderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("sbatch command failed: %w\nstdout: %s", err, stdoutStr)
	}

	// Parse job ID from output
	outputStr := strings.TrimSpace(stdout.String())
	// Expected format: "Submitted batch job 12345"
	parts := strings.Fields(outputStr)
	if len(parts) >= 4 {
		return parts[3], nil
	}

	return "", fmt.Errorf("failed to parse job ID from output: %s", outputStr)
}

// waitForCompletion waits for the Slurm job to complete
func (e *SlurmEngine) waitForCompletion() error {
	if e.status.JobID == "" {
		return fmt.Errorf("no job ID available")
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := e.checkJobStatus(e.status.JobID)
			if err != nil {
				logger.Warnf("Failed to check job status: %v", err)
				continue
			}

			e.status.Progress = status.Progress
			e.status.Message = status.Message

			switch status.State {
			case StatusCompleted:
				e.status.State = StatusCompleted
				e.status.EndTime = time.Now()
				logger.Info("Slurm job completed successfully")
				return nil
			case StatusFailed:
				e.status.State = StatusFailed
				e.status.EndTime = time.Now()
				return fmt.Errorf("Slurm job failed")
			case StatusRunning:
				logger.Debugf("Slurm job running: %s", status.Message)
			}
		}
	}
}

// checkJobStatus checks the status of a Slurm job
func (e *SlurmEngine) checkJobStatus(jobID string) (*Status, error) {
	cmd := exec.Command("squeue", "-j", jobID, "-o", "%T,%L")
	output, err := cmd.Output()
	if err != nil {
		// Job might have completed
		return &Status{
			State:   StatusCompleted,
			Message: "Job completed",
		}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	parts := strings.SplitN(outputStr, ",", 2)

	if len(parts) >= 2 {
		state := parts[0]
		message := parts[1]

		switch state {
		case "RUNNING":
			return &Status{
				State:   StatusRunning,
				Message: message,
			}, nil
		case "FAILED":
			return &Status{
				State:   StatusFailed,
				Message: message,
			}, nil
		}
	}

	return &Status{
		State:   StatusRunning,
		Message: "Checking...",
	}, nil
}

// collectJobOutput collects the output of a completed Slurm job
func (e *SlurmEngine) collectJobOutput(jobID string) (string, error) {
	// Use sacct to get job output
	cmd := exec.Command("sacct", "-j", jobID, "-X", "-o", "JobID,JobName,State,ExitCode", "--parsable2")
	output, err := cmd.Output()
	if err != nil {
		logger.Warnf("Failed to collect job output: %v", err)
		return "", fmt.Errorf("failed to collect job output: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	logger.Infof("Job output collected: %s", outputStr)

	return outputStr, nil
}
