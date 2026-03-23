package handler

import (
	"fmt"
	"io"
	"learning-go/internal/shared/dto"
	"learning-go/internal/shared/logger"
	"learning-go/internal/shared/response"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type VocabularyHandler struct {
	vocabCmd  port.VocabularyCommandPort
	vocabQry  port.VocabularyQueryPort
	folderCmd port.FolderCommandPort
	folderQry port.FolderQueryPort
	topicQry  port.TopicQueryPort
	ocrCmd    port.OCRCommandPort
	importCmd port.ImportCommandPort
}

func NewVocabularyHandler(
	vocabCmd port.VocabularyCommandPort,
	vocabQry port.VocabularyQueryPort,
	folderCmd port.FolderCommandPort,
	folderQry port.FolderQueryPort,
	topicQry port.TopicQueryPort,
	ocrCmd port.OCRCommandPort,
	importCmd port.ImportCommandPort,
) *VocabularyHandler {
	return &VocabularyHandler{
		vocabCmd:  vocabCmd,
		vocabQry:  vocabQry,
		folderCmd: folderCmd,
		folderQry: folderQry,
		topicQry:  topicQry,
		ocrCmd:    ocrCmd,
		importCmd: importCmd,
	}
}

// --- Vocabulary endpoints ---

func (handler *VocabularyHandler) CreateVocabulary(c *gin.Context) {
	var req vdto.CreateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.vocabCmd.CreateVocabulary(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, res)
}

func (handler *VocabularyHandler) GetVocabulary(c *gin.Context) {
	res, err := handler.vocabQry.GetVocabulary(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

func (handler *VocabularyHandler) GetVocabularyDetail(c *gin.Context) {
	res, err := handler.vocabQry.GetVocabularyDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

func (handler *VocabularyHandler) ListByHSKLevel(c *gin.Context) {
	level, err := strconv.Atoi(c.Param("level"))
	if err != nil || level < 1 || level > 9 {
		response.BadRequest(c, "common.bad_request")
		return
	}

	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.vocabQry.ListByHSKLevel(c.Request.Context(), level, pagination)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMetadata(c, http.StatusOK, res.Items, res.Metadata)
}

func (handler *VocabularyHandler) ListByTopic(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "common.bad_request")
		return
	}

	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.vocabQry.ListByTopic(c.Request.Context(), slug, pagination)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMetadata(c, http.StatusOK, res.Items, res.Metadata)
}

func (handler *VocabularyHandler) SearchVocabulary(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "common.bad_request")
		return
	}

	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.vocabQry.SearchVocabulary(c.Request.Context(), query, pagination)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMetadata(c, http.StatusOK, res.Items, res.Metadata)
}

func (handler *VocabularyHandler) UpdateVocabulary(c *gin.Context) {
	var req vdto.UpdateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.vocabCmd.UpdateVocabulary(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

func (handler *VocabularyHandler) DeleteVocabulary(c *gin.Context) {
	if err := handler.vocabCmd.DeleteVocabulary(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, nil, "common.deleted_successfully")
}

// --- Topic endpoints ---

func (handler *VocabularyHandler) ListTopics(c *gin.Context) {
	res, err := handler.topicQry.ListTopics(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

// --- OCR endpoints ---

func (handler *VocabularyHandler) ProcessOCRScan(c *gin.Context) {
	const maxImageSize = 5 << 20 // 5MB

	var httpReq vdto.OCRScanHTTPRequest
	if err := c.ShouldBindJSON(&httpReq); err != nil {
		logger.WithContext(c.Request.Context()).Warn("[OCR] invalid request body", zap.Error(err))
		response.ValidationError(c, err)
		return
	}

	imageBytes, err := downloadImage(httpReq.ImageURL, maxImageSize)
	if err != nil {
		logger.WithContext(c.Request.Context()).Warn("[OCR] download image failed",
			zap.String("image_url", httpReq.ImageURL),
			zap.Error(err),
		)
		response.BadRequest(c, "ocr.image_download_failed")
		return
	}

	ocrType := httpReq.Type
	if ocrType == "" {
		ocrType = "auto"
	}
	language := httpReq.Language
	if language == "" {
		language = "zh"
	}

	req := vdto.OCRScanRequest{
		Image:    imageBytes,
		Type:     ocrType,
		Language: language,
	}

	res, err := handler.ocrCmd.ProcessOCRScan(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	logger.WithContext(c.Request.Context()).Info("[OCR] scan completed",
		zap.Int("new_items", len(res.NewItems)),
		zap.Int("existing_items", len(res.ExistingItems)),
		zap.Int("low_confidence_items", len(res.LowConfidenceItems)),
		zap.String("engine", res.Metadata.EngineUsed),
		zap.Int64("processing_ms", res.Metadata.ProcessingTimeMs),
	)
	response.Success(c, http.StatusOK, res)
}

func downloadImage(imageURL string, maxSize int64) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download image: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("image exceeds %d bytes", maxSize)
	}

	return data, nil
}

// --- Import endpoints ---

func (handler *VocabularyHandler) ImportVocabularies(c *gin.Context) {
	var req vdto.BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.importCmd.ImportVocabularies(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

// --- Folder endpoints ---

func (handler *VocabularyHandler) CreateFolder(c *gin.Context) {
	var req vdto.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	res, err := handler.folderCmd.CreateFolder(c.Request.Context(), userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, res)
}

func (handler *VocabularyHandler) ListFolders(c *gin.Context) {
	userID := c.GetString("user_id")
	res, err := handler.folderQry.ListFolders(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

func (handler *VocabularyHandler) UpdateFolder(c *gin.Context) {
	var req vdto.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	res, err := handler.folderCmd.UpdateFolder(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, res)
}

func (handler *VocabularyHandler) DeleteFolder(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := handler.folderCmd.DeleteFolder(c.Request.Context(), c.Param("id"), userID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, nil, "common.deleted_successfully")
}

func (handler *VocabularyHandler) AddVocabularyToFolder(c *gin.Context) {
	var req vdto.FolderVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	if err := handler.folderCmd.AddVocabulary(c.Request.Context(), c.Param("id"), req.VocabularyID, userID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, nil)
}

func (handler *VocabularyHandler) RemoveVocabularyFromFolder(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := handler.folderCmd.RemoveVocabulary(c.Request.Context(), c.Param("id"), c.Param("vocab_id"), userID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, nil, "common.deleted_successfully")
}

func (handler *VocabularyHandler) ListFolderVocabularies(c *gin.Context) {
	userID := c.GetString("user_id")
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := handler.folderQry.ListVocabularies(c.Request.Context(), c.Param("id"), userID, pagination)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMetadata(c, http.StatusOK, res.Items, res.Metadata)
}
