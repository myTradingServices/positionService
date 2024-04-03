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
	GetAllChanForSymb(symb string) ([]chan model.Price, error)
	GetOrCreate(key model.SymbOperDTO) (chan model.Price, bool)
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

func (s *symbUserPrice) GetOrCreate(key model.SymbOperDTO) (ch chan model.Price, ok bool) {
	s.mut.RLock()
	_, ok = s.symbUserMap[key.Symbol]
	s.mut.RUnlock()

	if !ok {
		underlying := make(map[string]chan model.Price)
		ch = make(chan model.Price)
		underlying[key.UserID] = ch

		s.mut.Lock()
		s.symbUserMap[key.Symbol] = underlying
		s.mut.Unlock()

		return
	}

	s.mut.RLock()
	ch, ok = s.symbUserMap[key.Symbol][key.UserID]
	s.mut.RUnlock()
	if ok {
		return
	}

	ch = make(chan model.Price)
	return
}

func (s *symbUserPrice) GetAllChanForSymb(symb string) (res []chan model.Price, _ error) {
	s.mut.RLock()
	chMap, ok := s.symbUserMap[symb]
	if !ok {
		s.mut.RUnlock()
		return nil, ErrGetElementThatNotExist
	}

	for _, val := range chMap {
		res = append(res, val)
	}
	s.mut.RUnlock()

	return res, nil
}
