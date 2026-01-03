package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pdf-rag-system/backend/internal/client"
	"github.com/pdf-rag-system/backend/internal/domain"
	"github.com/pdf-rag-system/backend/internal/repository"
	"github.com/pdf-rag-system/backend/pkg/config"
	"github.com/pgvector/pgvector-go"
)

type DocumentService struct {
	docRepo         *repository.DocumentRepository
	chunkRepo       *repository.ChunkRepository
	docreaderClient *client.DocReaderClient
	llmClient       *client.LLMClient
	config          *config.Config
}

func NewDocumentService(
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	docreaderClient *client.DocReaderClient,
	cfg *config.Config,
) *DocumentService {
	llmClient := client.NewLLMClient(cfg.Embedding.APIBaseURL, cfg.Embedding.APIKey, cfg.Embedding.Model)

	return &DocumentService{
		docRepo:         docRepo,
		chunkRepo:       chunkRepo,
		docreaderClient: docreaderClient,
		llmClient:       llmClient,
		config:          cfg,
	}
}

func (s *DocumentService) Upload(ctx context.Context, file io.Reader, filename string, fileSize int64) (*domain.Document, error) {
	fmt.Printf("=== UPLOAD SERVICE START ===\n")
	fmt.Printf("Filename: %s, Size: %d bytes (%.2f MB)\n", filename, fileSize, float64(fileSize)/(1024*1024))

	// Create document record
	doc := &domain.Document{
		ID:         uuid.New().String(),
		Filename:   filename,
		FileSize:   fileSize,
		UploadTime: time.Now(),
		Status:     "processing",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	fmt.Printf("Created document record with ID: %s\n", doc.ID)

	// Save file
	uploadDir := s.config.Upload.Dir
	fmt.Printf("Upload directory: %s\n", uploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create upload directory: %v\n", err)
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	filePath := filepath.Join(uploadDir, doc.ID+".pdf")
	fmt.Printf("Saving file to: %s\n", filePath)
	outFile, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("ERROR: Failed to create file: %v\n", err)
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	fmt.Println("Reading file content...")
	fileContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("ERROR: Failed to read file: %v\n", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	fmt.Printf("Read %d bytes from file\n", len(fileContent))

	fmt.Println("Writing file to disk...")
	if _, err := outFile.Write(fileContent); err != nil {
		fmt.Printf("ERROR: Failed to write file: %v\n", err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	fmt.Println("File written successfully")

	doc.FilePath = filePath

	// Save document to DB
	fmt.Println("Saving document to database...")
	if err := s.docRepo.Create(ctx, doc); err != nil {
		fmt.Printf("ERROR: Failed to save document to DB: %v\n", err)
		return nil, fmt.Errorf("failed to save document: %w", err)
	}
	fmt.Println("Document saved to database")

	// Parse PDF via gRPC (async in background)
	fmt.Println("Starting PDF processing in background...")
	go s.processPDF(context.Background(), doc.ID, fileContent, filename)

	fmt.Printf("=== UPLOAD SERVICE COMPLETE ===\n")
	return doc, nil
}

func (s *DocumentService) processPDF(ctx context.Context, docID string, fileContent []byte, filename string) {
	fmt.Printf("\n=== PROCESS PDF START (ID: %s) ===\n", docID)
	fmt.Printf("File: %s, Size: %d bytes (%.2f MB)\n", filename, len(fileContent), float64(len(fileContent))/(1024*1024))

	// Create a context with timeout for large PDFs (10 minutes)
	pdfCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Call docreader to parse PDF
	fmt.Println("Calling docreader gRPC service...")
	startTime := time.Now()
	resp, err := s.docreaderClient.ParsePDF(pdfCtx, fileContent, filename, 500, 50)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("ERROR: Docreader ParsePDF failed after %v: %v\n", duration, err)
		s.updateDocumentStatus(ctx, docID, "error")
		return
	}
	fmt.Printf("Docreader response received in %v. Total pages: %d, Chunks: %d\n", duration, resp.TotalPages, len(resp.Chunks))

	if resp.Error != "" {
		fmt.Printf("ERROR: Docreader returned error: %s\n", resp.Error)
		s.updateDocumentStatus(ctx, docID, "error")
		return
	}

	// Update total pages
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		fmt.Printf("ERROR: Failed to get document %s: %v\n", docID, err)
		s.updateDocumentStatus(ctx, docID, "error")
		return
	}
	doc.TotalPages = int(resp.TotalPages)
	if err := s.docRepo.Update(ctx, doc); err != nil {
		fmt.Printf("ERROR: Failed to update document %s: %v\n", docID, err)
	}

	// Process chunks
	chunks := make([]*domain.Chunk, 0, len(resp.Chunks))
	totalChunks := len(resp.Chunks)
	fmt.Printf("Generating embeddings for %d chunks...\n", totalChunks)

	for i, pbChunk := range resp.Chunks {
		// Progress logging every 100 chunks
		if i%100 == 0 {
			fmt.Printf("Progress: %d/%d chunks (%.1f%%)\n", i, totalChunks, float64(i)*100/float64(totalChunks))
		}

		// Generate embedding
		embedding, err := s.llmClient.GetEmbedding(pbChunk.Content, s.config.Embedding.Model)
		if err != nil {
			fmt.Printf("WARNING: Failed to generate embedding for chunk %d in document %s: %v\n", pbChunk.ChunkIndex, docID, err)
			continue
		}

		// Convert float64 to float32 for pgvector
		embedding32 := make([]float32, len(embedding))
		for i, v := range embedding {
			embedding32[i] = float32(v)
		}

		chunk := &domain.Chunk{
			ID:         uuid.New().String(),
			DocumentID: docID,
			Content:    pbChunk.Content,
			ChunkIndex: int(pbChunk.ChunkIndex),
			PageNumber: int(pbChunk.PageNumber),
			StartPos:   int(pbChunk.StartPos),
			EndPos:     int(pbChunk.EndPos),
			Embedding:  pgvector.NewVector(embedding32),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Add bbox if present
		if pbChunk.Bbox != nil {
			x1, y1, x2, y2 := float64(pbChunk.Bbox.X1), float64(pbChunk.Bbox.Y1), float64(pbChunk.Bbox.X2), float64(pbChunk.Bbox.Y2)
			chunk.BboxX1 = &x1
			chunk.BboxY1 = &y1
			chunk.BboxX2 = &x2
			chunk.BboxY2 = &y2
		}

		chunks = append(chunks, chunk)
	}

	// Check if we have any chunks
	if len(chunks) == 0 {
		fmt.Printf("ERROR: No chunks created for document %s (all embeddings failed)\n", docID)
		s.updateDocumentStatus(ctx, docID, "error")
		return
	}

	// Save chunks
	if err := s.chunkRepo.BatchCreate(ctx, chunks); err != nil {
		fmt.Printf("ERROR: Failed to save chunks for document %s: %v\n", docID, err)
		s.updateDocumentStatus(ctx, docID, "error")
		return
	}

	// Update status to completed
	fmt.Printf("SUCCESS: Saved %d chunks for document %s\n", len(chunks), docID)
	s.updateDocumentStatus(ctx, docID, "completed")
}

func (s *DocumentService) updateDocumentStatus(ctx context.Context, docID, status string) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return
	}
	doc.Status = status
	s.docRepo.Update(ctx, doc)
}

func (s *DocumentService) List(ctx context.Context) ([]*domain.Document, error) {
	return s.docRepo.List(ctx)
}

func (s *DocumentService) Get(ctx context.Context, id string) (*domain.Document, error) {
	return s.docRepo.GetByID(ctx, id)
}

func (s *DocumentService) Delete(ctx context.Context, id string) error {
	// Delete chunks first
	if err := s.chunkRepo.DeleteByDocumentID(ctx, id); err != nil {
		return err
	}

	// Delete document
	return s.docRepo.Delete(ctx, id)
}

func (s *DocumentService) RenderPageImage(ctx context.Context, docID, pageNum, bboxX1, bboxY1, bboxX2, bboxY2 string) ([]byte, error) {
	// Get document
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Build Python command
	args := []string{
		"-c",
		`import sys; sys.path.append('/app'); from pdf_renderer import render_pdf_page_to_image; ` +
			`bbox = None if len(sys.argv) <= 6 or not all([sys.argv[3], sys.argv[4], sys.argv[5], sys.argv[6]]) else {"x1": float(sys.argv[3]), "y1": float(sys.argv[4]), "x2": float(sys.argv[5]), "y2": float(sys.argv[6])}; ` +
			`result = render_pdf_page_to_image(sys.argv[1], int(sys.argv[2]), bbox); ` +
			`sys.stdout.buffer.write(result)`,
		doc.FilePath,
		pageNum,
		bboxX1,
		bboxY1,
		bboxX2,
		bboxY2,
	}

	cmd := exec.CommandContext(ctx, "python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("ERROR: Failed to render page: %s\nOutput: %s\n", err, string(output))
		return nil, fmt.Errorf("failed to render page: %s (output: %s)", err, string(output))
	}

	return output, nil
}
