package test

import (
	"testing"
	"yuko_chat/pkg/zlog"
)

func TestInfo(t *testing.T) {
	zlog.Info("我是zlog.info()")
}

func TestWarn(t *testing.T) {
	zlog.Warn("我是zlog.Warn()")
}

func TestError(t *testing.T) {
	zlog.Error("我是zlog.Error()")
}
