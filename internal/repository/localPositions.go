package repository

import (
	"sync"

	"github.com/mmfshirokan/positionService/internal/model"
)

type Positions struct {
	pst map[string]chan model.Position
	mut sync.RWMutex
}

func NewLocalPosition(pst map[string]chan model.Position) *Positions {
	return &Positions{
		pst: pst,
	}
}

func (p *Positions) Add(userID string, ch chan model.Position) {
	p.mut.Lock()
	defer p.mut.Unlock()

	p.pst[userID] = ch
}

func (p *Positions) Get(userID string) (chan model.Position, bool) {
	p.mut.RLock()
	defer p.mut.RUnlock()

	ch, ok := p.pst[userID]

	return ch, ok
}
func (p *Positions) Delete(userID string) (wasDeleted bool) {
	p.mut.Lock()
	defer p.mut.Unlock()

	if _, ok := p.pst[userID]; !ok {
		return false
	}

	delete(p.pst, userID)
	return true
}
