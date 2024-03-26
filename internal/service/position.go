package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type dbRepo struct {
	repo repository.DBInterface
}

type DBInterface interface {
	Add(ctx context.Context, position model.Position) error
	Deleete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
	Update(ctx context.Context, position model.Position) error
	GetAllOpend(ctx context.Context) ([]model.Position, error)
	GetOneState(ctx context.Context, operID uuid.UUID) (bool, error)
}

func NewPositionService(repo repository.DBInterface) DBInterface {
	return &dbRepo{
		repo: repo,
	}
}

func (r *dbRepo) Add(ctx context.Context, position model.Position) error {
	return r.repo.Add(ctx, position)
}

func (r *dbRepo) Deleete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Deleete(ctx, id)
}

func (r *dbRepo) Get(ctx context.Context, id uuid.UUID) ([]model.Position, error) {
	return r.repo.Get(ctx, id)
}

func (r *dbRepo) Update(ctx context.Context, position model.Position) error {
	return r.repo.Update(ctx, position)
}

func (r *dbRepo) GetAllOpend(ctx context.Context) ([]model.Position, error) {
	return r.repo.GetAllOpend(ctx)
}

func (r *dbRepo) GetOneState(ctx context.Context, operID uuid.UUID) (bool, error) {
	return r.repo.GetOneState(ctx, operID)
}
