package consumer

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Opener interface {
	Open(ctx context.Context)
}

type opener struct {
	positionsCH service.PositionMapInterface
	pricesCH    service.PriceMapInterface

	dbpool  service.DBInterface
	lisenCh chan model.Position
}

func NewOpener(
	positionsCH service.PositionMapInterface,
	pricesCH service.PriceMapInterface,
	dbPool service.DBInterface,
	lisenCh chan model.Position) Opener {
	return &opener{
		positionsCH: positionsCH,
		pricesCH:    pricesCH,
		dbpool:      dbPool,
		lisenCh:     lisenCh,
	}
}

func (o *opener) Open(ctx context.Context) {

	consumerCore := func(posCh chan model.Position, userID string) {
		pricesForPNL := make(map[string]model.Position)
		priceCh := make(chan model.Price)
		totalPnl := decimal.Decimal{}

		defer func() {
			defer close(priceCh)
			defer close(posCh)
			log.Infof("User: %v is closed", userID)
		}()

		for {
			select {
			case p := <-posCh:
				{
					if p.ClosePrice.IsZero() {
						pricesForPNL[p.Symbol] = model.Position{
							OpenPrice: p.OpenPrice,
							Long:      p.Long,
						}
						o.pricesCH.Add(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: p.UserID.String(),
						}, priceCh)

					} else {
						o.pricesCH.Delete(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: userID,
						})
						delete(pricesForPNL, p.Symbol)
					}
				}
			case p := <-priceCh:
				{
					pnl := computePNL(pricesForPNL[p.Symbol].OpenPrice, p, pricesForPNL[p.Symbol].Long)

					if pricesForPNL[p.Symbol].Long {
						opTmp := pricesForPNL[p.Symbol].OpenPrice
						pricesForPNL[p.Symbol] = model.Position{
							OpenPrice:  opTmp,
							Long:       true,
							ClosePrice: p.Ask,
						}
					} else {
						opTmp := pricesForPNL[p.Symbol].OpenPrice
						pricesForPNL[p.Symbol] = model.Position{
							OpenPrice:  opTmp,
							Long:       true,
							ClosePrice: p.Ask,
						}
					}

					log.WithFields(log.Fields{
						"UserID: ": userID,
						"Symbol: ": p.Symbol,
						"PNL: ":    pnl,
					}).Info("Profit and loss info")

					if totalPnl = totalPnl.Add(pnl); totalPnl.IsNegative() { // close all popsirion // exit go rutine
						log.Info("Warning proffit is less than thero, closing all positions for user: ", userID)

						o.positionsCH.Deleete(userID)

						for symb, pos := range pricesForPNL {
							err := o.dbpool.Update(ctx, model.Position{
								UserID:     uuid.MustParse(userID),
								Symbol:     symb,
								ClosePrice: pos.ClosePrice,
							})
							if err != nil {
								log.WithFields(
									log.Fields{
										"UserID: ":     userID,
										"Symbol: ":     symb,
										"ClosePrice: ": pos.ClosePrice,
									},
								).Error("Can't close position")
								continue
							}
						}

						return
					}
				}
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
		o.positionsCH.Add(opened.UserID.String(), tmpCh)
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
				if ch, ok := o.positionsCH.Get(pos.UserID.String()); ok {
					go consumerCore(ch, pos.UserID.String())
					continue
				}

				tmpCh := make(chan model.Position)
				o.positionsCH.Add(pos.UserID.String(), tmpCh)
				go consumerCore(tmpCh, pos.UserID.String())
			}
		}
	}
}

func computePNL(openPrice decimal.Decimal, currentPrice model.Price, long bool) decimal.Decimal {
	if long {
		return currentPrice.Ask.Add(openPrice.Neg())
	}

	return openPrice.Add(currentPrice.Bid.Neg())
}
