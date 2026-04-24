package main

import (
	"fmt"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/route"
	"yuko_chat/internal/service/chat"
	"yuko_chat/pkg/zlog"

	"go.uber.org/zap"
)

func main() {
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
	defer dao.CloseDB()

	// 初始化 Redis 连接
	if err := dao.InitRedis(); err != nil {
		zlog.Error("初始化 Redis 连接失败", zap.String("error", err.Error()))
		return
	}
	defer dao.CloseRedis()

	// 初始化 Kafka 连接
	if err := dao.InitKafka(); err != nil {
		zlog.Error("初始化 Kafka 连接失败", zap.String("error", err.Error()))
		return
	}
	defer dao.CloseKafka()

	// 启动 Server
	go chat.Myserver.Start()

	// 开启 web 服务
	port := fmt.Sprintf(":%d", config.Cfg.MainConfig.Port)
	zlog.Info("backend port", zap.String("port", port))
	route.InitRoute()
	if err := route.GE.Run(port); err != nil {
		zlog.Error("启动 Web 服务失败", zap.String("error", err.Error()))
		return
	}
}
