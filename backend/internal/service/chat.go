package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/pdf-rag-system/backend/internal/client"
	"github.com/pdf-rag-system/backend/internal/domain"
	"github.com/pdf-rag-system/backend/internal/repository"
	"github.com/pdf-rag-system/backend/pkg/config"
)

type ChatService struct {
	chunkRepo *repository.ChunkRepository
	llmClient *client.LLMClient
	config    *config.Config
}

func NewChatService(chunkRepo *repository.ChunkRepository, cfg *config.Config) *ChatService {
	llmClient := client.NewLLMClient(cfg.LLM.APIBaseURL, cfg.LLM.APIKey, cfg.LLM.Model)

	return &ChatService{
		chunkRepo: chunkRepo,
		llmClient: llmClient,
		config:    cfg,
	}
}

type QueryRequest struct {
	Query       string   `json:"query"`
	DocumentIDs []string `json:"document_ids"`
}

type QueryResponse struct {
	Answer    string                   `json:"answer"`
	Citations []*domain.SearchResult `json:"citations"`
}

func (s *ChatService) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	fmt.Printf("\n=== QUERY START ===\n")
	fmt.Printf("Query: %s\n", req.Query)
	fmt.Printf("Document IDs: %v\n", req.DocumentIDs)

	// Generate query embedding
	embeddingClient := client.NewLLMClient(s.config.Embedding.APIBaseURL, s.config.Embedding.APIKey, s.config.Embedding.Model)
	queryEmbedding, err := embeddingClient.GetEmbedding(req.Query, s.config.Embedding.Model)
	if err != nil {
		fmt.Printf("ERROR: Failed to generate query embedding: %v\n", err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	fmt.Printf("Query embedding generated (dim: %d, first 5 values: [%.4f, %.4f, %.4f, %.4f, %.4f])\n",
		len(queryEmbedding), queryEmbedding[0], queryEmbedding[1], queryEmbedding[2], queryEmbedding[3], queryEmbedding[4])

	// Vector search - get more results for better coverage
	fmt.Printf("Performing vector search (top 10 results)...\n")
	searchResults, err := s.chunkRepo.VectorSearch(ctx, queryEmbedding, req.DocumentIDs, 10)
	if err != nil {
		fmt.Printf("ERROR: Vector search failed: %v\n", err)
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	fmt.Printf("Found %d search results\n", len(searchResults))

	if len(searchResults) == 0 {
		fmt.Println("WARNING: No search results found")
		return &QueryResponse{
			Answer:    "No relevant information found in the documents.",
			Citations: []*domain.SearchResult{},
		}, nil
	}

	// Filter by similarity threshold (only keep results with score > 0.3)
	const similarityThreshold = 0.3
	var filteredResults []*domain.SearchResult
	for _, result := range searchResults {
		if result.Score > similarityThreshold {
			filteredResults = append(filteredResults, result)
		}
	}

	// Log search results with scores
	fmt.Println("\nTop search results:")
	for i, result := range searchResults {
		status := "✓"
		if result.Score <= similarityThreshold {
			status = "✗ (filtered out)"
		}
		fmt.Printf("  %s %d. Page %d, Chunk %d, Score: %.4f, Content: %.100s...\n",
			status, i+1, result.PageNumber, result.ChunkIndex, result.Score, result.Content)
	}

	// Use filtered results or return no relevant info
	if len(filteredResults) == 0 {
		fmt.Printf("WARNING: All results below similarity threshold (%.2f)\n", similarityThreshold)
		return &QueryResponse{
			Answer:    "No sufficiently relevant information found in the documents. The query may not be related to the document content.",
			Citations: []*domain.SearchResult{},
		}, nil
	}
	searchResults = filteredResults
	fmt.Printf("\nUsing %d results above threshold (%.2f)\n", len(filteredResults), similarityThreshold)

	// Build context from top results
	var contextParts []string
	for i, result := range searchResults {
		contextParts = append(contextParts, fmt.Sprintf(
			"[Source %d - %s, Page %d]:\n%s",
			i+1, result.Filename, result.PageNumber, result.Content,
		))
	}
	context := strings.Join(contextParts, "\n\n")
	fmt.Printf("\nContext built (total length: %d chars)\n", len(context))

	// Create prompt with strong hallucination prevention
	systemPrompt := `You are a precise document assistant. Your task is to answer questions STRICTLY based on the provided context.

CRITICAL RULES:
1. ONLY use information explicitly stated in the context
2. If the answer is not in the context, say "The provided context does not contain information about this question"
3. DO NOT make assumptions or add information from your general knowledge
4. ALWAYS cite the source number [Source X] when using information
5. If you're uncertain, acknowledge it clearly
6. Quote relevant parts of the context when possible`

	userPrompt := fmt.Sprintf(`Context from document:
%s

Question: %s

Instructions:
- Answer ONLY based on the context above
- Cite sources using [Source X] format
- If the context doesn't answer the question, say so explicitly
- Do not use external knowledge`, context, req.Query)

	// Call LLM
	messages := []client.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	fmt.Println("Calling LLM for answer generation...")
	answer, err := s.llmClient.Chat(messages)
	if err != nil {
		fmt.Printf("ERROR: LLM call failed: %v\n", err)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	fmt.Printf("Answer generated (length: %d chars)\n", len(answer))
	fmt.Printf("=== QUERY COMPLETE ===\n\n")

	return &QueryResponse{
		Answer:    answer,
		Citations: searchResults,
	}, nil
}
