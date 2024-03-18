package service

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
)

type MapInterface[T model.SymbOperDTO | string] interface {
	Add(key T, val chan model.Price) error
	Get(key T) (chan model.Price, error)
	Delete(key T) error
	Contains(key T) bool
	GetKeys() []T
}

type symbPriceMap struct {
	repo repository.MapInterface[string]
}

func NewSymbPriceMap(repo repository.MapInterface[string]) MapInterface[string] {
	return &symbPriceMap{
		repo: repo,
	}
}

func (s *symbPriceMap) Add(key string, val chan model.Price) error {
	return s.repo.Add(key, val)
}
func (s *symbPriceMap) Get(key string) (chan model.Price, error) {
	return s.repo.Get(key)
}
func (s *symbPriceMap) Delete(key string) error {
	return s.repo.Delete(key)
}

func (s *symbPriceMap) Contains(key string) bool {
	return s.repo.Contains(key)
}

func (s *symbPriceMap) GetKeys() []string {
	return s.repo.GetKeys()
}

type symbOperPriceMap struct {
	repo repository.MapInterface[model.SymbOperDTO]
}

func NewSymbOperMap(repo repository.MapInterface[model.SymbOperDTO]) MapInterface[model.SymbOperDTO] {
	return &symbOperPriceMap{
		repo: repo,
	}
}
func (s *symbOperPriceMap) Add(key model.SymbOperDTO, val chan model.Price) error {
	return s.repo.Add(key, val)
}
func (s *symbOperPriceMap) Get(key model.SymbOperDTO) (chan model.Price, error) {
	return s.repo.Get(key)
}
func (s *symbOperPriceMap) Delete(key model.SymbOperDTO) error {
	return s.repo.Delete(key)
}

func (s *symbOperPriceMap) Contains(key model.SymbOperDTO) bool {
	return s.repo.Contains(key)
}

func (s *symbOperPriceMap) GetKeys() []model.SymbOperDTO {
	return s.repo.GetKeys()
}
