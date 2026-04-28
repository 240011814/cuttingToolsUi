package service

import (
	"embed"
	"fmt"
	"log"

	"backend/config"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

//go:embed db/migrations/*.sql
var embedMigrations embed.FS

// InitDB 初始化数据库连接并自动运行迁移
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 运行数据库迁移 (Goose)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %v", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("mysql"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %v", err)
	}

	log.Println("Running database migrations...")
	if err := goose.Up(sqlDB, "db/migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	DB = db
	log.Println("Database connection established and migrations completed")
	return db, nil
}
