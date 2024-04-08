package repository

import (
	"sync"

	"github.com/mmfshirokan/positionService/internal/model"
)

// type PriceMapInterface interface {
// 	GetAllChanForSymb(symb string) (res []chan model.Price, err error)
// 	Get(key model.SymbOperDTO) chan model.Price

// 	Add(key model.SymbOperDTO, ch chan model.Price)
// 	Delete(key model.SymbOperDTO)
// }

type Prices struct {
	prc map[string]map[string]chan model.Price
	mut sync.RWMutex
}

func NewPrices(prc map[string]map[string]chan model.Price) *Prices {
	return &Prices{
		prc: prc,
	}
}

func (s *Prices) GetAllChanForSymb(symb string) (chanels []chan model.Price, isSuccessfull bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	uids, ok := s.prc[symb]
	if !ok {
		return nil, false
	}

	var chs []chan model.Price
	for _, val := range uids {
		chs = append(chs, val)
	}

	return chs, true
}

func (s *Prices) Add(key model.SymbOperDTO, ch chan model.Price) {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, ok := s.prc[key.Symbol]

	if !ok {
		underlying := make(map[string]chan model.Price)
		underlying[key.UserID] = ch
		s.prc[key.Symbol] = underlying

		return
	}

	s.prc[key.Symbol][key.UserID] = ch
}

func (s *Prices) Get(key model.SymbOperDTO) (ch chan model.Price, isSuccessfull bool) {
	s.mut.Lock()
	defer s.mut.Unlock()

	ch, ok := s.prc[key.Symbol][key.UserID]
	return ch, ok
}

func (s *Prices) Delete(key model.SymbOperDTO) (wasDeleted bool) {
	s.mut.Lock()
	defer s.mut.Unlock()

	under, ok := s.prc[key.Symbol]
	if !ok {
		return false
	}

	_, ok = under[key.UserID]
	if !ok {
		return false
	}

	delete(under, key.UserID)
	if len(under) == 0 {
		delete(s.prc, key.Symbol)
	}

	return true
}
