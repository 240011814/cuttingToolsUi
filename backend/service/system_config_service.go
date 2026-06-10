package service

import (
	"backend/model"
)

// Mem0Config mem0 服务配置
type Mem0Config struct {
	APIKey  string
	BaseURL string
}

type SystemConfigService struct{}

func NewSystemConfigService() *SystemConfigService {
	return &SystemConfigService{}
}

func (s *SystemConfigService) GetAll() ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := DB.Find(&configs).Error
	return configs, err
}

func (s *SystemConfigService) GetValue(key string) (string, error) {
	var config model.SystemConfig
	err := DB.Where("`key` = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

func (s *SystemConfigService) SetValue(key, value, remark string) error {
	var cfg model.SystemConfig
	err := DB.Where("`key` = ?", key).First(&cfg).Error
	if err != nil {
		cfg = model.SystemConfig{Key: key, Value: value, Remark: remark}
		return DB.Create(&cfg).Error
	}
	cfg.Value = value
	if remark != "" {
		cfg.Remark = remark
	}
	return DB.Save(&cfg).Error
}

// GetMem0Config 从数据库读取 mem0 配置
func (s *SystemConfigService) GetMem0Config() Mem0Config {
	apiKey, _ := s.GetValue("mem0_api_key")
	baseURL, _ := s.GetValue("mem0_base_url")
	if baseURL == "" {
		baseURL = "https://api.mem0.ai/v1"
	}
	return Mem0Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}
