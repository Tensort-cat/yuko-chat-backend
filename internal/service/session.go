package service

import (
	"errors"
	"fmt"
	"time"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/dto/respond"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	contact_enum "yuko_chat/pkg/enum/contact"
	group_enum "yuko_chat/pkg/enum/group"
	user_enum "yuko_chat/pkg/enum/user"
	"yuko_chat/pkg/util"
	"yuko_chat/pkg/zlog"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type sessionService struct {
}

var SessionService = new(sessionService)

func (s *sessionService) OpenSession(req request.OpenSessionRequest) (string, int, string) {
	// 先看 redis 有没有
	redisKey := fmt.Sprintf("session:%s:%s", req.SendId, req.ReceiveId)
	sessionId, err := dao.GetKeyNilIsErr(redisKey)
	if err != nil {
		if err == redis.Nil {
			var session model.Session
			err := dao.DB.Where("send_id = ? and receive_id = ?", req.SendId, req.ReceiveId).First(&session).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) { // 创建新会话
					zlog.Info("会话没有找到，创建新会话")
					createReq := request.CreateSessionRequest{
						SendId:    req.SendId,
						ReceiveId: req.ReceiveId,
					}
					return s.CreateSession(createReq)
				}
				return constant.SYS_ERR_MSG, -1, ""
			}
			// 放 redis
			dao.SetKeyEx(redisKey, session.Uuid, constant.SESSION_TIMEOUT)
			return "会话打开成功", 0, session.Uuid
		}
		return constant.SYS_ERR_MSG, -1, ""
	}

	return "会话打开成功", 0, sessionId
}

func (s *sessionService) CreateSession(req request.CreateSessionRequest) (string, int, string) {
	// 取出对方的用户名和头像
	// 看是私聊还是群聊
	var session model.Session
	switch req.ReceiveId[0] {
	case 'U': // 私聊
		{
			var to model.UserInfo
			err := dao.DB.First(&to, "uuid = ?", req.ReceiveId).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					zlog.Info("用户不存在")
					return "用户不存在", 0, ""
				}
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1, ""
			}
			session = model.Session{
				Uuid:        util.GenUUID("S"),
				SendId:      req.SendId,
				ReceiveId:   req.ReceiveId,
				ReceiveName: to.Nickname,
				Avatar:      to.Avatar,
				CreatedAt:   time.Now(),
			}
		}
	case 'G': // 群聊
		{
			var to model.GroupInfo
			err := dao.DB.First(&to, "uuid = ?", req.ReceiveId).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					zlog.Info("群聊不存在")
					return "群聊不存在", 0, ""
				}
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1, ""
			}
			session = model.Session{
				Uuid:        util.GenUUID("S"),
				SendId:      req.SendId,
				ReceiveId:   req.ReceiveId,
				ReceiveName: to.Name,
				Avatar:      to.Avatar,
				CreatedAt:   time.Now(),
			}
		}
	}

	err := dao.DB.Create(&session).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, ""
	}

	return "创建会话成功", 0, session.Uuid
}

// GetUserSessionList 获取用户会话列表
func (s *sessionService) GetUserSessionList(req request.OwnlistRequest) (string, int, []respond.UserSessionListRespond) {
	var sessions []model.Session
	err := dao.DB.Order("created_at desc").Where("send_id = ?", req.OwnerId).Find(&sessions).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("暂时没有用户会话")
			return "暂时没有用户会话", 0, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.UserSessionListRespond
	for _, session := range sessions {
		if session.ReceiveId[0] == 'U' {
			res = append(res, respond.UserSessionListRespond{
				SessionId: session.Uuid,
				Avatar:    session.Avatar,
				UserId:    session.ReceiveId,
				Username:  session.ReceiveName,
			})
		}
	}

	return "获取用户会话列表成功", 0, res
}

// GetGroupSessionList 获取群聊会话列表
func (s *sessionService) GetGroupSessionList(req request.OwnlistRequest) (string, int, []respond.GroupSessionListRespond) {
	var sessions []model.Session
	err := dao.DB.Order("created_at desc").Where("send_id = ?", req.OwnerId).Find(&sessions).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("暂时没有群聊会话")
			return "暂时没有群聊会话", 0, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.GroupSessionListRespond
	for _, session := range sessions {
		if session.ReceiveId[0] == 'G' {
			res = append(res, respond.GroupSessionListRespond{
				SessionId: session.Uuid,
				GroupName: session.ReceiveName,
				GroupId:   session.ReceiveId,
				Avatar:    session.Avatar,
			})
		}
	}

	return "获取群聊列表成功", 0, res
}

// DeleteSession 删除会话
func (s *sessionService) DeleteSession(req request.DeleteSessionRequest) (string, int) {
	// 先删 redis 缓存
	var session model.Session
	err := dao.DB.First(&session, "uuid = ?", req.SessionId).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	redisKey := fmt.Sprintf("session:%s:%s", session.SendId, session.ReceiveId)
	err = dao.DelKeyIfExists(redisKey)
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	var deleted_at gorm.DeletedAt
	deleted_at.Time = time.Now()
	deleted_at.Valid = true

	err = dao.DB.Model(&model.Session{}).Where("uuid = ?", req.SessionId).Update("deleted_at", deleted_at).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "删除成功", 0
}

// CheckOpenSessionAllowed 检查是否允许发起会话
func (s *sessionService) CheckOpenSessionAllowed(req request.CreateSessionRequest) (string, int, bool) {
	// 被对方拉黑，拉黑了对方，对方被封禁都不能发起会话，群聊除了没有拉黑都类似
	var contact model.UserContact
	err := dao.DB.Where("user_id = ? and contact_id = ?", req.SendId, req.ReceiveId).First(&contact).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "不能发起对话", -2, false
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, false
	}

	switch contact.Status {
	case contact_enum.BLACK: // 把对方拉黑了
		return "对方被你拉黑，解除拉黑后才可发起会话", -2, false
	case contact_enum.BE_BLACK: // 被对方拉黑了
		return "你已被对方拉黑", -2, false
	}

	// 对方是否被封禁
	switch req.ReceiveId[0] {
	case 'U': // 用户
		var user model.UserInfo
		err := dao.DB.First(&user, "uuid = ?", req.ReceiveId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, false
		}
		if user.Status == user_enum.DISABLE {
			return "对方已被封禁", -2, false
		}

	case 'G': // 群聊
		var group model.GroupInfo
		err := dao.DB.First(&group, "uuid = ?", req.ReceiveId).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1, false
		}
		switch group.Status {
		case group_enum.DISABLE:
			return "群聊已被封禁", -2, false
		case group_enum.DISSMIS:
			return "群聊已被解散", -2, false
		}
	}

	return "可以发起会话", 0, true
}
