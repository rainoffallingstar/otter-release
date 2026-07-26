package site

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func ValidateSlurmPartition(partition string) error {
	if strings.TrimSpace(partition) == "" {
		return fmt.Errorf("partition name cannot be empty")
	}
	cmd := exec.Command("sinfo", "-h", "-p", partition)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("partition %q is not accessible: %w", partition, err)
	}
	if len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("partition %q does not exist", partition)
	}
	return nil
}

func ValidateSlurmAccount(account string) error {
	if strings.TrimSpace(account) == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	cmd := exec.Command("sacctmgr", "show", "account", "where", "account="+account, "-s", "-n", "-o", "Account")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("account %q validation failed: %w", account, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == account {
			return nil
		}
	}
	return fmt.Errorf("account %q does not exist", account)
}

func ValidateSlurmAccountPartition(account string, partition string) error {
	if strings.TrimSpace(account) == "" || strings.TrimSpace(partition) == "" {
		return fmt.Errorf("account and partition must both be provided")
	}
	cmd := exec.Command("sacctmgr", "show", "assoc", "where",
		"account="+account, "partition="+partition,
		"-s", "-n", "-o", "Account",
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("account-partition association check failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("account %q cannot access partition %q", account, partition)
	}
	return nil
}

func ValidateSlurmQOS(qos string) error {
	if strings.TrimSpace(qos) == "" {
		return fmt.Errorf("qos name cannot be empty")
	}
	cmd := exec.Command("sacctmgr", "show", "qos", "where", "name="+qos, "-s", "-n", "-o", "Name")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("qos %q validation failed: %w", qos, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == qos {
			return nil
		}
	}
	return fmt.Errorf("qos %q does not exist", qos)
}

var slurmTimePattern = regexp.MustCompile(`^(?:(\d+)-)?(\d{1,2}):(\d{2}):(\d{2})$`)

func ValidateSlurmTimeFormat(timeSpec string) error {
	if strings.TrimSpace(timeSpec) == "" {
		return nil
	}
	if !slurmTimePattern.MatchString(timeSpec) {
		return fmt.Errorf("invalid SLURM time format %q; expected D-HH:MM:SS or HH:MM:SS", timeSpec)
	}
	return nil
}

func ValidateSlurmMaxJobs(maxJobs int) error {
	if maxJobs < 0 {
		return fmt.Errorf("max_jobs cannot be negative, got %d", maxJobs)
	}
	if maxJobs == 0 {
		return nil
	}
	return nil
}

func CheckComputeNodePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf("test -d %q && test -r %q", path, path))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("path %q is not accessible on the login node", path)
	}
	return nil
}

func ValidateComputeNodePath(path string, partition string, account string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}
	args := []string{
		"--ntasks=1",
		"--partition=" + partition,
		"--account=" + account,
		"bash", "-c", fmt.Sprintf("test -d %q && test -r %q", path, path),
	}
	cmd := exec.Command("srun", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("path %q is not accessible on compute nodes (partition %q, account %q): %w\n%s",
			path, partition, account, err, string(output))
	}
	return nil
}
