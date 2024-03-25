package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type DbInterface struct {
	repo repository.DbInterface
}

type DBInterface interface {
	Add(ctx context.Context, position model.Position) error
	Deleete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
	Update(ctx context.Context, position model.Position) error
}

func NewPositionService(repo repository.DbInterface) DBInterface {
	return &DbInterface{
		repo: repo,
	}
}

func (r *DbInterface) Add(ctx context.Context, position model.Position) error {
	return r.repo.Add(ctx, position)
}

func (r *DbInterface) Deleete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Deleete(ctx, id)
}

func (r *DbInterface) Get(ctx context.Context, id uuid.UUID) ([]model.Position, error) {
	return r.repo.Get(ctx, id)
}

func (r *DbInterface) Update(ctx context.Context, position model.Position) error {
	return r.repo.Update(ctx, position)
}
