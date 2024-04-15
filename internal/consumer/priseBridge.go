package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
)

type PriceBridge struct {
	priceMap PriceGeter
	chPrice  chan model.Price
}

type PriceGeter interface {
	GetAllChanForSymb(symb string) (chanels []chan model.Price, isSuccessfull bool)
}

func NewPriceBridge(chPrice chan model.Price, priceMap PriceGeter) *PriceBridge {
	return &PriceBridge{
		chPrice:  chPrice,
		priceMap: priceMap,
	}
}

func (p *PriceBridge) PriceBridge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case prc := <-p.chPrice:
			{
				writeTo, ok := p.priceMap.GetAllChanForSymb(prc.Symbol)
				if !ok {
					log.Error("can't get chanels via method GetAllChanForSymb. ok: ", ok)
					continue
				}

				for _, writeChanel := range writeTo {
					writeChanel <- prc
				}
			}
		}
	}
}
