package consumer

import (
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type price struct {
	chResever <-chan chan model.Price
	mapServ   service.MapInterface[string]
}

func NewPriceConsumer(chResever <-chan chan model.Price, mapServ service.MapInterface[string]) PriceInterface {
	return &price{
		chResever: chResever,
		mapServ:   mapServ,
	}
}

type PriceInterface interface {
	ConsumePrice()
}

func (p *price) ConsumePrice() {
	for {
		ch := <-p.chResever
		price := <-ch

		if ok := p.mapServ.Contains(price.Symbol); !ok {
			if err := p.mapServ.Add(price.Symbol, ch); err != nil {
				log.Errorf("Error in adding chanel: %v, exiting the consumer loop now", err)
				break
			}
		}
	}
}
