package service

import (
	"backend/config"
	"backend/model"
)

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

// GetMem0Config 从数据库读取 mem0 配置，未配置则使用默认值
func (s *SystemConfigService) GetMem0Config(fallback config.Mem0Config) config.Mem0Config {
	apiKey, _ := s.GetValue("mem0_api_key")
	baseURL, _ := s.GetValue("mem0_base_url")

	if apiKey == "" {
		apiKey = fallback.APIKey
	}
	if baseURL == "" {
		baseURL = fallback.BaseURL
	}

	return config.Mem0Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}
