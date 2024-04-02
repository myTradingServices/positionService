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
	positionMap service.PositionMapInterface
	chPosition  chan model.Position
}

func NewPositionBridge(chPosition chan model.Position, positionMap service.PositionMapInterface) PositionBridger {
	return &positionBridge{
		chPosition:  chPosition,
		positionMap: positionMap,
	}
}

func (p *positionBridge) PositionBridge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tmpPosition := <-p.chPosition:
			{
				writeChanel, ok := p.positionMap.Get(tmpPosition.UserID.String())
				if !ok {
					log.Error("chanel is not stored")
					return // ?
				}

				select {
				case writeChanel <- tmpPosition:
				default:
				}

			}
		}
	}
}
