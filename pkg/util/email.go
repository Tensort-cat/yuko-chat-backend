package util

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"yuko_chat/internal/config"
)

type smtpConfig struct {
	host     string
	port     string
	username string
	password string
	from     string
	to       string
}

func SendEmail(to string) (string, error) {
	if to == "" {
		return "", fmt.Errorf("recipient email address is required")
	}
	emailCfg := config.Cfg.EmailConfig
	cfg := smtpConfig{
		host:     emailCfg.Host,
		port:     strconv.Itoa(emailCfg.Port),
		username: emailCfg.Username,
		password: emailCfg.Password,
		from:     emailCfg.From,
		to:       to,
	}

	code := generateVerifyCode(6)
	subject := "YukoChat 邮箱验证码"
	body := fmt.Sprintf("你的验证码是：%s\n5 分钟内有效。", code)

	if err := sendEmail(cfg, subject, body); err != nil {
		return "", fmt.Errorf("send email failed: %v", err)
	}
	return code, nil
}

func generateVerifyCode(length int) string {
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	digits := make([]byte, length)
	for i := range digits {
		digits[i] = byte('0' + source.Intn(10))
	}
	return string(digits)
}

func sendEmail(cfg smtpConfig, subject string, body string) error {
	addr := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	message := buildMessage(cfg.from, cfg.to, subject, body)

	// 465 通常使用 SSL 直连，其他端口优先尝试标准 SMTP 发信。
	if cfg.port == "465" {
		return sendByTLS(addr, cfg, message)
	}

	auth := smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)
	return smtp.SendMail(addr, auth, cfg.from, []string{cfg.to}, []byte(message))
}

func sendByTLS(addr string, cfg smtpConfig, message string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: cfg.host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		return err
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)
	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(cfg.from); err != nil {
		return err
	}
	if err = client.Rcpt(cfg.to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write([]byte(message))
	return err
}

func buildMessage(from string, to string, subject string, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}
	return strings.Join(headers, "\r\n")
}
