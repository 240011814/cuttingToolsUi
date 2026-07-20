package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"backend/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramService struct {
	sysCfgService *SystemConfigService
	bot           *tgbotapi.BotAPI
	stopCh        chan struct{}
}

func NewTelegramService(sysCfgService *SystemConfigService) *TelegramService {
	s := &TelegramService{
		sysCfgService: sysCfgService,
		stopCh:        make(chan struct{}),
	}
	return s
}

// StartBot 启动 Telegram Bot
func (s *TelegramService) StartBot() error {
	token := s.sysCfgService.GetTelegramBotToken()
	if token == "" {
		log.Println("Telegram Bot Token not configured, skipping bot start")
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	s.bot = bot
	log.Printf("Telegram Bot authorized as @%s", bot.Self.UserName)

	go s.listenUpdates()
	return nil
}

// StopBot 停止 Bot
func (s *TelegramService) StopBot() {
	close(s.stopCh)
	if s.bot != nil {
		s.bot.StopReceivingUpdates()
	}
}

// RestartBot 重启 Bot（配置更新后调用）
func (s *TelegramService) RestartBot() error {
	s.StopBot()
	s.stopCh = make(chan struct{})
	return s.StartBot()
}

// listenUpdates 监听消息更新
func (s *TelegramService) listenUpdates() {
	if s.bot == nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	for {
		select {
		case <-s.stopCh:
			return
		case update := <-updates:
			if update.Message != nil {
				s.handleMessage(update.Message)
			}
		}
	}
}

// handleMessage 处理用户消息
func (s *TelegramService) handleMessage(msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return
	}

	switch msg.Command() {
	case "start":
		s.handleStart(msg)
	case "bind":
		s.handleBind(msg)
	case "unbind":
		s.handleUnbind(msg)
	case "status":
		s.handleStatus(msg)
	default:
		s.reply(msg.Chat.ID, "未知命令。可用命令：\n/bind <code> - 绑定账号\n/unbind - 解绑\n/status - 查看状态")
	}
}

// handleStart 处理 /start 命令
func (s *TelegramService) handleStart(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	user, _ := s.getUserByChatID(chatID)

	if user != nil {
		s.reply(chatID, fmt.Sprintf("你好 %s！你的账号已绑定。", user.Nickname))
	} else {
		s.reply(chatID, "你好！请先在网页端生成绑定码，然后使用 /bind <code> 绑定你的账号。")
	}
}

// handleBind 处理 /bind 命令
func (s *TelegramService) handleBind(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	args := msg.CommandArguments()

	if args == "" {
		s.reply(chatID, "请提供绑定码。用法：/bind <6位数字码>")
		return
	}

	// 检查是否已绑定
	existingUser, _ := s.getUserByChatID(chatID)
	if existingUser != nil {
		s.reply(chatID, "你的 Telegram 已绑定账号："+existingUser.Username)
		return
	}

	// 查找绑定码匹配的用户
	var user model.User
	now := time.Now()
	result := DB.Where("telegram_bind_code = ? AND telegram_bind_code_expires_at > ?", args, now).First(&user)
	if result.Error != nil {
		s.reply(chatID, "绑定码无效或已过期，请重新生成。")
		return
	}

	// 完成绑定
	telegramUsername := msg.From.UserName
	if telegramUsername == "" {
		telegramUsername = msg.From.FirstName
	}

	updates := map[string]interface{}{
		"telegram_chat_id":            chatID,
		"telegram_username":           telegramUsername,
		"telegram_bind_code":          nil,
		"telegram_bind_code_expires_at": nil,
	}

	if err := DB.Model(&user).Updates(updates).Error; err != nil {
		s.reply(chatID, "绑定失败，请稍后重试。")
		return
	}

	s.reply(chatID, fmt.Sprintf("绑定成功！你好 %s，现在可以使用学习助手功能了。", user.Nickname))
}

// handleUnbind 处理 /unbind 命令
func (s *TelegramService) handleUnbind(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "你的 Telegram 尚未绑定任何账号。")
		return
	}

	updates := map[string]interface{}{
		"telegram_chat_id":    nil,
		"telegram_username":   nil,
	}

	if err := DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		s.reply(chatID, "解绑失败，请稍后重试。")
		return
	}

	s.reply(chatID, "已成功解绑。")
}

// handleStatus 处理 /status 命令
func (s *TelegramService) handleStatus(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "你的 Telegram 尚未绑定任何账号。请使用 /bind <code> 绑定。")
		return
	}

	s.reply(chatID, fmt.Sprintf("已绑定账号：%s (%s)", user.Username, user.Nickname))
}

// reply 发送回复消息
func (s *TelegramService) reply(chatID int64, text string) {
	if s.bot == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	s.bot.Send(msg)
}

// getUserByChatID 通过 Telegram Chat ID 获取用户
func (s *TelegramService) getUserByChatID(chatID int64) (*model.User, error) {
	var user model.User
	result := DB.Where("telegram_chat_id = ?", chatID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GenerateBindCode 为用户生成绑定码
func (s *TelegramService) GenerateBindCode(userID uint) (*model.TelegramBindCodeResponse, error) {
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 如果已绑定，返回错误
	if user.TelegramChatID != nil {
		return nil, errors.New("Telegram 已绑定，请先解绑")
	}

	// 生成 6 位随机数字码
	code, err := generateRandomCode(6)
	if err != nil {
		return nil, errors.New("生成绑定码失败")
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	// 保存绑定码
	updates := map[string]interface{}{
		"telegram_bind_code":           code,
		"telegram_bind_code_expires_at": expiresAt,
	}

	if err := DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, errors.New("保存绑定码失败")
	}

	botName := ""
	if s.bot != nil {
		botName = s.bot.Self.UserName
	}

	return &model.TelegramBindCodeResponse{
		BindCode:  code,
		ExpiresAt: expiresAt.Unix(),
		BotName:   botName,
	}, nil
}

// GetTelegramStatus 获取用户的 Telegram 绑定状态
func (s *TelegramService) GetTelegramStatus(userID uint) (*model.TelegramStatusResponse, error) {
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.TelegramChatID == nil {
		return &model.TelegramStatusResponse{IsBound: false}, nil
	}

	tgUsername := ""
	if user.TelegramUsername != nil {
		tgUsername = *user.TelegramUsername
	}

	return &model.TelegramStatusResponse{
		IsBound:          true,
		TelegramUsername: tgUsername,
	}, nil
}

// UnbindTelegram 解绑用户的 Telegram
func (s *TelegramService) UnbindTelegram(userID uint) error {
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.TelegramChatID == nil {
		return errors.New("Telegram 未绑定")
	}

	updates := map[string]interface{}{
		"telegram_chat_id":    nil,
		"telegram_username":   nil,
		"telegram_bind_code":          nil,
		"telegram_bind_code_expires_at": nil,
	}

	return DB.Model(&user).Updates(updates).Error
}

// generateRandomCode 生成随机数字码
func generateRandomCode(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}
