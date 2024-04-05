package service

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type PriceMapInterface interface {
	GetAllChanForSymb(symb string) (res []chan model.Price, err error)
	Add(key model.SymbOperDTO, ch chan model.Price)
	Get(key model.SymbOperDTO) chan model.Price
	Delete(key model.SymbOperDTO)
}

type priceMap struct {
	repo repository.PriceMapInterface
}

func NewSymbOperMap(repo repository.PriceMapInterface) PriceMapInterface {
	return &priceMap{
		repo: repo,
	}
}

func (s *priceMap) GetAllChanForSymb(symb string) (res []chan model.Price, _ error) {
	return s.repo.GetAllChanForSymb(symb)
}

func (s *priceMap) Add(key model.SymbOperDTO, ch chan model.Price) {
	s.repo.Add(key, ch)
}
func (s *priceMap) Get(key model.SymbOperDTO) chan model.Price {
	return s.Get(key)
}

func (s *priceMap) Delete(key model.SymbOperDTO) {
	s.repo.Delete(key)
}
