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

type MapInterface[T model.SymbOperDTO | string] interface {
	Add(key T, value chan model.Price) error
	Get(key T) (chan model.Price, error)
	Delete(key T) error
	Contains(key T) bool
	GetKeys() (keyArr []T)
}

type symbPrice struct {
	symbPriceMap map[string]chan model.Price
	mut          sync.RWMutex
}

func NewStringPrice(spMap map[string]chan model.Price) MapInterface[string] {
	return &symbPrice{
		symbPriceMap: spMap,
	}
}

func (s *symbPrice) Add(key string, val chan model.Price) error {
	s.mut.RLock()
	_, ok := s.symbPriceMap[key]
	s.mut.RUnlock()

	if ok {
		return ErrAddElementThatAlreadyExist
	}
	s.mut.Lock()
	s.symbPriceMap[key] = val
	s.mut.Unlock()

	return nil
}

func (s *symbPrice) Get(key string) (chan model.Price, error) {
	s.mut.RLock()
	val, ok := s.symbPriceMap[key]
	s.mut.RUnlock()

	if !ok {
		return nil, ErrGetElementThatNotExist
	}

	return val, nil
}
func (s *symbPrice) Delete(key string) error {
	s.mut.RLock()
	_, ok := s.symbPriceMap[key]
	s.mut.RUnlock()

	if !ok {
		return ErrDeleteElementThatNotExist
	}
	s.mut.Lock()
	//close(val)
	delete(s.symbPriceMap, key)
	s.mut.Unlock()

	return nil
}

func (s *symbPrice) Contains(key string) bool {
	s.mut.RLock()
	_, ok := s.symbPriceMap[key]
	s.mut.RUnlock()

	return ok
}

func (s *symbPrice) GetKeys() (keyArr []string) {
	s.mut.RLock()
	for symb := range s.symbPriceMap {
		keyArr = append(keyArr, symb)
	}
	s.mut.RUnlock()
	return keyArr
}

type symbOperPrice struct {
	symbOperMap map[string]map[string]chan model.Price
	mut         sync.RWMutex
}

func NewSymbOperMap(sopMap map[string]map[string]chan model.Price) MapInterface[model.SymbOperDTO] {
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

func (s *symbOperPrice) Get(key model.SymbOperDTO) (chan model.Price, error) {
	s.mut.RLock()
	val, ok := s.symbOperMap[key.Symbol][key.Operation]
	s.mut.RUnlock()

	if !ok {
		return nil, ErrGetElementThatNotExist
	}

	return val, nil
}

func (s *symbOperPrice) Delete(key model.SymbOperDTO) error {
	s.mut.RLock()
	_, ok := s.symbOperMap[key.Symbol][key.Operation]
	s.mut.RUnlock()

	if !ok {
		return ErrDeleteElementThatNotExist
	}

	s.mut.Lock()
	// close(val)
	delete(s.symbOperMap[key.Symbol], key.Operation)
	s.mut.Unlock()

	return nil
}

func (s *symbOperPrice) Contains(key model.SymbOperDTO) bool {
	s.mut.RLock()
	_, ok := s.symbOperMap[key.Symbol][key.Operation]
	s.mut.RUnlock()

	return ok
}

func (s *symbOperPrice) GetKeys() (keyArr []model.SymbOperDTO) {
	s.mut.RLock()
	for symb, priceMap := range s.symbOperMap {
		for oper := range priceMap {
			keyArr = append(keyArr, model.SymbOperDTO{
				Symbol:    symb,
				Operation: oper,
			})
		}
	}
	s.mut.RUnlock()

	return keyArr
}
