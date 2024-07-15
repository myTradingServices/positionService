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

	mainCH chan model.Position
}

type PositionController interface {
	Update(ctx context.Context, position model.Position) error
	GetAllOpened(ctx context.Context) ([]model.Position, error)
}

type LPositionController interface {
	Add(userID string, ch chan model.Position)
	Delete(userID string) (wasDeleted bool)
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
		mainCH:    lisenCh,
	}
}

func (o *Open) Open(ctx context.Context) {
	log.Info("Open consumer started")
	defer log.Info("Open consumer exited")

	consume := func(posCh chan model.Position, userID string) {
		log.Info("New gorutine started")

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
						log.Info("Price chanel added: ", p.Symbol)
						pnlPrices[p.Symbol] = model.Position{
							OpenPrice: p.OpenPrice,
							Long:      p.Long,
						}
						o.prices.Add(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: p.UserID.String(),
						}, priceCh)

					} else {
						log.Info("Price chanel deleted: ", p.Symbol)
						o.prices.Delete(model.SymbOperDTO{
							Symbol: p.Symbol,
							UserID: userID,
						})
						delete(pnlPrices, p.Symbol)
					}
				}
			case p := <-priceCh:
				{
					log.Info("price chanel pnl handling started for symbol, user: ", p.Symbol, userID)
					pnl := computePNL(pnlPrices[p.Symbol].OpenPrice, p, pnlPrices[p.Symbol].Long)
					log.Info("PNL: ", pnl)

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

						o.localPos.Delete(userID)

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

		log.Info("Send position via ch (open 1)")
		posCh <- opened
		log.Info("Send position via ch (open 1) done")
	}

	for {
		select {
		case <-ctx.Done():
			{
				log.Info("Consumer Open Shut down")
				return
			}
		case pos := <-o.mainCH:
			{
				ch, ok := o.localPos.Get(pos.UserID.String())
				if !ok {
					posCh := make(chan model.Position)
					o.localPos.Add(pos.UserID.String(), posCh)

					go consume(posCh, pos.UserID.String())

					log.Info("Send position via ch (open 2)")
					posCh <- pos
					log.Info("Send position via ch (open 2) done")
					continue
				}

				log.Info("Send position via ch (open 3)")
				ch <- pos
				log.Info("Send position via ch (open 3) done")
			}
		}
	}
}

func computePNL(openPrice decimal.Decimal, currentPrice model.Price, long bool) decimal.Decimal {
	log.Info("ComputePnl called")
	if long {
		return currentPrice.Ask.Add(openPrice.Neg())
	}
	return openPrice.Add(currentPrice.Bid.Neg())
}
