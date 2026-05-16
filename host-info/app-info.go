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
	// MemDisplay string  `json:"memDisplay"`
	MemoryUsed  uint64 `json:"memoryUsed"`
	MemoryTotal uint64 `json:"memoryTotal"`
	Order       int    `json:"order"`
}

type ContainerInfoResponse struct {
	DockerVersion  string           `json:"dockerVersion"`
	ComposeVersion string           `json:"composeVersion"`
	Running        bool             `json:"running"` // false 表示未运行
	Error          string           `json:"error"`   // 未运行错误
	Containers     *[]ContainerInfo `json:"containers"`
}

var containerInfoCache = &ContainerInfoResponse{
	DockerVersion:  "",
	ComposeVersion: "",
	Running:        false,
	Error:          "未初始化",
	Containers:     nil,
}

func UpdateDockerContainerInfo() {
	ctx := context.Background()

	var cifq = &ContainerInfoResponse{
		Running:        false,
		Error:          "未初始化",
		DockerVersion:  "",
		ComposeVersion: "",
	}

	defer func() {
		containerInfoCache = cifq
	}()

	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli, err := client.New(client.FromEnv)
	if err != nil {
		cifq.Error = fmt.Sprintf("无法创建 Docker 客户端: %v", err)
		logger.Error("无法创建 Docker 客户端", err)
		return
	}
	defer cli.Close()

	// 获取 Docker 版本信息
	dockerVersion, err := cli.ServerVersion(ctx, client.ServerVersionOptions{})
	if err != nil {
		cifq.Error = fmt.Sprintf("无法获取 Docker 版本信息: %v", err)
		logger.Error("无法获取 Docker 版本信息", err)
		return
	}
	cifq.DockerVersion = dockerVersion.Version

	logger.Debug("Docker 版本: ", dockerVersion.Version)

	// 获取 Docker Compose 版本信息
	composeVersion, err := getDockerComposeVersion()
	if err != nil {
		cifq.Error = fmt.Sprintf("无法获取 Docker Compose 版本信息: %v", err)
		logger.Warn("无法获取 Docker Compose 版本信息: ", err)
	}
	cifq.ComposeVersion = composeVersion

	logger.Debug("Docker Compose 版本: ", composeVersion)

	result, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		cifq.Error = fmt.Sprintf("无法列出容器: %v", err)
		logger.Error("无法列出容器", err)
		return
	}

	totalContainers := len(result.Items)
	if totalContainers == 0 {
		cifq.Error = "没有找到任何容器。"
		logger.Warn("没有找到任何容器。")
		return
	}

	cifq.Running = true
	cifq.Error = ""

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
			// MemDisplay: "0.00 B / 0.00 B",
			Order:       i,
			MemoryUsed:  0,
			MemoryTotal: 0,
		}

		if info.State == "running" {
			wg.Add(1)
			go func(cid string, baseInfo ContainerInfo) {
				defer wg.Done()
				cpu, memUsed, memTotal := getPreciseContainerStats(ctx, cli, cid)
				baseInfo.CPUPercent = cpu
				baseInfo.MemoryUsed = memUsed
				baseInfo.MemoryTotal = memTotal
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

	cifq.Containers = &collected

	for _, info := range collected {
		logger.Info(fmt.Sprintf("容器：%s  镜像： %s 版本： %s  状态： %s CPU使用率：%.2f 内存占用：%s\n",
			info.Name, info.ImageName, info.ImageTag, info.State, info.CPUPercent, fmt.Sprintf("%s / %s", formatBytes(info.MemoryUsed), formatBytes(info.MemoryTotal))))
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

func getPreciseContainerStats(ctx context.Context, cli *client.Client, containerID string) (float64, uint64, uint64) {
	stats, err := cli.ContainerStats(ctx, containerID, client.ContainerStatsOptions{Stream: true})
	if err != nil {
		return 0, 0, 0
	}
	defer stats.Body.Close()

	decoder := json.NewDecoder(stats.Body)
	var firstSample container.StatsResponse
	var secondSample container.StatsResponse

	// 读取第一帧
	if err := decoder.Decode(&firstSample); err != nil {
		return 0, 0, 0
	}
	// 读取第二帧（此步会天然阻塞 1 秒，让 CPU 产生可算的数据差）
	if err := decoder.Decode(&secondSample); err != nil {
		return calculateStats(&firstSample, &firstSample)
	}

	// 传入第二帧(最新)和第一帧(历史)
	return calculateStats(&secondSample, &firstSample)
}

func calculateStats(current, previous *container.StatsResponse) (float64, uint64, uint64) {
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

	return cpuPercent, memUsage, memLimit
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

func GetDockerContainerInfo() *ContainerInfoResponse {
	return containerInfoCache
}
