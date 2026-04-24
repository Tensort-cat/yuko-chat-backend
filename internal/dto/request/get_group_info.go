package request

type GetGroupInfoRequest struct {
	GroupId string `json:"group_id" binding:"required"` // 群聊id
}
