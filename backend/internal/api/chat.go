package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pdf-rag-system/backend/internal/service"
)

type ChatHandler struct {
	service *service.ChatService
}

func NewChatHandler(service *service.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) Query(c *gin.Context) {
	var req service.QueryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
		return
	}

	if len(req.DocumentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document IDs required"})
		return
	}

	resp, err := h.service.Query(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"answer":  resp.Answer,
		"citations": resp.Citations,
	})
}
