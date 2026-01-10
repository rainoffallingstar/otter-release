package tui

import (
	"fmt"
	"os"
	"time"
)

// SimpleTUI represents a simple terminal UI
type SimpleTUI struct {
	Running bool
}

// NewSimpleTUI creates a new simple TUI
func NewSimpleTUI() *SimpleTUI {
	return &SimpleTUI{
		Running: false,
	}
}

// Run starts the TUI interface
func (tui *SimpleTUI) Run() error {
	fmt.Println("xdxtools TUI - Starting...")
	fmt.Println("Available options:")
	fmt.Println("  1. Create New Project")
	fmt.Println("  2. Run Workflow")
	fmt.Println("  3. View Dashboard")
	fmt.Println("  4. Configuration")
	fmt.Println("  5. Exit")
	fmt.Println()

	// Simple input loop
	for tui.Running {
		fmt.Print("Select option (1-5): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			fmt.Println("Create New Project - Feature coming soon!")
		case "2":
			fmt.Println("Run Workflow - Feature coming soon!")
		case "3":
			fmt.Println("View Dashboard - Feature coming soon!")
		case "4":
			fmt.Println("Configuration - Feature coming soon!")
		case "5":
			fmt.Println("Exiting...")
			tui.Running = false
			return nil
		default:
			fmt.Println("Invalid option. Please try again.")
		}
		fmt.Println()
	}

	return nil
}

// RunWithBubble starts the TUI using Bubble Tea (if available)
func (tui *SimpleTUI) RunWithBubble() error {
	// This is a placeholder for future Bubble Tea implementation
	// For now, we'll use the simple version
	return tui.Run()
}

// Model represents the TUI model (simplified for now)
type Model struct {
	Page       string
	WorkflowID string
	ConfigPath string
}

// RunInteractiveMenu runs an interactive menu
func RunInteractiveMenu() error {
	tui := NewSimpleTUI()
	tui.Running = true
	return tui.Run()
}

// DisplayWelcome displays a welcome message
func DisplayWelcome() {
	fmt.Println("==================================================")
	fmt.Println("  xdxtools - Bioinformatics Workflow Manager")
	fmt.Println("  Terminal User Interface (TUI)")
	fmt.Println("==================================================")
	fmt.Println()
}

// DisplayStatus displays workflow status
func DisplayStatus(workflowID string, status string) {
	fmt.Printf("Workflow ID: %s\n", workflowID)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
}

// RunDashboard runs the dashboard view
func RunDashboard() {
	fmt.Println("Workflow Dashboard")
	fmt.Println("==================")
	fmt.Println()

	// TODO: Display real workflow status from system
	DisplayStatus("demo-workflow", "running")
	fmt.Println("Features coming soon:")
	fmt.Println("  - Real-time workflow monitoring")
	fmt.Println("  - Sample progress tracking")
	fmt.Println("  - Error reporting")
	fmt.Println()
}

// DisplayHelp displays help information
func DisplayHelp() {
	fmt.Println("Help - xdxtools TUI")
	fmt.Println("===================")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  create  - Create a new analysis project")
	fmt.Println("  run     - Execute a workflow")
	fmt.Println("  status  - View workflow status")
	fmt.Println("  config  - Manage configuration")
	fmt.Println("  help    - Show this help")
	fmt.Println("  quit    - Exit the application")
	fmt.Println()
}

// StartTUI starts the TUI interface
func StartTUI() error {
	DisplayWelcome()
	return RunInteractiveMenu()
}

// Error displays an error message
func Error(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// Info displays an information message
func Info(msg string) {
	fmt.Printf("Info: %s\n", msg)
}

// Warn displays a warning message
func Warn(msg string) {
	fmt.Printf("Warning: %s\n", msg)
}
