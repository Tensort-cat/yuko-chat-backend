package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

// OpenSession 打开会话
func OpenSession(c *gin.Context) {
	var req request.OpenSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}

	msg, ret, sessionId := service.SessionService.OpenSession(req)
	JsonBack(c, msg, ret, sessionId)
}

// GetUserSessionList 获取用户会话列表
func GetUserSessionList(c *gin.Context) {
	var req request.OwnlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, sessions := service.SessionService.GetUserSessionList(req)
	JsonBack(c, message, ret, sessions)
}

// GetGroupSessionList 获取用户群聊列表
func GetGroupSessionList(c *gin.Context) {
	var req request.OwnlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, groupList := service.SessionService.GetGroupSessionList(req)
	JsonBack(c, message, ret, groupList)
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {
	var req request.DeleteSessionRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.SessionService.DeleteSession(req)
	JsonBack(c, message, ret, nil)
}

// CheckOpenSessionAllowed 检查是否可以打开会话
func CheckOpenSessionAllowed(c *gin.Context) {
	var req request.CreateSessionRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, res := service.SessionService.CheckOpenSessionAllowed(req)
	JsonBack(c, message, ret, res)
}
