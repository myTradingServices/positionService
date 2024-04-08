package service

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type LocalPositions struct {
	pst *repository.Positions
}

type LPstGeter interface {
	Get(userID string) (chan model.Position, bool)
}

type LPstController interface {
	LPstGeter
	Add(userID string, ch chan model.Position)
	Deleete(userID string) (wasDeleted bool)
}

func NewLocalPositions(pst *repository.Positions) *LocalPositions {
	return &LocalPositions{
		pst: pst,
	}
}

func (lp *LocalPositions) Add(userID string, ch chan model.Position) {
	lp.pst.Add(userID, ch)
}
func (lp *LocalPositions) Get(userID string) (chan model.Position, bool) {
	return lp.pst.Get(userID)
}

func (lp *LocalPositions) Deleete(userID string) (wasDeleted bool) {
	return lp.pst.Delete(userID)
}
