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
	mapServ service.MapInterface
}

type Pricer interface {
	ConsumePrice(ctx context.Context)
}

func NewPositionConsumer(db service.DBInterface, mapServ service.MapInterface) Pricer {
	return &position{
		db:      db,
		mapServ: mapServ,
	}
}

func (p *position) ConsumePrice(ctx context.Context) {
	t := time.NewTicker(time.Second * 2)
	defer t.Stop()

	positions, err := p.db.GetAllOpend(ctx)
	if err != nil {
		log.Errorf("repository error: %v, exiting consumer", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			{
				timeBigest := time.Time{}

				for _, pos := range positions {
					key := model.SymbOperDTO{
						Symbol:    pos.Symbol,
						Operation: pos.OperationID.String(),
					}

					priceChan := make(chan model.Price)
					err := p.mapServ.Add(key, priceChan)
					if err != nil {
						log.Error("Map repository error:", err)
						return
					}

					if timeBigest.Before(pos.CreatedAt) {
						timeBigest = pos.CreatedAt
					}

					go func(ch chan model.Price, openPrice decimal.Decimal, buy bool) {
						for {
							price := <-ch
							pnl := computePNL(openPrice, price, buy)

							log.WithFields(log.Fields{
								"Symbol: ": price.Symbol,
								"PNL: ":    pnl,
							}).Info("Profit and loss info")
						}
					}(priceChan, pos.OpenPrice, pos.Buy)
				}

				positions, err = p.db.GetLaterThen(ctx, timeBigest)
				if err != nil {
					log.Error("repository error:", err)
					return
				}
			}
		}
	}
}

func computePNL(openPrice decimal.Decimal, currentPrice model.Price, buy bool) decimal.Decimal {
	if buy {
		return currentPrice.Ask.Add(openPrice.Neg())
	}

	return openPrice.Add(currentPrice.Bid.Neg())
}
