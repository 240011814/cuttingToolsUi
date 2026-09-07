package service

import (
	iface "backend/interface"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	configSvc *SystemConfigService
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(configSvc *SystemConfigService) *EmailNotifier {
	return &EmailNotifier{configSvc: configSvc}
}

// Name 返回通知器名称
func (n *EmailNotifier) Name() string {
	return "email"
}

// Send 发送邮件通知
func (n *EmailNotifier) Send(to string, msg iface.NotifyMessage) error {
	host, _ := n.configSvc.GetValue("smtp_host")
	portStr, _ := n.configSvc.GetValue("smtp_port")
	encryption, _ := n.configSvc.GetValue("smtp_encryption")
	user, _ := n.configSvc.GetValue("smtp_user")
	password, _ := n.configSvc.GetValue("smtp_password")
	from, _ := n.configSvc.GetValue("smtp_from")
	fromName, _ := n.configSvc.GetValue("smtp_from_name")

	if host == "" || user == "" || password == "" || from == "" {
		return fmt.Errorf("SMTP 配置不完整，请先填写 SMTP 主机、用户名、密码和发件人邮箱")
	}

	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	if fromName == "" {
		fromName = "系统通知"
	}

	data := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\nMIME-Version: 1.0\r\n\r\n%s",
		fromName, from, to, msg.Subject, msg.Body)

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", user, password, host)

	switch encryption {
	case "ssl":
		return n.sendSSL(addr, auth, from, to, data, host, port)
	case "starttls":
		return n.sendStartTLS(addr, auth, from, to, data, host)
	default:
		return smtp.SendMail(addr, auth, from, []string{to}, []byte(data))
	}
}

func (n *EmailNotifier) sendSSL(addr string, auth smtp.Auth, from, to, data, host string, port int) error {
	tlsConfig := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("SSL 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据 writer 失败: %w", err)
	}
	if _, err = w.Write([]byte(data)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭数据 writer 失败: %w", err)
	}

	return client.Quit()
}

func (n *EmailNotifier) sendStartTLS(addr string, auth smtp.Auth, from, to, data, host string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("启动 STARTTLS 失败: %w", err)
		}
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据 writer 失败: %w", err)
	}
	if _, err = w.Write([]byte(data)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭数据 writer 失败: %w", err)
	}

	return client.Quit()
}
