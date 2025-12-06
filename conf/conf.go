package conf

var SOCKET_PATH = "/run/host-agent/agent.sock"

var SOCKET_KEY = []byte("BastionX!@#qwe12")

// API响应结构 模拟http返回与请求 正常 status 0 其他值为错误
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`        // 错误信息
	Data    []byte `json:"data,omitempty"` // 返回内容
}

type Request struct {
	Path string `json:"path"`           // 请求路径
	Data []byte `json:"data,omitempty"` // 请求内容
}
