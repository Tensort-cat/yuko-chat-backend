package test

import (
	"fmt"
	"testing"
	"yuko_chat/pkg/util"
)

func TestPhoneNumberVerify(t *testing.T) {
	num1 := "12358944441"
	vaild := util.IsValidPhoneNumber(num1)
	fmt.Println(num1, "valid:", vaild) // 输出: 12358944441 valid: false

	num2 := "18985052504"
	vaild = util.IsValidPhoneNumber(num2)
	fmt.Println(num2, "valid:", vaild) // 输出: 18985052504 valid: true
}

func TestEmailNumberVerify(t *testing.T) {
	email1 := "test@example.com"
	vaild := util.IsValidEmail(email1)
	fmt.Println(email1, "valid:", vaild) // 输出: test@example.com valid: true

	email2 := "invalid-email"
	vaild = util.IsValidEmail(email2)
	fmt.Println(email2, "valid:", vaild) // 输出: invalid-email valid: false
}
