package database

import (
	"fmt"

	"github.com/pdf-rag-system/backend/internal/domain"
	"github.com/pdf-rag-system/backend/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&domain.Document{}, &domain.Chunk{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
