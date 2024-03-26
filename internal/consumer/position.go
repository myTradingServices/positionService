package consumer

import (
	"context"
	"time"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type position struct {
	db      service.DBInterface
	mapServ service.MapInterface[model.SymbOperDTO]
}

type PosirionIntrerface interface {
	ConsumePrice(ctx context.Context)
}

func NewPositionConsumer(db service.DBInterface, mapServ service.MapInterface[model.SymbOperDTO]) PosirionIntrerface {
	return &position{
		db:      db,
		mapServ: mapServ,
	}
}

func (p *position) ConsumePrice(ctx context.Context) {
	positions, err := p.db.GetLaterThen(ctx, time.Now().Add(-time.Minute))
	if err != nil {
		log.Errorf("repository error: %v, exiting consumer", err)
		return
	}

	lastTime := time.Now()

	for {
		for _, pos := range positions {
			priceChan, err := p.mapServ.Get(model.SymbOperDTO{
				Symbol:    pos.Symbol,
				Operation: pos.OperationID.String(),
			})
			if err != nil {
				log.Error("repository error:", err)
				continue
			}

			go func(ch chan model.Price, openPrice decimal.Decimal, buy bool) {
				for {
					price := <-ch
					pnl := computePNL(openPrice, price, buy)

					log.WithFields(log.Fields{
						"Symbol: ": price.Symbol,
						"PNL: ":    pnl,
					}).Info("Profit and loss info")

					time.Sleep(time.Second)
				}
			}(priceChan, pos.OpenPrice, pos.Buy)
		}

		positions, err = p.db.GetLaterThen(ctx, lastTime)
		if err != nil {
			log.Errorf("repository error: %v, exiting consumer", err)
			break
		}
		lastTime = time.Now()

		time.Sleep(time.Millisecond * 3)
	}
}

func computePNL(openPrice decimal.Decimal, currentPrice model.Price, buy bool) decimal.Decimal {
	if buy {
		return currentPrice.Ask.Add(openPrice.Neg())
	}

	return openPrice.Add(currentPrice.Bid.Neg())
}
