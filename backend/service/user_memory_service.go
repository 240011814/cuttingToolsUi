package service

import (
	"backend/model"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MemoryExtractionConfig 画像/经历抽取配置(来自 system_config)
type MemoryExtractionConfig struct {
	Enabled         bool
	Model           string
	IdleMinutes     int
	MinUserMessages int
}

// UserMemoryService 用户画像与经历: 会话级抽取、画像合并、注入与 CRUD
type UserMemoryService struct {
	agentSvc *AIAgentService
	sysCfg   *SystemConfigService
	mu       sync.Mutex
}

func NewUserMemoryService(agentSvc *AIAgentService, sysCfg *SystemConfigService) *UserMemoryService {
	return &UserMemoryService{agentSvc: agentSvc, sysCfg: sysCfg}
}

func (s *UserMemoryService) getConfig() MemoryExtractionConfig {
	cfg := MemoryExtractionConfig{
		Enabled:         true,
		IdleMinutes:     15,
		MinUserMessages: 6,
	}
	if s.sysCfg == nil {
		return cfg
	}
	if v, err := s.sysCfg.GetValue("memory_extraction_enabled"); err == nil {
		cfg.Enabled = v != "false"
	}
	if v, err := s.sysCfg.GetValue("memory_extraction_model"); err == nil {
		cfg.Model = v
	}
	if v, err := s.sysCfg.GetValue("memory_session_idle_minutes"); err == nil {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.IdleMinutes = n
		}
	}
	if v, err := s.sysCfg.GetValue("memory_min_min_user_messages"); err == nil {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MinUserMessages = n
		}
	}
	return cfg
}

// ============ 画像 ============

func (s *UserMemoryService) GetOrCreatePortrait(userID uint) (*model.UserPortrait, error) {
	p := model.UserPortrait{UserID: userID}
	err := DB.Where("user_id = ?", userID).
		Attrs(model.UserPortrait{ExtractionEnabled: true, Dimensions: "{}", Tags: "[]"}).
		FirstOrCreate(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *UserMemoryService) GetPortrait(userID uint) (*model.UserPortrait, error) {
	var p model.UserPortrait
	if err := DB.Where("user_id = ?", userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// BuildProfilePrompt 构造注入对话的画像文本, 无画像/低置信/未启用时返回空串
func (s *UserMemoryService) BuildProfilePrompt(userID uint) string {
	p, err := s.GetPortrait(userID)
	if err != nil || p == nil || !p.ExtractionEnabled {
		return ""
	}

	var dims map[string]string
	_ = json.Unmarshal([]byte(p.Dimensions), &dims)
	var tags []string
	_ = json.Unmarshal([]byte(p.Tags), &tags)

	if strings.TrimSpace(p.Summary) == "" && len(dims) == 0 && len(tags) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## 用户画像（供参考，不要直接复述）\n")
	if strings.TrimSpace(p.Summary) != "" {
		b.WriteString(strings.TrimSpace(p.Summary) + "\n")
	}
	for k, v := range dims {
		if strings.TrimSpace(v) != "" {
			b.WriteString(fmt.Sprintf("- %s：%s\n", k, v))
		}
	}
	if len(tags) > 0 {
		b.WriteString("标签：" + strings.Join(tags, "、") + "\n")
	}

	return b.String()
}

func (s *UserMemoryService) UpdatePortrait(userID uint, req model.UpdateUserPortraitRequest) error {
	if _, err := s.GetOrCreatePortrait(userID); err != nil {
		return err
	}
	updates := map[string]interface{}{"is_user_edited": true}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Dimensions != nil {
		b, _ := json.Marshal(*req.Dimensions)
		updates["dimensions"] = string(b)
	}
	if req.Tags != nil {
		b, _ := json.Marshal(*req.Tags)
		updates["tags"] = string(b)
	}
	if req.ExtractionEnabled != nil {
		updates["extraction_enabled"] = *req.ExtractionEnabled
	}
	return DB.Model(&model.UserPortrait{}).Where("user_id = ?", userID).Updates(updates).Error
}

// ============ 经历 CRUD ============

func (s *UserMemoryService) ListExperiences(userID uint, page, pageSize int, category, keyword string) (int64, []model.UserExperience, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := DB.Model(&model.UserExperience{}).Where("user_id = ?", userID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.UserExperience
	if err := q.Order("occurred_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return 0, nil, err
	}
	return total, list, nil
}

func (s *UserMemoryService) SearchExperiences(userID uint, query string, limit int) ([]model.UserExperience, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	q := DB.Where("user_id = ? AND status = ?", userID, "active")
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", like, like, like)
	}
	var list []model.UserExperience
	if err := q.Order("occurred_at DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *UserMemoryService) AddExperience(userID uint, exp *model.UserExperience) (*model.UserExperience, error) {
	exp.ID = 0
	exp.UserID = userID
	exp.Status = "active"
	exp.IsUserEdited = true
	if exp.OccurredAt == nil {
		now := time.Now()
		exp.OccurredAt = &now
	}
	if err := DB.Create(exp).Error; err != nil {
		return nil, err
	}
	return exp, nil
}

func (s *UserMemoryService) CreateExperience(userID uint, req model.CreateUserExperienceRequest) (*model.UserExperience, error) {
	tags, _ := json.Marshal(req.Tags)
	exp := &model.UserExperience{
		UserID:     userID,
		Category:   req.Category,
		Title:      req.Title,
		Content:    req.Content,
		OccurredAt: req.OccurredAt,
		Tags:       string(tags),
		Status:     "active",
	}
	return s.AddExperience(userID, exp)
}

func (s *UserMemoryService) UpdateExperience(userID, id uint, req model.UpdateUserExperienceRequest) error {
	updates := map[string]interface{}{"is_user_edited": true}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.OccurredAt != nil {
		updates["occurred_at"] = *req.OccurredAt
	}
	if req.Tags != nil {
		b, _ := json.Marshal(*req.Tags)
		updates["tags"] = string(b)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	res := DB.Model(&model.UserExperience{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *UserMemoryService) DeleteExperience(userID, id uint) error {
	res := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.UserExperience{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ============ 抽取 ============

type eligibleSession struct {
	HistoryID    uint      `gorm:"column:history_id"`
	UserID       uint      `gorm:"column:user_id"`
	Title        string    `gorm:"column:title"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UserMsgCount int       `gorm:"column:user_msg_count"`
	MaxSortOrder int       `gorm:"column:max_sort_order"`
	LastMsgAt    time.Time `gorm:"column:last_msg_at"`
}

type extractedExperience struct {
	Category   string   `json:"category"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	Confidence float64  `json:"confidence"`
}

type extractionResult struct {
	IsSubstantial bool                 `json:"is_substantial"`
	Experience    *extractedExperience `json:"experience"`
	ProfileFacts  []string             `json:"profile_facts"`
}

// RunExtraction 定时任务: 扫描满足条件的会话并抽取
func (s *UserMemoryService) RunExtraction() error {
	cfg := s.getConfig()
	if !cfg.Enabled {
		log.Println("[UserMemory] 抽取未启用, 跳过")
		return nil
	}
	if s.agentSvc == nil || s.agentSvc.activeModel == nil {
		return errors.New("AI 模型未配置, 无法抽取")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sessions, err := s.collectEligible(0, cfg, true)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return nil
	}
	log.Printf("[UserMemory] 本轮待抽取会话 %d 个", len(sessions))

	var firstErr error
	for _, sess := range sessions {
		if err := s.extractSession(sess, cfg); err != nil {
			log.Printf("[UserMemory] 会话 %d 抽取失败: %v", sess.HistoryID, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// ExtractUserSessions 手动触发某用户全部满足条件的会话抽取(忽略静默等待)
func (s *UserMemoryService) ExtractUserSessions(userID uint) (int, error) {
	cfg := s.getConfig()
	if !cfg.Enabled {
		return 0, errors.New("抽取未启用")
	}
	if s.agentSvc == nil || s.agentSvc.activeModel == nil {
		return 0, errors.New("AI 模型未配置, 无法抽取")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sessions, err := s.collectEligible(userID, cfg, false)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, sess := range sessions {
		if err := s.extractSession(sess, cfg); err != nil {
			log.Printf("[UserMemory] 会话 %d 抽取失败: %v", sess.HistoryID, err)
			continue
		}
		done++
	}
	return done, nil
}

// collectEligible 收集可抽取会话; userID=0 表示全部用户; useIdle 是否要求会话已静默
func (s *UserMemoryService) collectEligible(userID uint, cfg MemoryExtractionConfig, useIdle bool) ([]eligibleSession, error) {
	var rows []eligibleSession

	sql := `SELECT h.id AS history_id, h.user_id, h.title, h.created_at,
       COUNT(CASE WHEN m.role = 'user' THEN 1 END) AS user_msg_count,
       MAX(m.sort_order) AS max_sort_order,
       MAX(m.created_at) AS last_msg_at
FROM training_histories h
JOIN training_messages m ON m.history_id = h.id`
	args := []interface{}{}
	if userID > 0 {
		sql += " WHERE h.user_id = ?"
		args = append(args, userID)
	}
	sql += " GROUP BY h.id, h.user_id, h.title, h.created_at HAVING user_msg_count > ?"
	args = append(args, cfg.MinUserMessages)
	if useIdle {
		sql += " AND last_msg_at <= ?"
		args = append(args, time.Now().Add(-time.Duration(cfg.IdleMinutes)*time.Minute))
	}

	if err := DB.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	// 过滤: 正在处理/已完成且无新消息的会话
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.HistoryID)
	}
	var states []model.MemoryExtractionState
	DB.Where("history_id IN ?", ids).Find(&states)
	stateMap := make(map[uint]model.MemoryExtractionState, len(states))
	for _, st := range states {
		stateMap[st.HistoryID] = st
	}

	result := make([]eligibleSession, 0, len(rows))
	for _, r := range rows {
		st, ok := stateMap[r.HistoryID]
		if ok {
			if st.Status == model.MemoryExtractionStatusProcessing {
				continue
			}
			if (st.Status == model.MemoryExtractionStatusDone || st.Status == model.MemoryExtractionStatusSkipped) &&
				st.LastSortOrder >= r.MaxSortOrder {
				continue
			}
		}
		result = append(result, r)
		if len(result) >= 50 {
			break
		}
	}
	return result, nil
}

func (s *UserMemoryService) extractSession(sess eligibleSession, cfg MemoryExtractionConfig) error {
	var msgs []model.TrainingMessage
	if err := DB.Where("history_id = ?", sess.HistoryID).Order("sort_order ASC").Find(&msgs).Error; err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	s.setState(sess.HistoryID, sess.UserID, model.MemoryExtractionStatusProcessing, 0, "")

	transcript := buildTranscript(msgs)
	res, err := s.extractWithLLM(sess.Title, transcript, cfg.Model)
	if err != nil {
		s.setState(sess.HistoryID, sess.UserID, model.MemoryExtractionStatusFailed, 0, err.Error())
		return err
	}

	if !res.IsSubstantial || res.Experience == nil || strings.TrimSpace(res.Experience.Title) == "" {
		// 无实质内容: 不产生经历, 但仍贡献画像事实
		if len(res.ProfileFacts) > 0 {
			if err := s.MergeProfileFacts(sess.UserID, res.ProfileFacts); err != nil {
				log.Printf("[UserMemory] 画像合并失败 user=%d err=%v", sess.UserID, err)
			}
		}
		s.setState(sess.HistoryID, sess.UserID, model.MemoryExtractionStatusSkipped, sess.MaxSortOrder, "")
		return nil
	}

	if err := s.upsertSessionExperience(sess, res.Experience); err != nil {
		s.setState(sess.HistoryID, sess.UserID, model.MemoryExtractionStatusFailed, 0, err.Error())
		return err
	}
	if len(res.ProfileFacts) > 0 {
		if err := s.MergeProfileFacts(sess.UserID, res.ProfileFacts); err != nil {
			log.Printf("[UserMemory] 画像合并失败 user=%d err=%v", sess.UserID, err)
		}
	}
	s.setState(sess.HistoryID, sess.UserID, model.MemoryExtractionStatusDone, sess.MaxSortOrder, "")
	return nil
}

func (s *UserMemoryService) upsertSessionExperience(sess eligibleSession, exp *extractedExperience) error {
	hid := sess.HistoryID
	tags, _ := json.Marshal(exp.Tags)
	occurred := sess.CreatedAt

	var existing model.UserExperience
	err := DB.Where("history_id = ?", hid).First(&existing).Error
	if err == nil {
		if existing.IsUserEdited {
			return nil
		}
		return DB.Model(&model.UserExperience{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"category":    exp.Category,
			"title":       exp.Title,
			"content":     exp.Content,
			"tags":        string(tags),
			"confidence":  exp.Confidence,
			"occurred_at": occurred,
			"status":      "active",
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	record := &model.UserExperience{
		UserID:     sess.UserID,
		HistoryID:  &hid,
		Category:   exp.Category,
		Title:      exp.Title,
		Content:    exp.Content,
		OccurredAt: &occurred,
		Tags:       string(tags),
		Confidence: exp.Confidence,
		Status:     "active",
	}
	return DB.Create(record).Error
}

func (s *UserMemoryService) setState(historyID, userID uint, status string, lastSort int, errMsg string) {
	var st model.MemoryExtractionState
	err := DB.Where("history_id = ?", historyID).First(&st).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		st = model.MemoryExtractionState{
			HistoryID: historyID,
			UserID:    userID,
			Status:    status,
		}
		if lastSort > 0 {
			st.LastSortOrder = lastSort
		}
		if errMsg != "" {
			st.Error = errMsg
		}
		if status == model.MemoryExtractionStatusDone {
			now := time.Now()
			st.ExtractedAt = &now
		}
		DB.Create(&st)
		return
	}
	updates := map[string]interface{}{"status": status, "error": errMsg}
	if lastSort > 0 {
		updates["last_sort_order"] = lastSort
	}
	if status == model.MemoryExtractionStatusDone {
		updates["extracted_at"] = time.Now()
	}
	DB.Model(&model.MemoryExtractionState{}).Where("history_id = ?", historyID).Updates(updates)
}

// ============ LLM ============

const extractSystemPrompt = `你是一个用户记忆抽取助手。请从一段用户与 AI 的对话中，抽取可用于构建长期用户画像的信息，并判断这段对话是否构成一段"值得记录的用户经历"。

要求：
1. 只输出一个 JSON 对象，不要输出解释，也不要使用 markdown 代码块。
2. JSON 结构：
{
  "is_substantial": true,
  "experience": {
    "category": "work|project|study|achievement|challenge|other",
    "title": "不超过20字、能概括这段经历的标题",
    "content": "对该经历的客观总结，2~4句",
    "tags": ["关键词"],
    "confidence": 0.8
  },
  "profile_facts": ["关于用户的稳定事实或偏好，例如职业、目标、技能水平、兴趣、习惯"]
}
3. is_substantial 表示对话是否包含实质内容(项目/学习/工作/决策/成就/困难等)；寒暄、纯事实问答、无实质信息时为 false，此时 experience 为 null。
4. profile_facts 只记录与"用户本人"相关的信息，没有则为空数组。`

const mergeSystemPrompt = `你是用户画像维护助手。给定【当前画像】和【新增事实】，请合并成最新画像。

要求：
1. 只输出一个 JSON 对象，不要输出解释，也不要使用 markdown 代码块。
2. JSON 结构：
{"summary":"一段话整体画像，不超过200字","dimensions":{"维度名":"值"},"tags":["标签"],"confidence":0.8}
3. 保留仍然有效的信息，用新增事实修正冲突项，不要编造不存在的信息。`

func (s *UserMemoryService) extractWithLLM(title, transcript, modelOverride string) (*extractionResult, error) {
	userPrompt := fmt.Sprintf("会话标题：%s\n\n对话内容：\n%s", title, transcript)
	raw, err := s.agentSvc.GenerateText(modelOverride, extractSystemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}
	var res extractionResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &res); err != nil {
		return nil, fmt.Errorf("解析抽取结果失败: %w", err)
	}
	return &res, nil
}

// MergeProfileFacts 将新增事实合并进用户画像; 用户手动编辑过的画像不覆盖
func (s *UserMemoryService) MergeProfileFacts(userID uint, facts []string) error {
	if len(facts) == 0 {
		return nil
	}
	cfg := s.getConfig()
	if !cfg.Enabled {
		return nil
	}
	if s.agentSvc == nil || s.agentSvc.activeModel == nil {
		return errors.New("AI 模型未配置")
	}
	p, err := s.GetOrCreatePortrait(userID)
	if err != nil {
		return err
	}
	if p.IsUserEdited {
		return nil
	}

	current := map[string]interface{}{
		"summary":    p.Summary,
		"dimensions": json.RawMessage(orDefault(p.Dimensions, "{}")),
		"tags":       json.RawMessage(orDefault(p.Tags, "[]")),
	}
	curJSON, _ := json.Marshal(current)
	factsJSON, _ := json.Marshal(facts)
	userPrompt := fmt.Sprintf("【当前画像】\n%s\n\n【新增事实】\n%s", string(curJSON), string(factsJSON))

	raw, err := s.agentSvc.GenerateText(cfg.Model, mergeSystemPrompt, userPrompt)
	if err != nil {
		return err
	}
	var merged struct {
		Summary    string            `json:"summary"`
		Dimensions map[string]string `json:"dimensions"`
		Tags       []string          `json:"tags"`
		Confidence float64           `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &merged); err != nil {
		return fmt.Errorf("解析画像合并结果失败: %w", err)
	}
	dims, _ := json.Marshal(merged.Dimensions)
	tags, _ := json.Marshal(merged.Tags)

	return DB.Model(&model.UserPortrait{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"summary":    merged.Summary,
		"dimensions": string(dims),
		"tags":       string(tags),
		"confidence": merged.Confidence,
	}).Error
}

// ============ 定时任务 ============

func (s *UserMemoryService) RegisterCronTasks(js *JobScheduler) {
	if js == nil {
		return
	}
	js.RegisterTask("memory.extract_sessions", "抽取用户画像与经历(会话增量)", json.RawMessage(`{}`),
		func(_ *model.JobDefinition, _ json.RawMessage) error {
			return s.RunExtraction()
		})
}

// ============ 辅助 ============

func buildTranscript(msgs []model.TrainingMessage) string {
	var b strings.Builder
	for _, m := range msgs {
		if m.Role == "system" {
			continue
		}
		role := "用户"
		if m.Role == "assistant" {
			role = "AI"
		}
		b.WriteString(role + "：" + m.Content + "\n")
	}
	out := b.String()
	const maxLen = 12000
	runes := []rune(out)
	if len(runes) > maxLen {
		out = string(runes[:maxLen])
	}
	return out
}

// extractJSON 去掉 LLM 输出中可能包裹的 markdown 代码块, 并截取首个 JSON 对象
func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "json"))
		if j := strings.Index(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	return s
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}