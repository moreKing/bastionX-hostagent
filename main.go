package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"host-agent/conf"
	"host-agent/cron"
	hostinfo "host-agent/host-info"
	"host-agent/logger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultNetworkName = "bastionx-net"

func main() {

	port := flag.Int("port", conf.DEFAULT_PORT, "HTTP 服务监听端口")
	logPath := flag.String("log-path", logger.DefaultLogPath, "日志输出目录")
	logLevel := flag.String("log-level", "info", "日志级别：debug/info/warn/error")
	networkName := flag.String("network", defaultNetworkName, "docker compose 自定义网络名称，监听其网关地址")
	flag.Parse()

	level, err := logger.ParseLevel(*logLevel)
	if err != nil {
		fmt.Println(err)
		return
	}
	logger.Init(*logPath, level)

	logger.Debug("启动定时器")

	cron.GetCron().Add("get-network-info", 60*time.Second, hostinfo.UpdateNetInfoTrends, true)
	cron.GetCron().Add("get-cpu-info", 60*time.Second, hostinfo.UpdateCpuInfoTrends, true)
	cron.GetCron().Add("get-memory-info", 60*time.Second, hostinfo.UpdateMemoryTrends, true)
	cron.GetCron().Add("get-docker-container-info", 30*time.Second, hostinfo.UpdateDockerContainerInfo, true)

	logger.Debug("定时器启动完成")

	http.HandleFunc("/get-system-info", requireGet(getSystemInfo))
	http.HandleFunc("/get-system-trends", requireGet(getSystemTrends))

	// 等待 docker compose 创建自定义网络，监听网关地址，使容器内可访问本服务
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gateway, err := waitForNetworkGateway(ctx, *networkName)
	if err != nil {
		logger.Error("等待 docker 网络失败: ", err)
		panic(err)
	}
	logger.Info("获取到网络网关: ", *networkName, " -> ", gateway)

	addr := net.JoinHostPort(gateway, fmt.Sprintf("%d", *port))
	logger.Info("HTTP 服务启动，监听:", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("HTTP 服务启动失败: ", err)
		panic(err)
	}
}

// requireGet 限制接口仅支持 GET 请求
func requireGet(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeResponse(w, conf.Response{
				Status:  -1,
				Message: "仅支持 GET 请求",
			})
			return
		}
		handler(w, r)
	}
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	sysInfo, err := hostinfo.GetSystemInfo()
	if err != nil {
		writeResponse(w, conf.Response{Status: -1, Message: err.Error()})
		return
	}

	resBytes, err := json.Marshal(sysInfo)
	if err != nil {
		writeResponse(w, conf.Response{Status: -1, Message: err.Error()})
		return
	}

	logger.Debug("返回信息: ", string(resBytes))
	writeResponse(w, conf.Response{Status: 0, Data: resBytes})
}

func getSystemTrends(w http.ResponseWriter, r *http.Request) {
	datas, err := hostinfo.GetSystemAllTrends()
	if err != nil {
		writeResponse(w, conf.Response{Status: -1, Message: err.Error()})
		return
	}

	resBytes, err := json.Marshal(datas)
	if err != nil {
		writeResponse(w, conf.Response{Status: -1, Message: err.Error()})
		return
	}

	logger.Debug("返回信息: ", string(resBytes))
	writeResponse(w, conf.Response{Status: 0, Data: resBytes})
}

// writeResponse 响应体为 Response 序列化后经 AES-GCM 加密的密文
func writeResponse(w http.ResponseWriter, response conf.Response) {
	resBytes, err := json.Marshal(&response)
	if err != nil {
		logger.Error("响应数据json序列化错误: ", err)
		return
	}

	logger.Debug("返回信息： ", string(resBytes))

	ciphertext, err := conf.EncryptAESGCM(resBytes)
	if err != nil {
		logger.Error("加密返回信息失败: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(ciphertext); err != nil {
		logger.Error("发送响应失败: ", err)
	}
}
