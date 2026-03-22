package usecase

import (
	"context"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"
	"time"

	"go.uber.org/zap"
)

const (
	ocrConfidenceConfirmed    = 0.80
	ocrConfidenceLowThreshold = 0.70
)

type OCRCommand struct {
	vocabRepo port.VocabularyRepositoryPort
	engines   port.OCREngineRegistry
}

func NewOCRCommand(vocabRepo port.VocabularyRepositoryPort, engines port.OCREngineRegistry) port.OCRCommandPort {
	return &OCRCommand{vocabRepo: vocabRepo, engines: engines}
}

// resolveEngine chọn engine dựa trên type + language (matching plan_ocr_engine.md section 5.1).
//
// Routing rules:
//
//	printed (any lang)        → google_vision
//	handwritten + zh          → baidu_ocr
//	handwritten + other       → google_vision
//	auto                      → google_vision (primary)
//
// Fallback: nếu engine được chọn không có trong registry → dùng engine đầu tiên có sẵn.
func (useCase *OCRCommand) resolveEngine(ocrType, language string) (port.OCRServicePort, port.OCREngineKey) {
	preferred := port.OCREngineGoogleVision

	switch ocrType {
	case "handwritten":
		if language == "zh" {
			preferred = port.OCREngineBaiduOCR
		}
	}

	if engine, ok := useCase.engines[preferred]; ok {
		return engine, preferred
	}

	// Fallback: dùng engine nào có sẵn
	for key, engine := range useCase.engines {
		return engine, key
	}

	return nil, ""
}

func (useCase *OCRCommand) ProcessOCRScan(ctx context.Context, req vdto.OCRScanRequest) (*vdto.OCRScanResponse, error) {
	hanziList := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		hanziList = append(hanziList, item.Hanzi)
	}

	existing, err := useCase.vocabRepo.FindByHanziList(ctx, hanziList)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		logger.WithContext(ctx).Error("[OCR] error finding existing vocabularies", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	existingMap := make(map[string]*domain.Vocabulary, len(existing))
	for _, v := range existing {
		existingMap[v.Hanzi] = v
	}

	var newItems []vdto.VocabularyListResponse
	var existingItems []vdto.VocabularyListResponse

	for _, item := range req.Items {
		if v, found := existingMap[item.Hanzi]; found {
			existingItems = append(existingItems, toVocabularyListResponse(v))
		} else {
			newItems = append(newItems, vdto.VocabularyListResponse{
				Hanzi: item.Hanzi,
			})
		}
	}

	if newItems == nil {
		newItems = []vdto.VocabularyListResponse{}
	}
	if existingItems == nil {
		existingItems = []vdto.VocabularyListResponse{}
	}

	return &vdto.OCRScanResponse{
		NewItems:      newItems,
		ExistingItems: existingItems,
	}, nil
}

func (useCase *OCRCommand) ProcessOCRImage(ctx context.Context, req vdto.OCRImageRequest) (*vdto.OCRImageResponse, error) {
	start := time.Now()

	engine, _ := useCase.resolveEngine(req.Type, req.Language)
	if engine == nil {
		logger.WithContext(ctx).Error("[OCR] no engine available")
		return nil, sharederror.ErrServiceUnavailable
	}

	ocrResult, err := engine.Recognize(ctx, port.OCRRequest{
		Image:    req.Image,
		Language: req.Language,
	})
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		logger.WithContext(ctx).Error("[OCR] error extracting characters from image", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	totalDetected := len(ocrResult.Characters)

	// Classify by confidence (matching plan_ocr_engine.md thresholds)
	var confirmed []port.OCRCharacter
	var lowConfidenceItems []vdto.OCRImageCharacterItem

	for _, ch := range ocrResult.Characters {
		if ch.Confidence < ocrConfidenceLowThreshold {
			lowConfidenceItems = append(lowConfidenceItems, vdto.OCRImageCharacterItem{
				Hanzi:      ch.Text,
				Confidence: ch.Confidence,
				Candidates: ch.Candidates,
			})
		} else {
			confirmed = append(confirmed, ch)
		}
	}

	// Check confirmed items against DB
	hanziList := make([]string, 0, len(confirmed))
	for _, ch := range confirmed {
		hanziList = append(hanziList, ch.Text)
	}

	var existingMap map[string]*domain.Vocabulary
	if len(hanziList) > 0 {
		existing, err := useCase.vocabRepo.FindByHanziList(ctx, hanziList)
		if err != nil {
			if _, ok := sharederror.IsAppError(err); ok {
				return nil, err
			}
			logger.WithContext(ctx).Error("[OCR] error finding existing vocabularies", zap.Error(err))
			return nil, sharederror.ErrInternal
		}
		existingMap = make(map[string]*domain.Vocabulary, len(existing))
		for _, v := range existing {
			existingMap[v.Hanzi] = v
		}
	}

	var newItems []vdto.OCRImageCharacterItem
	var existingItems []vdto.OCRImageExistingItem

	for _, ch := range confirmed {
		if v, found := existingMap[ch.Text]; found {
			existingItems = append(existingItems, vdto.OCRImageExistingItem{
				VocabularyListResponse: toVocabularyListResponse(v),
				Confidence:             ch.Confidence,
				Candidates:             ch.Candidates,
			})
		} else {
			newItems = append(newItems, vdto.OCRImageCharacterItem{
				Hanzi:      ch.Text,
				Confidence: ch.Confidence,
				Candidates: ch.Candidates,
			})
		}
	}

	if newItems == nil {
		newItems = []vdto.OCRImageCharacterItem{}
	}
	if existingItems == nil {
		existingItems = []vdto.OCRImageExistingItem{}
	}
	if lowConfidenceItems == nil {
		lowConfidenceItems = []vdto.OCRImageCharacterItem{}
	}

	return &vdto.OCRImageResponse{
		NewItems:           newItems,
		ExistingItems:      existingItems,
		LowConfidenceItems: lowConfidenceItems,
		Metadata: vdto.OCRImageMetadata{
			EngineUsed:       ocrResult.Engine,
			TotalDetected:    totalDetected,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		},
	}, nil
}

func toVocabularyListResponse(v *domain.Vocabulary) vdto.VocabularyListResponse {
	return vdto.VocabularyListResponse{
		ID:        v.ID.String(),
		Hanzi:     v.Hanzi,
		Pinyin:    v.Pinyin,
		MeaningVI: v.MeaningVI,
		MeaningEN: v.MeaningEN,
		HSKLevel:  v.HSKLevel,
	}
}
