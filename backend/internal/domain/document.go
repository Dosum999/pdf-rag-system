package domain

import "time"

// Document represents a PDF document
type Document struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Filename    string    `json:"filename" gorm:"type:varchar(255);not null"`
	FilePath    string    `json:"file_path" gorm:"type:varchar(512);not null"`
	FileSize    int64     `json:"file_size" gorm:"not null"`
	TotalPages  int       `json:"total_pages" gorm:"default:0"`
	UploadTime  time.Time `json:"upload_time" gorm:"not null;default:CURRENT_TIMESTAMP"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'processing'"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (Document) TableName() string {
	return "documents"
}
