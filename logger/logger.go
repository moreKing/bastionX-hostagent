package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var serveLevel *slog.LevelVar

var serverLogger *slog.Logger

const logLevel = slog.LevelDebug

func init() {

	const logFilePath = "/var/log/bastionx/agent"
	// logConf := config.GetLogger()
	// 创建按时间切割的日志文件
	logFile, err := rotatelogs.New(
		path.Join(logFilePath, "host-agent-%Y-%m-%d.log"),                 // 日志文件路径和命名规则
		rotatelogs.WithLinkName(path.Join(logFilePath, "host-agent.log")), // 创建软链接指向最新日志文件
		rotatelogs.WithMaxAge(30*24*time.Hour),                            // 保留7天的日志
		rotatelogs.WithRotationTime(24*time.Hour),                         // 每天切割一次
	)
	if err != nil {
		panic(err)
	}

	// 创建MultiWriter，同时写入控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	serveLevel = &slog.LevelVar{}
	serveLevel.Set(logLevel)

	// 创建TextHandler，使用MultiWriter
	handler := slog.NewTextHandler(multiWriter, createOptions(serveLevel, true))
	customHandler := &CustomHandler{Handler: handler, writer: multiWriter, addSource: true, depth: 5}
	// 创建Logger并设置为默认
	logger := slog.New(customHandler)
	serverLogger = logger
	// slog.SetDefault(logger)

	serverLogger.Info("完成logger初始化，时区指定为东八区（北京时间）")

}

func SetLevel(l slog.Level) {
	serveLevel.Set(l)
}

func Debug(args ...any) {
	result := ""
	for i, v := range args {
		if i == 0 {
			result = fmt.Sprintf("%+v", v)
			continue
		}
		result = fmt.Sprintf("%s %v", result, v)
	}
	serverLogger.Debug(result)
}

func Info(args ...any) {
	result := ""
	for i, v := range args {
		if i == 0 {
			result = fmt.Sprintf("%+v", v)
			continue
		}
		result = fmt.Sprintf("%s %v", result, v)
	}
	serverLogger.Info(result)
}

func Error(args ...any) {
	result := ""
	for i, v := range args {
		if i == 0 {
			result = fmt.Sprintf("%+v", v)
			continue
		}
		result = fmt.Sprintf("%s %v", result, v)
	}
	serverLogger.Error(result)
}

func Warn(args ...any) {
	result := ""
	for i, v := range args {
		if i == 0 {
			result = fmt.Sprintf("%+v", v)
			continue
		}
		result = fmt.Sprintf("%s %v", result, v)
	}
	serverLogger.Warn(result)
}

func GetLogger() *slog.Logger {
	return serverLogger
}
