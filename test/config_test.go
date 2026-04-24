package test

import (
	"fmt"
	"testing"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
	"yuko_chat/pkg/zlog"

	"go.uber.org/zap"
)

func TestCfg(t *testing.T) {
	// 初始化配置信息
	if err := config.InitConfig(); err != nil {
		zlog.Error("初始化配置信息失败", zap.String("error", err.Error()))
		return
	}

	// 初始化 MySQL 连接
	if err := dao.InitDB(); err != nil {
		zlog.Error("初始化 MySQL 连接失败", zap.String("error", err.Error()))
		return
	}

	// 初始化 Redis 连接
	if err := dao.InitRedis(); err != nil {
		zlog.Error("初始化 Redis 连接失败", zap.String("error", err.Error()))
		return
	}

	// 初始化 Kafka 连接
	if err := dao.InitKafka(); err != nil {
		zlog.Error("初始化 Kafka 连接失败", zap.String("error", err.Error()))
		return
	}

	fmt.Println(config.Cfg.StaticSrcConfig.StaticAvatarPath)
	fmt.Println(config.Cfg.StaticSrcConfig.StaticFilePath)
}
