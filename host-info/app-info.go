package hostinfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"host-agent/logger"
	"os/exec"
	"strings"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type ContainerInfo struct {
	Name       string  `json:"name"`
	ImageName  string  `json:"imageName"`
	ImageTag   string  `json:"imageTag"`
	State      string  `json:"state"`
	CPUPercent float64 `json:"cpuPercent"`
	MemDisplay string  `json:"memDisplay"`
	Order      int     `json:"order"`
}

type ContainerInfoResponse struct {
	DockerVersion  string          `json:"dockerVersion"`
	ComposeVersion string          `json:"composeVersion"`
	Containers     []ContainerInfo `json:"containers"`
}

var containerInfoCache *[]ContainerInfo

func GetDockerContainerInfo() {
	ctx := context.Background()

	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli, err := client.New(client.FromEnv)
	if err != nil {
		logger.Error("无法创建 Docker 客户端", err)
	}
	defer cli.Close()

	// 获取 Docker 版本信息
	dockerVersion, err := cli.ServerVersion(ctx, client.ServerVersionOptions{})
	if err != nil {
		logger.Error("无法获取 Docker 版本信息", err)
	}

	logger.Debug("Docker 版本: ", dockerVersion.Version)

	// 获取 Docker Compose 版本信息
	composeVersion, err := getDockerComposeVersion()
	if err != nil {
		logger.Warn("无法获取 Docker Compose 版本信息: ", err)
	}

	logger.Debug("Docker Compose 版本: ", composeVersion)

	result, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		logger.Error("无法列出容器", err)
	}

	totalContainers := len(result.Items)
	if totalContainers == 0 {
		logger.Warn("没有找到任何容器。")
		return
	}

	ch := make(chan ContainerInfo, totalContainers)
	var wg sync.WaitGroup

	for i, c := range result.Items {
		imageName, imageTag := parseImageString(c.Image)
		containerName := ""
		if len(c.Names) > 0 {
			containerName = strings.TrimPrefix(c.Names[0], "/")
		}

		info := ContainerInfo{
			Name:       containerName,
			ImageName:  imageName,
			ImageTag:   imageTag,
			State:      string(c.State),
			CPUPercent: 0.0,
			MemDisplay: "0.00 B / 0.00 B",
			Order:      i,
		}

		if info.State == "running" {
			wg.Add(1)
			go func(cid string, baseInfo ContainerInfo) {
				defer wg.Done()
				cpu, mem := getPreciseContainerStats(ctx, cli, cid)
				baseInfo.CPUPercent = cpu
				baseInfo.MemDisplay = mem
				ch <- baseInfo
			}(c.ID, info)
		} else {
			ch <- info
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	collected := make([]ContainerInfo, totalContainers)
	for info := range ch {
		collected[info.Order] = info
	}

	logger.Info(fmt.Sprintf("%-20s %-25s %-10s %-10s %-10s %-25s\n", "容器名称", "镜像名称", "版本", "状态", "CPU%", "内存占用 (实际/限额)"))
	logger.Info(strings.Repeat("-", 105))

	for _, info := range collected {
		logger.Info(fmt.Sprintf("%-20s %-25s %-10s %-10s %-10.2f %-25s\n",
			info.Name, info.ImageName, info.ImageTag, info.State, info.CPUPercent, info.MemDisplay))
	}
}

func parseImageString(imageStr string) (string, string) {
	if strings.Contains(imageStr, "@sha256:") {
		parts := strings.Split(imageStr, "@")
		return parts[0], "sha256"
	}
	parts := strings.Split(imageStr, ":")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return imageStr, "latest"
}

func getPreciseContainerStats(ctx context.Context, cli *client.Client, containerID string) (float64, string) {
	stats, err := cli.ContainerStats(ctx, containerID, client.ContainerStatsOptions{Stream: true})
	if err != nil {
		return 0, "0.00 B / 0.00 B"
	}
	defer stats.Body.Close()

	decoder := json.NewDecoder(stats.Body)
	var firstSample container.StatsResponse
	var secondSample container.StatsResponse

	// 读取第一帧
	if err := decoder.Decode(&firstSample); err != nil {
		return 0, "0.00 B / 0.00 B"
	}
	// 读取第二帧（此步会天然阻塞 1 秒，让 CPU 产生可算的数据差）
	if err := decoder.Decode(&secondSample); err != nil {
		return calculateStats(&firstSample, &firstSample)
	}

	// 传入第二帧(最新)和第一帧(历史)
	return calculateStats(&secondSample, &firstSample)
}

func calculateStats(current, previous *container.StatsResponse) (float64, string) {
	// --- 内存占用计算 ---
	memUsage := current.MemoryStats.Usage
	if cache, ok := current.MemoryStats.Stats["inactive_file"]; ok {
		if memUsage > cache {
			memUsage -= cache
		}
	} else if cache, ok := current.MemoryStats.Stats["cache"]; ok {
		if memUsage > cache {
			memUsage -= cache
		}
	}
	memLimit := current.MemoryStats.Limit
	memDisplay := fmt.Sprintf("%s / %s", formatBytes(memUsage), formatBytes(memLimit))

	// --- 核心修复：CPU 动态率计算 ---
	cpuPercent := 0.0

	// 1. 容器在 1 秒内的 CPU 耗时差
	cpuDelta := float64(current.CPUStats.CPUUsage.TotalUsage) - float64(previous.CPUStats.CPUUsage.TotalUsage)
	// 2. 宿主机系统在 1 秒内的总时钟差（修正点：直接用两帧的 SystemUsage 相减）
	systemDelta := float64(current.CPUStats.SystemUsage) - float64(previous.CPUStats.SystemUsage)

	onlineCPUs := float64(current.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(current.CPUStats.CPUUsage.PercpuUsage))
	}

	// 3. 计算百分比
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	return cpuPercent, memDisplay
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffix := []string{"KiB", "MiB", "GiB", "TiB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), suffix)
}

func getDockerComposeVersion() (string, error) {
	cmd := exec.Command("docker", "compose", "version")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		cmdOld := exec.Command("docker-compose", "--version")
		out.Reset()
		cmdOld.Stdout = &out
		if errOld := cmdOld.Run(); errOld != nil {
			return "", fmt.Errorf("未检测到 docker compose 插件或独立二进制文件")
		}
	}

	result := strings.TrimSpace(out.String())
	if strings.Contains(result, "version ") {
		parts := strings.Split(result, "version ")
		if len(parts) > 1 {
			return parts[1], nil
		}
	}

	return result, nil
}
