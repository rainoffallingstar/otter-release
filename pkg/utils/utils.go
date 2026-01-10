package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dir string) error {
	if !DirExists(dir) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetAbsolutePath returns the absolute path of a file or directory
func GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// SafeString returns a safe string for filenames
func SafeString(s string) string {
	// Replace special characters with underscores
	safe := s
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "*", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "|", "_")
	return safe
}

// DeriveSuffix2 auto-derives suffix2 from suffix1
func DeriveSuffix2(suffix1 string) string {
	// Try replacing '1' with '2'
	if strings.Contains(suffix1, "1") {
		return strings.Replace(suffix1, "1", "2", 1)
	}

	// Try replacing "R1" with "R2"
	if strings.Contains(suffix1, "R1") {
		return strings.Replace(suffix1, "R1", "R2", 1)
	}

	// Fallback: return original
	return suffix1
}

// FormatError formats an error with context
func FormatError(context string, err error) error {
	return fmt.Errorf("%s: %w", context, err)
}
