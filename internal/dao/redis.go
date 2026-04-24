package dao

import (
	"errors"
	"fmt"
	"time"
	"yuko_chat/internal/config"
	"yuko_chat/pkg/zlog"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var redisClient *redis.Client

func InitRedis() error {
	host, port := config.Cfg.RedisConfig.Host, config.Cfg.RedisConfig.Port
	db := config.Cfg.RedisConfig.DB
	pwd := config.Cfg.RedisConfig.Password
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		DB:       db,
		Password: pwd,
	})

	_, err := rc.Ping().Result()
	if err != nil {
		zlog.Error("failed to connect redis", zap.Error(err))
		return err
	}
	redisClient = rc
	return nil
}

func SetKeyEx(key string, value string, expiration time.Duration) error {
	_, err := redisClient.Set(key, value, expiration).Result()
	if err != nil {
		zlog.Error("failed to set redis key", zap.Error(err))
		return err
	}
	return nil
}

// 获取key，如果不存在，err = nil
func GetKey(key string) (string, error) {
	val, err := redisClient.Get(key).Result()
	if err != nil {
		if err == redis.Nil { // key 不存在
			zlog.Info("redis key does not exist", zap.String("key", key))
			return "", nil
		}
		zlog.Error("failed to get redis key", zap.Error(err))
		return "", err
	}
	return val, nil
}

// 获取key，如果不存在，err ≠ nil
func GetKeyNilIsErr(key string) (string, error) {
	value, err := redisClient.Get(key).Result()
	if err != nil {
		zlog.Error("failed to get redis key", zap.Error(err))
		return "", err
	}
	return value, nil
}

// 获取有prefix前缀的key，如果不存在，err ≠ nil
func GetKeyWithPrefixNilIsErr(prefix string) (string, error) {
	keys, err := redisClient.Keys(prefix + "*").Result()
	if err != nil {
		zlog.Error("failed to get redis keys with prefix", zap.String("prefix", prefix), zap.Error(err))
		return "", err
	}
	switch len(keys) {
	case 0: // 没有找到任何匹配的key
		zlog.Info("no redis keys found with prefix", zap.String("prefix", prefix))
		return "", errors.New("no keys found with prefix")
	case 1: // 找到一个匹配的key，返回它
		return keys[0], nil
	default: // 找到了数量大于1的key，查找异常
		zlog.Error("found multiple keys with prefix", zap.String("prefix", prefix))
		return "", errors.New("multiple keys found with prefix")
	}
}

// 获取有suffix后缀的key，如果不存在，err ≠ nil
func GetKeyWithSuffixNilIsErr(suffix string) (string, error) {
	keys, err := redisClient.Keys("*" + suffix).Result()
	if err != nil {
		zlog.Error("failed to get redis keys with suffix", zap.String("suffix", suffix), zap.Error(err))
		return "", err
	}
	switch len(keys) {
	case 0: // 没有找到任何匹配的key
		zlog.Info("no redis keys found with suffix", zap.String("suffix", suffix))
		return "", errors.New("no keys found with suffix")
	case 1: // 找到一个匹配的key，返回它
		return keys[0], nil
	default: // 找到了数量大于1的key，查找异常
		zlog.Error("found multiple keys with suffix", zap.String("suffix", suffix))
		return "", errors.New("multiple keys found with suffix")
	}
}

func DelKeyIfExists(key string) error {
	exists, err := redisClient.Exists(key).Result()
	if err != nil {
		zlog.Error("failed to check if redis key exists", zap.String("key", key), zap.Error(err))
		return err
	}
	if exists == 1 {
		_, err := redisClient.Del(key).Result()
		if err != nil {
			zlog.Error("failed to delete redis key", zap.String("key", key), zap.Error(err))
			return err
		}
	}
	// 无论键是否存在，都不返回错误
	return nil
}

// 如果有该前缀的key，则删除
func DelKeysWithPrefix(prefix string) error {
	keys, err := redisClient.Keys(prefix + "*").Result()
	if err != nil {
		zlog.Error("failed to get redis keys with prefix for deletion", zap.String("prefix", prefix), zap.Error(err))
		return err
	}
	for _, key := range keys {
		_, err := redisClient.Del(key).Result()
		if err != nil {
			zlog.Error("failed to delete redis key with prefix", zap.String("key", key), zap.Error(err))
			return err
		}
	}
	return nil
}

// 如果有该后缀的key，则删除
func DelKeysWithSuffix(suffix string) error {
	keys, err := redisClient.Keys("*" + suffix).Result()
	if err != nil {
		zlog.Error("failed to get redis keys with suffix for deletion", zap.String("suffix", suffix), zap.Error(err))
		return err
	}
	for _, key := range keys {
		_, err := redisClient.Del(key).Result()
		if err != nil {
			zlog.Error("failed to delete redis key with suffix", zap.String("key", key), zap.Error(err))
			return err
		}
	}
	return nil
}

// 删除所有的 key
func FlushAllKeys() error {
	_, err := redisClient.FlushDB().Result()
	if err != nil {
		zlog.Error("failed to flush all redis keys", zap.Error(err))
		return err
	}
	return nil
}

// 关闭 redis 连接
func CloseRedis() error {
	err := redisClient.Close()
	return err
}
