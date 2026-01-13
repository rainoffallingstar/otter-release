package enva

import (
	"os/exec"
)

// IsAvailable 检查 enva 是否在 PATH 中
func IsAvailable() bool {
	_, err := exec.LookPath("enva")
	return err == nil
}

// BuildCondaCommand 构建 conda/enva 命令
// 如果 enva 可用，返回 "enva run <env> -- <command> <args...>"
// 否则返回 "conda run -n <env> <command> <args...>"
func BuildCondaCommand(envName string, command string, args ...string) []string {
	if IsAvailable() {
		// 使用 enva: enva run <env> -- <command> <args...>
		result := []string{"enva", "run", envName, "--", command}
		result = append(result, args...)
		return result
	}

	// 回退到 conda: conda run -n <env> <command> <args...>
	result := []string{"conda", "run", "-n", envName}
	result = append(result, command)
	result = append(result, args...)
	return result
}

// BuildCondaCommandWithFlags 类似 BuildCondaCommand，但支持在 command 后添加额外 flags
func BuildCondaCommandWithFlags(envName string, command string, flags []string, args ...string) []string {
	if IsAvailable() {
		// 使用 enva: enva run <env> -- <command> <flags...> <args...>
		result := []string{"enva", "run", envName, "--", command}
		result = append(result, flags...)
		result = append(result, args...)
		return result
	}

	// 回退到 conda: conda run -n <env> <flags...> <command> <args...>
	result := []string{"conda", "run", "-n", envName}
	if len(flags) > 0 {
		result = append(result, flags...)
	}
	result = append(result, command)
	result = append(result, args...)
	return result
}
