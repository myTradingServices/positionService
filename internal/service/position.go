package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type service struct {
	repo repository.Interface
}

type Interface interface {
	Add(ctx context.Context, position model.Position) error
	Deleete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
}

func NewPositionService(repo repository.Interface) Interface {
	return &service{
		repo: repo,
	}
}

func (s *service) Add(ctx context.Context, position model.Position) error {
	return s.repo.Add(ctx, position)
}

func (s *service) Deleete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Deleete(ctx, id)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) ([]model.Position, error) {
	return s.repo.Get(ctx, id)
}
