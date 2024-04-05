package repository

import (
	"sync"

	"github.com/mmfshirokan/positionService/internal/model"
)

type PositionMapInterface interface {
	Add(userID string, ch chan model.Position)
	Get(userID string) (chan model.Position, bool)

	Delete(userID string)
}

type posMap struct {
	pMap map[string]chan model.Position
	mut  sync.RWMutex
}

func NewPositionMap(pMap map[string]chan model.Position) PositionMapInterface {
	return &posMap{
		pMap: pMap,
	}
}

func (p *posMap) Add(userID string, ch chan model.Position) {
	p.mut.Lock()
	p.pMap[userID] = ch
	p.mut.Unlock()
}

func (p *posMap) Get(userID string) (ch chan model.Position, ok bool) {
	p.mut.RLock()
	ch, ok = p.pMap[userID]
	p.mut.RUnlock()

	return
}
func (p *posMap) Delete(userID string) {
	p.mut.Lock()
	delete(p.pMap, userID)
	p.mut.Unlock()
}
