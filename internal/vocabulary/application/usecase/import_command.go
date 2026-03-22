package usecase

import (
	"context"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"

	"go.uber.org/zap"
)

type ImportCommand struct {
	vocabRepo port.VocabularyRepositoryPort
}

func NewImportCommand(vocabRepo port.VocabularyRepositoryPort) port.ImportCommandPort {
	return &ImportCommand{vocabRepo: vocabRepo}
}

func (useCase *ImportCommand) ImportVocabularies(ctx context.Context, req vdto.BulkImportRequest) (*vdto.BulkImportResponse, error) {
	// Collect all hanzi to check for existing
	hanziList := make([]string, 0, len(req.Vocabularies))
	for _, v := range req.Vocabularies {
		hanziList = append(hanziList, v.Hanzi)
	}

	existing, err := useCase.vocabRepo.FindByHanziList(ctx, hanziList)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		logger.WithContext(ctx).Error("[VOCABULARY] error checking existing vocabularies", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	existingSet := make(map[string]bool, len(existing))
	for _, v := range existing {
		existingSet[v.Hanzi] = true
	}

	// Filter out duplicates and create domain entities
	var newVocabs []*domain.Vocabulary
	skipped := 0
	for _, v := range req.Vocabularies {
		if existingSet[v.Hanzi] {
			skipped++
			continue
		}

		vocab, err := domain.NewVocabularyFromParams(domain.VocabularyParams{
			Hanzi:           v.Hanzi,
			Pinyin:          v.Pinyin,
			MeaningVI:       v.MeaningVI,
			MeaningEN:       v.MeaningEN,
			HSKLevel:        v.HSKLevel,
			AudioURL:        v.AudioURL,
			Examples:        toExampleEntities(v.Examples),
			Radicals:        v.Radicals,
			StrokeCount:     v.StrokeCount,
			StrokeDataURL:   v.StrokeDataURL,
			RecognitionOnly: v.RecognitionOnly,
			FrequencyRank:   v.FrequencyRank,
		})
		if err != nil {
			skipped++
			continue
		}
		newVocabs = append(newVocabs, vocab)
	}

	imported := 0
	if len(newVocabs) > 0 {
		imported, err = useCase.vocabRepo.SaveBatch(ctx, newVocabs)
		if err != nil {
			if _, ok := sharederror.IsAppError(err); ok {
				return nil, err
			}
			logger.WithContext(ctx).Error("[VOCABULARY] error batch saving vocabularies", zap.Error(err))
			return nil, sharederror.ErrInternal
		}
	}

	return &vdto.BulkImportResponse{
		Imported: imported,
		Skipped:  skipped,
		Total:    len(req.Vocabularies),
	}, nil
}
