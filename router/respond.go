package router

import (
	"encoding/json"
	"host-agent/conf"
	"host-agent/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondOK 将 data 序列化后作为 Response.Data 返回
func respondOK(c *gin.Context, data any) {
	resBytes, err := json.Marshal(data)
	if err != nil {
		respondError(c, err)
		return
	}
	logger.Debug("返回信息: ", string(resBytes))
	respond(c, conf.Response{Status: 0, Data: resBytes})
}

func respondError(c *gin.Context, err error) {
	respond(c, conf.Response{Status: -1, Message: err.Error()})
}

// respond 响应体为 Response 序列化后经 AES-GCM 加密的密文
func respond(c *gin.Context, response conf.Response) {
	resBytes, err := json.Marshal(&response)
	if err != nil {
		logger.Error("响应数据json序列化错误: ", err)
		c.Abort()
		return
	}

	logger.Debug("返回信息： ", string(resBytes))

	ciphertext, err := conf.EncryptAESGCM(resBytes)
	if err != nil {
		logger.Error("加密返回信息失败: ", err)
		c.Abort()
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", ciphertext)
}
