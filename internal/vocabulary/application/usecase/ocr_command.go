package usecase

import (
	"context"
	sharederror "learning-go/internal/shared/error"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"
)

type OCRCommand struct {
	vocabRepo port.VocabularyRepositoryPort
}

func NewOCRCommand(vocabRepo port.VocabularyRepositoryPort) port.OCRCommandPort {
	return &OCRCommand{vocabRepo: vocabRepo}
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
		return nil, sharederror.NewInternal(ctx, "ocr.find_existing_failed", err)
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
