package service

import (
	"encoding/json"
	"errors"
	"slices"
	"time"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/dto/respond"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	contact_enum "yuko_chat/pkg/enum/contact"
	group_enum "yuko_chat/pkg/enum/group"
	"yuko_chat/pkg/util"
	"yuko_chat/pkg/zlog"

	"gorm.io/gorm"
)

type groupInfoService struct {
}

var GroupInfoService = new(groupInfoService)

func (s *groupInfoService) CreateGroup(req request.CreateGroupRequest) (string, int) {
	group := model.GroupInfo{
		Uuid:      util.GenUUID("G"),
		OwnerId:   req.OwnerId,
		Name:      req.Name,
		Notice:    req.Notice,
		AddMode:   req.AddMode,
		Avatar:    req.Avatar,
		CreatedAt: time.Now(),
		MemberCnt: 1,
	}

	// 群聊一开始只有群主一个成员，成员列表是一个字符串数组的json格式
	members := []string{req.OwnerId}
	membersBytes, err := json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return "创建群聊失败", -2
	}
	group.Members = membersBytes

	err = dao.DB.Create(&group).Error
	if err != nil {
		return "创建群聊失败", -2
	}

	// 创建好群聊后，还要创建一个 user_contact 记录，表示群主加入了这个群聊
	contact := model.UserContact{
		UserId:      req.OwnerId,
		ContactId:   group.Uuid,
		ContactType: contact_enum.GROUP, // 群聊
		Status:      contact_enum.NORMAL,
		UpdateAt:    time.Now(),
		CreatedAt:   time.Now(),
	}
	err = dao.DB.Create(&contact).Error

	return "群聊创建成功", 0
}

func (s *groupInfoService) GetMyGroups(req request.OwnlistRequest) (string, int, []respond.GetMyGroupsRespond) {
	var groups []model.GroupInfo
	err := dao.DB.Where("owner_id = ?", req.OwnerId).Find(&groups).Error
	if err != nil {
		return "获取群聊列表失败", -2, nil
	}

	var res []respond.GetMyGroupsRespond
	for _, group := range groups {
		rep := respond.GetMyGroupsRespond{
			GroupId:   group.Uuid,
			GroupName: group.Name,
			Avatar:    group.Avatar,
		}
		res = append(res, rep)
	}
	return "获取群聊列表成功", 0, res
}

func (s *groupInfoService) CheckGroupAddMode(req request.CheckGroupAddModeRequest) (string, int, int8) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.GroupId).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, -1
	}
	return "获取群聊加群方式成功", 0, group.AddMode

}

func (s *groupInfoService) EnterGroupDirectly(req request.EnterGroupDirectlyRequest) (string, int) {
	// 检查群聊存不存在
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.Owner_id).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 反序列化群成员列表
	var members []string
	err = json.Unmarshal(group.Members, &members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 检查用户是否已经在群里了
	if slices.Contains(members, req.ContactId) {
		return "你已经在群里了", -2
	}

	// 将用户加入群成员列表
	members = append(members, req.ContactId)
	group.MemberCnt++
	// 序列化更新后的成员列表
	membersBytes, err := json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 更新群成员列表
	group.Members = membersBytes
	err = dao.DB.Save(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 创建新的 user_contact 记录，表示用户加入了这个群聊
	contact := model.UserContact{
		UserId:      req.ContactId,
		ContactId:   group.Uuid,
		ContactType: contact_enum.GROUP, // 群聊
		Status:      contact_enum.NORMAL,
		UpdateAt:    time.Now(),
		CreatedAt:   time.Now(),
	}
	err = dao.DB.Create(&contact).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "加入群聊成功", 0

}

// LeaveGroup 退出群聊
func (s *groupInfoService) LeaveGroup(req request.LeaveGroupRequest) (string, int) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.GroupId).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 反序列化群成员列表
	var members []string
	err = json.Unmarshal(group.Members, &members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除用户在群成员列表中的记录
	for i, member := range members {
		if member == req.UserId {
			members = append(members[:i], members[i+1:]...)
			break
		}
	}
	group.MemberCnt--

	// 序列化更新后的成员列表
	membersBytes, err := json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	group.Members = membersBytes

	// 更新群成员列表
	err = dao.DB.Save(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除用户和群聊之间的会话 (软删除)
	var deleteAt gorm.DeletedAt
	deleteAt.Time = time.Now()
	deleteAt.Valid = true
	err = dao.DB.Model(&model.Session{}).
		Where("send_id = ? AND receive_id = ?", req.UserId, req.GroupId).
		Update("deleted_at", deleteAt).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除用户和群聊之间的联系人关系 (软删除)
	err = dao.DB.Model(&model.UserContact{}).
		Where("user_id = ? AND contact_id = ?", req.UserId, req.GroupId).
		Update("deleted_at", deleteAt).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除用户申请加入群聊的记录
	err = dao.DB.Model(&model.ContactApply{}).
		Where("contact_id = ? AND user_id = ?", req.GroupId, req.UserId).
		Update("deleted_at", deleteAt).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "退出群聊成功", 0
}

// DismissGroup 解散群聊
func (s *groupInfoService) DismissGroup(req request.DismissGroupRequest) (string, int) {
	var deleteAt gorm.DeletedAt
	deleteAt.Time = time.Now()
	deleteAt.Valid = true
	err := dao.DB.Model(&model.GroupInfo{}).
		Where("uuid = ?", req.GroupId).
		Update("deleted_at", deleteAt).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除所有相关的会话记录 (软删除)
	var sessions []model.Session
	err = dao.DB.Model(&model.Session{}).
		Where("receive_id = ?", req.GroupId).
		Find(&sessions).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	for _, session := range sessions {
		err = dao.DB.Model(&session).
			Updates(map[string]any{
				"deleted_at": deleteAt,
			}).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	// 删除所有和该群相关的联系方式
	var contacts []model.UserContact
	err = dao.DB.Model(&model.UserContact{}).
		Where("contact_id = ?", req.GroupId).
		Find(&contacts).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	for _, contact := range contacts {
		err = dao.DB.Model(&contact).
			Updates(map[string]any{
				"deleted_at": deleteAt,
			}).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	// 删除所有入群申请
	var applies []model.ContactApply
	err = dao.DB.Model(&model.ContactApply{}).
		Where("contact_id = ?", req.GroupId).
		Find(&applies).
		Error
	if err != nil && err != gorm.ErrRecordNotFound {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	for _, apply := range applies {
		err = dao.DB.Model(&apply).
			Updates(map[string]any{
				"deleted_at": deleteAt,
			}).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	return "解散群聊成功", 0
}

func (s *groupInfoService) GetGroupInfo(req request.GetGroupInfoRequest) (string, int, *respond.GetGroupInfoRespond) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.GroupId).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	rep := &respond.GetGroupInfoRespond{
		Uuid:      group.Uuid,
		Name:      group.Name,
		Notice:    group.Notice,
		MemberCnt: group.MemberCnt,
		OwnerId:   group.OwnerId,
		AddMode:   group.AddMode,
		Status:    group.Status,
		Avatar:    group.Avatar,
	}
	if group.DeletedAt.Valid {
		rep.IsDeleted = true
	} else {
		rep.IsDeleted = false
	}
	return "获取群聊信息成功", 0, rep

}

// UpdateGroupInfo 更新群聊信息
func (s *groupInfoService) UpdateGroupInfo(req request.UpdateGroupInfoRequest) (string, int) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.Uuid).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	if req.Name != "" {
		group.Name = req.Name
	}
	if req.AddMode != -1 {
		group.AddMode = req.AddMode
	}
	if req.Notice != "" {
		group.Notice = req.Notice
	}
	if req.Avatar != "" {
		group.Avatar = req.Avatar
	}

	err = dao.DB.Save(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 修改群聊相关的会话
	var sessions []model.Session
	err = dao.DB.Where("receive_id = ?", group.Uuid).Find(&sessions).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	for _, session := range sessions {
		session.ReceiveName = group.Name
		session.Avatar = group.Avatar
		err = dao.DB.Save(&session).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	return "更新群聊信息成功", 0

}

// GetGroupMembers 获取群聊成员列表
func (s *groupInfoService) GetGroupMembers(req request.GetGroupMembersRequest) (string, int, []respond.GetGroupMembersRespond) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.GroupId).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}
	var members []string
	err = json.Unmarshal(group.Members, &members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var reps []respond.GetGroupMembersRespond
	for _, userId := range members {
		var user model.UserInfo
		err = dao.DB.Where("uuid = ?", userId).First(&user).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, nil
		}
		reps = append(reps, respond.GetGroupMembersRespond{
			UserId:   user.Uuid,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return "获取群聊成员列表成功", 0, reps
}

// RemoveGroupMember 移除群聊成员
func (s *groupInfoService) RemoveGroupMembers(req request.RemoveGroupMembersRequest) (string, int) {
	var group model.GroupInfo
	err := dao.DB.Where("uuid = ?", req.GroupId).First(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 反序列化群成员列表
	var members []string
	err = json.Unmarshal(group.Members, &members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	for _, uuid := range req.UuidList {
		if uuid == req.OwnerId {
			zlog.Error("不能移除群主")
			return "不能移除群主", -1
		}

		for i, member := range members {
			if member == uuid {
				members = append(members[:i], members[i+1:]...)
				group.MemberCnt--
				break
			}
		}

		// 删除用户和群聊之间的会话 (软删除)
		var deleteAt gorm.DeletedAt
		deleteAt.Time = time.Now()
		deleteAt.Valid = true
		err = dao.DB.Model(&model.Session{}).
			Where("send_id = ? AND receive_id = ?", uuid, req.GroupId).
			Update("deleted_at", deleteAt).
			Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 删除用户和群聊之间的联系人关系 (软删除)
		err = dao.DB.Model(&model.UserContact{}).
			Where("user_id = ? AND contact_id = ?", uuid, req.GroupId).
			Update("deleted_at", deleteAt).
			Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 删除用户申请加入群聊的记录
		err = dao.DB.Model(&model.ContactApply{}).
			Where("contact_id = ? AND user_id = ?", req.GroupId, uuid).
			Update("deleted_at", deleteAt).
			Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	// 序列化更新后的成员列表
	membersBytes, err := json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	group.Members = membersBytes

	// 更新群成员列表
	err = dao.DB.Save(&group).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	return "移除群聊成员成功", 0
}

// GetGroupInfoList 获取群聊列表 - 管理员
// 管理员少，而且如果用户更改了，那么管理员会一直频繁删除redis，更新redis，比较麻烦，所以管理员暂时不使用redis缓存
func (g *groupInfoService) GetGroupInfoList() (string, int, []respond.GetGroupListRespond) {
	var groupList []model.GroupInfo
	err := dao.DB.Unscoped().Find(&groupList).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.GetGroupListRespond
	for _, group := range groupList {
		rp := respond.GetGroupListRespond{
			Uuid:    group.Uuid,
			Name:    group.Name,
			OwnerId: group.OwnerId,
			Status:  group.Status,
		}
		if group.DeletedAt.Valid {
			rp.IsDeleted = true
		} else {
			rp.IsDeleted = false
		}

		res = append(res, rp)
	}

	return "获取群聊列表成功", 0, res
}

// DeleteGroups 删除列表中群聊 - 管理员
func (g *groupInfoService) DeleteGroups(req request.DeleteGroupsRequest) (string, int) {
	for _, uuid := range req.UuidList {
		var deleted_at gorm.DeletedAt
		deleted_at.Time = time.Now()
		deleted_at.Valid = true

		err := dao.DB.Model(&model.GroupInfo{}).Where("uuid = ?", uuid).Update("deleted_at", deleted_at).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 删除会话
		var sessions []model.Session
		err = dao.DB.Model(&model.Session{}).Where("receive_id = ?", uuid).Find(&sessions).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		for _, session := range sessions {
			err := dao.DB.Model(&session).Update("deleted_at", deleted_at).Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		// 删除联系方式
		var contactList []model.UserContact
		err = dao.DB.Model(&model.UserContact{}).Where("contact_id = ?", uuid).Find(&contactList).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		for _, contact := range contactList {
			err := dao.DB.Model(&contact).Update("deleted_at", deleted_at).Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		// 删除申请记录
		var applies []model.ContactApply
		err = dao.DB.Model(model.ContactApply{}).Where("contact_id = ?", uuid).Find(&applies).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				zlog.Info("没有申请记录需要删除")
			} else {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		for _, apply := range applies {
			err := dao.DB.Model(&apply).Update("deleted_at", deleted_at).Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}
	}

	return "解散/删除群聊成功", 0
}

// SetGroupsStatus 设置群聊是否启用
func (g *groupInfoService) SetGroupsStatus(req request.SetGroupsStatusRequest) (string, int) {
	var deleted_at gorm.DeletedAt
	deleted_at.Time = time.Now()
	deleted_at.Valid = true

	for _, uuid := range req.UuidList {
		err := dao.DB.Model(&model.GroupInfo{}).Where("uuid = ?", uuid).Update("status", req.Status).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 如果是禁用群聊，要删除所有会话
		if req.Status == group_enum.DISABLE {
			var sessions []model.Session
			err := dao.DB.Model(&model.Session{}).Where("receive_id = ?", uuid).Find(&sessions).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}

			for _, session := range sessions {
				err := dao.DB.Model(&session).Update("deleted_at", deleted_at).Error
				if err != nil {
					zlog.Error(err.Error())
					return constant.SYS_ERR_MSG, -1
				}
			}
		}
	}

	return "设置成功", 0
}
