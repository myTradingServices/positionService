package service

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type PositionMapInterface interface {
	Add(userID string, ch chan model.Position)
	Get(userID string) (chan model.Position, bool)
}

type posMap struct {
	userPosChMap repository.PositionMapInterface
}

func NewPositionMap(userPosChMap repository.PositionMapInterface) PositionMapInterface {
	return &posMap{
		userPosChMap: userPosChMap,
	}
}

func (p *posMap) Add(userID string, ch chan model.Position) {
	p.userPosChMap.Add(userID, ch)
}
func (p *posMap) Get(userID string) (chan model.Position, bool) {
	return p.userPosChMap.Get(userID)
}
