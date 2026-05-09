package model

import "time"

type CustomTraining struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	SystemPrompt     string    `json:"system_prompt"`
	Icon             string    `json:"icon"`
	Color            string    `json:"color"`
	InitialMessage   string    `json:"initial_message"`
	InputPlaceholder string    `json:"input_placeholder"`
	SpeechLang       string    `json:"speech_lang"`
	SpeechRate       float64   `json:"speech_rate"`
	IsFavorite       bool      `json:"is_favorite"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateCustomTrainingRequest struct {
	Title            string  `json:"title" binding:"required"`
	Description      string  `json:"description"`
	SystemPrompt     string  `json:"system_prompt" binding:"required"`
	Icon             string  `json:"icon"`
	Color            string  `json:"color"`
	InitialMessage   string  `json:"initial_message"`
	InputPlaceholder string  `json:"input_placeholder"`
	SpeechLang       string  `json:"speech_lang"`
	SpeechRate       float64 `json:"speech_rate"`
}

type UpdateCustomTrainingRequest struct {
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	SystemPrompt     string  `json:"system_prompt"`
	Icon             string  `json:"icon"`
	Color            string  `json:"color"`
	InitialMessage   string  `json:"initial_message"`
	InputPlaceholder string  `json:"input_placeholder"`
	SpeechLang       string  `json:"speech_lang"`
	SpeechRate       float64 `json:"speech_rate"`
}
