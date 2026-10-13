//go:build linux

package hostinfo

import (
	"fmt"
	"os"

	"github.com/zcalusic/sysinfo"
)

func applyPlatformSystemExtras(info *SystemInfo) error {
	var si sysinfo.SysInfo
	si.GetSysInfo()

	// Board info from DMI/SMBIOS
	info.BoardSerialNumber = si.Board.Serial
	info.BoardName = si.Board.Name
	info.BoardVendor = si.Board.Vendor

	// Product info
	info.ProductName = si.Product.Name
	info.ProductUUID = si.Product.UUID.String()

	// 系统创建时间：使用 /boot 目录的修改时间作为近似值
	bootStat, err := os.Stat("/boot")
	if err != nil {
		fmt.Println("bootCreateTime err:", err)
		return err
	}
	info.SystemCreateTime = fmt.Sprintf("%d", bootStat.ModTime().Unix())

	return nil
}
