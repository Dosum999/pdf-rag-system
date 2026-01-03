package domain

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// Chunk represents a text chunk from a document
type Chunk struct {
	ID         string          `json:"id" gorm:"type:varchar(36);primaryKey"`
	DocumentID string          `json:"document_id" gorm:"type:varchar(36);not null"`
	Content    string          `json:"content" gorm:"type:text;not null"`
	ChunkIndex int             `json:"chunk_index" gorm:"not null"`
	PageNumber int             `json:"page_number" gorm:"default:0"`
	StartPos   int             `json:"start_pos"`
	EndPos     int             `json:"end_pos"`
	BboxX1     *float64        `json:"bbox_x1,omitempty" gorm:"type:float"`
	BboxY1     *float64        `json:"bbox_y1,omitempty" gorm:"type:float"`
	BboxX2     *float64        `json:"bbox_x2,omitempty" gorm:"type:float"`
	BboxY2     *float64        `json:"bbox_y2,omitempty" gorm:"type:float"`
	Embedding  pgvector.Vector `json:"-" gorm:"type:vector(1536)"`
	CreatedAt  time.Time       `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relations
	Document Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

func (Chunk) TableName() string {
	return "chunks"
}

// BoundingBox represents bbox coordinates
type BoundingBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

// GetBoundingBox returns the bounding box if all coordinates are set
func (c *Chunk) GetBoundingBox() *BoundingBox {
	if c.BboxX1 == nil || c.BboxY1 == nil || c.BboxX2 == nil || c.BboxY2 == nil {
		return nil
	}
	return &BoundingBox{
		X1: *c.BboxX1,
		Y1: *c.BboxY1,
		X2: *c.BboxX2,
		Y2: *c.BboxY2,
	}
}

// SearchResult represents a search result with citation
type SearchResult struct {
	Chunk
	Score    float64 `json:"score"`
	Filename string  `json:"filename"`
}
