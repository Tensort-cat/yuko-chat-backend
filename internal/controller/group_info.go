package controller

import (
	"net/http"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/service"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

// 创建群聊
func CreateGroup(ctx *gin.Context) {
	var req request.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.CreateGroup(req)
	JsonBack(ctx, msg, ret, nil)
}

// 获取我创建的群聊列表
func GetMyGroups(ctx *gin.Context) {
	var req request.OwnlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, data := service.GroupInfoService.GetMyGroups(req)
	JsonBack(ctx, msg, ret, data)
}

// CheckGroupAddMode 检查群聊加群方式
// 该api接口用于在申请入群时，判断是否可以直接加入，还是等待群主审核。
// 如果是直接加入的话，就不会涉及后续新建ContactApply联系人申请记录；
// 如果是群主审核，就会涉及新建ContactApply联系人申请记录，在群主的管理组件会进一步的审核
func CheckGroupAddMode(ctx *gin.Context) {
	var req request.CheckGroupAddModeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, addMode := service.GroupInfoService.CheckGroupAddMode(req)
	JsonBack(ctx, msg, ret, addMode)
}

// EnterGroupDirectly 直接入群
func EnterGroupDirectly(ctx *gin.Context) {
	var req request.EnterGroupDirectlyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.EnterGroupDirectly(req)
	JsonBack(ctx, msg, ret, nil)
}

// LeaveGroup 退出群聊
func LeaveGroup(ctx *gin.Context) {
	var req request.LeaveGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.LeaveGroup(req)
	JsonBack(ctx, msg, ret, nil)
}

// DismissGroup 解散群聊
func DismissGroup(ctx *gin.Context) {
	var req request.DismissGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.DismissGroup(req)
	JsonBack(ctx, msg, ret, nil)
}

// GetGroupInfo 获取群聊详情
func GetGroupInfo(ctx *gin.Context) {
	var req request.GetGroupInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, groupInfo := service.GroupInfoService.GetGroupInfo(req)
	JsonBack(ctx, msg, ret, groupInfo)
}

// UpdateGroupInfo 更新群聊信息
func UpdateGroupInfo(ctx *gin.Context) {
	var req request.UpdateGroupInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.UpdateGroupInfo(req)
	JsonBack(ctx, msg, ret, nil)
}

// GetGroupMembers 获取群成员列表
func GetGroupMembers(ctx *gin.Context) {
	var req request.GetGroupMembersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret, members := service.GroupInfoService.GetGroupMembers(req)
	JsonBack(ctx, msg, ret, members)
}

// RemoveGroupMembers 移除群成员
func RemoveGroupMembers(ctx *gin.Context) {
	var req request.RemoveGroupMembersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	msg, ret := service.GroupInfoService.RemoveGroupMembers(req)
	JsonBack(ctx, msg, ret, nil)
}

// GetGroupInfoList 获取群聊列表 - 管理员
func GetGroupInfoList(c *gin.Context) {
	message, ret, groupList := service.GroupInfoService.GetGroupInfoList()
	JsonBack(c, message, ret, groupList)
}

// DeleteGroups 删除列表中群聊 - 管理员
func DeleteGroups(c *gin.Context) {
	var req request.DeleteGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.GroupInfoService.DeleteGroups(req)
	JsonBack(c, message, ret, nil)
}

// SetGroupsStatus 设置群聊是否启用
func SetGroupsStatus(c *gin.Context) {
	var req request.SetGroupsStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": constant.SYS_ERR_CODE,
			"msg":  constant.SYS_ERR_MSG,
		})
		return
	}
	message, ret := service.GroupInfoService.SetGroupsStatus(req)
	JsonBack(c, message, ret, nil)
}
