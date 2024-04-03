package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Opener interface {
	Open(ctx context.Context)
}

type opener struct {
	posMap   service.PositionMapInterface
	priceMap service.PriceMapInterface

	dbpool  service.DBInterface
	lisenCh chan model.Position
}

func NewOpener(dbPool service.DBInterface, lisenCh chan model.Position) Opener {
	return &opener{
		dbpool:  dbPool,
		lisenCh: lisenCh,
	}
}

func (o *opener) Open(ctx context.Context) {
	consumerCore := func(posCh chan model.Position, userID string) {
		opPrice := make(map[string]model.Position)
		priceCh := make(chan model.Price)
		totalPnl := decimal.Decimal{}

		defer close(priceCh) //\\
		defer close(posCh)

		for {
			select {
			case p := <-posCh:
				{
					if !p.OpenPrice.IsZero() {
						opPrice[p.Symbol] = model.Position{
							OpenPrice: p.OpenPrice,
							Long:      p.Long,
						}

						o.priceMap.Add(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: p.UserID.String(),
						}, priceCh)
					} else {
						o.priceMap.Delete(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: userID,
						})
						delete(opPrice, p.Symbol)
					}
				}
			default:
			}

			select {
			case p := <-priceCh:
				{
					pnl := computePNL(opPrice[p.Symbol].OpenPrice, p, opPrice[p.Symbol].Long)

					log.WithFields(log.Fields{
						"UserID: ": userID,
						"Symbol: ": p.Symbol,
						"PNL: ":    pnl,
					}).Info("Profit and loss info")

					if totalPnl = totalPnl.Add(pnl); totalPnl.IsNegative() {
						log.WithFields(log.Fields{
							"UserID: ": userID,
							"Symbol: ": p.Symbol,
							"PNL: ":    pnl,
						}).Info("Profit and loss info")
					}
				}
			default:
			}
		}
	}

	allOpened, err := o.dbpool.GetAllOpened(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	for _, opened := range allOpened {
		tmpCh := make(chan model.Position)
		o.posMap.Add(opened.UserID.String(), tmpCh) // !NOTE: can combine Add and chanel creation; what's the diff?

		go consumerCore(tmpCh, opened.UserID.String())
	}

	for {
		select {
		case <-ctx.Done():
			{
				log.Info("Consumer Open Shut down")
				return
			}
		case pos := <-o.lisenCh:
			{
				if ch, ok := o.posMap.Get(pos.UserID.String()); ok {
					go consumerCore(ch, pos.UserID.String())
					continue
				}

				tmpCh := make(chan model.Position)
				o.posMap.Add(pos.UserID.String(), tmpCh)
				go consumerCore(tmpCh, pos.UserID.String())
			}

		default:
		}
	}
}

func computePNL(openPrice decimal.Decimal, currentPrice model.Price, long bool) decimal.Decimal {
	if long {
		return currentPrice.Ask.Add(openPrice.Neg())
	}

	return openPrice.Add(currentPrice.Bid.Neg())
}
