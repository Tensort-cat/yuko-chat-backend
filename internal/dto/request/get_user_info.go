package request

type GetUserInfoRequest struct {
	Uuid string `json:"uuid" binding:"required"`
}
