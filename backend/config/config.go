package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Auth       AuthConfig       `yaml:"auth"`
	AI         AIConfig         `yaml:"ai"`
	Baostock   BaostockConfig   `yaml:"baostock"`
	ClickHouse ClickHouseConfig `yaml:"clickhouse"`
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

type AIConfig struct {
	TimeoutMinutes int `yaml:"timeout_minutes"`
}

type BaostockConfig struct {
	URL string `yaml:"url"`
}

// ClickHouseConfig ClickHouse 分析库配置 (MySQL->CH 数据复制的下游副本)
type ClickHouseConfig struct {
	Enabled  bool   `yaml:"enabled"` // 未启用时跳过连接与同步
	Host     string `yaml:"host"`    // 原生协议端口 9000
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
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

	// Baostock ENV
	if envURL := os.Getenv("BAOSTOCK_API_URL"); envURL != "" {
		config.Baostock.URL = envURL
	}

	// ClickHouse ENV
	if v := os.Getenv("CH_ENABLED"); v == "1" || strings.EqualFold(v, "true") {
		config.ClickHouse.Enabled = true
	}
	if v := os.Getenv("CH_HOST"); v != "" {
		config.ClickHouse.Host = v
	}
	if v := os.Getenv("CH_PORT"); v != "" {
		config.ClickHouse.Port = v
	}
	if v := os.Getenv("CH_USER"); v != "" {
		config.ClickHouse.User = v
	}
	if v := os.Getenv("CH_PASSWORD"); v != "" {
		config.ClickHouse.Password = v
	}
	if v := os.Getenv("CH_DATABASE"); v != "" {
		config.ClickHouse.Database = v
	}

	// 3. 提供默认值兜底
	if config.Database.Host == "" {
		config.Database.Host = "127.0.0.1"
	}
	if config.Database.Port == "" {
		config.Database.Port = "3306"
	}
	if config.Auth.JWTSecret == "" {
		config.Auth.JWTSecret = "soybean-admin-secret"
	}
	if config.AI.TimeoutMinutes <= 0 {
		config.AI.TimeoutMinutes = 5
	}
	if config.Baostock.URL == "" {
		config.Baostock.URL = "http://127.0.0.1:3002"
	}
	if config.ClickHouse.Host == "" {
		config.ClickHouse.Host = "127.0.0.1"
	}
	if config.ClickHouse.Port == "" {
		config.ClickHouse.Port = "9000"
	}
	if config.ClickHouse.User == "" {
		config.ClickHouse.User = "default"
	}
	if config.ClickHouse.Database == "" {
		config.ClickHouse.Database = "default"
	}

	return &config, nil
}
