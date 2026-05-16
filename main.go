package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"host-agent/conf"
	"host-agent/cron"
	hostinfo "host-agent/host-info"
	"host-agent/logger"
	"io"
	"log"
	"net"
	"os"
	"path"
	"time"
)

func main() {

	logger.Debug("启动定时器")

	cron.GetCron().Add("get-network-info", 60*time.Second, hostinfo.UpdateNetInfoTrends, true)
	cron.GetCron().Add("get-cpu-info", 60*time.Second, hostinfo.UpdateCpuInfoTrends, true)
	cron.GetCron().Add("get-memory-info", 60*time.Second, hostinfo.UpdateMemoryTrends, true)

	logger.Debug("定时器启动完成")

	// 定义 socket 文件路径,
	err := os.MkdirAll(path.Dir(conf.SOCKET_PATH), 0666)
	if err != nil {
		panic(err)
	}
	// fmt.Println("目录创建成功", path.Dir(conf.SOCKET_PATH))
	logger.Info("目录创建成功", path.Dir(conf.SOCKET_PATH))
	// 移除已存在的 socket 文件
	os.Remove(conf.SOCKET_PATH)
	// 创建 Unix 地址

	// addr, err := net.ResolveUnixAddr("unixgram", conf.SOCKET_PATH)
	// if err != nil {
	// 	panic(err)
	// }

	// // 创建监听连接
	// conn, err := net.ListenUnixgram("unixgram", addr)
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close()

	lis, err := net.Listen("unix", conf.SOCKET_PATH)
	if err != nil {
		log.Fatal("listen:", err)
	}
	defer lis.Close()

	// 设置 socket 文件权限
	os.Chmod(conf.SOCKET_PATH, 0666)

	logger.Info("Unix Socket 服务器启动，监听:", conf.SOCKET_PATH)

	for {
		// 接收消息（会返回发送方地址）
		conn, err := lis.Accept()
		if err != nil {
			logger.Error("accept:", err)
			continue
		}
		// n, clientAddr, err := conn.ReadFromUnix(buffer)
		// if err != nil {
		// 	logger.Error("接收消息失败:", err)
		// 	continue
		// }

		go reqUnixHandle(conn)

	}

}

func sendMsg(w io.Writer, data []byte) error {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

func recvMsg(r io.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	l := binary.BigEndian.Uint32(lenBuf[:])
	if l > 64*1024*1024 { // 可自定义最大帧长
		return nil, fmt.Errorf("frame too large: %d", l)
	}
	data := make([]byte, l)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

func reqUnixHandle(conn net.Conn) {

	reqBytes, err := recvMsg(conn)
	if err != nil {
		return
	}

	var response = conf.Response{
		Status:  -1,
		Message: "",
		Data:    nil,
	}

	defer func() {
		// 发送响应给特定客户端
		resBytes, err := json.Marshal(&response)
		if err != nil {
			logger.Error("响应数据json序列化错误: %v\n", err)
			return
		}

		logger.Debug("返回信息： ", string(resBytes))

		// 加密返回
		ciphertext, err := conf.EncryptAESGCM(resBytes)
		if err != nil {
			logger.Error("加密返回信息失败: %v", err)
			return
		}

		if err := sendMsg(conn, ciphertext); err != nil {
			logger.Error("发送响应失败:", err)
			return
		}
	}()

	plaintext, err := conf.DecryptAESGCM(reqBytes)
	if err != nil {
		response.Message = "非法的请求内容：" + err.Error()
		logger.Error("解密请求内容失败: %v", err.Error())
		return
	}

	// fmt.Printf("收到消息: %s\n", string(plaintext))
	logger.Debug("收到消息: ", string(plaintext))

	var req conf.Request
	if err := json.Unmarshal(plaintext, &req); err != nil {
		response.Message = err.Error()
		logger.Error("请求数据json解析错误: %v\n", err)
		return
	}

	switch req.Path {
	case "get-system-info":
		sysInfo, err := hostinfo.GetSystemInfo()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(sysInfo)
		if err != nil {
			response.Message = err.Error()
			return
		}

		logger.Debug("返回信息: ", string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-system-trends":
		datas, err := hostinfo.GetSystemAllTrends()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}
		logger.Debug("返回信息: ", string(resBytes))
		response.Status = 0
		response.Data = resBytes

	default:
		response.Message = "未知的请求路径"
	}

}
