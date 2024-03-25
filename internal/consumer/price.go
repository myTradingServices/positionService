package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type price struct {
	chResever <-chan chan model.Price
	mapServ   service.MapInterface[string]
}

func NewPositionConsumer(chResever <-chan chan model.Price, mapServ service.MapInterface[string]) PositionConsumer {
	return &price{
		chResever: chResever,
		mapServ:   mapServ,
	}
}

type PositionConsumer interface {
	Consume(ctx context.Context)
}

func (p *price) Consume(ctx context.Context) {
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
