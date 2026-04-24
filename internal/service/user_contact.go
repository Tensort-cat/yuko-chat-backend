package service

import (
	"encoding/json"
	"errors"
	"time"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/dto/respond"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	apply_enum "yuko_chat/pkg/enum/apply"
	contact_enum "yuko_chat/pkg/enum/contact"
	group_enum "yuko_chat/pkg/enum/group"
	user_enum "yuko_chat/pkg/enum/user"
	"yuko_chat/pkg/util"
	"yuko_chat/pkg/zlog"

	"gorm.io/gorm"
)

var UserContactService = new(userContactService)

type userContactService struct {
}

func (s *userContactService) GetContactList(req request.GetContactListRequest) (string, int, []respond.UserListResponse) {
	var contactList []model.UserContact
	err := dao.DB.Model(&model.UserContact{}).Where("user_id = ? and status not in (?, ?)",
		req.Owner_id, contact_enum.DELETE, contact_enum.BE_BLACK).
		Find(&contactList).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			msg := "目前没有联系人"
			zlog.Info(msg)
			return msg, 0, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.UserListResponse
	for _, contact := range contactList {
		// 首先联系人是用户不是群聊
		if contact.ContactId[0] == 'U' {
			// 获取用户信息
			var user model.UserInfo
			err := dao.DB.Where("uuid = ?", contact.ContactId).First(&user).Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1, nil
			}
			res = append(res, respond.UserListResponse{
				UserId:   user.Uuid,
				UserName: user.Nickname,
				Avatar:   user.Avatar,
			})
		}
	}

	return "获取联系人列表成功", 0, res
}

// LoadMyJoinedGroup 获取我加入的群聊
func (s *userContactService) LoadMyJoinedGroup(req request.GetContactListRequest) (string, int, []respond.LoadMyJoinedGroupRespond) {
	// 先找出用户的 contact 记录
	var contactList []model.UserContact
	err := dao.DB.Where("user_id = ?", req.Owner_id).Find(&contactList).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	// 找出所有关联的群聊
	var res []respond.LoadMyJoinedGroupRespond
	for _, contact := range contactList {
		if contact.ContactType == contact_enum.GROUP &&
			contact.Status == contact_enum.NORMAL { // 只处理未退出和未被踢出的群聊
			var group model.GroupInfo
			err = dao.DB.Model(&model.GroupInfo{}).
				Where("uuid = ?", contact.ContactId).
				First(&group).
				Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1, nil
			}
			res = append(res, respond.LoadMyJoinedGroupRespond{
				GroupId:   group.Uuid,
				GroupName: group.Name,
				Avatar:    group.Avatar,
			})
		}
	}

	return "获取群聊成功", 0, res

}

// GetContactInfo 获取联系人信息
func (s *userContactService) GetContactInfo(req request.GetContactInfoRequest) (string, int, *respond.GetContactInfoRespond) {
	if req.ContactId[0] == 'G' { // 是群聊
		var group model.GroupInfo
		err := dao.DB.First(&group, "uuid = ?", req.ContactId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, nil
		}
		if group.Status != group_enum.DISABLE {
			return "获取联系人成功", 0, &respond.GetContactInfoRespond{
				ContactId:        group.Uuid,
				ContactName:      group.Name,
				ContactAvatar:    group.Avatar,
				ContactNotice:    group.Notice,
				ContactAddMode:   group.AddMode,
				ContactMembers:   group.Members,
				ContactMemberCnt: group.MemberCnt,
				ContactOwnerId:   group.OwnerId,
			}
		}
		msg := "群聊处于禁用状态"
		zlog.Error(msg)
		return msg, -2, nil
	} else { // 是用户
		var user model.UserInfo
		err := dao.DB.First(&user, "uuid = ?", req.ContactId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, nil
		}
		if user.Status != user_enum.DISABLE {
			return "获取联系人成功", 0, &respond.GetContactInfoRespond{
				ContactId:        user.Uuid,
				ContactName:      user.Nickname,
				ContactAvatar:    user.Avatar,
				ContactBirthday:  user.Birthday,
				ContactEmail:     user.Email,
				ContactPhone:     user.Telephone,
				ContactGender:    user.Gender,
				ContactSignature: user.Signature,
			}
		}
		msg := "用户处于禁用状态"
		zlog.Info(msg)
		return msg, -2, &respond.GetContactInfoRespond{}
	}
}

// DeleteContact 删除联系人
func (s *userContactService) DeleteContact(req request.DeleteContactRequest) (string, int) {
	// 涉及到联系，会话，申请记录三个表，且每对关系都是对应一对记录的
	var deletedAt gorm.DeletedAt
	deletedAt.Time = time.Now()
	deletedAt.Valid = true

	// 删联系人
	err := dao.DB.Model(&model.UserContact{}).
		Where("user_id = ? and contact_id = ?", req.OwnerId, req.ContactId).
		Updates(map[string]any{
			"deleted_at": deletedAt,
			"status":     contact_enum.DELETE,
		}).Error

	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	err = dao.DB.Model(&model.UserContact{}).
		Where("user_id = ? and contact_id = ?", req.ContactId, req.OwnerId).
		Updates(map[string]any{
			"deleted_at": deletedAt,
			"status":     contact_enum.BE_DELETE,
		}).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删会话
	err = dao.DB.Model(&model.Session{}).
		Where("send_id = ? and receive_id = ?", req.OwnerId, req.ContactId).
		Update("deleted_at", deletedAt).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	err = dao.DB.Model(&model.Session{}).
		Where("send_id = ? and receive_id = ?", req.ContactId, req.OwnerId).
		Update("deleted_at", deletedAt).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删申请记录
	err = dao.DB.Model(&model.ContactApply{}).
		Where("user_id = ? and contact_id = ?", req.OwnerId, req.ContactId).
		Update("deleted_at", deletedAt).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	err = dao.DB.Model(&model.ContactApply{}).
		Where("user_id = ? and contact_id = ?", req.ContactId, req.OwnerId).
		Update("deleted_at", deletedAt).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "删除联系人成功", 0
}

// ApplyContact 申请添加联系人 (用户或群聊)
func (s *userContactService) ApplyContact(req request.ApplyContactRequest) (string, int) {
	switch req.ContactId[0] {
	case 'U': // 申请加好友
		var user model.UserInfo
		// 判断用户是否存在
		err := dao.DB.First(&user, "uuid = ?", req.ContactId).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				zlog.Info("用户不存在")
				return "用户不存在", -2
			}
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 用户是否被禁用
		if user.Status == user_enum.DISABLE {
			zlog.Info("该用户已被封禁")
			return "该用户已被封禁", -2
		}

		// 创建申请记录
		var apply model.ContactApply
		// 由于申请可能发过了，是重复发的申请，检查一下
		err = dao.DB.
			Where("user_id = ? and contact_id = ?", req.OwnerId, req.ContactId).
			First(&apply).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { // 没有申请过，创建申请记录
				apply = model.ContactApply{
					Uuid:        util.GenUUID("A"),
					UserId:      req.OwnerId,
					ContactId:   req.ContactId,
					ContactType: apply_enum.USER,
					Status:      apply_enum.APPLING,
					Message:     req.Message,
					LastApplyAt: time.Now(),
				}
				err = dao.DB.Create(&apply).Error
				if err != nil {
					zlog.Error(err.Error())
					return constant.SYS_ERR_MSG, -1
				}
				return "申请成功", 0
			} else {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}
		// err = nil 说明已经申请过了
		apply.LastApplyAt = time.Now()
		err = dao.DB.Save(&apply).Error
		if err != nil {
			return constant.SYS_ERR_MSG, -1
		}
		if apply.Status == apply_enum.BE_BLACK { // 被拉黑了
			zlog.Info("你被拉黑了")
			return "你被拉黑了", -2
		}
		zlog.Info("请勿重复申请")
		return "请勿重复申请", -2

	case 'G': // 申请入群
		var group model.GroupInfo
		// 判断群是否存在
		err := dao.DB.First(&group, "uuid = ?", req.ContactId).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { //群聊不存在
				zlog.Info("群聊不存在")
				return "群聊不存在", -2
			}
			return constant.SYS_ERR_MSG, -1
		}

		// 群存在，但可能被禁用或解散了
		switch group.Status {
		case group_enum.DISABLE: // 群聊被封禁
			zlog.Info("该群已被封禁")
			return "该群已被封禁", -2
		case group_enum.DISSMIS:
			zlog.Info("该群已解散")
			return "该群已解散", -2
		}

		// 现在确定群存在且状态正常了
		var apply model.ContactApply
		// 由于申请可能发过了，是重复发的申请，检查一下
		err = dao.DB.
			Where("user_id = ? and contact_id = ?", req.OwnerId, req.ContactId).
			First(&apply).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { // 确实没申请过
				apply = model.ContactApply{
					Uuid:        util.GenUUID("A"),
					UserId:      req.OwnerId,
					ContactId:   req.ContactId,
					ContactType: apply_enum.GROUP,
					Status:      apply_enum.APPLING,
					Message:     req.Message,
					LastApplyAt: time.Now(),
				}
				err = dao.DB.Create(&apply).Error
				if err != nil {
					zlog.Error(err.Error())
					return constant.SYS_ERR_MSG, -1
				}
				return "申请成功", 0
			} else {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		// 申请过了
		apply.LastApplyAt = time.Now()
		err = dao.DB.Create(&apply).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		return "请勿重复申请", -2

	default:
		return "用户/群聊不存在", -2
	}
}

// GetNewContactList 获取新的联系人申请列表
func (s *userContactService) GetNewContactList(req request.OwnlistRequest) (string, int, []respond.NewContactListRespond) {
	var applies []model.ContactApply
	err := dao.DB.Model(&model.ContactApply{}).
		Where("contact_id = ? and status = ?", req.OwnerId, apply_enum.APPLING).
		Find(&applies).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("没有申请中的联系人")
			return "没有申请中的联系人", -2, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.NewContactListRespond
	for _, apply := range applies {
		var msg string
		if apply.Message == "" {
			msg = "申请理由: 无"
		} else {
			msg = "申请理由: " + apply.Message
		}
		newContact := respond.NewContactListRespond{
			ContactId: apply.UserId,
			Message:   msg,
		}

		// 找出申请用户的昵称和头像
		var user model.UserInfo
		err := dao.DB.First(&user, "uuid = ?", apply.UserId).Error
		if err != nil {
			return constant.SYS_ERR_MSG, -1, nil
		}
		newContact.ContactName = user.Nickname
		newContact.ContactAvatar = user.Avatar

		res = append(res, newContact)
	}

	return "获取成功", 0, res
}

func (s *userContactService) PassContactApply(req request.PassContactApplyRequest) (string, int) {
	// ownerId 如果是用户的话就是登录用户，如果是群聊的话就是群聊id
	// 先看看是加好友还是加群
	var apply model.ContactApply
	err := dao.DB.Where("contact_id = ? and user_id = ?", req.OwnerId, req.ContactId).First(&apply).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	switch req.OwnerId[0] {
	case 'U': // 好友申请
		// 对方可能已经被封禁了
		var user model.UserInfo
		err = dao.DB.First(&user, "uuid = ?", req.ContactId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		if user.Status == user_enum.DISABLE { // 对方已经被封了
			zlog.Info("对方已被封禁")
			return "对方已被封禁", -2
		}

		// 标记申请已经通过
		apply.Status = apply_enum.ACCEPT
		err := dao.DB.Save(&apply).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 创建新的联系记录 (注意用户间的 UserContact 是一对)
		contactMy := model.UserContact{
			UserId:      req.ContactId,
			ContactId:   req.OwnerId,
			ContactType: contact_enum.USER,
			Status:      contact_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		contactYour := model.UserContact{
			UserId:      req.OwnerId,
			ContactId:   req.ContactId,
			ContactType: contact_enum.USER,
			Status:      contact_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		err = dao.DB.Create(&contactMy).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		err = dao.DB.Create(&contactYour).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		return "添加好友成功", 0

	case 'G': // 入群申请
		// 先看群有没有被解散或封禁
		var group model.GroupInfo
		err := dao.DB.First(&group, "uuid = ?", req.OwnerId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		switch group.Status {
		case group_enum.DISABLE: // 群被封禁了
			zlog.Info("群已被封禁")
			return "群已被封禁", -2
		case group_enum.DISSMIS: // 群被解散了
			zlog.Info("群已被解散")
			return "群已被解散", -2
		}

		// 群正常
		// 将申请标记为通过
		apply.Status = apply_enum.ACCEPT
		err = dao.DB.Save(&apply).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 创建联系记录 (注意用户与群聊间的 UserContact 只有一行)
		contact := model.UserContact{
			UserId:      req.ContactId,
			ContactId:   req.OwnerId,
			ContactType: contact_enum.GROUP,
			Status:      contact_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		err = dao.DB.Create(&contact).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 添加群成员
		var members []string
		err = json.Unmarshal(group.Members, &members)
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		members = append(members, req.ContactId)
		group.MemberCnt++
		group.Members, err = json.Marshal(members)
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		err = dao.DB.Save(&group).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		return "入群申请已通过", 0
	}

	return constant.SYS_ERR_MSG, -1
}

// BlackContact 拉黑联系人
func (u *userContactService) BlackContact(req request.BlackContactRequest) (string, int) {
	// 拉黑
	err := dao.DB.Model(&model.UserContact{}).
		Where("user_id = ? and contact_id = ?", req.OwnerId, req.ContactId).
		Updates(map[string]any{
			"status":    contact_enum.BLACK,
			"update_at": time.Now(),
		}).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 被拉黑
	err = dao.DB.Model(&model.UserContact{}).
		Where("user_id = ? and contact_id = ?", req.ContactId, req.OwnerId).
		Updates(map[string]any{
			"status":    contact_enum.BE_BLACK,
			"update_at": time.Now(),
		}).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 删除会话
	var deletedAt gorm.DeletedAt
	deletedAt.Time = time.Now()
	deletedAt.Valid = true
	err = dao.DB.Model(&model.Session{}).
		Where("send_id = ? and receive_id = ?", req.OwnerId, req.ContactId).
		Update("deleted_at", deletedAt).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "已拉黑该联系人", 0
}

// CancelBlackContact 取消拉黑联系人
func (u *userContactService) CancelBlackContact(req request.BlackContactRequest) (string, int) {
	// 因为前端的设定，这里需要判断一下ownerId和contactId是不是有拉黑和被拉黑的状态
	var blackContact model.UserContact
	if res := dao.DB.Where("user_id = ? AND contact_id = ?", req.OwnerId, req.ContactId).First(&blackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	if blackContact.Status != contact_enum.BLACK {
		return "未拉黑该联系人，无需解除拉黑", -2
	}
	var beBlackContact model.UserContact
	if res := dao.DB.Where("user_id = ? AND contact_id = ?", req.ContactId, req.OwnerId).First(&beBlackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	if beBlackContact.Status != contact_enum.BE_BLACK {
		return "该联系人未被拉黑，无需解除拉黑", -2
	}

	// 取消拉黑
	blackContact.Status = contact_enum.NORMAL
	beBlackContact.Status = contact_enum.NORMAL
	if res := dao.DB.Save(&blackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	if res := dao.DB.Save(&beBlackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	return "已解除拉黑该联系人", 0
}

// GetAddGroupList 获取新的加群列表
// 前端已经判断调用接口的用户是群主，也只有群主才能调用这个接口
func (u *userContactService) GetAddGroupList(req request.AddGroupListRequest) (string, int, []respond.AddGroupListRespond) {
	var applies []model.ContactApply
	err := dao.DB.Where("contact_id = ? and status = ?", req.GroupId, apply_enum.APPLING).Find(&applies).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("目前没有加群申请")
			return "目前没有加群申请", 0, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.AddGroupListRespond
	for _, apply := range applies {
		var user model.UserInfo
		err := dao.DB.First(&user, "uuid = ?", apply.UserId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, nil
		}
		var msg string
		if apply.Message == "" {
			msg = "申请理由: 无"
		} else {
			msg = "申请理由: " + apply.Message
		}
		res = append(res, respond.AddGroupListRespond{
			ContactId:     user.Uuid,
			ContactName:   user.Nickname,
			Message:       msg,
			ContactAvatar: user.Avatar,
		})
	}
	return "获取成功", 0, res
}

// RefuseContactApply 拒绝联系人申请
func (u *userContactService) RefuseContactApply(req request.PassContactApplyRequest) (string, int) {
	// ownerId 如果是用户的话就是登录用户，如果是群聊的话就是群聊id
	var apply model.ContactApply
	if res := dao.DB.Where("contact_id = ? AND user_id = ?", req.OwnerId, req.ContactId).First(&apply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	apply.Status = apply_enum.REJECT
	if res := dao.DB.Save(&apply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constant.SYS_ERR_MSG, -1
	}
	if req.OwnerId[0] == 'U' {
		return "已拒绝该联系人申请", 0
	} else {
		return "已拒绝该加群申请", 0
	}

}

func (u *userContactService) BlackApply(req request.BlackApplyRequest) (string, int) {
	var apply model.ContactApply
	err := dao.DB.Where("user_id = ? and contact_id = ?", req.ContactId, req.OwnerId).First(&apply).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	apply.Status = apply_enum.BE_BLACK
	err = dao.DB.Save(&apply).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "已拉黑该申请", 0
}
