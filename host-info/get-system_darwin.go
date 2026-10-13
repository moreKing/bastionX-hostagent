//go:build darwin

package hostinfo

import (
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
)

func applyPlatformSystemExtras(info *SystemInfo) error {
	// ProductUUID：gopsutil host.HostID() 在 macOS 上读取 IOPlatformUUID
	if hostID, err := host.HostID(); err == nil && hostID != "" {
		info.ProductUUID = hostID
	}

	// ProductName：sysctl 获取硬件型号
	if model, err := sysctlRead("hw.model"); err == nil {
		info.ProductName = model
	}

	// 主板信息：gopsutil 不支持，使用 system_profiler
	boardVendor, boardName, boardSerial := getBoardInfoFromProfiler()
	info.BoardVendor = boardVendor
	info.BoardName = boardName
	info.BoardSerialNumber = boardSerial

	// 系统创建时间
	info.SystemCreateTime = getDarwinSystemCreateTime()

	return nil
}

// sysctlRead 读取 sysctl 值
func sysctlRead(key string) (string, error) {
	out, err := exec.Command("sysctl", "-n", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getBoardInfoFromProfiler 通过 system_profiler 获取主板信息
func getBoardInfoFromProfiler() (vendor, name, serial string) {
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Model Name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "Model Name:"))
		}
		if strings.HasPrefix(line, "Serial Number") {
			serial = strings.TrimSpace(strings.TrimPrefix(line, "Serial Number (system):"))
			serial = strings.TrimSpace(strings.TrimPrefix(serial, "Serial Number:"))
		}
	}
	vendor = "Apple Inc."
	return
}

// getDarwinSystemCreateTime 获取 macOS 系统安装时间
func getDarwinSystemCreateTime() string {
	out, err := exec.Command("stat", "-f", "%m", "/var/db/.AppleSetupDone").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
