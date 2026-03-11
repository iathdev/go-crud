package usecase

import (
	"context"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vdto "learning-go/internal/vocabulary/application/dto"
	"learning-go/internal/vocabulary/application/port"
	"learning-go/internal/vocabulary/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FolderCommand struct {
	folderRepo port.FolderRepositoryPort
	vocabRepo  port.VocabularyRepositoryPort
}

func NewFolderCommand(folderRepo port.FolderRepositoryPort, vocabRepo port.VocabularyRepositoryPort) port.FolderCommandPort {
	return &FolderCommand{folderRepo: folderRepo, vocabRepo: vocabRepo}
}

func (useCase *FolderCommand) CreateFolder(ctx context.Context, userID string, req vdto.CreateFolderRequest) (*vdto.FolderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, sharederror.ErrInvalidInput
	}

	folder, err := domain.NewFolder(uid, req.Name, req.Description)
	if err != nil {
		return nil, sharederror.ErrInvalidInput
	}

	if err := useCase.folderRepo.Save(ctx, folder); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		logger.WithContext(ctx).Error("error saving folder", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	return toFolderResponse(folder), nil
}

func (useCase *FolderCommand) UpdateFolder(ctx context.Context, id string, userID string, req vdto.UpdateFolderRequest) (*vdto.FolderResponse, error) {
	folder, err := getOwnedFolder(ctx, useCase.folderRepo, id, userID)
	if err != nil {
		return nil, err
	}

	if err := folder.Update(req.Name, req.Description); err != nil {
		return nil, sharederror.ErrInvalidInput
	}

	if err := useCase.folderRepo.Update(ctx, folder); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return nil, err
		}
		logger.WithContext(ctx).Error("error updating folder", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	return toFolderResponse(folder), nil
}

func (useCase *FolderCommand) DeleteFolder(ctx context.Context, id string, userID string) error {
	folder, err := getOwnedFolder(ctx, useCase.folderRepo, id, userID)
	if err != nil {
		return err
	}

	if err := useCase.folderRepo.Delete(ctx, folder.ID); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return err
		}
		logger.WithContext(ctx).Error("error deleting folder", zap.Error(err))
		return sharederror.ErrInternal
	}
	return nil
}

func (useCase *FolderCommand) AddVocabulary(ctx context.Context, folderID string, vocabID string, userID string) error {
	folder, err := getOwnedFolder(ctx, useCase.folderRepo, folderID, userID)
	if err != nil {
		return err
	}

	vid, err := uuid.Parse(vocabID)
	if err != nil {
		return sharederror.ErrInvalidInput
	}

	if _, err := useCase.vocabRepo.FindByID(ctx, vid); err != nil {
		return classifyRepoError(ctx, err, "error finding vocabulary")
	}

	if err := useCase.folderRepo.AddVocabulary(ctx, folder.ID, vid); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return err
		}
		logger.WithContext(ctx).Error("error adding vocabulary to folder", zap.Error(err))
		return sharederror.ErrInternal
	}
	return nil
}

func (useCase *FolderCommand) RemoveVocabulary(ctx context.Context, folderID string, vocabID string, userID string) error {
	folder, err := getOwnedFolder(ctx, useCase.folderRepo, folderID, userID)
	if err != nil {
		return err
	}

	vid, err := uuid.Parse(vocabID)
	if err != nil {
		return sharederror.ErrInvalidInput
	}

	if err := useCase.folderRepo.RemoveVocabulary(ctx, folder.ID, vid); err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			return err
		}
		logger.WithContext(ctx).Error("error removing vocabulary from folder", zap.Error(err))
		return sharederror.ErrInternal
	}
	return nil
}
