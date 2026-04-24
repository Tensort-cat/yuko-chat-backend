package constant

import "time"

const (
	SYS_ERR_CODE    = 500
	SYS_ERR_MSG     = "system error"
	FILE_MAX_SIZE   = 50000
	CHANNEL_SIZE    = 100 // 通道大小
	REDIS_TIMEOUT   = 1 * time.Minute
	SESSION_TIMEOUT = 24 * time.Hour
)
