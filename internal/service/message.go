package service

import (
	"io"
	"os"
	"path/filepath"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/dto/respond"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type messageService struct {
}

var MessageService = new(messageService)

// GetMessageList 获取聊天记录
func (m *messageService) GetMessageList(req request.GetMessageListRequest) (string, int, []respond.GetMessageListRespond) {
	var messages []model.Message
	err := dao.DB.
		Order("created_at asc").
		Where(
			"(send_id = ? and receive_id = ?) or (send_id = ? and receive_id = ?)", req.UserOneId, req.UserTwoId, req.UserTwoId, req.UserOneId).
		Find(&messages).
		Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.GetMessageListRespond
	for _, msg := range messages {
		res = append(res, respond.GetMessageListRespond{
			SendId:     msg.SendId,
			SendName:   msg.SendName,
			SendAvatar: msg.SendAvatar,
			ReceiveId:  msg.ReceiveId,
			Type:       msg.Type,
			Content:    msg.Content,
			Url:        msg.Url,
			FileType:   msg.FileType,
			FileName:   msg.FileName,
			FileSize:   msg.FileSize,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return "获取聊天记录成功", 0, res
}

// GetGroupMessageList 获取群聊消息记录
func (m *messageService) GetGroupMessageList(req request.GetGroupMessageListRequest) (string, int, []respond.GetGroupMessageListRespond) {
	var messages []model.Message
	err := dao.DB.Where("receive_id = ?", req.GroupId).Order("created_at asc").Find(&messages).Error
	if err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1, nil
	}

	var res []respond.GetGroupMessageListRespond
	for _, msg := range messages {
		res = append(res, respond.GetGroupMessageListRespond{
			SendId:     msg.SendId,
			SendName:   msg.SendName,
			SendAvatar: msg.SendAvatar,
			ReceiveId:  msg.ReceiveId,
			Type:       msg.Type,
			Content:    msg.Content,
			Url:        msg.Url,
			FileType:   msg.FileType,
			FileSize:   msg.FileSize,
			FileName:   msg.FileName,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return "获取群聊聊天记录成功", 0, res
}

// todo: 文件上传与下载
// UploadAvator 上传头像
func (m *messageService) UploadAvatar(c *gin.Context) (string, int) {
	if err := m.uploadFile(c, config.Cfg.StaticSrcConfig.StaticAvatarPath); err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	return "上传成功", 0
}

// UploadFile 上传文件
func (m *messageService) UploadFile(c *gin.Context) (string, int) {
	if err := m.uploadFile(c, config.Cfg.StaticSrcConfig.StaticFilePath); err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}
	return "上传成功", 0
}

func (m *messageService) uploadFile(c *gin.Context, path string) error {
	if err := c.Request.ParseMultipartForm(constant.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return err
	}
	mForm := c.Request.MultipartForm
	for key := range mForm.File {
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return err
		}
		defer file.Close()
		zlog.Info("上传文件信息", zap.String("文件名", fileHeader.Filename), zap.Int64("文件大小", fileHeader.Size))
		// 原来Filename应该是213451545.xxx，将Filename修改为avatar_ownerId.xxx
		// 获取文件后缀
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info("文件后缀", zap.String("ext", ext))
		localFileName := path + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return err
		}
		zlog.Info("完成文件上传")
	}

	return nil
}
