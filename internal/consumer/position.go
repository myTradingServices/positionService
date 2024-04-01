package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type position struct {
	ch      chan model.Position
	mapServ service.MapInterface
}

type Pricer interface {
	ConsumePrice(ctx context.Context)
}

func NewPositionConsumer(ch chan model.Position, mapServ service.MapInterface) Pricer {
	return &position{
		ch:      ch,
		mapServ: mapServ,
	}
}

func (p *position) ConsumePrice(ctx context.Context) {
	for {

		select {
		case <-ctx.Done():
			return
		default:
			{

				select {
				case pos := <-p.ch:
					{
						key := model.SymbOperDTO{
							Symbol:    pos.Symbol,
							Operation: pos.OperationID.String(),
						}

						priceChan := make(chan model.Price)
						err := p.mapServ.Add(key, priceChan)
						if err != nil {
							log.Error("Map repository error:", err)
						}

						go func(ch chan model.Price, openPrice decimal.Decimal, buy bool) {
							for {
								price, ok := <-ch
								if !ok {
									return
								}

								pnl := computePNL(openPrice, price, buy)

								log.WithFields(log.Fields{
									"Symbol: ": price.Symbol,
									"PNL: ":    pnl,
								}).Info("Profit and loss info")
							}
						}(priceChan, pos.OpenPrice, pos.Long)
					}
				default:
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

// timeBigest := time.Time{}

// for _, pos := range positions {
// 	key := model.SymbOperDTO{
// 		Symbol:    pos.Symbol,
// 		Operation: pos.OperationID.String(),
// 	}

// 	priceChan := make(chan model.Price)
// 	err := p.mapServ.Add(key, priceChan)
// 	if err != nil {
// 		log.Error("Map repository error:", err)
// 		return
// 	}

// 	if timeBigest.After(pos.CreatedAt) {
// 		timeBigest = pos.CreatedAt
// 	}

// 	go func(ch chan model.Price, openPrice decimal.Decimal, buy bool) {
// 		for {
// 			price := <-ch
// 			pnl := computePNL(openPrice, price, buy)

// 			log.WithFields(log.Fields{
// 				"Symbol: ": price.Symbol,
// 				"PNL: ":    pnl,
// 			}).Info("Profit and loss info")
// 		}
// 	}(priceChan, pos.OpenPrice, pos.Long)
// }

// positions, err = p.db.GetLaterThen(ctx, timeBigest)
// if err != nil {
// 	log.Error("repository error:", err)
// 	return
// }
