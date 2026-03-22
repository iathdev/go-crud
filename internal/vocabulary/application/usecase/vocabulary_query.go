package usecase

import (
	"context"
	"learning-go/internal/shared/dto"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"
	"math"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VocabularyQuery struct {
	vocabRepo   port.VocabularyRepositoryPort
	topicRepo   port.TopicRepositoryPort
	grammarRepo port.GrammarPointRepositoryPort
}

func NewVocabularyQuery(
	vocabRepo port.VocabularyRepositoryPort,
	topicRepo port.TopicRepositoryPort,
	grammarRepo port.GrammarPointRepositoryPort,
) port.VocabularyQueryPort {
	return &VocabularyQuery{
		vocabRepo:   vocabRepo,
		topicRepo:   topicRepo,
		grammarRepo: grammarRepo,
	}
}

func (useCase *VocabularyQuery) GetVocabulary(ctx context.Context, id string) (*vdto.VocabularyResponse, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, sharederror.NewInvalidInput("vocabulary.invalid_id")
	}

	vocab, err := useCase.vocabRepo.FindByID(ctx, uuidID)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	return toVocabularyResponse(vocab), nil
}

func (useCase *VocabularyQuery) GetVocabularyDetail(ctx context.Context, id string) (*vdto.VocabularyDetailResponse, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, sharederror.NewInvalidInput("vocabulary.invalid_id")
	}

	vocab, err := useCase.vocabRepo.FindByID(ctx, uuidID)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	// Fetch related topics
	var topicResponses []vdto.TopicResponse
	topics, err := useCase.findVocabTopics(ctx, uuidID)
	if err != nil {
		logger.WithContext(ctx).Warn("[VOCABULARY] error fetching topics for vocabulary", zap.Error(err))
		topicResponses = []vdto.TopicResponse{}
	} else {
		topicResponses = make([]vdto.TopicResponse, 0, len(topics))
		for _, t := range topics {
			topicResponses = append(topicResponses, toTopicResponse(t))
		}
	}

	// Fetch related grammar points
	var gpResponses []vdto.GrammarPointResponse
	grammarPoints, err := useCase.grammarRepo.FindByVocabularyID(ctx, uuidID)
	if err != nil {
		logger.WithContext(ctx).Warn("[VOCABULARY] error fetching grammar points for vocabulary", zap.Error(err))
		gpResponses = []vdto.GrammarPointResponse{}
	} else {
		gpResponses = make([]vdto.GrammarPointResponse, 0, len(grammarPoints))
		for _, gp := range grammarPoints {
			gpResponses = append(gpResponses, toGrammarPointResponse(gp))
		}
	}

	return &vdto.VocabularyDetailResponse{
		VocabularyResponse: *toVocabularyResponse(vocab),
		Topics:             topicResponses,
		GrammarPoints:      gpResponses,
	}, nil
}

func (useCase *VocabularyQuery) findVocabTopics(ctx context.Context, vocabID uuid.UUID) ([]*domain.Topic, error) {
	return useCase.topicRepo.FindByVocabularyID(ctx, vocabID)
}

func (useCase *VocabularyQuery) ListByHSKLevel(ctx context.Context, level int, pagination dto.PaginationRequest) (*dto.PaginatedResponse, error) {
	normalizePagination(&pagination)
	offset := (pagination.Page - 1) * pagination.PageSize

	total, err := useCase.vocabRepo.CountByHSKLevel(ctx, level)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	vocabs, err := useCase.vocabRepo.FindByHSKLevel(ctx, level, offset, pagination.PageSize)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	return toPaginatedResponse(vocabs, total, pagination), nil
}

func (useCase *VocabularyQuery) ListByTopic(ctx context.Context, slug string, pagination dto.PaginationRequest) (*dto.PaginatedResponse, error) {
	topic, err := useCase.topicRepo.FindBySlug(ctx, slug)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "topic.query_failed", err)
	}

	normalizePagination(&pagination)
	offset := (pagination.Page - 1) * pagination.PageSize

	total, err := useCase.vocabRepo.CountByTopicID(ctx, topic.ID)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	vocabs, err := useCase.vocabRepo.FindByTopicID(ctx, topic.ID, offset, pagination.PageSize)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	return toPaginatedResponse(vocabs, total, pagination), nil
}

func (useCase *VocabularyQuery) SearchVocabulary(ctx context.Context, query string, pagination dto.PaginationRequest) (*dto.PaginatedResponse, error) {
	normalizePagination(&pagination)
	offset := (pagination.Page - 1) * pagination.PageSize

	total, err := useCase.vocabRepo.CountSearch(ctx, query)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	vocabs, err := useCase.vocabRepo.Search(ctx, query, offset, pagination.PageSize)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		return nil, sharederror.NewInternal(ctx, "vocabulary.query_failed", err)
	}

	return toPaginatedResponse(vocabs, total, pagination), nil
}

func toVocabularyResponse(v *domain.Vocabulary) *vdto.VocabularyResponse {
	var examples []vdto.ExampleDTO
	if len(v.Examples) > 0 {
		examples = make([]vdto.ExampleDTO, 0, len(v.Examples))
		for _, e := range v.Examples {
			examples = append(examples, vdto.ExampleDTO{
				SentenceCN: e.SentenceCN,
				SentenceVI: e.SentenceVI,
				AudioURL:   e.AudioURL,
			})
		}
	}

	return &vdto.VocabularyResponse{
		ID:              v.ID.String(),
		Hanzi:           v.Hanzi,
		Pinyin:          v.Pinyin,
		MeaningVI:       v.MeaningVI,
		MeaningEN:       v.MeaningEN,
		HSKLevel:        v.HSKLevel,
		AudioURL:        v.AudioURL,
		Examples:        examples,
		Radicals:        v.Radicals,
		StrokeCount:     v.StrokeCount,
		StrokeDataURL:   v.StrokeDataURL,
		RecognitionOnly: v.RecognitionOnly,
		FrequencyRank:   v.FrequencyRank,
		CreatedAt:       v.CreatedAt,
	}
}

func toTopicResponse(t *domain.Topic) vdto.TopicResponse {
	return vdto.TopicResponse{
		ID:     t.ID.String(),
		NameCN: t.NameCN,
		NameVI: t.NameVI,
		NameEN: t.NameEN,
		Slug:   t.Slug,
	}
}

func toGrammarPointResponse(gp *domain.GrammarPoint) vdto.GrammarPointResponse {
	return vdto.GrammarPointResponse{
		ID:            gp.ID.String(),
		Code:          gp.Code,
		Pattern:       gp.Pattern,
		ExampleCN:     gp.ExampleCN,
		ExampleVI:     gp.ExampleVI,
		Rule:          gp.Rule,
		CommonMistake: gp.CommonMistake,
		HSKLevel:      gp.HSKLevel,
	}
}

func toPaginatedResponse(vocabs []*domain.Vocabulary, total int64, pagination dto.PaginationRequest) *dto.PaginatedResponse {
	items := make([]*vdto.VocabularyResponse, 0, len(vocabs))
	for _, v := range vocabs {
		items = append(items, toVocabularyResponse(v))
	}
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	return &dto.PaginatedResponse{
		Items: items,
		Metadata: dto.PaginationMeta{
			Total:      total,
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalPages: totalPages,
		},
	}
}

func normalizePagination(p *dto.PaginationRequest) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}
