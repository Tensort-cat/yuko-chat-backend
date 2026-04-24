package util

import (
	"fmt"
	"sync"
	"time"
)

const (
	snowflakeEpoch     int64 = 1704067200000
	snowflakeNodeID    int64 = 1
	snowflakeNodeBits  uint8 = 10
	snowflakeStepBits  uint8 = 12
	snowflakeNodeMax   int64 = -1 ^ (-1 << snowflakeNodeBits)
	snowflakeStepMask  int64 = -1 ^ (-1 << snowflakeStepBits)
	snowflakeTimeShift       = snowflakeNodeBits + snowflakeStepBits
	snowflakeNodeShift       = snowflakeStepBits
)

type snowflakeGenerator struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	sequence  int64
}

var userUUIDGenerator = newSnowflakeGenerator(snowflakeNodeID)

func newSnowflakeGenerator(nodeID int64) *snowflakeGenerator {
	if nodeID < 0 || nodeID > snowflakeNodeMax {
		nodeID = 0
	}
	return &snowflakeGenerator{nodeID: nodeID}
}

func (g *snowflakeGenerator) nextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := currentMillis()
	if now == g.timestamp {
		g.sequence = (g.sequence + 1) & snowflakeStepMask
		if g.sequence == 0 {
			now = g.waitNextMillis(now)
		}
	} else {
		g.sequence = 0
	}

	g.timestamp = now
	return ((now - snowflakeEpoch) << snowflakeTimeShift) | (g.nodeID << snowflakeNodeShift) | g.sequence
}

func (g *snowflakeGenerator) waitNextMillis(lastTimestamp int64) int64 {
	now := currentMillis()
	for now <= lastTimestamp {
		now = currentMillis()
	}
	return now
}

func currentMillis() int64 {
	return time.Now().UnixMilli()
}

func GenUUID(prefix string) string {
	return prefix + fmt.Sprintf("%019d", userUUIDGenerator.nextID())
}
