package interfaces

// NotifyMessage 通知消息
type NotifyMessage struct {
	Subject string // 邮件主题/通知标题
	Body    string // 通知内容
}

// Notifier 消息通知接口
type Notifier interface {
	// Send 发送通知
	Send(to string, msg NotifyMessage) error
	// Name 返回通知器名称
	Name() string
}
