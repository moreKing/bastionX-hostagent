package main

import (
	"context"
	"fmt"
	"host-agent/logger"
	"time"

	"github.com/moby/moby/client"
)

// waitForNetworkGateway 轮询等待 docker compose 创建自定义网络，
// 返回该网络的网关地址，用于 HTTP 监听，保证容器内可通过网关访问本服务
func waitForNetworkGateway(ctx context.Context, networkName string) (string, error) {
	logger.Info("等待 docker 网络创建: ", networkName)

	for {
		gateway, err := lookupNetworkGateway(ctx, networkName)
		if err == nil {
			return gateway, nil
		}

		logger.Warn("等待 docker 网络 ", networkName, " 就绪: ", err, "，3 秒后重试")
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func lookupNetworkGateway(ctx context.Context, networkName string) (string, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return "", fmt.Errorf("创建 Docker 客户端失败: %w", err)
	}
	defer cli.Close()

	res, err := cli.NetworkInspect(ctx, networkName, client.NetworkInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("获取网络信息失败: %w", err)
	}

	for _, c := range res.Network.IPAM.Config {
		if c.Gateway.IsValid() {
			return c.Gateway.String(), nil
		}
	}

	return "", fmt.Errorf("网络 %s 尚未分配网关地址", networkName)
}
