package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service"
	"yuko_chat/pkg/constant"

	"github.com/gin-gonic/gin"
)

// GetMessageList 获取聊天记录
func GetMessageList(c *gin.Context) {
	var req request.GetMessageListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, data := service.MessageService.GetMessageList(req)
	JsonBack(c, message, ret, data)
}

// GetGroupMessageList 获取群聊消息记录
func GetGroupMessageList(c *gin.Context) {
	var req request.GetGroupMessageListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, data := service.MessageService.GetGroupMessageList(req)
	JsonBack(c, message, ret, data)
}

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	message, ret := service.MessageService.UploadAvatar(c)
	JsonBack(c, message, ret, nil)
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	message, ret := service.MessageService.UploadFile(c)
	JsonBack(c, message, ret, nil)
}
