package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"host-agent/conf"
	hostinfo "host-agent/host-info"
	"io"
	"log"
	"net"
	"os"
	"path"
)

func main() {

	// 定义 socket 文件路径,
	err := os.MkdirAll(path.Dir(conf.SOCKET_PATH), 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("目录创建成功", path.Dir(conf.SOCKET_PATH))
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

	fmt.Println("Unix Socket 服务器启动，监听:", conf.SOCKET_PATH)

	for {
		// 接收消息（会返回发送方地址）
		conn, err := lis.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		// n, clientAddr, err := conn.ReadFromUnix(buffer)
		// if err != nil {
		// 	fmt.Println("接收消息失败:", err)
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
			fmt.Println(err)
			return
		}

		fmt.Println("返回信息： ", string(resBytes))

		// 加密返回
		ciphertext, err := conf.EncryptAESGCM(resBytes)
		if err != nil {
			fmt.Println(string(resBytes))
			fmt.Println("加密返回信息失败:", err)
			return
		}

		if err := sendMsg(conn, ciphertext); err != nil {
			fmt.Println("发送响应失败:", err)
		}
	}()

	plaintext, err := conf.DecryptAESGCM(reqBytes)
	if err != nil {
		response.Message = "非法的请求内容：" + err.Error()
		return
	}

	fmt.Printf("收到消息: %s\n", string(plaintext))

	var req conf.Request
	if err := json.Unmarshal(plaintext, &req); err != nil {
		response.Message = err.Error()
		fmt.Printf("请求数据json解析错误: %v\n", err)
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

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-cpu":
		datas, err := hostinfo.GetCPU()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-net":
		datas, err := hostinfo.GetNetInfo()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-disk":
		datas, err := hostinfo.GetDisk()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-host":
		datas, err := hostinfo.GetHost()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return

	case "get-mem":
		datas, err := hostinfo.GetMemory()
		if err != nil {
			response.Message = err.Error()
			return
		}

		resBytes, err := json.Marshal(datas)
		if err != nil {
			response.Message = err.Error()
			return
		}

		fmt.Println(string(resBytes))

		response.Status = 0
		response.Data = resBytes
		return
	default:

	}

}
