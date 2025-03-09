package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logger *log.Logger
)

// Init 初始化日志配置
func Init(logDir string) error {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 使用当前日期作为日志文件名
	logFileName := filepath.Join(logDir, fmt.Sprintf("lottery_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 设置日志输出同时写入文件和控制台
	writer := io.MultiWriter(os.Stdout, logFile)

	// 配置日志格式，包含时间、文件位置等信息
	logger = log.New(writer, "", log.Ldate|log.Ltime|log.Lshortfile)

	// 替换标准库的默认logger
	log.SetOutput(writer)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	Info("日志系统初始化完成，日志文件：" + logFileName)
	return nil
}

// Debug 输出调试日志
func Debug(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info 输出信息日志
func Info(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	}
}

// Warn 输出警告日志
func Warn(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[WARN] "+format, v...)
	}
}

// Error 输出错误日志
func Error(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	}
}

// Fatal 输出致命错误日志并退出程序
func Fatal(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[FATAL] "+format, v...)
	}
	os.Exit(1)
}
