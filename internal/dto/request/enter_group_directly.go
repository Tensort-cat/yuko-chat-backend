package request

type EnterGroupDirectlyRequest struct {
	Owner_id  string `json:"owner_id"`
	ContactId string `json:"contact_id"`
}
