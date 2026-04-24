package zlog

import (
	"os"
	"path"
	"runtime"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 简单封装一下对 zap 日志库的使用
// 使用方式：
// zlog.Debug("hello", zap.String("name", "Kevin"), zap.Any("arbitraryObj", dummyObject))
// zlog.Info("hello", zap.String("name", "Kevin"), zap.Any("arbitraryObj", dummyObject))
// zlog.Warn("hello", zap.String("name", "Kevin"), zap.Any("arbitraryObj", dummyObject))
var logger *zap.Logger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	// 设置日志记录中时间的格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 日志Encoder 还是JSONEncoder，把日志行格式化成JSON格式的
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	fileWriteSyncer := getFileLogWriter()

	core := zapcore.NewTee(
		// 同时向控制台和文件写日志， 生产环境记得把控制台写入去掉，日志记录的基本是Debug 及以上，生产环境记得改成Info
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel), // 写控制台
		zapcore.NewCore(encoder, fileWriteSyncer, zapcore.DebugLevel),            // 写文件
	)

	// 生产环境版本
	// core := zapcore.NewTee(
	// 	zapcore.NewCore(encoder, fileWriteSyncer, zapcore.InfoLevel), // 写文件
	// )

	logger = zap.New(core)
}

// Zap 本身不支持日志切割，可以借助另外一个库 lumberjack 协助完成切割
func getFileLogWriter() (writeSyncer zapcore.WriteSyncer) {
	// 使用 lumberjack 实现 log rotate
	// 文件超过 100MB 文件名会改成“原文件名 + 时间戳 + 原扩展名”的形式
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "../logs/app.log",
		MaxSize:    100, // 单个文件最大100M
		MaxBackups: 60,  // 多于 60 个日志文件后，清理较旧的日志
		MaxAge:     1,   // 一天一切割
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func Info(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Error(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Warn(message, fields...)
}

func getCallerInfoForLog() (callerFields []zap.Field) {

	pc, file, line, ok := runtime.Caller(2) // 回溯两层，拿到写日志的调用方的函数信息
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName) //Base函数返回路径的最后一个元素，只保留函数名

	callerFields = append(callerFields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))
	return
}
