package tools

import (
	interfaces "backend/interface"
	"backend/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var userMemorySvc interfaces.UserMemoryService

// SetUserMemoryService 注册用户画像/经历服务实例供工具使用
func SetUserMemoryService(svc interfaces.UserMemoryService) { userMemorySvc = svc }

type userInfoTool struct{}

type userInfoRequest struct {
	Action   string   `json:"action" jsonschema:"description=操作类型: get_profile 获取画像, search_experience 搜索经历, add_experience 记录经历, update_profile 更新画像"`
	Query    string   `json:"query,omitempty" jsonschema:"description=搜索关键词（search_experience 时使用）"`
	Category string   `json:"category,omitempty" jsonschema:"description=经历分类: work/project/study/achievement/challenge/other"`
	Title    string   `json:"title,omitempty" jsonschema:"description=经历标题（add_experience 时必填）"`
	Content  string   `json:"content,omitempty" jsonschema:"description=经历内容（add_experience 时使用）"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=经历标签"`
	Facts    []string `json:"facts,omitempty" jsonschema:"description=要合并进画像的新事实（update_profile 时必填）"`
}

const userInfoDesc = "管理用户的画像与个人经历。get_profile 获取用户画像; search_experience 搜索用户过去的经历; add_experience 记录一段用户经历; update_profile 更新用户画像事实。当需要了解用户是什么样的人、有什么目标/偏好，或用户提到过往经历、让你记住某件事时使用。"

func init() {
	Register("user_info", "用户画像与经历", userInfoDesc, nil,
		func(config map[string]any) (tool.BaseTool, error) {
			return &userInfoTool{}, nil
		})
}

func (t *userInfoTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "user_info",
		Desc: userInfoDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Type:     schema.String,
				Desc:     "操作类型: get_profile 获取画像, search_experience 搜索经历, add_experience 记录经历, update_profile 更新画像",
				Required: true,
			},
			"query": {
				Type: schema.String,
				Desc: "搜索关键词（search_experience 时使用）",
			},
			"category": {
				Type: schema.String,
				Desc: "经历分类: work/project/study/achievement/challenge/other",
			},
			"title": {
				Type: schema.String,
				Desc: "经历标题（add_experience 时必填）",
			},
			"content": {
				Type: schema.String,
				Desc: "经历内容（add_experience 时使用）",
			},
			"tags": {
				Type: schema.Array,
				Desc: "经历标签",
			},
			"facts": {
				Type: schema.Array,
				Desc: "要合并进画像的新事实（update_profile 时必填）",
			},
		}),
	}, nil
}

func (t *userInfoTool) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	var req userInfoRequest
	if err := json.Unmarshal([]byte(arguments), &req); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}
	if userMemorySvc == nil {
		return "", fmt.Errorf("用户画像服务未初始化")
	}

	sessionValues := adk.GetSessionValues(ctx)
	userIDVal, ok := sessionValues["user_id"]
	if !ok {
		return "", fmt.Errorf("无法获取用户 ID")
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return "", fmt.Errorf("用户 ID 类型错误")
	}

	switch req.Action {
	case "get_profile":
		profile := userMemorySvc.BuildProfilePrompt(userID)
		if profile == "" {
			return "暂无用户画像", nil
		}
		return profile, nil

	case "search_experience":
		list, err := userMemorySvc.SearchExperiences(userID, req.Query, 10)
		if err != nil {
			return "", fmt.Errorf("搜索经历失败: %w", err)
		}
		if len(list) == 0 {
			return "未找到相关经历", nil
		}
		result, _ := json.Marshal(toExperienceViews(list))
		return string(result), nil

	case "add_experience":
		if req.Title == "" {
			return "", fmt.Errorf("经历标题不能为空")
		}
		tags, _ := json.Marshal(req.Tags)
		exp, err := userMemorySvc.AddExperience(userID, &model.UserExperience{
			Category: req.Category,
			Title:    req.Title,
			Content:  req.Content,
			Tags:     string(tags),
		})
		if err != nil {
			return "", fmt.Errorf("记录经历失败: %w", err)
		}
		return fmt.Sprintf("经历已记录, ID: %d", exp.ID), nil

	case "update_profile":
		if len(req.Facts) == 0 {
			return "", fmt.Errorf("facts 不能为空")
		}
		if err := userMemorySvc.MergeProfileFacts(userID, req.Facts); err != nil {
			return "", fmt.Errorf("更新画像失败: %w", err)
		}
		return "用户画像已更新", nil

	default:
		return "", fmt.Errorf("不支持的操作: %s，支持的操作: get_profile, search_experience, add_experience, update_profile", req.Action)
	}
}

type experienceView struct {
	ID         uint           `json:"id"`
	Category   string         `json:"category"`
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	Tags       []string       `json:"tags"`
	OccurredAt interface{}    `json:"occurred_at"`
}

func toExperienceViews(list []model.UserExperience) []experienceView {
	views := make([]experienceView, 0, len(list))
	for _, e := range list {
		var tags []string
		_ = json.Unmarshal([]byte(e.Tags), &tags)
		views = append(views, experienceView{
			ID:         e.ID,
			Category:   e.Category,
			Title:      e.Title,
			Content:    e.Content,
			Tags:       tags,
			OccurredAt: e.OccurredAt,
		})
	}
	return views
}