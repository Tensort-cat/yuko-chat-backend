package respond

type UserListResponse struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
}
