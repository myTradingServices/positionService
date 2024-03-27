package repository

import (
	"errors"
	"sync"

	"github.com/mmfshirokan/positionService/internal/model"
)

var (
	ErrAddElementThatAlreadyExist error = errors.New("attempt of Adding eliment for already existing key")
	ErrGetElementThatNotExist     error = errors.New("attempt of Geting eliment that do not exist")
	ErrDeleteElementThatNotExist  error = errors.New("attempt of Deleting eliment that do not exist")
)

type MapInterface interface {
	Add(key model.SymbOperDTO, value chan model.Price) error
	GetAllChanForSymb(symb string) (res []chan model.Price, _ error)
}

type symbOperPrice struct {
	symbOperMap map[string]map[string]chan model.Price
	mut         sync.RWMutex
}

func NewSymbOperMap(sopMap map[string]map[string]chan model.Price) MapInterface {
	return &symbOperPrice{
		symbOperMap: sopMap,
	}
}

func (s *symbOperPrice) Add(key model.SymbOperDTO, val chan model.Price) error {
	s.mut.RLock()
	_, ok := s.symbOperMap[key.Symbol]
	s.mut.RUnlock()

	if !ok {
		underlying := make(map[string]chan model.Price)
		underlying[key.Operation] = val

		s.mut.Lock()
		s.symbOperMap[key.Symbol] = underlying
		s.mut.Unlock()

		return nil
	}

	s.mut.RLock()
	_, ok = s.symbOperMap[key.Symbol][key.Operation]
	s.mut.RUnlock()

	if ok {
		return ErrAddElementThatAlreadyExist
	}

	s.mut.Lock()
	underlying := s.symbOperMap[key.Symbol]
	underlying[key.Operation] = val
	s.mut.Unlock()

	return nil
}

func (s *symbOperPrice) GetAllChanForSymb(symb string) (res []chan model.Price, _ error) {
	s.mut.RLock()
	chMap, ok := s.symbOperMap[symb]

	if !ok {
		return nil, ErrGetElementThatNotExist
	}
	for _, val := range chMap {
		res = append(res, val)
	}
	s.mut.RUnlock()

	return res, nil
}
