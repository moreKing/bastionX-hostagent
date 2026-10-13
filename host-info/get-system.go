package hostinfo

import (
	"fmt"
	stdnet "net"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/net"
)

type Ethernet struct {
	Name string   `json:"name"` // 网络接口名称
	Mac  string   `json:"mac"`  // mac地址
	IPv4 []string `json:"ipv4"` // ip地址
	IPv6 []string `json:"ipv6"` // ip地址
}

type SystemInfo struct {
	Hostname string `json:"hostname"` // 主机名
	OS       string `json:"os"`       // 操作系统
	Arch     string `json:"arch"`     // 架构
	Kernel   string `json:"kernel"`   // 内核版本

	SystemName string `json:"systemName"` // 系统名称

	ProductUUID string `json:"productUUID"` // 操作系统UUID
	ProductName string `json:"productName"` // 操作系统名称

	BoardSerialNumber string `json:"boardSerialNumber"` // 主板序列号
	BoardName         string `json:"boardName"`         // 主板产品名称
	BoardVendor       string `json:"boardVendor"`       // 主板供应商

	SystemCreateTime string `json:"systemCreateTime"` // 系统创建时间

	Ethernets []Ethernet `json:"ethernets"`

	CpuModelName string `json:"cpuModelName"` // cpu型号
	CpuCores     int    `json:"cpuCores"`     // cpu核心数
	CpuFamily    string `json:"cpuFamily"`    // cpu家族
	CpuMHz       int64  `json:"cpuMHz"`       // cpu主频
}

// readFileTrimmed 读取文件内容并去掉首尾空白
func readFileTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func GetSystemInfo() (*SystemInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("hostInfo err:", err)
		return nil, err
	}

	systemInfo := &SystemInfo{
		SystemName: "bastionx",
		Hostname:   hostInfo.Hostname,
		OS:         hostInfo.Platform + " " + hostInfo.PlatformVersion,
		Arch:       hostInfo.KernelArch,
		Kernel:     hostInfo.KernelVersion,
		Ethernets:  []Ethernet{},
	}

	// CPU 信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Println("cpuInfo err:", err)
		return nil, err
	}
	cpuCores, err := cpu.Counts(true)
	if err != nil {
		fmt.Println("cpuCores err:", err)
		return nil, err
	}
	if len(cpuInfo) > 0 {
		systemInfo.CpuModelName = cpuInfo[0].ModelName
		systemInfo.CpuCores = int(cpuCores)
		systemInfo.CpuFamily = cpuInfo[0].Family
		systemInfo.CpuMHz = int64(cpuInfo[0].Mhz)
	}

	// 平台差异：board/product/create time（由 _linux.go / _darwin.go 实现）
	if err := applyPlatformSystemExtras(systemInfo); err != nil {
		return nil, err
	}

	// 网络接口
	systemInfo.Ethernets = collectEthernets()

	return systemInfo, nil
}

// collectEthernets 收集物理网卡信息
func collectEthernets() []Ethernet {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("interfaces err:", err)
		return []Ethernet{}
	}

	var ethernets []Ethernet
	for _, iface := range ifaces {
		if iface.Name == "lo" || iface.Name == "" {
			continue
		}

		// 判断是否物理接口（仅 Linux 有效，macOS 会跳过）
		device := filepath.Join("/sys/class/net", iface.Name, "device")
		if _, err := os.Stat(device); err != nil {
			continue
		}

		ethernet := Ethernet{
			Name: iface.Name,
			Mac:  iface.HardwareAddr,
			IPv4: []string{},
			IPv6: []string{},
		}

		for _, addr := range iface.Addrs {
			ipStr := strings.Split(addr.Addr, "/")[0]
			ip := stdnet.ParseIP(ipStr)
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				ethernet.IPv4 = append(ethernet.IPv4, ip.String())
			} else if ip.To16() != nil {
				ethernet.IPv6 = append(ethernet.IPv6, ip.String())
			}
		}

		ethernets = append(ethernets, ethernet)
	}

	return ethernets
}
