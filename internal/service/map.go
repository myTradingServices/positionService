package service

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type MapInterface interface {
	Add(key model.SymbOperDTO, val chan model.Price) error
	GetAllChanForSymb(symb string) (res []chan model.Price, err error)
	Get(key model.SymbOperDTO) (res chan model.Price, err error)
}

type symbOperPriceMap struct {
	repo repository.MapInterface
}

func NewSymbOperMap(repo repository.MapInterface) MapInterface {
	return &symbOperPriceMap{
		repo: repo,
	}
}
func (s *symbOperPriceMap) Add(key model.SymbOperDTO, val chan model.Price) error {
	return s.repo.Add(key, val)
}

func (s *symbOperPriceMap) GetAllChanForSymb(symb string) (res []chan model.Price, _ error) {
	return s.repo.GetAllChanForSymb(symb)
}

func (s *symbOperPriceMap) Get(key model.SymbOperDTO) (res chan model.Price, err error) {
	return s.repo.Get(key)
}
