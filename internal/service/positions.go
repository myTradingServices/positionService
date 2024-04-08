package service

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type Positons struct {
	repo *repository.Postgres
}

type PstManipulator interface {
	Add(ctx context.Context, position model.Position) error
	Update(ctx context.Context, position model.Position) error
}

type PstController interface {
	PstManipulator
	GetAllOpened(ctx context.Context) ([]model.Position, error)
}

func NewPosition(repo *repository.Postgres) *Positons {
	return &Positons{
		repo: repo,
	}
}

func (p *Positons) Add(ctx context.Context, pst model.Position) error {
	return p.repo.Add(ctx, pst)
}

func (p *Positons) Update(ctx context.Context, pst model.Position) error {
	return p.repo.Update(ctx, pst)
}

func (p *Positons) GetAllOpened(ctx context.Context) ([]model.Position, error) {
	return p.repo.GetAllOpened(ctx)
}
