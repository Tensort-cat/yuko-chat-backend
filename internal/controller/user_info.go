package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

// 普通用户功能
func Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, userInfo, ret := service.UserInfoService.Login(req.Telephone, req.Password)
	JsonBack(ctx, msg, ret, userInfo)
}

func Register(ctx *gin.Context) {
	var req request.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, userInfo, ret := service.UserInfoService.Register(req)
	JsonBack(ctx, msg, ret, userInfo)
}

func SendVerificationCode(ctx *gin.Context) {
	var req request.SendVerificationCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.UserInfoService.SendVerificationCode(req.Email)
	JsonBack(ctx, msg, ret, nil)
}

func GetUserInfo(ctx *gin.Context) {
	var req request.GetUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, userInfo, ret := service.UserInfoService.GetUserInfo(req.Uuid)
	JsonBack(ctx, msg, ret, userInfo)
}

func UpdateUserInfo(ctx *gin.Context) {
	var req request.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.UserInfoService.UpdateUserInfo(req)
	JsonBack(ctx, msg, ret, nil)
}

// 管理员功能
// GetUserInfoList 获取用户列表
func GetUserInfoList(c *gin.Context) {
	var req request.GetUserInfoListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, userList := service.UserInfoService.GetUserInfoList(req)
	JsonBack(c, message, ret, userList)
}

// AbleUsers 启用用户
func AbleUsers(c *gin.Context) {
	var req request.AbleUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserInfoService.AbleUsers(req)
	JsonBack(c, message, ret, nil)
}

// DisableUsers 禁用用户
func DisableUsers(c *gin.Context) {
	var req request.AbleUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserInfoService.DisableUsers(req)
	JsonBack(c, message, ret, nil)
}

// DeleteUsers 删除用户
func DeleteUsers(c *gin.Context) {
	var req request.AbleUsersRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserInfoService.DeleteUsers(req)
	JsonBack(c, message, ret, nil)
}

// SetAdmin 设置管理员
func SetAdmin(c *gin.Context) {
	var req request.AbleUsersRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserInfoService.SetAdmin(req)
	JsonBack(c, message, ret, nil)
}

// SmsLogin 短信验证码登录
func SmsLogin(c *gin.Context) {
	var req request.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, userInfo, ret := service.UserInfoService.SmsLogin(req)
	JsonBack(c, msg, ret, userInfo)
}
