package dao

import (
	"fmt"
	"yuko_chat/internal/config"
	"yuko_chat/pkg/zlog"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 初始化数据库连接
func InitDB() error {
	host := config.Cfg.MysqlConfig.Host
	port := config.Cfg.MysqlConfig.Port
	user := config.Cfg.MysqlConfig.User
	password := config.Cfg.MysqlConfig.Password
	database := config.Cfg.MysqlConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zlog.Error("Failed to connect to MySQL", zap.Error(err))
		return err
	}
	DB = db
	return nil
}

func CloseDB() error {
	db_, _ := DB.DB()
	err := db_.Close()
	return err
}
