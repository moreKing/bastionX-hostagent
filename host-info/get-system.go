package hostinfo

import (
	"fmt"
	stdnet "net"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
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

// execute command and return output
func getCommandOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
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

	// Get board info
	systemInfo.BoardSerialNumber, _ = getCommandOutput("cat", "/sys/class/dmi/id/board_serial")
	systemInfo.BoardName, _ = getCommandOutput("cat", "/sys/class/dmi/id/board_name")
	systemInfo.BoardVendor, _ = getCommandOutput("cat", "/sys/class/dmi/id/board_vendor")
	systemInfo.ProductUUID, _ = getCommandOutput("cat", "/sys/class/dmi/id/product_uuid")

	systemInfo.ProductName, _ = getCommandOutput("cat", "/sys/class/dmi/id/product_name")

	// 获取指定文件夹的创建时间
	bootCreateTime, err := getCommandOutput("stat", "-c", "%W", "/boot")
	if err != nil {
		fmt.Println("bootCreateTime err:", err)
		return nil, err
	}
	systemInfo.SystemCreateTime = bootCreateTime

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("interfaces err:", err)
		return nil, err
	}

	for _, iface := range interfaces {

		if iface.Name == "lo" {
			continue
		}

		// 判断是否物理接口
		device := filepath.Join("/sys/class/net", iface.Name, "device")
		if _, err := os.Stat(device); err != nil {
			fmt.Println(device, "地址不存在")
			continue
		}

		ethernet := Ethernet{
			Name: iface.Name,
			Mac:  iface.HardwareAddr,
			IPv4: []string{},
			IPv6: []string{},
		}

		if iface.HardwareAddr != "" {
			ethernet.Mac = iface.HardwareAddr
		}
		for _, addr := range iface.Addrs {
			ipStr := strings.Split(addr.Addr, "/")[0]
			ip := stdnet.ParseIP(ipStr)
			if ip != nil && !ip.IsLoopback() {
				if ip.To4() != nil {
					ethernet.IPv4 = append(ethernet.IPv4, ip.String())
				} else if ip.To16() != nil {
					ethernet.IPv6 = append(ethernet.IPv6, ip.String())
				}
			}
		}
		systemInfo.Ethernets = append(systemInfo.Ethernets, ethernet)
	}

	return systemInfo, nil
}
