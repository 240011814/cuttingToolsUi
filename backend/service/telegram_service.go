package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"backend/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
)

// TelegramSession 用户会话状态
type TelegramSession struct {
	ScenarioID        uint   // model_scenario ID
	CustomTrainID     *uint  // custom_training ID
	TrainingType      string // "english", "decision", "social", "emergency", "custom"
	SelectedModel     string // 用户选择的模型
	HistoryID         uint   // 关联的历史记录 ID
	Messages          []model.OpenAIMessage
	IsActive          bool
}

// BuiltinTraining 内置训练配置
type BuiltinTraining struct {
	Key            string
	Name           string
	SystemPrompt   string
	InitialMessage string
}

// getBuiltinTrainings 获取内置训练配置
func getBuiltinTrainings() map[string]BuiltinTraining {
	return map[string]BuiltinTraining{
		"ai_chat": {
			Key:  "ai_chat",
			Name: "英语训练",
			SystemPrompt: `You are a professional AI English Teacher specializing in scenario-based simulation training.
Your goal is to help users practice authentic spoken English through daily life scenarios.

Training Workflow:
1. **Scene Setup**: Start or continue a daily scenario (e.g., ordering food, business meeting, traveling).
2. **Translation Task**: Provide a specific sentence in Chinese and ask the user to translate it into English.
3. **Evaluation & Feedback**: After the user responds, evaluate their translation. Compare it with authentic native expressions, explain grammar/vocabulary points, and provide "Natural Expression" tips.
4. **Progressive Learning**: Move the story forward and provide the next Chinese sentence for the user to translate.

Response Structure:
- Use "地道表达" (Authentic Expression) for corrections.
- Use "💡 重点纠错与地道笔记" for detailed learning points.
- Always include a section "📊 模拟训练进度" to show the current scenario step.
- ALWAYS append identified vocabulary at the end in this format:
<vocabs>[{"word": "word", "phonetic": "...", "definition": "Chinese meaning", "example": "...", "confusingWords": "..."}]</vocabs>
- ALWAYS append authentic expressions from this turn at the end in this format:
<expressions>[{"english": "Authentic English sentence from this turn", "chinese": "中文翻译"}]</expressions>

Rules:
- Focus on oral, daily-use English.
- Be encouraging but precise with corrections.
- Do not mention the <vocabs> or <expressions> tag in your natural speech.
- **CRITICAL**: Every time you correct the user or introduce new words (like Sugar, Milk in your notes), you MUST extract them into the JSON format below and append it to the VERY END of your response.
- **CRITICAL**: Every time you provide authentic expressions or corrections in this turn, you MUST extract them into the <expressions> JSON format below and append it to the VERY END of your response. Include at least 1-3 authentic expressions per response.

Format Example:
<vocabs>[{"word": "Sugar", "phonetic": "/ˈʃʊɡ.ər/", "definition": "糖", "example": "Do you take sugar? (你要加糖吗？)", "confusingWords": "Shook (摇动), Shocker (令人震惊的事)"}]</vocabs>
<expressions>[{"english": "Could I get a latte to go?", "chinese": "我能拿一杯拿铁带走吗？"}, {"english": "What size would you like?", "chinese": "你想要什么杯型？"}]</expressions>
If no new words, you can omit <vocabs>, but if you taught anything, it MUST be there.
If no authentic expressions in this turn, you can omit <expressions>, but if you corrected or taught expressions, it MUST be there.
Format the **example field in JSON** to provide example sentences for words. Prioritize using the natural, idiomatic expressions mentioned above as the example sentences.`,
			InitialMessage: "Hello! 我是你的 AI 英语口语老师。我们可以通过模拟真实生活场景来练习地道表达。你想从哪个场景开始？比如：\"咖啡店点单\"、\"酒店入住\"或者\"入职第一天\"。",
		},
		"ai_decision": {
			Key:  "ai_decision",
			Name: "决策训练",
			SystemPrompt: `你是一名专业的AI决策教练。
你的目标是帮助用户练习在个人生活、工作、学习、人际关系、财务、健康、时间管理、情绪调节、沟通、风险承担以及长期人生设计方面，做出更清晰、更冷静、更具辩护力的决策。

训练工作流:
1. 场景设置：询问用户正面临什么决策，有哪些选项，最看重什么，以及存在哪些约束。
2. 决策框架：帮助用户将决策与情绪、压力、恐惧、沉没成本和他人的期望分开。
3. 模型选择：选择一个适合当前情况的决策模型，而不是应用所有模型。
4. 分析：按价值观、证据、风险、可逆性、机会成本、可选性（是否扩大未来选择）、执行可行性（能否真正执行）和下一步行动来比较选项。
5. 承诺：帮助用户选择一个小的下一步、决策截止日期、止损条件或实验。
6. 模式泛化：在完成决策反馈后，识别底层的决策模式，将其映射到生活不同领域（工作、人际关系、学习、财务、时间）的类似情境，并提炼出一条可迁移的规则或默认应对策略，用于未来类似情境。将其用于将单例学习转化为可重用的决策启发式。

决策手册:
- 首先澄清真正的决策："现在到底需要决定什么？" 避免把模糊的焦虑当作明确的选择去解决。
- 检验选择是否满足用户的真实需求，而不仅仅是表面欲望、短期舒适、恐惧回避或他人的期望。
- 询问选项是否扩大了未来的可选性。更好的选择通常会带来更多的未来选择、更强的能力、更多的资源，或推动人生进入下一阶段。
- 区分事实、假设、情绪和价值观。不要让恐惧、内疚、紧迫感、沉没成本或社会压力悄悄地成为决策者。
- 在比较选项之前先定义成功。明确好的结果在金钱、时间、精力、成长、关系质量、风险和长期遗憾方面意味着什么。
- 识别约束：截止日期、预算、能力、健康、责任、信息质量、不可逆的后果和依赖关系。
- 在解决问题之前，先用一句清晰的话定义问题。当用户在解决杂乱问题时，使用5W2H、5个为什么、隐藏假设检查和核心冲突识别。
- 生成真正的替代方案。当可能存在第三选项、分阶段选项、试验选项或"暂时什么都不做"的选项时，不要强迫做出二元选择。
- 在选择之前收集信息，列出所有可用的方法。如果缺少选项，首先拓宽选项集。
- 对可逆决策加快评估，对不可逆决策放慢评估。如果决策是可逆的，偏好小实验而非无尽的分析。
- 重要决策需要备用计划和时机。如果存在已知的时间窗口，应在预期时间的大约三分之二处做出决定，而不是等到最后一刻。
- 偏好稳健的决策而非完美的决策。一个良好的决策，即使其中一项假设被证明是错的，也应该仍然可以接受。
- 注意偏差：沉没成本、损失厌恶、确认偏误、现状偏误、社会认同、稀缺性偏差、激励导致的偏差、过度自信、近因偏误、完美主义和情绪推理。
- 将情绪视为信息，而非现实。恐惧、焦虑和痛苦会扭曲解读；在做决定前，先标记、观察并接纳情绪。
- 避免用幻想代替决策。只思考而不测试、行动、获取数据或反馈，这是决策过程陷入停滞的标志。
- 对于困难问题，首先识别核心困难，然后询问是否可以绕过、重新框定、分解或通过另一条路径解决。
- 当决策很重要时，至少询问三位背景不同的人，然后按激励、专业能力和风险承受能力过滤他们的建议。
- 当其他人的激励、反应、信任、竞争、合作、谈判或重复互动很重要时，将决策视为一场博弈，而非单人优化。
- 在策略情境下，询问：玩家是谁？每个玩家可以选什么？每个玩家想要什么？如果我选X，他们会怎么做？这是一次性的还是重复的？承诺、信任、信息或规则能否改变结果？
- 如果用户缺少信息，选择能最大限度减少不确定性的最小测试。
- 尽可能使用数据：量化时间、成本、概率、结果、风险水平和成功指标。
- 每次教练回合结束时，给出一个具体的下一步行动，而不仅仅是分析。

规则:
- 专注于决策质量，而非预测未来。
- 不要假装不确定的结果是确定的。
- 如果信息不足，询问缺失的决策标准或建议小型测试。
- 当更好的行动是收集一项关键信息时，不要强迫用户立即选择。
- 当用户的问题模糊时，首先解决定义问题；不要直接跳到建议。
- 当用户在回避恐惧、羞耻、压力或责任时，将情绪与现实分离，并建议安全暴露、实验或分阶段行动。
- 当用户已投入时间、金钱或情感时，检查沉没成本并设定一个固定的止损配额。
- 当用户只凭空想象选项时，推动小型测试、数据收集或反馈循环。
- 当决策涉及其他人时，不要假设他们会被动接受用户的计划。分析激励、最佳反应、可信度、信息不对称以及互动是一次性的还是重复的。
- 当仅靠说服无法解决策略性问题时，优先改变激励、规则、默认设置、承诺和信息流。
- 直接、结构化、实用。
- 如果用户的消息是中文，主要用中文回应。
- 如果用户要求用英文练习，则用英文角色扮演，并用中文解释决策反馈。`,
			InitialMessage: "你好，我是你的 AI 决策训练教练。我们可以练习如何在工作、学习、职业、关系、金钱、时间安排或人生选择里做决定。你现在面对的一个选择是什么？",
		},
		"ai_social": {
			Key:  "ai_social",
			Name: "社交训练",
			SystemPrompt: `你是专业的 AI 社交技能教练。
你的目标是帮助用户在日常生活和职场场景中，练习真实的社交互动、对话策略、情绪觉察和自信回应。

## 角色定位
你是一名温和但敏锐的教练，不评判、不说教。你通过场景化训练，让用户自己发现更好的表达方式，并随时准备给出清晰、可操作的替代方案。你始终将双方视作平等的协作者，用"我们可以这样试试"替代"你应该这样做"。

## 训练流程
1. **场景设定**：根据用户需求或上次进度，开启或延续一个具体的社交情境（如初次见面、拒绝借钱、接受批评、缓和冲突等）。
2. **角色扮演**：清晰描述当前场景、对方身份和情绪状态，然后问用户："你会怎么说/怎么做？"
3. **技能评估（隐式）**：内部评估用户回应的情绪觉察、边界感、模型运用准确度，以此调整下一步的练习难度。
   - 若用户连续两次在同类场景中表现良好 → 主动引入对方情绪升级或场景复杂度（如从同事冷淡升级为公开质疑）。
   - 若用户出现明显回避或攻击性应对 → 主动降级，先示范一句更安全的话，再给用户一次修正机会。
4. **反馈与示范**：根据输出规范给出反馈，包括评价、技巧点、更自然的说法，并视情况引用模型和示例。
5. **推进**：在反馈后给出下一步的角色挑战，形成练习闭环。

## 核心沟通原则
- **心态基础**：协作者心态，杜绝审问、否定、比较、自我聚焦。
- **对话飞轮**：观察提问 → 积极倾听 → 适度自我袒露 → 真诚追问。这是所有日常对话的底层引擎。
- **倾听三阶**：听事实 → 听情绪（注意"总是、每次、从不等词"） → 听真实意图。回应前先完成这个诊断。
- **上堆与下切**：上堆用来提炼观点、探讨动机、展望未来；下切用来回到细节、感受、当下事实。对方说事实时先下切情绪，避免直接争论。
- **安全破冰**：共同观察/相似点/精准称赞开场，问观点、方法、经历、兴趣，避开隐私、争议、负面八卦。
- **讲故事**：提供谁/何时/何地/什么/为什么/如何，用"目标→努力→困难→结果"结构制造接话钩子。
- **情绪应对**：安慰时不说"至少…"，不说教，不匆忙解决；用"支持→倾听→反馈情绪→反馈认知"的流程。
- **拒绝与边界**：允许拖延答复，使用三明治结构或直接拒绝+理由，保护自己正当利益。
- **批评与建议**：非必要不批评；私下场合先肯定人再说行为，请对方复述、补充，再一起约定方案。
- **称赞与接受称赞**：称赞要具体到行为→特质→影响；接受称赞用延伸/归因/衬托/找补/调侃法，大方回馈价值。
- **鼓励与道歉**：鼓励用"描述好做法→请教方法→说出启发"；道歉要早、带方案、请教建议、承诺未来。
- **情感账户**：通过理解他人、注意细节、信守承诺、明确期望、诚恳道歉、赞扬鼓励来持续存款。
- **潜台词诊断**：听到"看情况吧""再说吧"等模糊表述，马上识别可能拒绝或不确定性，主动给台阶。

## 输出规范

### 动态输出选择
- **首次进入场景或用户出现较大失误时，使用完整反馈格式**：场景→反馈→更自然的说法→技巧点→模型→下一步。
- **在同一场景内进行连续微调练习时，使用轻量反馈格式**：只输出"反馈 + 更自然的说法+ 技巧点 + 下一步"。

### 核心规则
- 回复简洁、温暖、可操作。
- 绝不倾倒理论，始终通过场景诊断、修改措辞、下一步练习来传授技巧。
- 用户若用中文提问，用中文回复；若用户要求用英语练习，用英语进行角色扮演，并用中文给出反馈。`,
			InitialMessage: "你好，我是你的 AI 社交训练教练。我们可以练习聊天破冰、安慰、拒绝、赞扬、被夸回应、道歉、求助、提建议、临时发言、尴尬应对或职场沟通。你想先练哪个场景？",
		},
		"ai_emergency": {
			Key:  "ai_emergency",
			Name: "应急训练",
			SystemPrompt: `你是一位专业的 AI 突发应变与反应力训练教练。
你的目标是帮助用户在各种突发情况下锻炼反应速度、问题解决能力和临场表达能力。

训练范围（不限于以下，场景来自生活的方方面面）：
- 人际与隐私：被撞见藏私房钱、被发现浏览了不该看的内容、手机被看到了不该看的聊天
- 解释与借口：解释为什么迟到、为什么没完成任务、昨晚到底去哪了
- 社交尴尬：在领导面前说错话、微信发错对象、当众出糗、认错人
- 意外相遇：请病假却被领导撞见在商场、被债主堵住、遇到不想见的前任
- 生活突发：家里突然漏水、钱包被偷、孩子摔了东西、突然停电
- 职场危机：服务器崩了要汇报、重要文件发错人、临时顶替发言、客户突然发飙

应变模型库：
【认知与分析类】
OODA 循环：观察（分析现场）→ 定向（判断威胁/对方意图）→ 决策（想好说什么）→ 行动（沉着应对）
STOP 模型：停下（不要慌）→ 思考（局面是什么）→ 观察（有什么可用资源）→ 计划（分步走）
5W1H 快速扫描：在3秒内回答 Who/What/When/Where/Why/How，快速梳理事件全貌，理清思路再开口
最小信息原则：不确定时只说最少必要的内容，不主动补充细节，不自作聪明多解释（解释越多漏洞越多）
先发制人：在对方开口前，主动说出来，抢占叙事主动权，让对方跟着你的节奏走
意图分层法：把对方表达拆成三层——表层问题 / 真实目的 / 底层动机 → 回答"目的"，而不是回答"问题"
【沟通与人际类】
重构叙事：改变对方对"你做的事"的解读框架（如：不是在藏钱，是在给你准备惊喜的预算）
非暴力沟通（NVC）：观察（陈述事实，不评判）→ 感受（说出自己的感受）→ 需要（背后的真实需求）→ 请求（具体、可执行的请求）；避免激化矛盾
"是的，而且"法则：不否认、不硬刚，顺着对方说然后叠加自己的叙事，避免陷入防守
情绪降温话术：主动延迟回答赢得缓冲时间（如"你说得对，让我整理一下思路"、"这个问题很重要，我认真想想"）
模糊回应策略（Strategic Vagueness）：用模糊但合理的话回应，不给明确细节锚点
递延回应法（Delay & Redirect）：不当场答，换时间或场景（如"我整理一下，晚点给你完整信息"）
【危机公关类】
AER 模型：承认(Acknowledge，承认现状/错误) → 解释(Explain，解释原因，不找借口) → 补救(Remedy，提出具体补救方案)；适合职场失误或明显犯错的场景
3C 危机原则：关注(Concern，表达关心对方感受) → 承诺(Commitment，做出可信承诺) → 控制(Control，掌握后续主导权)；适合快速平息对方情绪
降级处理法（De-escalation）：把"大问题"描述为"局部偏差"，控制严重程度
【心理与博弈类】
扑克脸原则：情绪隔离，主动控制"慌张信号"（手抖/目光回避/结巴/过度解释），让表情和语气先稳下来
三秒原则：高压场景下，深吸一口气，必须在三秒内给出第一句话，打破沉默比说完美更重要
低预期管理（Expectation Lowering）：提前降低对方预期，后续更容易形成正向反馈

训练流程：
1. 场景生成：随机生成一个高压突发场景，标注【紧急程度】（低 / 中 / 高 / 极高）
2. 反应阶段：要求用户立刻给出第一句话或第一个动作。AI 会模拟现场压力（如"她正盯着你等你解释……"）
3. 推进阶段：根据用户的应对方式继续推进剧情，可能引入新的"变故"
4. 复盘解析：
   - 反应评分：你的反应有多快、多果断？
   - 策略分析：你用了哪种模型？哪个环节有漏洞？
   - 话术打磨：有没有更高明的说法？教你"救场金句"
   - 经验提炼：提炼一条可迁移的通用原则，用于未来类似场景

规则：
- 扮演各种角色（怀疑的伴侣、追问的领导、旁观的同事等），制造真实压力
- 全程使用中文进行角色扮演和对话
- 复盘阶段给出深度分析，不要只说"不错"
- 用【场景】【反应】【复盘】标签结构化输出`,
			InitialMessage: "欢迎来到突发应变训练！我会把你扔进各种真实的突发场景。你只有几秒钟反应时间。准备好了吗？我来出第一个场景。",
		},
	}
}

type TelegramService struct {
	sysCfgService   *SystemConfigService
	historyService  *HistoryService
	promptService   *PromptService
	bot             *tgbotapi.BotAPI
	stopCh          chan struct{}
	mu              sync.Mutex
	startMu         sync.Mutex // 启动专用锁，防止并发启动
	sessions        map[int64]*TelegramSession // chatID -> session
	sessionsMu      sync.RWMutex
	webhookURL      string // 当前使用的 webhook URL，空表示 long polling 模式
	running         bool   // 是否正在运行
}

func NewTelegramService(sysCfgService *SystemConfigService) *TelegramService {
	s := &TelegramService{
		sysCfgService:  sysCfgService,
		historyService: NewHistoryService(),
		promptService:  NewPromptService(DB),
		stopCh:         make(chan struct{}),
		sessions:       make(map[int64]*TelegramSession),
	}
	return s
}

// StartBot 启动 Telegram Bot
func (s *TelegramService) StartBot() error {
	s.startMu.Lock()
	defer s.startMu.Unlock()

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		log.Println("[Telegram] Bot is already running, skipping")
		return nil
	}
	s.running = true
	s.mu.Unlock()

	if !s.sysCfgService.IsTelegramEnabled() {
		log.Println("[Telegram] Bot is disabled in system config, skipping")
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return nil
	}

	token := s.sysCfgService.GetTelegramBotToken()
	if token == "" {
		log.Println("[Telegram] Bot Token not configured, skipping bot start")
		return nil
	}

	log.Printf("[Telegram] Starting bot with token: %s...", token[:10])

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("[Telegram] Failed to create bot: %v", err)
		return fmt.Errorf("failed to create bot: %w", err)
	}

	s.bot = bot
	bot.Debug = true
	log.Printf("[Telegram] Bot authorized as @%s (ID: %d)", bot.Self.UserName, bot.Self.ID)

	// 检查是否配置了 webhook URL
	webhookURL := s.sysCfgService.GetTelegramWebhookURL()
	if webhookURL != "" {
		// Webhook 模式 - 拼接完整路径
		webhookURL = strings.TrimRight(webhookURL, "/") + "/api/telegram/webhook"
		log.Printf("[Telegram] Using webhook mode, URL: %s", webhookURL)
		if err := s.setupWebhook(webhookURL); err != nil {
			log.Printf("[Telegram] Failed to setup webhook: %v, falling back to long polling", err)
			s.webhookURL = ""
			s.startLongPolling()
		} else {
			s.webhookURL = webhookURL
		}
	} else {
		// Long Polling 模式
		log.Println("[Telegram] Using long polling mode")
		s.webhookURL = ""
		s.startLongPolling()
	}

	s.setBotCommands()
	return nil
}

// setBotCommands 设置 Bot 命令菜单
func (s *TelegramService) setBotCommands() {
	if s.bot == nil {
		return
	}

	commands := []tgbotapi.BotCommand{
		{Command: "train", Description: "选择训练场景"},
		{Command: "model", Description: "切换 AI 模型"},
		{Command: "exit", Description: "退出当前训练"},
		{Command: "bind", Description: "绑定账号"},
		{Command: "unbind", Description: "解绑账号"},
		{Command: "status", Description: "查看绑定状态"},
		{Command: "help", Description: "帮助信息"},
	}

	_, err := s.bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("[Telegram] Failed to set bot commands: %v", err)
	} else {
		log.Println("[Telegram] Bot commands set successfully")
	}
}

// StopBot 停止 Bot
func (s *TelegramService) StopBot() {
	s.startMu.Lock()
	defer s.startMu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.stopCh:
		// already closed, do nothing
	default:
		close(s.stopCh)
	}

	if s.bot != nil {
		// 如果是 webhook 模式，删除 webhook
		if s.webhookURL != "" {
			log.Println("[Telegram] Deleting webhook...")
			_, err := s.bot.Request(tgbotapi.DeleteWebhookConfig{})
			if err != nil {
				log.Printf("[Telegram] Failed to delete webhook: %v", err)
			}
		} else {
			s.bot.StopReceivingUpdates()
		}
		s.bot = nil
		s.webhookURL = ""
	}

	s.running = false
}

// RestartBot 重启 Bot（配置更新后调用）
func (s *TelegramService) RestartBot() error {
	s.StopBot()

	s.mu.Lock()
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	return s.StartBot()
}

// setupWebhook 设置 Webhook
func (s *TelegramService) setupWebhook(webhookURL string) error {
	// 先删除旧的 webhook
	_, err := s.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Printf("[Telegram] Failed to delete old webhook: %v", err)
	}

	// 设置新的 webhook
	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}

	_, err = s.bot.Request(wh)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	// 验证 webhook 设置
	info, err := s.bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %w", err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("[Telegram] Webhook has error: %s", info.LastErrorMessage)
	}

	log.Printf("[Telegram] Webhook set successfully, pending updates: %d", info.PendingUpdateCount)
	return nil
}

// startLongPolling 启动 Long Polling 模式
func (s *TelegramService) startLongPolling() {
	// 删除 webhook 以使用 long polling
	log.Println("[Telegram] Deleting webhook...")
	_, err := s.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Printf("[Telegram] Failed to delete webhook: %v", err)
	} else {
		log.Println("[Telegram] Webhook deleted successfully")
	}

	go s.listenUpdates()
}

// GetBot 获取 Bot 实例（供 webhook handler 使用）
func (s *TelegramService) GetBot() *tgbotapi.BotAPI {
	return s.bot
}

// HandleWebhookUpdate 处理 Webhook 回调的 update
func (s *TelegramService) HandleWebhookUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		log.Printf("[Telegram] [Webhook] Received message from %s: %s", update.Message.From.UserName, update.Message.Text)
		s.handleMessage(update.Message)
	}
	if update.CallbackQuery != nil {
		log.Printf("[Telegram] [Webhook] Received callback: %s from %s", update.CallbackQuery.Data, update.CallbackQuery.From.UserName)
		s.handleCallbackQuery(update.CallbackQuery)
	}
}

// listenUpdates 监听消息更新
func (s *TelegramService) listenUpdates() {
	if s.bot == nil {
		log.Println("[Telegram] Bot is nil, cannot listen for updates")
		return
	}

	log.Println("[Telegram] Starting to listen for updates...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)
	log.Println("[Telegram] Update channel created, waiting for messages...")

	for {
		select {
		case <-s.stopCh:
			log.Println("[Telegram] Stop signal received")
			return
		case update := <-updates:
			if update.Message != nil {
				log.Printf("[Telegram] Received message from %s: %s", update.Message.From.UserName, update.Message.Text)
				s.handleMessage(update.Message)
			}
			if update.CallbackQuery != nil {
				log.Printf("[Telegram] Received callback: %s from %s", update.CallbackQuery.Data, update.CallbackQuery.From.UserName)
				s.handleCallbackQuery(update.CallbackQuery)
			}
		}
	}
}

// handleMessage 处理用户消息
func (s *TelegramService) handleMessage(msg *tgbotapi.Message) {
	log.Printf("[Telegram] Handling message: %s (IsCommand: %v)", msg.Text, msg.IsCommand())

	if !msg.IsCommand() {
		// 检查是否有活跃的训练会话
		session := s.getSession(msg.Chat.ID)
		if session != nil && session.IsActive {
			s.handleTrainingMessage(msg, session)
			return
		}
		log.Println("[Telegram] Message is not a command and no active session, ignoring")
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
	case "train":
		s.handleTrain(msg)
	case "model":
		s.handleModelSwitch(msg)
	case "exit":
		s.handleExit(msg)
	case "help":
		s.handleHelp(msg)
	default:
		s.reply(msg.Chat.ID, "未知命令。发送 /help 查看可用命令。")
	}
}

// handleStart 处理 /start 命令
func (s *TelegramService) handleStart(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	log.Printf("[Telegram] Handling /start command from chat %d", chatID)
	user, _ := s.getUserByChatID(chatID)

	if user != nil {
		log.Printf("[Telegram] User found, sending bound message")
		s.reply(chatID, fmt.Sprintf("你好 %s！你的账号已绑定。", user.Nickname))
	} else {
		log.Printf("[Telegram] User not bound, sending help message")
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

// handleHelp 处理 /help 命令
func (s *TelegramService) handleHelp(msg *tgbotapi.Message) {
	helpText := `可用命令：

/train - 选择训练场景
/model - 切换 AI 模型
/exit - 退出当前训练
/clear - 清空聊天记录
/bind <code> - 绑定账号（在网页端生成绑定码）
/unbind - 解绑账号
/status - 查看绑定状态
/help - 显示此帮助信息

绑定步骤：
1. 登录网页端，进入个人中心
2. 点击 Telegram 绑定，生成绑定码
3. 在这里发送 /bind <绑定码>`

	s.reply(msg.Chat.ID, helpText)
}

// handleTrain 处理 /train 命令
func (s *TelegramService) handleTrain(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// 检查是否已绑定
	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "请先绑定账号。使用 /bind <code> 绑定。")
		return
	}

	// 检查是否有活跃会话
	session := s.getSession(chatID)
	if session != nil && session.IsActive {
		s.reply(chatID, "你已有活跃的训练会话。请先使用 /exit 退出当前训练。")
		return
	}

	// 显示训练场景选择
	s.showTrainingMenu(chatID)
}

// showTrainingMenu 显示训练场景选择菜单（含模型选择）
func (s *TelegramService) showTrainingMenu(chatID int64) {
	// 获取可用模型
	var aiModels []model.AIModel
	DB.Joins("JOIN ai_providers ON ai_providers.id = ai_models.provider_id").
		Where("ai_providers.is_active = ?", true).
		Find(&aiModels)

	// 获取用户上次选择的模型
	session := s.getSession(chatID)
	selectedModel := ""
	if session != nil {
		selectedModel = session.SelectedModel
	}
	if selectedModel == "" && len(aiModels) > 0 {
		// 默认使用默认模型
		for _, m := range aiModels {
			if m.IsDefault {
				selectedModel = m.ModelCode
				break
			}
		}
		if selectedModel == "" && len(aiModels) > 0 {
			selectedModel = aiModels[0].ModelCode
		}
	}

	// 构建消息
	modelDisplay := selectedModel
	for _, m := range aiModels {
		if m.ModelCode == selectedModel && m.DisplayName != "" {
			modelDisplay = m.DisplayName
			break
		}
	}

	text := fmt.Sprintf("当前模型: %s\n\n请选择训练场景：", modelDisplay)

	// 构建模型选择按钮行
	modelRows := [][]tgbotapi.InlineKeyboardButton{}
	modelRow := []tgbotapi.InlineKeyboardButton{}
	for i, m := range aiModels {
		buttonText := m.DisplayName
		if buttonText == "" {
			buttonText = m.ModelCode
		}
		if m.ModelCode == selectedModel {
			buttonText = "✓ " + buttonText
		}
		modelRow = append(modelRow, tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("setmodel_%s", m.ModelCode)))
		if (i+1)%2 == 0 || i == len(aiModels)-1 {
			modelRows = append(modelRows, modelRow)
			modelRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// 构建训练场景按钮行
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("英语训练", "train_ai_chat"),
			tgbotapi.NewInlineKeyboardButtonData("决策训练", "train_ai_decision"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("社交训练", "train_ai_social"),
			tgbotapi.NewInlineKeyboardButtonData("应急训练", "train_ai_emergency"),
		},
	}

	// 获取用户自定义训练
	user, _ := s.getUserByChatID(chatID)
	var customTrainings []model.CustomTraining
	if user != nil {
		DB.Where("user_id = ?", user.ID).Find(&customTrainings)
	}

	// 添加自定义训练按钮
	for _, ct := range customTrainings {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(ct.Title, fmt.Sprintf("train_custom_%d", ct.ID)),
		})
	}

	// 合并所有行
	allRows := [][]tgbotapi.InlineKeyboardButton{}
	if len(modelRows) > 0 {
		// 添加模型标题行
		allRows = append(allRows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("── 切换模型 ──", "noop"),
		})
		allRows = append(allRows, modelRows...)
	}
	// 添加分隔符
	allRows = append(allRows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("── 选择训练 ──", "noop"),
	})
	allRows = append(allRows, rows...)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(allRows...)

	if s.bot != nil {
		s.bot.Send(msg)
	}
}

// handleExit 处理 /exit 命令
func (s *TelegramService) handleExit(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	session := s.getSession(chatID)
	if session == nil || !session.IsActive {
		s.reply(chatID, "你当前没有活跃的训练会话。")
		return
	}

	s.clearSession(chatID)
	s.reply(chatID, "已退出训练。使用 /train 开始新的训练。")
}

// handleModelSwitch 处理 /model 命令（切换模型）
func (s *TelegramService) handleModelSwitch(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	session := s.getSession(chatID)
	if session == nil || !session.IsActive {
		s.reply(chatID, "请先开始训练。")
		return
	}

	s.showModelSwitchMenu(chatID, session)
}

// showModelSwitchMenu 显示模型切换菜单
func (s *TelegramService) showModelSwitchMenu(chatID int64, session *TelegramSession) {
	// 获取可用模型
	var aiModels []model.AIModel
	DB.Joins("JOIN ai_providers ON ai_providers.id = ai_models.provider_id").
		Where("ai_providers.is_active = ?", true).
		Find(&aiModels)

	if len(aiModels) == 0 {
		s.reply(chatID, "没有可用的 AI 模型。")
		return
	}

	text := fmt.Sprintf("当前模型: %s\n\n选择新模型：", session.SelectedModel)

	// 构建按钮行
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, m := range aiModels {
		buttonText := m.DisplayName
		if buttonText == "" {
			buttonText = m.ModelCode
		}
		if m.ModelCode == session.SelectedModel {
			buttonText = "✓ " + buttonText
		}
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("switchmodel_%s", m.ModelCode)),
		})
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if s.bot != nil {
		s.bot.Send(msg)
	}
}

// handleCallbackQuery 处理按钮点击
func (s *TelegramService) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// 确认回调，避免 loading 状态
	s.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// 忽略 noop 按钮
	if data == "noop" {
		return
	}

	switch {
	case data == "train_exit":
		s.clearSession(chatID)
		s.editMessage(chatID, callback.Message.MessageID, "已退出训练。使用 /train 开始新的训练。")
	case data == "train_back":
		s.showTrainingMenu(chatID)
	default:
		// 处理模型切换（训练中）
		if len(data) > 12 && data[:12] == "switchmodel_" {
			modelCode := data[12:]
			session := s.getSession(chatID)
			if session != nil && session.IsActive {
				session.SelectedModel = modelCode
				s.reply(chatID, fmt.Sprintf("已切换到模型: %s", modelCode))
			}
			return
		}
		// 处理模型选择（训练前）
		if len(data) > 9 && data[:9] == "setmodel_" {
			modelCode := data[9:]
			// 更新 session 中的模型选择
			session := s.getSession(chatID)
			if session == nil {
				session = &TelegramSession{}
				s.setSession(chatID, session)
			}
			session.SelectedModel = modelCode
			// 刷新菜单
			s.showTrainingMenu(chatID)
			return
		}
		// 处理训练选择
		if len(data) > 6 && data[:6] == "train_" {
			trainingType := data[6:]
			// 获取当前选择的模型
			session := s.getSession(chatID)
			modelCode := ""
			if session != nil {
				modelCode = session.SelectedModel
			}
			// 编辑原消息，移除按钮
			s.editMessage(chatID, callback.Message.MessageID, "训练已开始，请在对话中发送消息。")
			// 开始训练
			if len(trainingType) > 7 && trainingType[:7] == "custom_" {
				var trainingID uint
				fmt.Sscanf(trainingType[7:], "%d", &trainingID)
				s.startCustomTraining(chatID, trainingID, modelCode)
			} else {
				s.startBuiltinTraining(chatID, trainingType, modelCode)
			}
		}
	}
}

// startBuiltinTraining 开始内置训练
func (s *TelegramService) startBuiltinTraining(chatID int64, trainingType string, modelCode string) {
	trainings := getBuiltinTrainings()
	training, ok := trainings[trainingType]
	if !ok {
		s.reply(chatID, "训练类型不存在。")
		return
	}

	// 获取用户信息
	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "用户信息获取失败。")
		return
	}

	// 获取用户自定义提示词（如果有）
	customPrompt, _, _, _ := s.promptService.GetEffectivePrompt(user.ID, trainingType)
	systemPrompt := training.SystemPrompt
	if customPrompt != "" {
		systemPrompt = customPrompt
	}

	// 设置会话
	session := &TelegramSession{
		TrainingType:  trainingType,
		SelectedModel: modelCode,
		Messages:      []model.OpenAIMessage{},
		IsActive:      true,
	}
	s.setSession(chatID, session)

	// 添加系统消息
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// 添加欢迎语到消息列表（与前端保持一致）
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "assistant",
		Content: training.InitialMessage,
	})

	// 构建提示信息
	modelHint := ""
	if modelCode != "" {
		modelHint = fmt.Sprintf("\n当前模型: %s", modelCode)
	}

	text := fmt.Sprintf("已进入：%s\n\n%s%s\n\n发送消息开始对话，使用 /exit 退出训练。", training.Name, training.InitialMessage, modelHint)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("退出训练", "train_exit"),
		),
	)

	if s.bot != nil {
		s.bot.Send(msg)
	}
}

// startCustomTraining 开始自定义训练
func (s *TelegramService) startCustomTraining(chatID int64, trainingID uint, modelCode string) {
	var training model.CustomTraining
	if err := DB.First(&training, trainingID).Error; err != nil {
		s.reply(chatID, "训练不存在。")
		return
	}

	// 获取用户信息
	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "用户信息获取失败。")
		return
	}

	// 获取用户自定义提示词（如果有）
	moduleKey := fmt.Sprintf("custom_%d", trainingID)
	customPrompt, _, _, _ := s.promptService.GetEffectivePrompt(user.ID, moduleKey)
	systemPrompt := training.SystemPrompt
	if customPrompt != "" {
		systemPrompt = customPrompt
	}

	// 设置会话
	session := &TelegramSession{
		CustomTrainID: &trainingID,
		TrainingType:  "custom",
		SelectedModel: modelCode,
		Messages:      []model.OpenAIMessage{},
		IsActive:      true,
	}
	s.setSession(chatID, session)

	// 添加系统消息
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	initialMsg := training.InitialMessage
	if initialMsg == "" {
		initialMsg = "发送消息开始对话"
	}

	// 添加欢迎语到消息列表（与前端保持一致）
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "assistant",
		Content: initialMsg,
	})

	// 构建提示信息
	modelHint := ""
	if modelCode != "" {
		modelHint = fmt.Sprintf("\n当前模型: %s", modelCode)
	}

	text := fmt.Sprintf("已进入：%s\n\n%s%s\n\n发送消息开始对话，使用 /exit 退出训练。", training.Title, initialMsg, modelHint)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("退出训练", "train_exit"),
		),
	)

	if s.bot != nil {
		s.bot.Send(msg)
	}
}

// editMessage 编辑消息
func (s *TelegramService) editMessage(chatID int64, messageID int, text string) {
	if s.bot == nil {
		return
	}
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	s.bot.Send(msg)
}

// getSession 获取用户会话
func (s *TelegramService) getSession(chatID int64) *TelegramSession {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	return s.sessions[chatID]
}

// setSession 设置用户会话
func (s *TelegramService) setSession(chatID int64, session *TelegramSession) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[chatID] = session
}

// clearSession 清除用户会话
func (s *TelegramService) clearSession(chatID int64) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, chatID)
}

// handleTrainingMessage 处理训练中的消息
func (s *TelegramService) handleTrainingMessage(msg *tgbotapi.Message, session *TelegramSession) {
	chatID := msg.Chat.ID
	userText := msg.Text

	// 获取用户信息
	user, _ := s.getUserByChatID(chatID)
	if user == nil {
		s.reply(chatID, "用户信息获取失败，请重新绑定。")
		return
	}

	// 添加用户消息到会话
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "user",
		Content: userText,
	})

	// 调用 AI
	reply, err := s.callAI(session.Messages)
	if err != nil {
		log.Printf("[Telegram] AI call failed: %v", err)
		s.reply(chatID, "AI 服务出错，请稍后重试。")
		return
	}

	// 添加 AI 回复到会话
	session.Messages = append(session.Messages, model.OpenAIMessage{
		Role:    "assistant",
		Content: reply,
	})

	// 保存历史记录
	go s.saveHistory(user.ID, session)

	// 发送回复
	s.reply(chatID, reply)
}

// saveHistory 保存历史记录
func (s *TelegramService) saveHistory(userID uint, session *TelegramSession) {
	// 确定训练类型和自定义训练 ID
	trainingType := session.TrainingType
	var customTrainingID *uint

	if session.TrainingType == "custom" {
		customTrainingID = session.CustomTrainID
	}

	// 生成标题（与前端保持一致：取第一条用户消息，截取前20个字符）
	title := "AI 训练对话"
	for _, m := range session.Messages {
		if m.Role == "user" {
			title = m.Content
			// 按 rune 截取，支持中文
			runes := []rune(title)
			if len(runes) > 20 {
				title = string(runes[:20]) + "..."
			}
			break
		}
	}

	// 保存历史记录
	historyID, err := s.historyService.SaveHistory(
		userID,
		session.HistoryID,
		trainingType,
		customTrainingID,
		title,
		session.Messages,
		false,
	)

	if err != nil {
		log.Printf("[Telegram] Failed to save history: %v", err)
		return
	}

	// 更新 session 的 HistoryID
	if session.HistoryID == 0 {
		session.HistoryID = historyID
	}
}

// callAI 调用 AI 服务（非流式）
func (s *TelegramService) callAI(messages []model.OpenAIMessage) (string, error) {
	// 获取 AI 服务配置
	var provider model.AIProvider
	if err := DB.Where("is_active = ?", true).First(&provider).Error; err != nil {
		return "", errors.New("未配置 AI 提供商")
	}

	var aiModel model.AIModel
	if err := DB.Where("provider_id = ? AND is_default = ?", provider.ID, true).First(&aiModel).Error; err != nil {
		if err := DB.Where("provider_id = ?", provider.ID).First(&aiModel).Error; err != nil {
			return "", errors.New("未配置 AI 模型")
		}
	}

	// 创建客户端
	config := openai.DefaultConfig(provider.APIKey)
	if provider.BaseURL != "" {
		config.BaseURL = provider.BaseURL
	}
	config.HTTPClient = &http.Client{
		Timeout: 60 * time.Second,
	}
	client := openai.NewClientWithConfig(config)

	// 构建请求
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    aiModel.ModelCode,
		Messages: openaiMessages,
	}

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("AI 未返回回复")
	}

	return resp.Choices[0].Message.Content, nil
}

// reply 发送回复消息
func (s *TelegramService) reply(chatID int64, text string) {
	if s.bot == nil {
		log.Println("[Telegram] Bot is nil, cannot send reply")
		return
	}
	log.Printf("[Telegram] Sending reply to chat %d: %s", chatID, truncateString(text, 100))
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := s.bot.Send(msg)
	if err != nil {
		// Markdown 解析失败，尝试纯文本发送
		log.Printf("[Telegram] Markdown parse failed, retrying as plain text: %v", err)
		msg.ParseMode = ""
		_, err = s.bot.Send(msg)
		if err != nil {
			log.Printf("[Telegram] Failed to send message: %v", err)
		} else {
			log.Printf("[Telegram] Message sent successfully (plain text)")
		}
	} else {
		log.Printf("[Telegram] Message sent successfully")
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
