package vocabulary

import (
	"learning-go/internal/vocabulary/adapter/handler"
	"learning-go/internal/vocabulary/adapter/repository"
	"learning-go/internal/vocabulary/application/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	handler *handler.VocabularyHandler
}

func NewModule(db *gorm.DB) *Module {
	vocabRepo := repository.NewVocabularyRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	grammarRepo := repository.NewGrammarPointRepository(db)

	vocabCmd := usecase.NewVocabularyCommand(vocabRepo)
	vocabQry := usecase.NewVocabularyQuery(vocabRepo, topicRepo, grammarRepo)
	folderCmd := usecase.NewFolderCommand(folderRepo, vocabRepo)
	folderQry := usecase.NewFolderQuery(folderRepo)
	topicQry := usecase.NewTopicQuery(topicRepo)
	ocrCmd := usecase.NewOCRCommand(vocabRepo)
	importCmd := usecase.NewImportCommand(vocabRepo)

	vocabHandler := handler.NewVocabularyHandler(vocabCmd, vocabQry, folderCmd, folderQry, topicQry, ocrCmd, importCmd)

	return &Module{handler: vocabHandler}
}

func (module *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Topics (protected)
	protected.GET("/topics", module.handler.ListTopics)

	// Vocabulary CRUD
	protected.POST("/vocabularies", module.handler.CreateVocabulary)
	protected.GET("/vocabularies/:id", module.handler.GetVocabulary)
	protected.GET("/vocabularies/:id/detail", module.handler.GetVocabularyDetail)
	protected.GET("/vocabularies/hsk/:level", module.handler.ListByHSKLevel)
	protected.GET("/vocabularies/topic/:slug", module.handler.ListByTopic)
	protected.GET("/vocabularies/search", module.handler.SearchVocabulary)
	protected.PUT("/vocabularies/:id", module.handler.UpdateVocabulary)
	protected.DELETE("/vocabularies/:id", module.handler.DeleteVocabulary)

	// OCR
	protected.POST("/vocabularies/ocr-scan", module.handler.ProcessOCRScan)

	// Admin import
	protected.POST("/admin/vocabularies/import", module.handler.ImportVocabularies)

	// Folder CRUD
	protected.POST("/folders", module.handler.CreateFolder)
	protected.GET("/folders", module.handler.ListFolders)
	protected.PUT("/folders/:id", module.handler.UpdateFolder)
	protected.DELETE("/folders/:id", module.handler.DeleteFolder)

	// Folder-Vocabulary operations
	protected.POST("/folders/:id/vocabularies", module.handler.AddVocabularyToFolder)
	protected.DELETE("/folders/:id/vocabularies/:vocab_id", module.handler.RemoveVocabularyFromFolder)
	protected.GET("/folders/:id/vocabularies", module.handler.ListFolderVocabularies)
}
