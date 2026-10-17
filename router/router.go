package router

import "github.com/gin-gonic/gin"

// New 构建 HTTP 路由，注册所有接口
func New() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/get-system-info", getSystemInfo)
	r.GET("/get-system-trends", getSystemTrends)
	r.POST("/restart/container/:name", restartContainer)

	return r
}
