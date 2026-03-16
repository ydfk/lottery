/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 16:40:41
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 15:31:56
 */
package logger

import (
	"fmt"
	"go-fiber-starter/pkg/util"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger

func Init() error {

	if err := util.EnsureDir("log"); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	lumberjacklogger := &lumberjack.Logger{
		Filename:   "./log/log.json",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	// 1. 文件日志配置 - 不带颜色
	fileCfg := zap.NewProductionEncoderConfig()
	fileCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileCfg.EncodeLevel = zapcore.CapitalLevelEncoder // 不带颜色的大写级别

	// 2. 控制台配置 - 带颜色
	consoleCfg := zap.NewDevelopmentEncoderConfig()
	consoleCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 带颜色的级别
	consoleCfg.EncodeCaller = zapcore.ShortCallerEncoder
	consoleCfg.EncodeName = func(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("\x1b[36m" + loggerName + "\x1b[0m") // 青色，ANSI转义码
	}

	// 文件输出编码器 - 使用无颜色的配置
	fileEncoder := zapcore.NewJSONEncoder(fileCfg)

	// 控制台输出编码器 - 使用带颜色的配置
	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)

	// 创建多核心 - 同时写入文件和控制台
	core := zapcore.NewTee(
		zapcore.NewCore(
			fileEncoder,                       // 文件编码设置
			zapcore.AddSync(lumberjacklogger), // 输出到文件
			zap.InfoLevel,                     // 日志等级
		),
		zapcore.NewCore(
			consoleEncoder,             // 控制台编码设置
			zapcore.AddSync(os.Stdout), // 输出到控制台
			zap.InfoLevel,              // 日志等级
		),
	)

	// 添加调用者信息，便于调试
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	defer log.Sync()

	Logger = log.Sugar()
	Logger.Infof("日志系统初始化完成")
	return nil
}

// FiberLogWriter 实现 io.Writer 接口，用于接收 fiber 日志
type FiberLogWriter struct{}

// Write 实现 io.Writer 接口，将 fiber 的日志输出重定向到自定义 logger
func (w FiberLogWriter) Write(p []byte) (n int, err error) {
	// 移除末尾的换行符
	logLine := string(p)
	if len(logLine) > 0 && logLine[len(logLine)-1] == '\n' {
		logLine = logLine[:len(logLine)-1]
	}

	Info("%s", logLine)
	return len(p), nil
}

// GetFiberLogWriter 返回 fiber 日志适配器
func GetFiberLogWriter() io.Writer {
	return &FiberLogWriter{}
}

// Debug 输出调试日志
func Debug(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debugf(format, args...)
	}
}

// Info 输出信息日志
func Info(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(format, args...)
	}
}

// Warn 输出警告日志
func Warn(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(format, args...)
	}
}

// Error 输出错误日志
func Error(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
	}
}

// Fatal 输出致命错误日志并退出
func Fatal(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Fatalf(format, args...)
	}
}
