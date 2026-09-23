package model

import "time"

// AISkill 存储 Eino Skill Middleware 加载的 Skill 定义
// 对应 SKILL.md 的 YAML frontmatter 与 markdown 正文
type AISkill struct {
	ID int `json:"id" gorm:"primaryKey"`
	// Name 唯一标识, 即 frontmatter.name / SKILL.md 所在目录名
	Name string `json:"name" gorm:"uniqueIndex;size:100"`
	// Description 供模型判断何时使用该 Skill
	Description string `json:"description" gorm:"type:text"`
	// Context 执行模式: inline(默认) / fork / fork_with_context
	Context string `json:"context" gorm:"size:32"`
	// Agent fork 模式下使用的子 Agent 名称
	Agent string `json:"agent" gorm:"size:100"`
	// Model 指定 skill 使用的模型名称 (Eino ModelHub 解析)
	Model string `json:"model" gorm:"size:100"`
	// Content SKILL.md frontmatter 之后的 markdown 正文
	Content string `json:"content" gorm:"type:longtext"`
	// Enabled 是否参与动态加载
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AISkill) TableName() string {
	return "ai_skills"
}

type CreateAISkillRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Context     string `json:"context"`
	Agent       string `json:"agent"`
	Model       string `json:"model"`
	Content     string `json:"content"`
	Enabled     *bool  `json:"enabled"`
}

type UpdateAISkillRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Context     *string `json:"context"`
	Agent       *string `json:"agent"`
	Model       *string `json:"model"`
	Content     *string `json:"content"`
	Enabled     *bool   `json:"enabled"`
}
