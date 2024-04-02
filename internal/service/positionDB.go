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
	Update(ctx context.Context, position model.Position) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
	GetAllOpened() []model.Position
	Deleete(ctx context.Context, id uuid.UUID) error
}

func NewPositionService(repo repository.DBInterface) DBInterface {
	return &dbRepo{
		repo: repo,
	}
}

func (r *dbRepo) Add(ctx context.Context, position model.Position) error {
	return r.repo.Add(ctx, position)
}

func (r *dbRepo) Update(ctx context.Context, position model.Position) error {
	return r.repo.Update(ctx, position)
}

func (r *dbRepo) Get(ctx context.Context, id uuid.UUID) ([]model.Position, error) {
	return r.repo.Get(ctx, id)
}

// TODO: complite
func (r *dbRepo) GetAllOpened() []model.Position

func (r *dbRepo) Deleete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Deleete(ctx, id)
}
