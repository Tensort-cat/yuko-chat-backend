package request

type RemoveGroupMembersRequest struct {
	GroupId  string   `json:"group_id"`  // 群聊id
	OwnerId  string   `json:"owner_id"`  // 群主id
	UuidList []string `json:"uuid_list"` // 要移除的成员uuid列表
}
