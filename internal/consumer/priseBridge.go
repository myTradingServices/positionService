package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type PriceBridger interface {
	PriceBridge(ctx context.Context)
}

type priceBridge struct {
	priceMap service.PriceMapInterface
	chPrice  chan model.Price
}

func NewPriceBridge(chPrice chan model.Price, priceMap service.PriceMapInterface) PriceBridger {
	return &priceBridge{
		chPrice:  chPrice,
		priceMap: priceMap,
	}
}

func (p *priceBridge) PriceBridge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tmpPrice := <-p.chPrice:
			{
				writeChanels, err := p.priceMap.GetAllChanForSymb(tmpPrice.Symbol)
				if err != nil {
					log.Error("can't get chanels via method GetAllChanForSymb:", err)
					return
				}

				for _, writeChanel := range writeChanels {
					writeChanel <- tmpPrice
				}
			}
		}
	}
}
