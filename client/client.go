package client

import (
	"encoding/json"
	"fmt"
	"host-agent/conf"
	hostinfo "host-agent/host-info"
	"net"
	"os"
)

func Client() {

	clientPath := fmt.Sprintf("/tmp/client_%d.sock", os.Getpid())
	// 删除旧的客户端 socket 文件
	os.Remove(clientPath)

	// 创建客户端地址（重要：客户端必须有自己的地址）
	clientAddr, err := net.ResolveUnixAddr("unixgram", clientPath)
	if err != nil {
		panic(err)
	}

	// 创建服务器地址
	serverAddr, err := net.ResolveUnixAddr("unixgram", conf.SOCKET_PATH)
	if err != nil {
		panic(err)
	}

	// 创建客户端连接（绑定到客户端地址）
	conn, err := net.DialUnix("unixgram", clientAddr, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 发送消息\

	req := conf.Request{
		Path: "get-system-info",
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		fmt.Println(err)
		return
	}

	ciphertext, err := conf.EncryptAESGCM(reqBytes)
	if err != nil {
		fmt.Println("加密发送失败:", err)
		return
	}

	if _, err := conn.Write(ciphertext); err != nil {
		fmt.Println("发送失败:", err)
		return
	}

	// 接收响应
	buffer := make([]byte, 65536)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("接收失败:", err)
		return
	}

	// 解密
	responseBytes, err := conf.DecryptAESGCM(buffer[:n])
	if err != nil {
		return
	}

	// 解析
	var response conf.Response
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		fmt.Println("接收转码失败:", err)
		return
	}

	fmt.Printf("收到回显: %s\n", string(response.Data))
}

func SystemInfo() {

	conn, err := net.Dial("unix", conf.SOCKET_PATH)
	if err != nil {
		fmt.Println("请求错误：", conf.SOCKET_PATH, err)
		return
	}
	defer conn.Close()

	// 发送消息\

	req := conf.Request{
		Path: "get-system-info",
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		fmt.Println(err)
		return
	}

	ciphertext, err := conf.EncryptAESGCM(reqBytes)
	if err != nil {
		fmt.Println("加密发送失败:", err)
		return
	}

	if err := sendMsg(conn, ciphertext); err != nil {
		fmt.Println("发送失败:", err)
		return
	}

	// 接收响应
	resstr, err := recvMsg(conn)
	if err != nil {
		fmt.Println("接收失败:", err)
		return
	}

	// 解密
	responseBytes, err := conf.DecryptAESGCM(resstr)
	if err != nil {
		return
	}

	// 解析
	var response conf.Response
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		fmt.Println("接收转码失败:", err)
		return
	}

	// fmt.Printf("收到回显: %s\n", string(response.Data))

	var sysInfo hostinfo.SystemInfo
	if err := json.Unmarshal(response.Data, &sysInfo); err != nil {
		fmt.Println("解析json失败", err)
		return
	}

	fmt.Printf("解析成功：%#v\n", sysInfo)

}
