package usecase

import (
	"context"
	"errors"
	sharederror "learning-go/internal/shared/error"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"

	"github.com/google/uuid"
)

type VocabularyCommand struct {
	vocabRepo port.VocabularyRepositoryPort
}

func NewVocabularyCommand(vocabRepo port.VocabularyRepositoryPort) port.VocabularyCommandPort {
	return &VocabularyCommand{vocabRepo: vocabRepo}
}

func (useCase *VocabularyCommand) CreateVocabulary(ctx context.Context, req vdto.CreateVocabularyRequest) (*vdto.VocabularyResponse, error) {
	params := domain.VocabularyParams{
		Hanzi:           req.Hanzi,
		Pinyin:          req.Pinyin,
		MeaningVI:       req.MeaningVI,
		MeaningEN:       req.MeaningEN,
		HSKLevel:        req.HSKLevel,
		AudioURL:        req.AudioURL,
		Examples:        toExampleEntities(req.Examples),
		Radicals:        req.Radicals,
		StrokeCount:     req.StrokeCount,
		StrokeDataURL:   req.StrokeDataURL,
		RecognitionOnly: req.RecognitionOnly,
		FrequencyRank:   req.FrequencyRank,
	}

	vocab, err := domain.NewVocabularyFromParams(params)
	if err != nil {
		return nil, mapVocabEntityError(err)
	}

	if err := useCase.vocabRepo.Save(ctx, vocab); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.InternalServerError("vocabulary.save_failed", err)
	}

	return toVocabularyResponse(vocab), nil
}

func (useCase *VocabularyCommand) UpdateVocabulary(ctx context.Context, id string, req vdto.UpdateVocabularyRequest) (*vdto.VocabularyResponse, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, sharederror.BadRequest("vocabulary.invalid_id")
	}

	vocab, err := useCase.vocabRepo.FindByID(ctx, uuidID)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.InternalServerError("vocabulary.query_failed", err)
	}

	params := domain.VocabularyParams{
		Hanzi:           req.Hanzi,
		Pinyin:          req.Pinyin,
		MeaningVI:       req.MeaningVI,
		MeaningEN:       req.MeaningEN,
		HSKLevel:        req.HSKLevel,
		AudioURL:        req.AudioURL,
		Examples:        toExampleEntities(req.Examples),
		Radicals:        req.Radicals,
		StrokeCount:     req.StrokeCount,
		StrokeDataURL:   req.StrokeDataURL,
		RecognitionOnly: req.RecognitionOnly,
		FrequencyRank:   req.FrequencyRank,
	}

	if err := vocab.UpdateFromParams(params); err != nil {
		return nil, mapVocabEntityError(err)
	}

	if err := useCase.vocabRepo.Update(ctx, vocab); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.InternalServerError("vocabulary.update_failed", err)
	}

	// Set topics if provided
	if req.TopicIDs != nil {
		topicUUIDs, parseErr := parseUUIDs(req.TopicIDs)
		if parseErr != nil {
			return nil, sharederror.BadRequest("vocabulary.invalid_topic_id")
		}
		if err := useCase.vocabRepo.SetTopics(ctx, uuidID, topicUUIDs); err != nil {
			if _, ok := sharederror.IsAppError(err); ok {
				return nil, err
			}
			return nil, sharederror.InternalServerError("vocabulary.set_topics_failed", err)
		}
	}

	// Set grammar points if provided
	if req.GrammarPointIDs != nil {
		gpUUIDs, parseErr := parseUUIDs(req.GrammarPointIDs)
		if parseErr != nil {
			return nil, sharederror.BadRequest("vocabulary.invalid_grammar_point_id")
		}
		if err := useCase.vocabRepo.SetGrammarPoints(ctx, uuidID, gpUUIDs); err != nil {
			if _, ok := sharederror.IsAppError(err); ok {
				return nil, err
			}
			return nil, sharederror.InternalServerError("vocabulary.set_grammar_points_failed", err)
		}
	}

	return toVocabularyResponse(vocab), nil
}

func (useCase *VocabularyCommand) DeleteVocabulary(ctx context.Context, id string) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return sharederror.BadRequest("vocabulary.invalid_id")
	}

	if _, err := useCase.vocabRepo.FindByID(ctx, uuidID); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return err
		}
		return sharederror.InternalServerError("vocabulary.query_failed", err)
	}

	if err := useCase.vocabRepo.Delete(ctx, uuidID); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return err
		}
		return sharederror.InternalServerError("vocabulary.delete_failed", err)
	}

	return nil
}

func toExampleEntities(dtos []vdto.ExampleDTO) []domain.Example {
	if dtos == nil {
		return nil
	}
	examples := make([]domain.Example, 0, len(dtos))
	for _, d := range dtos {
		examples = append(examples, domain.Example{
			SentenceCN: d.SentenceCN,
			SentenceVI: d.SentenceVI,
			AudioURL:   d.AudioURL,
		})
	}
	return examples
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}

func mapVocabEntityError(err error) error {
	switch {
	case errors.Is(err, domain.ErrHanziRequired):
		return sharederror.UnprocessableEntity("vocabulary.hanzi_required")
	case errors.Is(err, domain.ErrPinyinRequired):
		return sharederror.UnprocessableEntity("vocabulary.pinyin_required")
	case errors.Is(err, domain.ErrMeaningRequired):
		return sharederror.UnprocessableEntity("vocabulary.meaning_required")
	case errors.Is(err, domain.ErrInvalidHSKLevel):
		return sharederror.UnprocessableEntity("vocabulary.invalid_hsk_level")
	default:
		return sharederror.InternalServerError("common.internal_server_error", err)
	}
}
