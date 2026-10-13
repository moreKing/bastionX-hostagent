package client

import (
	"encoding/json"
	"fmt"
	"host-agent/conf"
	hostinfo "host-agent/host-info"
	"io"
	"net/http"
)

// Get 发起 GET 请求并解密响应密文
func Get(path string) (*conf.Response, error) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", conf.DEFAULT_PORT, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ciphertext, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseBytes, err := conf.DecryptAESGCM(ciphertext)
	if err != nil {
		return nil, err
	}

	var response conf.Response
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func Client() {
	response, err := Get("/get-system-info")
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}

	fmt.Printf("收到回显: %s\n", string(response.Data))
}

func SystemInfo() {
	response, err := Get("/get-system-info")
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}

	var sysInfo hostinfo.SystemInfo
	if err := json.Unmarshal(response.Data, &sysInfo); err != nil {
		fmt.Println("解析json失败", err)
		return
	}

	fmt.Printf("解析成功：%#v\n", sysInfo)
}
