package consumer

import (
	"context"
	"strings"
	"time"

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
	log.Info("PriceBridgestarted")
	defer log.Info("PriceBridge exited")

	for {
		select {
		case <-ctx.Done():
			return
		case prc := <-p.chPrice:
			{
				symb := strings.Replace(prc.Symbol, "ol", "", 1)
				log.Info("Price bridge attempt to get symbol for: ", symb)
				writeTo, ok := p.priceMap.GetAllChanForSymb(symb)
				if !ok {
					log.Error("can't get chanels via method GetAllChanForSymb. ok: ", ok)
					time.Sleep(time.Second) //delete
					continue
				}

				for _, writeChanel := range writeTo {
					writeChanel <- prc
				}
			}
		}
	}
}
