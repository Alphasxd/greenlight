package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

// 初始化日志级别常量
const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// 返回日志级别字符串
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger 定义日志结构体，包含输出流、最小日志级别、互斥锁
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// New 初始化日志结构体，包含输出流、最小日志级别
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) PrintInfo(msg string, properties map[string]string) {
	_, err := l.print(LevelInfo, msg, properties)
	if err != nil {
		return
	}
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	_, err = l.print(LevelError, err.Error(), properties)
	if err != nil {
		return
	}
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	_, err = l.print(LevelFatal, err.Error(), properties)
	if err != nil {
		return
	}
	// FATAL level，退出程序
	os.Exit(1)
}

// 内置打印方法，包含日志级别、日志信息、日志属性
func (l *Logger) print(level Level, msg string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}
	// 匿名结构体，包含日志级别、时间、日志信息、日志属性、堆栈信息
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    msg,
		Properties: properties,
	}

	// 如果日志级别为ERROR以及FATAL，添加堆栈信息
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte
	// 将结构体转换为json字符串，如果转换失败，将日志级别和错误信息写入日志
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 将日志写入输出流
	return l.out.Write(append(line, '\n'))
}

// Write 实现io.Writer接口，将日志写入输出流
func (l *Logger) Write(msg []byte) (n int, err error) {
	return l.print(LevelInfo, string(msg), nil)
}
