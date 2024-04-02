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

// func (s *symbUserPrice) GetOrCreate(key model.SymbOperDTO) (ch chan model.Price, ok bool) {
// 	s.mut.RLock()
// 	_, ok = s.symbUserMap[key.Symbol]
// 	s.mut.RUnlock()

// 	if !ok {
// 		underlying := make(map[string]chan model.Price)
// 		ch = make(chan model.Price)
// 		underlying[key.UserID] = ch

// 		s.mut.Lock()
// 		s.symbUserMap[key.Symbol] = underlying
// 		s.mut.Unlock()

// 		return
// 	}

// 	s.mut.RLock()
// 	ch, ok = s.symbUserMap[key.Symbol][key.UserID]
// 	s.mut.RUnlock()
// 	if ok {
// 		return
// 	}

// 	var wg sync.WaitGroup

// 	s.mut.RLock()
// 	for k, val := range s.symbUserMap {
// 		if k == key.Symbol {
// 			continue
// 		}

// 		wg.Add(1)
// 		go func(v map[string]chan model.Price) { // Вопрос: правильно ли применять передачу по ссылке т.к. функция не буде тсоздовать копии и производительность будет выше
// 			defer wg.Done()
// 			for userID, chanel := range v {
// 				if userID == key.UserID {
// 					ch = chanel
// 					ok = true
// 					break
// 				}
// 			}
// 		}(val)
// 	}
// 	s.mut.RUnlock()
// 	wg.Wait()

// 	if ok {
// 		return
// 	}

// 	ch = make(chan model.Price)
// 	s.mut.Lock()
// 	s.symbUserMap[key.Symbol][key.UserID] = ch
// 	s.mut.Unlock()

// 	return
// }

// Returns true if chanel exists, returns false if not

// func (s *symbOperPrice) Get(key model.SymbOperDTO) (chan model.Price, error) {
// 	s.mut.RLock()
// 	val, ok := s.symbOperMap[key.Symbol][key.UserID]
// 	s.mut.RUnlock()

// 	if !ok {
// 		return nil, ErrGetElementThatNotExist
// 	}

// 	return val, nil
// }

// //Note: following method can be used insted of Get
// func (s *symbOperPrice) FindOne(key model.SymbOperDTO) (ch chan model.Price, ok bool) {
// 	s.mut.RLock()
// 	ch, ok = s.symbOperMap[key.Symbol][key.UserID]
// 	s.mut.RUnlock()

// 	if ok {
// 		return
// 	}

// 	var wg sync.WaitGroup

// 	s.mut.RLock()
// 	for k, val := range s.symbOperMap {
// 		if k == key.Symbol {
// 			continue
// 		}

// 		wg.Add(1)
// 		go func(v *map[string] chan model.Price) { // Вопрос: правильно ли применять передачу по ссылке т.к. функция не буде тсоздовать копии и производительность будет выше
// 			defer wg.Done()
// 			for userID, chanel := range *v {
// 				if userID == key.UserID {
// 					ch = chanel
// 					ok = true
// 					break
// 				}
// 			}
// 		}(&val)
// 	}
// 	s.mut.RUnlock()
// 	wg.Wait()

// 	if ok {
// 		return
// 	}

// 	return nil, false
// }

// func (s *symbOperPrice) Add(key model.SymbOperDTO, val chan model.Price) error {
// 	s.mut.RLock()
// 	_, ok := s.symbOperMap[key.Symbol]
// 	s.mut.RUnlock()

// 	if !ok {
// 		underlying := make(map[string]chan model.Price)
// 		underlying[key.UserID] = val

// 		s.mut.Lock()
// 		s.symbOperMap[key.Symbol] = underlying
// 		s.mut.Unlock()

// 		return nil
// 	}

// 	s.mut.RLock()
// 	_, ok = s.symbOperMap[key.Symbol][key.UserID]
// 	s.mut.RUnlock()

// 	if ok {
// 		return ErrAddElementThatAlreadyExist
// 	}

// 	s.mut.Lock()
// 	underlying := s.symbOperMap[key.Symbol]
// 	underlying[key.UserID] = val
// 	s.mut.Unlock()

// 	return nil
// }
