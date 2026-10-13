package conf

// HTTP 服务默认监听端口，启动时可通过 -port 参数覆盖
var DEFAULT_PORT = 9090

var SOCKET_KEY = []byte("BastionX!@#qwe12")

// API响应结构 正常 status 0 其他值为错误
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`        // 错误信息
	Data    []byte `json:"data,omitempty"` // 返回内容
}
