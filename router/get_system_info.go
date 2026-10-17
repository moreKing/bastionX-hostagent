package router

import (
	hostinfo "host-agent/host-info"

	"github.com/gin-gonic/gin"
)

// getSystemInfo 获取主机系统信息
func getSystemInfo(c *gin.Context) {
	sysInfo, err := hostinfo.GetSystemInfo()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, sysInfo)
}
