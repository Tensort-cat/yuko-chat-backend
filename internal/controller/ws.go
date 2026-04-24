package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service/chat"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

// WsLogin wss登录 (GET 请求)
func WsLogin(c *gin.Context) {
	clientId := c.Query("client_id")
	if clientId == "" {
		zlog.Error("clientId 获取失败")
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}

	chat.ClientLogin(c, clientId)
}

// WsLogout wss登出
func WsLogout(c *gin.Context) {
	var req request.WsLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := chat.ClientLogout(req.OwnerId)
	JsonBack(c, message, ret, nil)
}
