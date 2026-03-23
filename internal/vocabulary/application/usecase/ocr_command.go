package usecase

import (
	"context"
	apperr "learning-go/internal/shared/error"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"
	"time"
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
	start := time.Now()

	engine, _ := useCase.resolveEngine(req.Type, req.Language)
	if engine == nil {
		return nil, apperr.ServiceUnavailable("ocr.no_engine_available", nil)
	}

	ocrResult, err := engine.Recognize(ctx, port.OCRRequest{
		Image:    req.Image,
		Language: req.Language,
	})
	if err != nil {
		if _, ok := apperr.IsAppError(err); ok {
			return nil, err
		}
		return nil, apperr.ServiceUnavailable("ocr.recognize_failed", err)
	}

	totalDetected := len(ocrResult.Characters)

	// Classify by confidence (matching plan_ocr_engine.md thresholds)
	var confirmed []port.OCRCharacter
	var lowConfidenceItems []vdto.OCRScanCharacterItem

	for _, ch := range ocrResult.Characters {
		if ch.Confidence < ocrConfidenceLowThreshold {
			lowConfidenceItems = append(lowConfidenceItems, vdto.OCRScanCharacterItem{
				Hanzi:      ch.Text,
				Pinyin:     ch.Pinyin,
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
			if _, ok := apperr.IsAppError(err); ok {
				return nil, err
			}
			return nil, apperr.InternalServerError("ocr.find_existing_failed", err)
		}
		existingMap = make(map[string]*domain.Vocabulary, len(existing))
		for _, v := range existing {
			existingMap[v.Hanzi] = v
		}
	}

	var newItems []vdto.OCRScanCharacterItem
	var existingItems []vdto.OCRScanExistingItem

	for _, ch := range confirmed {
		if v, found := existingMap[ch.Text]; found {
			existingItems = append(existingItems, vdto.OCRScanExistingItem{
				VocabularyListResponse: toVocabularyListResponse(v),
				Confidence:             ch.Confidence,
				Candidates:             ch.Candidates,
			})
		} else {
			newItems = append(newItems, vdto.OCRScanCharacterItem{
				Hanzi:      ch.Text,
				Pinyin:     ch.Pinyin,
				Confidence: ch.Confidence,
				Candidates: ch.Candidates,
			})
		}
	}

	if newItems == nil {
		newItems = []vdto.OCRScanCharacterItem{}
	}
	if existingItems == nil {
		existingItems = []vdto.OCRScanExistingItem{}
	}
	if lowConfidenceItems == nil {
		lowConfidenceItems = []vdto.OCRScanCharacterItem{}
	}

	return &vdto.OCRScanResponse{
		NewItems:           newItems,
		ExistingItems:      existingItems,
		LowConfidenceItems: lowConfidenceItems,
		Metadata: vdto.OCRScanMetadata{
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
