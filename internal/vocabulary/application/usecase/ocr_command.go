package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
	"go.uber.org/zap"

	apperr "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"
)

const (
	ocrConfidenceConfirmed    = 0.80
	ocrConfidenceLowThreshold = 0.70
	ocrCascadingThreshold     = 0.75 // auto mode: cascade to secondary engine if avg confidence < 75%
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
//	printed (any lang)             → google_vision (no fallback)
//	handwritten + zh               → baidu_ocr → fallback google_vision
//	handwritten + other            → google_vision (no fallback)
//	auto                           → google_vision (cascading handled in ProcessOCRScan)
//
// Returns (engine, key) or (nil, "") if no engine available.
func (useCase *OCRCommand) resolveEngine(ocrType, language string) (port.OCRServicePort, port.OCREngineKey) {
	switch ocrType {
	case "printed":
		return useCase.getEngine(port.OCREngineGoogleVision)

	case "handwritten":
		if language == "zh" {
			// Baidu primary → PaddleOCR fallback → Google Vision fallback
			return useCase.getFirstAvailable(
				port.OCREngineBaiduOCR,
				port.OCREnginePaddleOCR,
				port.OCREngineGoogleVision,
			)
		}
		return useCase.getEngine(port.OCREngineGoogleVision)

	default: // "auto"
		return useCase.getEngine(port.OCREngineGoogleVision)
	}
}

func (useCase *OCRCommand) getEngine(key port.OCREngineKey) (port.OCRServicePort, port.OCREngineKey) {
	if engine, ok := useCase.engines[key]; ok {
		return engine, key
	}
	return nil, ""
}

func (useCase *OCRCommand) getFirstAvailable(keys ...port.OCREngineKey) (port.OCRServicePort, port.OCREngineKey) {
	for _, key := range keys {
		if engine, ok := useCase.engines[key]; ok {
			return engine, key
		}
	}
	return nil, ""
}

func (useCase *OCRCommand) ProcessOCRScan(ctx context.Context, req vdto.OCRScanRequest) (*vdto.OCRScanResponse, error) {
	start := time.Now()

	engine, _ := useCase.resolveEngine(req.Type, req.Language)
	if engine == nil {
		return nil, apperr.ServiceUnavailable("ocr.no_engine_available", nil)
	}

	ocrReq := port.OCRRequest{Image: req.Image, Language: req.Language}

	ocrResult, err := engine.Recognize(ctx, ocrReq)
	if err != nil {
		if _, ok := apperr.IsAppError(err); ok {
			return nil, err
		}
		return nil, apperr.ServiceUnavailable("ocr.recognize_failed", err)
	}

	// Enrich pinyin for characters that don't have it (e.g., Google Vision)
	if req.Language == "zh" {
		for i := range ocrResult.Characters {
			if ocrResult.Characters[i].Pinyin == "" {
				ocrResult.Characters[i].Pinyin = convertToPinyin(ocrResult.Characters[i].Text)
			}
		}
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

	resp := &vdto.OCRScanResponse{
		NewItems:           newItems,
		ExistingItems:      existingItems,
		LowConfidenceItems: lowConfidenceItems,
		Metadata: vdto.OCRScanMetadata{
			EngineUsed:       ocrResult.Engine,
			TotalDetected:    totalDetected,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		},
	}

	logger.Info(ctx, "[OCR] scan completed",
		zap.Int("new_items", len(newItems)),
		zap.Int("existing_items", len(existingItems)),
		zap.Int("low_confidence_items", len(lowConfidenceItems)),
		zap.String("engine", ocrResult.Engine),
		zap.Int64("processing_ms", resp.Metadata.ProcessingTimeMs),
	)

	return resp, nil
}

func avgConfidence(chars []port.OCRCharacter) float64 {
	if len(chars) == 0 {
		return 0
	}
	var sum float64
	for _, ch := range chars {
		sum += ch.Confidence
	}
	return sum / float64(len(chars))
}

var pinyinArgs = func() pinyin.Args {
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone
	return a
}()

func convertToPinyin(hanzi string) string {
	result := pinyin.Pinyin(hanzi, pinyinArgs)
	parts := make([]string, 0, len(result))
	for _, r := range result {
		if len(r) > 0 {
			parts = append(parts, r[0])
		}
	}
	return strings.Join(parts, " ")
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
