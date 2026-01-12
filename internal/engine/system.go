package engine

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// SystemInfo 存储系统资源信息
type SystemInfo struct {
	TotalCores   int   // 总CPU核数
	AvailableMem int64 // 可用内存 (MB)
	TotalMem     int64 // 总内存 (MB)
}

// GetSystemInfo 获取系统资源信息
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// 获取CPU核数
	info.TotalCores = runtime.NumCPU()
	if info.TotalCores <= 0 {
		info.TotalCores = 1
		logger.Warn("Failed to detect CPU cores, using default: 1")
	}

	// 获取内存信息 - 从 /proc/meminfo 读取
	totalMemBytes, err := readTotalMemory()
	if err != nil {
		logger.Warnf("Failed to read memory from /proc/meminfo: %v", err)
		// 使用 fallback 方法
		var sysInfo syscall.Sysinfo_t
		if err2 := syscall.Sysinfo(&sysInfo); err2 != nil {
			return nil, fmt.Errorf("failed to get system memory info: %w", err)
		}
		// 假设页面大小为4096字节（大多数Linux系统）
		totalMemBytes = int64(sysInfo.Totalram) * 4096
	}

	info.TotalMem = totalMemBytes / (1024 * 1024) // 转换为MB

	// 计算可用内存 (总内存的80%)
	info.AvailableMem = int64(float64(info.TotalMem) * 0.8)

	logger.Debugf("System info: %d cores, %d MB total memory, %d MB available (80%%)",
		info.TotalCores, info.TotalMem, info.AvailableMem)

	return info, nil
}

// readTotalMemory 从 /proc/meminfo 读取总内存
func readTotalMemory() (int64, error) {
	file, err := open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 查找 "MemTotal:" 行
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// 提取数值（单位：kB）
				memKB, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("failed to parse memory value: %w", err)
				}
				// 转换为字节
				return memKB * 1024, nil
			}
		}
	}

	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

// open 是一个包装函数，在测试时可以模拟
var open = func(path string) (file, error) {
	f, err := os.Open(path)
	return file{f}, err
}

// file 是 os.File 的包装
type file struct {
	*os.File
}

// bufio.Scanner 需要的接口
func (f file) Read(p []byte) (n int, err error) {
	return f.File.Read(p)
}

// ParseMemory 解析内存字符串为MB
// 支持格式: "100G", "2000M", "10G", "512M", "8GB", "1000MB", "1T", "2TB"
func ParseMemory(memStr string) (int64, error) {
	if memStr == "" {
		return 0, nil
	}

	// 去除空格并转换为大写
	memStr = strings.TrimSpace(memStr)
	memStr = strings.ToUpper(memStr)

	// 提取数字部分
	var numStr, unitStr string
	for i, char := range memStr {
		if (char >= '0' && char <= '9') || char == '.' {
			numStr += string(char)
		} else {
			unitStr = memStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid memory format: %s", memStr)
	}

	// 解析数字
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value: %s", numStr)
	}

	// 处理单位
	unitMultiplier := int64(1)
	switch unitStr {
	case "", "B", "BYTE", "BYTES":
		unitMultiplier = 1
	case "K", "KB", "KILOBYTE", "KILOBYTES":
		unitMultiplier = 1024
	case "M", "MB", "MEGABYTE", "MEGABYTES":
		unitMultiplier = 1024 * 1024
	case "G", "GB", "GIGABYTE", "GIGABYTES":
		unitMultiplier = 1024 * 1024 * 1024
	case "T", "TB", "TERABYTE", "TERABYTES":
		unitMultiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown memory unit: %s", unitStr)
	}

	// 转换为字节，然后转为MB
	totalBytes := int64(value * float64(unitMultiplier))
	mb := totalBytes / (1024 * 1024)

	return mb, nil
}

// ValidateLocalResources 验证本地资源是否满足需求
func ValidateLocalResources(requestedCores int, requestedMemory string) error {
	// 获取系统信息
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	// 验证CPU
	if requestedCores <= 0 {
		return fmt.Errorf("invalid requested cores: %d", requestedCores)
	}

	if requestedCores > sysInfo.TotalCores {
		return fmt.Errorf("requested %d cores exceeds system available %d cores",
			requestedCores, sysInfo.TotalCores)
	}

	// 验证内存
	if requestedMemory != "" {
		requestedMemMB, err := ParseMemory(requestedMemory)
		if err != nil {
			return fmt.Errorf("invalid memory format: %w", err)
		}

		if requestedMemMB <= 0 {
			return fmt.Errorf("invalid requested memory: %s", requestedMemory)
		}

		if requestedMemMB > sysInfo.AvailableMem {
			return fmt.Errorf("requested %dMB exceeds available %dMB (80%% of %dMB total)",
				requestedMemMB, sysInfo.AvailableMem, sysInfo.TotalMem)
		}
	}

	// 警告：如果请求的CPU接近系统总核心数
	if requestedCores >= sysInfo.TotalCores {
		logger.Warnf("Requested %d cores equals total cores (%d). System may become unresponsive.",
			requestedCores, sysInfo.TotalCores)
	}

	logger.Infof("Resource validation passed: %d/%d cores, %s/%dMB memory available",
		requestedCores, sysInfo.TotalCores, requestedMemory, sysInfo.AvailableMem)

	return nil
}

// ValidateParallelJobs 验证并行任务数是否合理
func ValidateParallelJobs(parallelJobs int) error {
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	if parallelJobs <= 0 {
		return fmt.Errorf("invalid parallel jobs: %d", parallelJobs)
	}

	// 建议：并行任务不超过CPU核心数
	if parallelJobs > sysInfo.TotalCores {
		return fmt.Errorf("parallel jobs %d exceeds CPU cores %d",
			parallelJobs, sysInfo.TotalCores)
	}

	// 建议：并行任务不超过总核心数的75%
	maxRecommended := int(float64(sysInfo.TotalCores) * 0.75)
	if parallelJobs > maxRecommended {
		logger.Warnf("High parallelism (%d jobs on %d cores) may impact performance. Recommended: <= %d",
			parallelJobs, sysInfo.TotalCores, maxRecommended)
	} else {
		logger.Debugf("Parallel jobs validation passed: %d/%d cores", parallelJobs, sysInfo.TotalCores)
	}

	return nil
}
