package util

import (
	"regexp"
)

// IsValidPhoneNumber 验证手机号格式（中国大陆手机号）
func IsValidPhoneNumber(phone string) bool {
	// 中国手机号正则：以1开头，第二位3-9，后面9位数字
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(phone)
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	// 简单的邮箱正则：本地部分@域名.顶级域
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

