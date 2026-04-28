package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DeepSeek DeepSeekConfig `yaml:"deepseek"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

type DeepSeekConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

// LoadConfig 从指定的文件路径加载配置，并融合环境变量（环境变量优先级更高）
func LoadConfig(path string) (*Config, error) {
	var config Config

	// 1. 尝试从文件读取
	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}

	// 2. 环境变量优先级最高，如果存在则覆盖文件配置
	if envAPIKey := os.Getenv("DEEPSEEK_API_KEY"); envAPIKey != "" {
		config.DeepSeek.APIKey = envAPIKey
	}
	if envBaseURL := os.Getenv("DEEPSEEK_BASE_URL"); envBaseURL != "" {
		config.DeepSeek.BaseURL = envBaseURL
	}
	if envModel := os.Getenv("DEEPSEEK_MODEL"); envModel != "" {
		config.DeepSeek.Model = envModel
	}

	// Database ENV
	if envHost := os.Getenv("DB_HOST"); envHost != "" {
		config.Database.Host = envHost
	}
	if envPort := os.Getenv("DB_PORT"); envPort != "" {
		config.Database.Port = envPort
	}
	if envUser := os.Getenv("DB_USER"); envUser != "" {
		config.Database.User = envUser
	}
	if envPass := os.Getenv("DB_PASSWORD"); envPass != "" {
		config.Database.Password = envPass
	}
	if envDB := os.Getenv("DB_NAME"); envDB != "" {
		config.Database.DBName = envDB
	}

	// Auth ENV
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		config.Auth.JWTSecret = envSecret
	}

	// 3. 提供默认值兜底
	if config.DeepSeek.BaseURL == "" {
		config.DeepSeek.BaseURL = "https://api.deepseek.com/v1"
	}
	if config.DeepSeek.Model == "" {
		config.DeepSeek.Model = "deepseek-chat"
	}
	if config.Database.Host == "" {
		config.Database.Host = "127.0.0.1"
	}
	if config.Database.Port == "" {
		config.Database.Port = "3306"
	}
	if config.Auth.JWTSecret == "" {
		config.Auth.JWTSecret = "soybean-admin-secret"
	}

	return &config, nil
}
