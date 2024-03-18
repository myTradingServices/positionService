package rpc

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
)

type positionServer struct {
	priceDistribution service.MapInterface[model.SymbOperDTO]
	balance           chan<- model.Position
	serv              service.Interface

	pb.UnimplementedPositionServer
}

type Positioner interface {
	Positioner(pb.Position_PositionerServer) error
}

func NewPositionServer(
	priceDistribution service.MapInterface[model.SymbOperDTO],
	balance chan<- model.Position,
	serv service.Interface,
) pb.PositionServer {
	return &positionServer{
		priceDistribution: priceDistribution,
		balance:           balance,
		serv:              serv,
	}
}

func (s *positionServer) Positioner(stream pb.Position_PositionerServer) error {
	ctx := stream.Context()
	positionOpen := make(map[string]chan bool)

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Info("io.EOF recieved, exiting stream")
			break
		}
		if err != nil {
			log.Error(err)
			return err
		}

		ok := s.priceDistribution.Contains(model.SymbOperDTO{Symbol: recv.Symbol, Operation: recv.OperationID})

		if !ok {
			positionOpen[recv.OperationID] = make(chan bool)
			priceChan := make(chan model.Price)

			s.priceDistribution.Add(model.SymbOperDTO{
				Symbol:    recv.Symbol,
				Operation: recv.OperationID,
			}, priceChan)

			tmpOperUUID, err := uuid.Parse(recv.OperationID)
			if err != nil {
				log.Error("failed to parse operation UUID, wrong rpc data", err)
				return err
			}

			tmpUserUUID, err := uuid.Parse(recv.UserID)
			if err != nil {
				log.Error("failed to parse user UUID, wrong rpc data", err)
				return err
			}

			mod := model.Position{
				OperationID: tmpOperUUID,
				UserID:      tmpUserUUID,
				Symbol:      recv.Symbol,
			}

			go func(ctxx context.Context, openCh chan bool, tmpMod model.Position, open bool, buy bool) { // TODO add parralel symbo receiving
				priceChan, err := s.priceDistribution.Get(model.SymbOperDTO{
					Symbol:    tmpMod.Symbol,
					Operation: tmpMod.OperationID.String(),
				})
				if err != nil {
					log.Error("failed to get price channel, wrong rpc data", err)
					return
				}

				startTime := time.Now()
				tmpPrice := <-priceChan

				for {
					if startTime.Truncate(time.Second).Compare(tmpPrice.Date.Truncate(time.Second)) != 0 {
						tmpPrice = <-priceChan
						continue
					}

					if buy {
						tmpMod.OpenPrice = tmpPrice.Bid
					} else {
						tmpMod.OpenPrice = tmpPrice.Ask
					}

					break
				}

				var closeTime time.Time
				for {
					if !open {
						closeTime = time.Now()
						break
					}

					open = <-openCh
				}

				tmpPrice = <-priceChan

				for {
					if closeTime.Truncate(time.Second).Compare(tmpPrice.Date.Truncate(time.Second)) != 0 {
						tmpPrice = <-priceChan
						continue
					}

					if buy {
						tmpMod.ClosePrice = tmpPrice.Bid
					} else {
						tmpMod.ClosePrice = tmpPrice.Ask
					}

					break
				}

				s.serv.Add(ctxx, tmpMod)
				s.balance <- tmpMod
			}(ctx, positionOpen[recv.OperationID], mod, recv.Open, recv.Buy)

			continue
		}

		positionOpen[recv.OperationID] <- recv.Open
	}

	return nil
}

// reconectionCounter := 0
// for {
// 	s.mut.RLock()
// 	_, ok := s.priceDistribution[recv.Symbol]
// 	s.mut.RUnlock()

// 	if reconectionCounter > 14 {
// 		log.Error("ERROR failed to find symbol exiting gorutine")
// 		return
// 	}

// 	if !ok {
// 		log.Error("error: symbol not found, atempting to refind in 1 second")
// 		reconectionCounter++
// 		time.Sleep(time.Second)
// 		continue
// 	}

// 	break
// }
