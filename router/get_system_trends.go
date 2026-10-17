package router

import (
	hostinfo "host-agent/host-info"

	"github.com/gin-gonic/gin"
)

// getSystemTrends 获取主机系统趋势数据
func getSystemTrends(c *gin.Context) {
	datas, err := hostinfo.GetSystemAllTrends()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, datas)
}
