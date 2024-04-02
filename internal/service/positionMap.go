package service

import "github.com/mmfshirokan/positionService/internal/model"

type PositionMapInterface interface {
	Add(userID string, ch chan model.Position)
	Get(userID string) (chan model.Position, bool)
}

type posMap struct {
	userPosChMap map[string]chan model.Price
}

func NewPositionMap(userPosChMap map[string]chan model.Price) PositionMapInterface {
	return &posMap{
		userPosChMap: userPosChMap,
	}
}

func (p *posMap) Add(userID string, ch chan model.Position)
func (p *posMap) Get(userID string) (chan model.Position, bool)
