package route

import (
	"yuko_chat/internal/config"
	"yuko_chat/internal/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var GE *gin.Engine

func InitRoute() {
	GE = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "Authorization"}
	GE.Use(cors.New(corsConfig))

	api := GE.Group("/api")
	{
		api.Static("/static/avatars", config.Cfg.StaticSrcConfig.StaticAvatarPath)
		api.Static("/static/files", config.Cfg.StaticSrcConfig.StaticFilePath)
		api.POST("/login", controller.Login)
		api.POST("/register", controller.Register)

		user := api.Group("/user")
		{
			user.POST("/updateUserInfo", controller.UpdateUserInfo)
			user.POST("/getUserInfo", controller.GetUserInfo)
			user.POST("/sendSmsCode", controller.SendVerificationCode)
			user.POST("/wsLogout", controller.WsLogout)
			user.POST("/smsLogin", controller.SmsLogin)
			user.POST("/getUserInfoList", controller.GetUserInfoList)
			user.POST("/disableUsers", controller.DisableUsers)
			user.POST("/ableUsers", controller.AbleUsers)
			user.POST("/deleteUsers", controller.DeleteUsers)
			user.POST("/setAdmin", controller.SetAdmin)
		}

		group := api.Group("/group")
		{
			group.POST("/createGroup", controller.CreateGroup)
			group.POST("/loadMyGroup", controller.GetMyGroups)
			group.POST("/checkGroupAddMode", controller.CheckGroupAddMode)
			group.POST("/enterGroupDirectly", controller.EnterGroupDirectly)
			group.POST("/leaveGroup", controller.LeaveGroup)
			group.POST("/dismissGroup", controller.DismissGroup)
			group.POST("/getGroupInfo", controller.GetGroupInfo)
			group.POST("/updateGroupInfo", controller.UpdateGroupInfo)
			group.POST("/removeGroupMembers", controller.RemoveGroupMembers)
			group.POST("/getGroupMemberList", controller.GetGroupMembers)
			group.POST("/getGroupInfoList", controller.GetGroupInfoList)
			group.POST("/setGroupsStatus", controller.SetGroupsStatus)
			group.POST("/deleteGroups", controller.DeleteGroups)
		}

		message := api.Group("/message")
		{
			message.POST("/getMessageList", controller.GetMessageList)
			message.POST("/getGroupMessageList", controller.GetGroupMessageList)
			message.POST("/uploadAvatar", controller.UploadAvatar)
			message.POST("/uploadFile", controller.UploadFile)
		}

		session := api.Group("/session")
		{
			session.POST("/openSession", controller.OpenSession)
			session.POST("/getUserSessionList", controller.GetUserSessionList)
			session.POST("/getGroupSessionList", controller.GetGroupSessionList)
			session.POST("/deleteSession", controller.DeleteSession)
			session.POST("/checkOpenSessionAllowed", controller.CheckOpenSessionAllowed)
		}

		contact := api.Group("/contact")
		{
			contact.POST("/getContactList", controller.GetContactList)
			contact.POST("/loadMyJoinedGroup", controller.LoadMyJoinedGroup)
			contact.POST("/getContactInfo", controller.GetContactInfo)
			contact.POST("/deleteContact", controller.DeleteContact)
			contact.POST("/applyContact", controller.ApplyContact)
			contact.POST("/getNewContactList", controller.GetNewContactList)
			contact.POST("/passContactApply", controller.PassContactApply)
			contact.POST("/blackContact", controller.BlackContact)
			contact.POST("/cancelBlackContact", controller.CancelBlackContact)
			contact.POST("/getAddGroupList", controller.GetAddGroupList)
			contact.POST("/refuseContactApply", controller.RefuseContactApply)
			contact.POST("/blackApply", controller.BlackApply)
		}
	}

	GE.GET("/wss", controller.WsLogin)

}
