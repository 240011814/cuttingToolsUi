package service

import (
	iface "backend/interface"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"sync"
)

// SMTPConfig SMTP 配置缓存
type SMTPConfig struct {
	Host       string
	Port       int
	Encryption string
	User       string
	Password   string
	From       string
	FromName   string
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	configSvc *SystemConfigService
	mu        sync.RWMutex
	smtpCache *SMTPConfig
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(configSvc *SystemConfigService) *EmailNotifier {
	n := &EmailNotifier{
		configSvc: configSvc,
	}
	n.RefreshConfig()
	return n
}

// Name 返回通知器名称
func (n *EmailNotifier) Name() string {
	return "email"
}

// RefreshConfig 刷新 SMTP 配置缓存
func (n *EmailNotifier) RefreshConfig() {
	host, _ := n.configSvc.GetValue("smtp_host")
	portStr, _ := n.configSvc.GetValue("smtp_port")
	encryption, _ := n.configSvc.GetValue("smtp_encryption")
	user, _ := n.configSvc.GetValue("smtp_user")
	password, _ := n.configSvc.GetValue("smtp_password")
	from, _ := n.configSvc.GetValue("smtp_from")
	fromName, _ := n.configSvc.GetValue("smtp_from_name")

	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	if fromName == "" {
		fromName = "系统通知"
	}

	cfg := &SMTPConfig{
		Host:       host,
		Port:       port,
		Encryption: encryption,
		User:       user,
		Password:   password,
		From:       from,
		FromName:   fromName,
	}

	n.mu.Lock()
	n.smtpCache = cfg
	n.mu.Unlock()
}

// Send 发送邮件通知
func (n *EmailNotifier) Send(to string, msg iface.NotifyMessage) error {
	n.mu.RLock()
	cfg := n.smtpCache
	n.mu.RUnlock()

	if cfg.Host == "" || cfg.User == "" || cfg.Password == "" || cfg.From == "" {
		return fmt.Errorf("SMTP 配置不完整，请先填写 SMTP 主机、用户名、密码和发件人邮箱")
	}

	data := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\nMIME-Version: 1.0\r\n\r\n%s",
		cfg.FromName, cfg.From, to, msg.Subject, msg.Body)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	switch cfg.Encryption {
	case "ssl":
		return n.sendSSL(addr, auth, cfg.From, to, data, cfg.Host, cfg.Port)
	case "starttls":
		return n.sendStartTLS(addr, auth, cfg.From, to, data, cfg.Host)
	default:
		return smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(data))
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
