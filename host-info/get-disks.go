package hostinfo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var reGetDiskInfo = regexp.MustCompile(`(?m)^(.+?)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)%\s+(.+)$`)

func GetDisk() (*[]Disk, error) {
	output, err := exec.Command("df", "-k").Output()
	if err != nil {
		fmt.Println("GetDiskInfo", err)
		return nil, err
	}
	var diskList []Disk
	for _, m := range reGetDiskInfo.FindAllStringSubmatch(string(output), -1) {
		size, _ := strconv.ParseInt(m[2], 10, 64)
		used, _ := strconv.ParseInt(m[3], 10, 64)
		avail, _ := strconv.ParseInt(m[4], 10, 64)
		diskList = append(diskList, Disk{
			Total:      size,
			Used:       used,
			Free:       avail,
			FileSystem: m[1],
			MountPoint: m[6],
		})
	}

	return &diskList, nil
}
