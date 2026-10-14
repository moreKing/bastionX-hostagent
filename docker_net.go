package main

import (
	"context"
	"fmt"
	"host-agent/logger"
	"strings"
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

	// compose 创建的网络名会带项目名前缀，如 055_bastionx-net，故列出后按名称匹配
	res, err := cli.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return "", fmt.Errorf("获取网络列表失败: %w", err)
	}

	for _, n := range res.Items {
		if n.Name != networkName && !strings.HasSuffix(n.Name, "_"+networkName) {
			continue
		}

		for _, c := range n.IPAM.Config {
			if c.Gateway.IsValid() {
				logger.Debug("匹配到 docker 网络: ", n.Name, " 网关: ", c.Gateway.String())
				return c.Gateway.String(), nil
			}
		}
	}

	return "", fmt.Errorf("未找到网络 %s 或尚未分配网关地址", networkName)
}
