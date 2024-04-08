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
	priceMap service.PrcGeter
	chPrice  chan model.Price
}

func NewPriceBridge(chPrice chan model.Price, priceMap *service.Prices) PriceBridger {
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
		case prc := <-p.chPrice:
			{
				writeTo, ok := p.priceMap.GetAllChanForSymb(prc.Symbol)
				if !ok {
					log.Error("can't get chanels via method GetAllChanForSymb:", ok)
					return
				}

				for _, writeChanel := range writeTo {
					writeChanel <- prc
				}
			}
		}
	}
}
