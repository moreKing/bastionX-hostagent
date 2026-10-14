package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var serveLevel *slog.LevelVar

var serverLogger *slog.Logger

var rotateLog io.Closer

// DefaultLogPath 默认日志目录
const DefaultLogPath = "/var/bastionx/agent"

func init() {
	// 解析启动参数前先用仅输出到 stdout 的默认配置，
	// main 中 flag.Parse 之后调用 Init 以启用文件日志和自定义路径/级别
	serveLevel = &slog.LevelVar{}
	serveLevel.Set(slog.LevelInfo)

	handler := slog.NewTextHandler(os.Stdout, createOptions(serveLevel, true))
	customHandler := &CustomHandler{Handler: handler, writer: os.Stdout, addSource: true, depth: 5}
	serverLogger = slog.New(customHandler)
}

// Init 初始化日志，可在解析启动参数后调用以覆盖默认路径和级别
func Init(logFilePath string, level slog.Level) {
	if err := os.MkdirAll(logFilePath, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 创建按时间切割的日志文件
	logFile, err := rotatelogs.New(
		path.Join(logFilePath, "host-agent-%Y-%m-%d.log"),                 // 日志文件路径和命名规则
		rotatelogs.WithLinkName(path.Join(logFilePath, "host-agent.log")), // 创建软链接指向最新日志文件
		rotatelogs.WithMaxAge(30*24*time.Hour),                            // 保留30天的日志
		rotatelogs.WithRotationTime(24*time.Hour),                         // 每天切割一次
	)
	if err != nil {
		panic(err)
	}

	// 关闭上一次 Init 打开的文件
	if rotateLog != nil {
		rotateLog.Close()
	}
	rotateLog = logFile

	// 创建MultiWriter，同时写入控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	if serveLevel == nil {
		serveLevel = &slog.LevelVar{}
	}
	serveLevel.Set(level)

	// 创建TextHandler，使用MultiWriter
	handler := slog.NewTextHandler(multiWriter, createOptions(serveLevel, true))
	customHandler := &CustomHandler{Handler: handler, writer: multiWriter, addSource: true, depth: 5}
	// 创建Logger并设置为默认
	serverLogger = slog.New(customHandler)

	serverLogger.Info(fmt.Sprintf("完成logger初始化，目录: %s 级别: %s", logFilePath, level.String()))
}

// ParseLevel 解析日志级别字符串（debug/info/warn/error）
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("未知的日志级别: %s（支持 debug/info/warn/error）", s)
	}
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
