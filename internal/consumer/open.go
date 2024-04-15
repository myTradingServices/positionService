package consumer

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Open struct {
	positions PositionController
	localPos  LPositionController
	prices    PriceManipulator

	lisenCh chan model.Position
}

type PositionController interface {
	Update(ctx context.Context, position model.Position) error
	GetAllOpened(ctx context.Context) ([]model.Position, error)
}

type LPositionController interface {
	Add(userID string, ch chan model.Position)
	Deleete(userID string) (wasDeleted bool)
	Get(userID string) (chan model.Position, bool)
}

type PriceManipulator interface {
	Add(key model.SymbOperDTO, ch chan model.Price)
	Delete(key model.SymbOperDTO) (wasDeleted bool)
}

func NewOpener(
	localPositions LPositionController,
	prices PriceManipulator,
	positions PositionController,
	lisenCh chan model.Position,
) *Open {
	return &Open{
		localPos:  localPositions,
		prices:    prices,
		positions: positions,
		lisenCh:   lisenCh,
	}
}

func (o *Open) Open(ctx context.Context) {

	consume := func(posCh chan model.Position, userID string) {
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
						"UserID": userID,
						"Symbol": p.Symbol,
						"PNL":    pnl,
					}).Info("Profit and loss info")

					if totalPNL = totalPNL.Add(pnl); totalPNL.IsNegative() { // close all popsirion // exit go rutine
						log.Info("Warning proffit is less than thero, closing all positions for user: ", userID)

						o.localPos.Deleete(userID)

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

							o.prices.Delete(model.SymbOperDTO{
								Symbol: symb,
								UserID: userID,
							})
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
		o.localPos.Add(opened.UserID.String(), posCh)

		go consume(posCh, opened.UserID.String())

		posCh <- opened
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
				ch, ok := o.localPos.Get(pos.UserID.String())
				if !ok {
					posCh := make(chan model.Position)
					o.localPos.Add(pos.UserID.String(), posCh)

					go consume(posCh, pos.UserID.String())

					posCh <- pos
					continue
				}

				ch <- pos
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
