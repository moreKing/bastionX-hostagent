package hostinfo

import "github.com/shirou/gopsutil/v3/host"

func GetHost() (*DataOverview, error) {

	// 获取系统信息
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	// 获取磁盘信息
	diskList, err := GetDisk()
	if err != nil {
		return nil, err
	}

	return &DataOverview{
		Hostname:  hostInfo.Hostname,
		OS:        hostInfo.Platform + " " + hostInfo.PlatformVersion,
		Kernel:    hostInfo.KernelVersion,
		Arch:      hostInfo.KernelArch,
		StartTime: hostInfo.BootTime * 1000,
		RunTime:   hostInfo.Uptime,
		Disks:     *diskList,
	}, nil
}
