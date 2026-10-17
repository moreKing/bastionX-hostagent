package router

import (
	"context"
	"errors"
	"host-agent/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

// gracefulStopSeconds 重启容器前等待其优雅退出的秒数，超时后 Docker 会发送 SIGKILL
const gracefulStopSeconds = 10

// restartTimeout 重启整个请求的超时时间
const restartTimeout = 2 * time.Minute

// restartContainer 重启指定名称的容器
func restartContainer(c *gin.Context) {
	name := c.Param("name")

	ctx, cancel := context.WithTimeout(c.Request.Context(), restartTimeout)
	defer cancel()

	cli, err := client.New(client.FromEnv)
	if err != nil {
		respondError(c, errors.New("无法创建 Docker 客户端: "+err.Error()))
		return
	}
	defer cli.Close()

	logger.Info("收到容器重启请求: ", name)

	timeout := gracefulStopSeconds
	if _, err := cli.ContainerRestart(ctx, name, client.ContainerRestartOptions{Timeout: &timeout}); err != nil {
		respondError(c, errors.New("重启容器 "+name+" 失败: "+err.Error()))
		return
	}

	respondOK(c, gin.H{"name": name, "restarted": true})
}
