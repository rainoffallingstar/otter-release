package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/workflow"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [project-dir]",
	Short: "Show workflow status",
	Long: `Display the status of a running or completed workflow.

If no directory is specified, checks the current directory.
The status is read from the .xdxtools_state.json file.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Determine directory to check
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", dir)
	}

	// Create state manager
	state := workflow.NewState(dir, "")

	// Check if state file exists
	if !state.Exists() {
		fmt.Println("No workflow state found")
		fmt.Println("Run 'xdxtools run' to start a new workflow")
		return nil
	}

	// Load state
	if err := state.Load(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Print status
	printWorkflowStatus(state)

	return nil
}

func printWorkflowStatus(state *workflow.State) {
	data := state.GetData()

	// Calculate duration
	var duration time.Duration
	if !data.LastUpdate.IsZero() && !data.StartTime.IsZero() {
		duration = data.LastUpdate.Sub(data.StartTime)
	}

	fmt.Println()
	fmt.Println("Workflow Status")
	fmt.Println("================")
	fmt.Printf("Job ID:     %s\n", data.JobID)
	fmt.Printf("Status:     %s\n", formatStatus(data.Status))
	fmt.Printf("Started:    %s\n", data.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Update: %s\n", data.LastUpdate.Format("2006-01-02 15:04:05"))
	if duration > 0 {
		fmt.Printf("Duration:   %s\n", formatDuration(duration))
	}
	fmt.Println()

	// Configuration
	fmt.Println("Configuration:")
	fmt.Printf("  Mode:      %s\n", data.Config.WorkflowMode)
	fmt.Printf("  Species:   %s", data.Config.Species1)
	if data.Config.Species2 != "" {
		fmt.Printf(" + %s (PDX)", data.Config.Species2)
	}
	fmt.Println()
	fmt.Printf("  Samples:   %d\n", data.Config.SampleCount)
	fmt.Printf("  Engine:    %s\n", data.Config.EngineType)
	if data.Config.Partition != "" {
		fmt.Printf("  Partition: %s\n", data.Config.Partition)
	}
	fmt.Println()

	// Steps
	fmt.Println("Steps:")
	for _, step := range data.Steps {
		statusIcon := "○"
		switch step.Status {
		case "completed":
			statusIcon = "✓"
		case "running":
			statusIcon = "→"
		case "failed":
			statusIcon = "✗"
		case "pending":
			statusIcon = "○"
		}

		fmt.Printf("  %s Step %d (%s): %s", statusIcon, step.Step, step.Name, step.Status)

		// Add time info for completed/running steps
		if step.Status == "completed" && !step.EndTime.IsZero() {
			if !step.StartTime.IsZero() {
				stepDuration := step.EndTime.Sub(step.StartTime)
				fmt.Printf(" [%s]", formatDuration(stepDuration))
			}
		} else if step.Status == "running" && !step.StartTime.IsZero() {
			runningTime := time.Since(step.StartTime)
			fmt.Printf(" [%s running]", formatDuration(runningTime))
		}

		// Add SLURM job ID if available
		if step.JobID != "" {
			fmt.Printf(" (Job: %s)", step.JobID)
		}

		fmt.Println()
	}
	fmt.Println()

	// Overall progress
	completed := 0
	running := 0
	pending := 0
	failed := 0
	for _, step := range data.Steps {
		switch step.Status {
		case "completed":
			completed++
		case "running":
			running++
		case "pending":
			pending++
		case "failed":
			failed++
		}
	}

	fmt.Printf("Progress: %d completed, %d running, %d pending", completed, running, pending)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()

	// Next step
	if running > 0 {
		for _, step := range data.Steps {
			if step.Status == "running" {
				fmt.Printf("\nCurrently running: Step %d (%s)\n", step.Step, step.Name)
				fmt.Println("To resume after interruption, use:")
				fmt.Printf("  xdxtools run --config config.yaml --resume\n")
				break
			}
		}
	} else if completed < len(data.Steps) && failed == 0 {
		for _, step := range data.Steps {
			if step.Status == "pending" {
				fmt.Printf("\nNext step: Step %d (%s)\n", step.Step, step.Name)
				break
			}
		}
	} else if completed == len(data.Steps) {
		fmt.Println("\n✓ Workflow completed successfully!")
		fmt.Println("State file can be removed:")
		fmt.Printf("  rm %s\n", state.GetFilePath())
	}
}

func formatStatus(status string) string {
	switch status {
	case "completed":
		return "Completed"
	case "running":
		return "Running"
	case "failed":
		return "Failed"
	case "pending":
		return "Pending"
	default:
		return status
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	} else {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd%dh", days, hours)
	}
}
