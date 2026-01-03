package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pdf-rag-system/backend/internal/service"
)

type DocumentHandler struct {
	service *service.DocumentService
}

func NewDocumentHandler(service *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	log.Println("=== UPLOAD REQUEST RECEIVED ===")
	log.Printf("Content-Length: %d bytes", c.Request.ContentLength)
	log.Printf("Content-Type: %s", c.Request.Header.Get("Content-Type"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("ERROR: Failed to get file from form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	log.Printf("File received: %s (size: %d bytes, %.2f MB)", header.Filename, header.Size, float64(header.Size)/(1024*1024))

	doc, err := h.service.Upload(c.Request.Context(), file, header.Filename, header.Size)
	if err != nil {
		log.Printf("ERROR: Upload service failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("SUCCESS: Document uploaded with ID: %s", doc.ID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doc,
	})
}

func (h *DocumentHandler) List(c *gin.Context) {
	docs, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"documents": docs,
	})
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id := c.Param("id")

	doc, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doc,
	})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document deleted",
	})
}

func (h *DocumentHandler) GetFile(c *gin.Context) {
	id := c.Param("id")

	doc, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Set proper headers for PDF serving
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+doc.Filename+"\"")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	c.File(doc.FilePath)
}

func (h *DocumentHandler) GetPageImage(c *gin.Context) {
	id := c.Param("id")
	pageNum := c.Param("page")

	// Get query parameters for bbox
	bboxX1 := c.Query("bbox_x1")
	bboxY1 := c.Query("bbox_y1")
	bboxX2 := c.Query("bbox_x2")
	bboxY2 := c.Query("bbox_y2")

	imageData, err := h.service.RenderPageImage(c.Request.Context(), id, pageNum, bboxX1, bboxY1, bboxX2, bboxY2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set proper headers for image
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", imageData)
}
