package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type MysqlConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type LogConfig struct {
	LogPath string `toml:"logPath"`
}

type KafkaConfig struct {
	HostPort  string        `toml:"hostPort"`
	Topic     string        `toml:"topic"`
	Partition int           `toml:"partition"`
	Timeout   time.Duration `toml:"timeout"`
}

type EmailConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	From     string `toml:"from"`
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MainConfig      MainConfig      `toml:"mainConfig"`
	MysqlConfig     MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     RedisConfig     `toml:"redisConfig"`
	LogConfig       LogConfig       `toml:"logConfig"`
	KafkaConfig     KafkaConfig     `toml:"kafkaConfig"`
	EmailConfig     EmailConfig     `toml:"emailConfig"`
	StaticSrcConfig StaticSrcConfig `toml:"staticSrcConfig"`
}

var Cfg *Config

func InitConfig() error {
	cfg := new(Config)
	if _, err := toml.DecodeFile("D:\\dev_soft\\Go_WorkSpace\\project\\YukoChat\\configs\\config.toml", cfg); err != nil {
		log.Fatal(err.Error())
		return err
	}
	Cfg = cfg
	return nil
}
