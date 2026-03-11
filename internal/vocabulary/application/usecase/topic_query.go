package usecase

import (
	"context"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
)

type TopicQuery struct {
	topicRepo port.TopicRepositoryPort
}

func NewTopicQuery(topicRepo port.TopicRepositoryPort) port.TopicQueryPort {
	return &TopicQuery{topicRepo: topicRepo}
}

func (useCase *TopicQuery) ListTopics(ctx context.Context) ([]*vdto.TopicResponse, error) {
	topics, err := useCase.topicRepo.FindAll(ctx)
	if err != nil {
		return nil, classifyRepoError(ctx, err, "error listing topics")
	}

	result := make([]*vdto.TopicResponse, 0, len(topics))
	for _, t := range topics {
		resp := toTopicResponse(t)
		result = append(result, &resp)
	}
	return result, nil
}
