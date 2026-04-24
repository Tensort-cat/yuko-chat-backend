package request

type SendVerificationCodeRequest struct {
	Email string `json:"email"`
}
