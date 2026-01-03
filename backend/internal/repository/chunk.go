package repository

import (
	"context"
	"fmt"

	"github.com/pdf-rag-system/backend/internal/domain"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type ChunkRepository struct {
	db *gorm.DB
}

func NewChunkRepository(db *gorm.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) Create(ctx context.Context, chunk *domain.Chunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

func (r *ChunkRepository) BatchCreate(ctx context.Context, chunks []*domain.Chunk) error {
	return r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error
}

func (r *ChunkRepository) GetByDocumentID(ctx context.Context, documentID string) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("chunk_index ASC").
		Find(&chunks).Error
	return chunks, err
}

func (r *ChunkRepository) VectorSearch(ctx context.Context, embedding []float64, documentIDs []string, limit int) ([]*domain.SearchResult, error) {
	var results []*domain.SearchResult

	// Convert float64 to float32 for pgvector
	embedding32 := make([]float32, len(embedding))
	for i, v := range embedding {
		embedding32[i] = float32(v)
	}
	vector := pgvector.NewVector(embedding32)

	query := `
		SELECT
			c.id,
			c.document_id,
			c.content,
			c.chunk_index,
			c.page_number,
			c.start_pos,
			c.end_pos,
			c.bbox_x1,
			c.bbox_y1,
			c.bbox_x2,
			c.bbox_y2,
			d.filename,
			1 - (c.embedding <=> ?) as score
		FROM chunks c
		JOIN documents d ON c.document_id = d.id
		WHERE c.document_id IN (?)
		ORDER BY c.embedding <=> ?
		LIMIT ?
	`

	err := r.db.WithContext(ctx).Raw(query, vector, documentIDs, vector, limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	return results, nil
}

func (r *ChunkRepository) DeleteByDocumentID(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&domain.Chunk{}).Error
}
