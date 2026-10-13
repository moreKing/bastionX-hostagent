package hostinfo

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/disk"
)

func GetDisk() (*[]Disk, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Println("GetDiskInfo", err)
		return nil, err
	}

	var diskList []Disk
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			fmt.Println("GetDiskInfo", p.Mountpoint, err)
			continue
		}

		diskList = append(diskList, Disk{
			Total:      int64(usage.Total / 1024),
			Used:       int64(usage.Used / 1024),
			Free:       int64(usage.Free / 1024),
			FileSystem: p.Fstype,
			MountPoint: p.Mountpoint,
		})
	}

	return &diskList, nil
}
