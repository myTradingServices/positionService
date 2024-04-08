package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type PositionBridger interface {
	PositionBridge(ctx context.Context)
}

type positionBridge struct {
	localPositions service.LPstGeter
	ch             chan model.Position
}

func NewPositionBridge(ch chan model.Position, positionMap *service.LocalPositions) PositionBridger {
	return &positionBridge{
		ch:             ch,
		localPositions: positionMap,
	}
}

func (p *positionBridge) PositionBridge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case pst := <-p.ch:
			{
				writeTo, ok := p.localPositions.Get(pst.UserID.String())
				if !ok {
					log.Error("chanel is not stored")
					return
				}

				writeTo <- pst
			}
		}
	}
}
