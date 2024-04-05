package repository

import (
	"errors"
	"sync"

	"github.com/mmfshirokan/positionService/internal/model"
)

var (
	ErrGetElementThatNotExist error = errors.New("attempt of Geting eliment that do not exist")
)

type PriceMapInterface interface {
	GetAllChanForSymb(symb string) (res []chan model.Price, err error)
	Get(key model.SymbOperDTO) chan model.Price

	Add(key model.SymbOperDTO, ch chan model.Price)
	Delete(key model.SymbOperDTO)
}

type symbUserPrice struct {
	symbUserMap map[string]map[string]chan model.Price
	mut         sync.RWMutex
}

func NewSymbOperMap(symbUserMap map[string]map[string]chan model.Price) PriceMapInterface {
	return &symbUserPrice{
		symbUserMap: symbUserMap,
	}
}

func (s *symbUserPrice) GetAllChanForSymb(symb string) (res []chan model.Price, _ error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	chMap, ok := s.symbUserMap[symb]

	if !ok {
		return nil, ErrGetElementThatNotExist
	}

	for _, val := range chMap {
		res = append(res, val)
	}

	return res, nil
}

func (s *symbUserPrice) Add(key model.SymbOperDTO, ch chan model.Price) {
	s.mut.RLock()
	_, ok := s.symbUserMap[key.Symbol]
	s.mut.RUnlock()

	if !ok {
		underlying := make(map[string]chan model.Price)
		underlying[key.UserID] = ch

		s.mut.Lock()
		s.symbUserMap[key.Symbol] = underlying
		s.mut.Unlock()

		return
	}

	s.mut.Lock()
	s.symbUserMap[key.Symbol][key.UserID] = ch
	s.mut.Unlock()
}
func (s *symbUserPrice) Get(key model.SymbOperDTO) chan model.Price {
	s.mut.Lock()
	ch := s.symbUserMap[key.Symbol][key.UserID]
	s.mut.Unlock()

	return ch
}

func (s *symbUserPrice) Delete(key model.SymbOperDTO) {
	s.mut.RLock()
	underlying := s.symbUserMap[key.Symbol]
	s.mut.RUnlock()

	s.mut.Lock()
	delete(underlying, key.UserID)
	if len(underlying) == 0 {
		delete(s.symbUserMap, key.Symbol)
	}
	s.mut.Unlock()
}
