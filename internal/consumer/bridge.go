package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type BridgeInterface interface {
	Bridge(ctx context.Context)
}

type bridge struct {
	positionMap service.MapInterface
	ch          chan model.Price
}

func NewBridge(ch chan model.Price, positionMap service.MapInterface) BridgeInterface {
	return &bridge{
		ch:          ch,
		positionMap: positionMap,
	}
}

func (b *bridge) Bridge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tmpPrice := <-b.ch:
			{
				writeChanels, err := b.positionMap.GetAllChanForSymb(tmpPrice.Symbol)
				if err != nil {
					log.Error("can't get chanels via method GetAllChanForSymb:", err)
					return
				}

				for _, writeChanel := range writeChanels {
					select {
					case writeChanel <- tmpPrice:
					default:
					}
				}
			}
		}
	}
}
