package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

// 获取联系人列表
func GetContactList(ctx *gin.Context) {
	var req request.GetContactListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, contactList := service.UserContactService.GetContactList(req)
	JsonBack(ctx, msg, ret, contactList)
}

// LoadMyJoinedGroup 获取我加入的群聊
func LoadMyJoinedGroup(ctx *gin.Context) {
	var req request.GetContactListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, groupList := service.UserContactService.LoadMyJoinedGroup(req)
	JsonBack(ctx, msg, ret, groupList)
}

// GetContactInfo 获取联系人信息
func GetContactInfo(ctx *gin.Context) {
	var req request.GetContactInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, contactInfo := service.UserContactService.GetContactInfo(req)
	JsonBack(ctx, msg, ret, contactInfo)
}

// DeleteContact 删除联系人
func DeleteContact(ctx *gin.Context) {
	var req request.DeleteContactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.UserContactService.DeleteContact(req)
	JsonBack(ctx, msg, ret, nil)
}

// ApplyContact 申请添加联系人 (用户或群聊)
func ApplyContact(ctx *gin.Context) {
	var req request.ApplyContactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.UserContactService.ApplyContact(req)
	JsonBack(ctx, msg, ret, nil)
}

// GetNewContactList 获取新的联系人申请列表
func GetNewContactList(ctx *gin.Context) {
	var req request.OwnlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, data := service.UserContactService.GetNewContactList(req)
	JsonBack(ctx, message, ret, data)
}

// PassContactApply 通过联系人申请
func PassContactApply(c *gin.Context) {
	var req request.PassContactApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserContactService.PassContactApply(req)
	JsonBack(c, message, ret, nil)
}

// BlackContact 拉黑联系人
func BlackContact(ctx *gin.Context) {
	var req request.BlackContactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserContactService.BlackContact(req)
	JsonBack(ctx, message, ret, nil)
}

// CancelBlackContact 解除拉黑联系人
func CancelBlackContact(c *gin.Context) {
	var req request.BlackContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserContactService.CancelBlackContact(req)
	JsonBack(c, message, ret, nil)
}

// GetAddGroupList 获取新的群聊申请列表
func GetAddGroupList(c *gin.Context) {
	var req request.AddGroupListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret, data := service.UserContactService.GetAddGroupList(req)
	JsonBack(c, message, ret, data)
}

// RefuseContactApply 拒绝联系人申请
func RefuseContactApply(c *gin.Context) {
	var req request.PassContactApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    constant.SYS_ERR_CODE,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserContactService.RefuseContactApply(req)
	JsonBack(c, message, ret, nil)
}

// BlackApply 拉黑申请
func BlackApply(c *gin.Context) {
	var req request.BlackApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.UserContactService.BlackApply(req)
	JsonBack(c, message, ret, nil)
}
