package repository

import (
	"context"

	"github.com/pdf-rag-system/backend/internal/domain"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	var doc domain.Document
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error
	return &doc, err
}

func (r *DocumentRepository) List(ctx context.Context) ([]*domain.Document, error) {
	var docs []*domain.Document
	err := r.db.WithContext(ctx).Order("upload_time DESC").Find(&docs).Error
	return docs, err
}

func (r *DocumentRepository) Update(ctx context.Context, doc *domain.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Document{}, "id = ?", id).Error
}
