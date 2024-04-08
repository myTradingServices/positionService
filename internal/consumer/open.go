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
	localPositions service.LPstController
	prices         service.PrcManipulator
	positions      service.PstController

	lisenCh chan model.Position
}

func NewOpener(
	localPositions *service.LocalPositions,
	prices *service.Prices,
	positions *service.Positons,
	lisenCh chan model.Position,
) Opener {
	return &opener{
		localPositions: localPositions,
		prices:         prices,
		positions:      positions,
		lisenCh:        lisenCh,
	}
}

func (o *opener) Open(ctx context.Context) {

	consumerCore := func(posCh chan model.Position, userID string) {
		pnlPrices := make(map[string]model.Position)
		priceCh := make(chan model.Price)
		totalPNL := decimal.Decimal{}

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
						pnlPrices[p.Symbol] = model.Position{
							OpenPrice: p.OpenPrice,
							Long:      p.Long,
						}
						o.prices.Add(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: p.UserID.String(),
						}, priceCh)

					} else {
						o.prices.Delete(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: userID,
						})
						delete(pnlPrices, p.Symbol)
					}
				}
			case p := <-priceCh:
				{
					pnl := computePNL(pnlPrices[p.Symbol].OpenPrice, p, pnlPrices[p.Symbol].Long)

					if pnlPrices[p.Symbol].Long {
						openP := pnlPrices[p.Symbol].OpenPrice
						pnlPrices[p.Symbol] = model.Position{
							OpenPrice:  openP,
							Long:       true,
							ClosePrice: p.Ask,
						}
					} else {
						opTmp := pnlPrices[p.Symbol].OpenPrice
						pnlPrices[p.Symbol] = model.Position{
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

					if totalPNL = totalPNL.Add(pnl); totalPNL.IsNegative() { // close all popsirion // exit go rutine
						log.Info("Warning proffit is less than thero, closing all positions for user: ", userID)

						o.localPositions.Deleete(userID)

						for symb, pos := range pnlPrices {
							err := o.positions.Update(ctx, model.Position{
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

	allOpened, err := o.positions.GetAllOpened(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	for _, opened := range allOpened {
		posCh := make(chan model.Position)
		o.localPositions.Add(opened.UserID.String(), posCh)
		go consumerCore(posCh, opened.UserID.String())
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
				if ch, ok := o.localPositions.Get(pos.UserID.String()); ok {
					go consumerCore(ch, pos.UserID.String())
					continue
				}

				posCh := make(chan model.Position)
				o.localPositions.Add(pos.UserID.String(), posCh)
				go consumerCore(posCh, pos.UserID.String())
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
