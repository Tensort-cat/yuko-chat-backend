package respond

type AddGroupListRespond struct {
	ContactId     string `json:"contact_id"`
	ContactName   string `json:"contact_name"`
	Message       string `json:"message"`
	ContactAvatar string `json:"contact_avatar"`
}
