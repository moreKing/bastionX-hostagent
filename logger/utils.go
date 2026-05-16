package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

func createOptions(lever *slog.LevelVar, addSource bool) *slog.HandlerOptions {
	// 设置目标时区（例如：Asia/Shanghai）
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("加载时区失败: " + err.Error())
	}

	return &slog.HandlerOptions{Level: lever, AddSource: addSource, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

		fmt.Println(a.Key)
		if a.Key == slog.TimeKey {
			t := a.Value.Time().In(location)
			// 格式化时间（包含时区信息）
			return slog.String(a.Key, t.Format("2006-01-02 15:04:05"))
		}

		return a
	}}
}

// 自定义 Handler
type CustomHandler struct {
	slog.Handler
	writer    io.Writer
	addSource bool
	depth     int
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// 创建一个缓冲区用于构建日志字符串
	var buf bytes.Buffer
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("加载时区失败: " + err.Error())
	}

	fileLine := ""

	// if h.addSource {
	// 	fs := runtime.CallersFrames([]uintptr{r.PC})
	// 	frame, _ := fs.Next()
	// 	fileLine = fmt.Sprintf("%s:%d", frame.File, frame.Line)
	// }

	if h.addSource {

		// 获取调用栈的程序计数器（PC）数组
		pcs := make([]uintptr, 32) // 假设最多获取 32 层调用栈
		n := runtime.Callers(0, pcs)
		pcs = pcs[:n]

		// 创建 CallersFrames 迭代器
		frames := runtime.CallersFrames(pcs)

		// fs := runtime.CallersFrames([]uintptr{r.PC})

		// 跳过指定的深度
		for range h.depth {
			_, _ = frames.Next()
		}

		frame, _ := frames.Next()
		fileLine = fmt.Sprintf("%s:%d", frame.File, frame.Line)
	}

	// 格式化日志字符串
	buf.WriteString(
		fmt.Sprintf("[%s] %s %s %s",
			strings.ToUpper(r.Level.String()),
			r.Time.In(location).Format("2006-01-02 15:04:05"),
			fileLine,
			r.Message,
		))

	// 添加日志属性
	r.Attrs(func(attr slog.Attr) bool {
		buf.WriteString(fmt.Sprintf("%s=%v ", attr.Key, attr.Value.Any()))
		return true
	})

	// buf.WriteTo(h.Handler)
	// 输出到标准输出
	_, err = fmt.Fprintln(h.writer, buf.String())
	return err
}
