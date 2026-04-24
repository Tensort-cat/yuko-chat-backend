package test

import (
	"testing"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
)

func TestGetKeyEx(t *testing.T) {
	config.InitConfig()
	dao.InitRedis()

	v, err := dao.GetKey("message_list:Ufuck111:Ufuck222")
	if err != nil {
		t.Errorf("Error occurred while fetching key: %v", err)
	}
	t.Logf("Retrieved value: %s", v)
}
