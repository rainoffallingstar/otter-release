package tui

import (
	"fmt"
	"testing"
)

func TestNewSimpleTUI(t *testing.T) {
	tui := NewSimpleTUI()

	if tui == nil {
		t.Error("NewSimpleTUI returned nil")
	}

	if tui.Running {
		t.Error("Running should be false initially")
	}
}

func TestNewSimpleTUISetsRunningFalse(t *testing.T) {
	tui := NewSimpleTUI()

	if tui.Running != false {
		t.Errorf("Expected Running to be false, got %v", tui.Running)
	}
}

func TestDisplayWelcome(t *testing.T) {
	// Just verify the function doesn't panic
	DisplayWelcome()
}

func TestDisplayStatus(t *testing.T) {
	// Test with valid inputs
	DisplayStatus("test-workflow", "running")

	// Verify it doesn't panic with empty strings
	DisplayStatus("", "")
}

func TestRunDashboard(t *testing.T) {
	// Just verify the function doesn't panic
	RunDashboard()
}

func TestDisplayHelp(t *testing.T) {
	// Just verify the function doesn't panic
	DisplayHelp()
}

func TestError(t *testing.T) {
	// Test error function
	err := fmt.Errorf("test error")
	Error(err)
}

func TestInfo(t *testing.T) {
	// Test info function
	Info("test info")
}

func TestWarn(t *testing.T) {
	// Test warn function
	Warn("test warning")
}

func TestRunInteractiveMenu(t *testing.T) {
	// This would start an interactive menu, which we don't want in tests
	// Just verify the function exists
	// We can't actually call it as it would wait for input
}

func TestModelCreation(t *testing.T) {
	model := Model{
		Page:       "menu",
		WorkflowID: "test-123",
		ConfigPath: "/tmp/config.yaml",
	}

	if model.Page != "menu" {
		t.Errorf("Expected page 'menu', got '%s'", model.Page)
	}

	if model.WorkflowID != "test-123" {
		t.Errorf("Expected workflow ID 'test-123', got '%s'", model.WorkflowID)
	}

	if model.ConfigPath != "/tmp/config.yaml" {
		t.Errorf("Expected config path '/tmp/config.yaml', got '%s'", model.ConfigPath)
	}
}
