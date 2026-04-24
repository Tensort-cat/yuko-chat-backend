package request

type SmsLoginRequest struct {
	Email   string `json:"email"`
	SmsCode string `json:"sms_code"`
}
