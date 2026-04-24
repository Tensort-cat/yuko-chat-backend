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
	user_enum "yuko_chat/pkg/enum/user"
	"yuko_chat/pkg/util"
	"yuko_chat/pkg/zlog"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userInfoService struct {
}

var UserInfoService = new(userInfoService)

func (s *userInfoService) Login(telephone, password string) (string, *respond.LoginRespond, int) {
	var userInfo model.UserInfo
	err := dao.DB.First(&userInfo, "telephone = ?", telephone).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("手机号输入错误")
			return "手机号输入错误", nil, -2
		}
		return constant.SYS_ERR_MSG, nil, -1
	}

	if !util.CheckPwd(password, userInfo.Password) {
		zlog.Info("密码输入错误")
		return "密码输入错误", nil, -2
	}
	loginRsp := &respond.LoginRespond{
		Uuid:      userInfo.Uuid,
		Telephone: userInfo.Telephone,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		Birthday:  userInfo.Birthday,
		Signature: userInfo.Signature,
		IsAdmin:   userInfo.IsAdmin,
		Status:    userInfo.Status,
	}
	year, month, day := userInfo.CreatedAt.Date()
	loginRsp.CreatedAt = fmt.Sprintf("%d.%d.%d", year, month, day)
	return "登录成功", loginRsp, 0
}

func (s *userInfoService) Register(req request.RegisterRequest) (string, *respond.RegisterRespond, int) {
	// 检查手机号格式
	if vaild := util.IsValidPhoneNumber(req.Telephone); !vaild {
		zlog.Info("手机号格式错误")
		return "请输入合法的手机号", nil, -2
	}

	// 检查邮箱格式
	if vaild := util.IsValidEmail(req.Email); !vaild {
		zlog.Info("邮箱格式错误")
		return "请输入合法的邮箱", nil, -2
	}

	// 用过的邮箱或手机号不能重复注册
	result := dao.DB.Where("telephone = ? or email = ?", req.Telephone, req.Email).First(&model.UserInfo{})
	if err := result.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, nil, -1
		}
	} else { // err 为空，说明数据库中有记录
		zlog.Info("手机号或邮箱已被注册")
		return "手机号或邮箱已被注册", nil, -2
	}

	redisKey := fmt.Sprintf("verify:%s", req.Email)
	storedCode, err := dao.GetKeyNilIsErr(redisKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "验证码不存在或已过期", nil, -2
		}
		return "系统错误", nil, -1
	}
	if storedCode != req.SmsCode {
		return "验证码错误", nil, -1
	}

	// 加密密码
	encryptPwd, err := util.HashPassword(req.Password)
	if err != nil {
		return constant.SYS_ERR_MSG, nil, -1
	}
	userInfo := model.UserInfo{
		Uuid:      util.GenUUID("U"),
		Telephone: req.Telephone,
		Password:  encryptPwd,
		Nickname:  req.Nickname,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	err = dao.DB.Create(&userInfo).Error
	if err != nil {
		return "创建用户失败", nil, -2
	}

	registerRsp := &respond.RegisterRespond{
		Uuid:      userInfo.Uuid,
		Telephone: userInfo.Telephone,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		Birthday:  userInfo.Birthday,
		Signature: userInfo.Signature,
		IsAdmin:   userInfo.IsAdmin,
		Status:    userInfo.Status,
	}
	year, month, day := userInfo.CreatedAt.Date()
	registerRsp.CreatedAt = fmt.Sprintf("%d.%d.%d", year, month, day)

	return "注册成功", registerRsp, 0
}

func (s *userInfoService) SendVerificationCode(email string) (string, int) {
	code, err := util.SendEmail(email)
	if err != nil {
		return "发送邮箱验证码失败", -1
	}
	zlog.Info("邮箱验证码", zap.String("code", code))

	redisKey := fmt.Sprintf("verify:%s", email)
	if err = dao.SetKeyEx(redisKey, code, constant.REDIS_TIMEOUT); err != nil {
		return "申请验证码频率过快", -1
	}

	return "发送邮箱验证码成功", 0
}

func (s *userInfoService) GetUserInfo(uuid string) (string, *respond.GetUserInfoRespond, int) {
	var userInfo model.UserInfo
	err := dao.DB.Where("uuid = ?", uuid).First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "用户不存在", nil, -2
		}
		return constant.SYS_ERR_MSG, nil, -1
	}

	getUserInfoRsp := &respond.GetUserInfoRespond{
		Uuid:      userInfo.Uuid,
		Telephone: userInfo.Telephone,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		Birthday:  userInfo.Birthday,
		Signature: userInfo.Signature,
		IsAdmin:   userInfo.IsAdmin,
		Status:    userInfo.Status,
	}
	year, month, day := userInfo.CreatedAt.Date()
	getUserInfoRsp.CreatedAt = fmt.Sprintf("%d.%d.%d", year, month, day)
	return "获取用户信息成功", getUserInfoRsp, 0
}

func (s *userInfoService) UpdateUserInfo(req request.UpdateUserInfoRequest) (string, int) {
	updateData := map[string]any{}
	fields := map[string]string{
		"email":     req.Email,
		"nickname":  req.Nickname,
		"birthday":  req.Birthday,
		"signature": req.Signature,
		"avatar":    req.Avatar,
		"gender":    fmt.Sprintf("%d", req.Gender),
	}

	for key, value := range fields {
		if value != "" {
			updateData[key] = value
		}
	}

	if len(updateData) == 0 {
		return "没有需要更新的信息", -2
	}

	result := dao.DB.Model(&model.UserInfo{}).Where("uuid = ?", req.Uuid).Updates(updateData)
	if result.Error != nil {
		return "更新用户信息失败", -1
	}
	if result.RowsAffected == 0 {
		return "用户不存在", -2
	}

	return "更新用户信息成功", 0
}

func (s *userInfoService) GetUserInfoList(req request.GetUserInfoListRequest) (string, int, []respond.GetUserListRespond) {
	var users []model.UserInfo
	// Unscoped() 可以屏蔽软删除，不然看不到被软删除的用户了
	err := dao.DB.Unscoped().Where("is_admin = 0 and uuid != ?", req.OwnerId).Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("系统中还没有普通用户")
			return "系统中还没有普通用户", 0, nil
		}
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.GetUserListRespond
	for _, user := range users {
		res = append(res, respond.GetUserListRespond{
			Uuid:      user.Uuid,
			Nickname:  user.Nickname,
			Telephone: user.Telephone,
			Status:    user.Status,
			IsAdmin:   user.IsAdmin,
			IsDeleted: user.DeletedAt.Valid,
		})
	}

	return "获取用户列表成功", 0, res
}

// AbleUsers 启用用户
// 用户是否启用禁用需要实时更新contact_user_list状态，所以redis的contact_user_list需要删除
func (s *userInfoService) AbleUsers(req request.AbleUsersRequest) (string, int) {
	if req.IsAdmin == 0 {
		zlog.Warn("孩子们，有黑客")
		return "你不是管理员，你是谁？", -1
	}

	var users []model.UserInfo
	err := dao.DB.Model(&model.UserInfo{}).Where("uuid in (?)", req.UuidList).Find(&users).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	for _, user := range users {
		user.Status = user_enum.NORMAL
		if err := dao.DB.Save(&user).Error; err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	// 删除所有"contact_user_list"开头的key
	if err := dao.DelKeysWithPrefix("contact_user_list"); err != nil {
		zlog.Error(err.Error())
	}

	return "启用用户成功", 0
}

// DisableUsers 禁用用户
// 用户是否启用禁用需要实时更新contact_user_list状态，所以redis的contact_user_list需要删除
func (s *userInfoService) DisableUsers(req request.AbleUsersRequest) (string, int) {
	if req.IsAdmin == 0 {
		zlog.Warn("孩子们，有黑客")
		return "你不是管理员，你是谁？", -1
	}

	var users []model.UserInfo
	err := dao.DB.Model(&model.UserInfo{}).Where("uuid in (?)", req.UuidList).Find(&users).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	for _, user := range users {
		user.Status = user_enum.DISABLE
		if err := dao.DB.Save(&user).Error; err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 用户相关的会话都要删除
		var sessions []model.Session
		err := dao.DB.Where("send_id = ? or receive_id = ?", user.Uuid, user.Uuid).Find(&sessions).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		for _, session := range sessions {
			var deleted_at gorm.DeletedAt
			deleted_at.Time = time.Now()
			deleted_at.Valid = true
			session.DeletedAt = deleted_at

			err := dao.DB.Save(&session).Error
			if err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}
	}
	// 删除所有"contact_user_list"开头的key
	if err := dao.DelKeysWithPrefix("contact_user_list"); err != nil {
		zlog.Error(err.Error())
	}
	return "禁用用户成功", 0

}

// DeleteUsers 删除用户
// 用户是否启用禁用需要实时更新contact_user_list状态，所以redis的contact_user_list需要删除
func (u *userInfoService) DeleteUsers(req request.AbleUsersRequest) (string, int) {
	if req.IsAdmin == 0 {
		zlog.Warn("孩子们，有黑客")
		return "你不是管理员，你是谁？", -1
	}

	// 先把要删除的用户都取出来
	var users []model.UserInfo
	err := dao.DB.Model(&model.UserInfo{}).Where("uuid in (?)", req.UuidList).Find(&users).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 软删除
	for _, user := range users {
		user.DeletedAt.Time = time.Now()
		user.DeletedAt.Valid = true
		if err := dao.DB.Save(&user).Error; err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}

		// 删除用户相关的会话
		var sessions []model.Session
		err := dao.DB.Where("send_id = ? or receive_id = ?", user.Uuid, user.Uuid).Find(&sessions).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		for _, session := range sessions {
			var deleted_at gorm.DeletedAt
			deleted_at.Time = time.Now()
			deleted_at.Valid = true
			session.DeletedAt = deleted_at

			if err := dao.DB.Save(&session).Error; err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		// 删除用户相关的联系记录
		var contactList []model.UserContact
		err = dao.DB.Where("user_id = ? or contact_id = ?", user.Uuid, user.Uuid).Find(&contactList).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		for _, contact := range contactList {
			var deleted_at gorm.DeletedAt
			deleted_at.Time = time.Now()
			deleted_at.Valid = true
			contact.DeletedAt = deleted_at

			if err := dao.DB.Save(&contact).Error; err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}

		// 删除相关的申请记录
		var applies []model.ContactApply
		err = dao.DB.Model(&model.ContactApply{}).Where("user_id = ? or contact_id = ?", user.Uuid, user.Uuid).Find(&applies).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
		for _, apply := range applies {
			var deleted_at gorm.DeletedAt
			deleted_at.Time = time.Now()
			deleted_at.Valid = true
			apply.DeletedAt = deleted_at

			if err := dao.DB.Save(&apply).Error; err != nil {
				zlog.Error(err.Error())
				return constant.SYS_ERR_MSG, -1
			}
		}
	}

	// 删除所有"contact_user_list"开头的key
	if err := dao.DelKeysWithPrefix("contact_user_list"); err != nil {
		zlog.Error(err.Error())
	}

	return "删除用户成功", 0
}

// SetAdmin 设置管理员
func (u *userInfoService) SetAdmin(req request.AbleUsersRequest) (string, int) {
	var users []model.UserInfo
	err := dao.DB.Where("uuid in (?)", req.UuidList).Find(&users).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	for _, user := range users {
		user.IsAdmin = 1
		err := dao.DB.Save(&user).Error
		if err != nil {
			zlog.Error(err.Error())
			return constant.SYS_ERR_MSG, -1
		}
	}

	return "设置管理员成功", 0
}

// SmsLogin 短信验证码登录
func (s *userInfoService) SmsLogin(req request.SmsLoginRequest) (string, *respond.LoginRespond, int) {
	redisKey := fmt.Sprintf("verify:%s", req.Email)
	storedCode, err := dao.GetKeyNilIsErr(redisKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "验证码不存在或已过期", nil, -2
		}
		return "系统错误", nil, -1
	}
	if storedCode != req.SmsCode {
		return "验证码错误", nil, -2
	}

	var userInfo model.UserInfo
	err = dao.DB.Where("email = ?", req.Email).First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("邮箱输入错误")
			return "邮箱输入错误", nil, -2
		}
		return constant.SYS_ERR_MSG, nil, -1
	}

	loginRsp := &respond.LoginRespond{
		Uuid:      userInfo.Uuid,
		Telephone: userInfo.Telephone,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		Birthday:  userInfo.Birthday,
		Signature: userInfo.Signature,
		IsAdmin:   userInfo.IsAdmin,
		Status:    userInfo.Status,
	}

	return "登录成功", loginRsp, 0
}
