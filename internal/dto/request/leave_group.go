package request

type LeaveGroupRequest struct {
	GroupId string `json:"group_id"` // 要退出的群聊
	UserId  string `json:"user_id"`  // 退出群聊的用户id
}
