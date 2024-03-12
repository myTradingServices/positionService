package rpc

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
)

type positionServer struct {
	symbols      map[string]chan model.Price
	positionOpen map[string]chan bool

	mut  sync.RWMutex
	serv service.Interface

	pb.UnimplementedPositionServer
}

type Positioner interface {
	Positioner(pb.Position_PositionerServer) error
}

func NewPositionServer(symb map[string]chan model.Price, service service.Interface) pb.PositionServer {
	return &positionServer{
		symbols: symb,
		serv:    service,
	}
}

func (s *positionServer) Positioner(stream pb.Position_PositionerServer) error {
	ctx := stream.Context()

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

		openCh, ok := s.positionOpen[recv.OperationID]
		if !ok {
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

			go func(ctxx context.Context, tmpMod model.Position, symb string, open bool, buy bool) { // TODO add parralel symbo receiving
				reconectionCounter := 0
				for {
					s.mut.RLock()
					_, ok := s.symbols[symb]
					s.mut.RUnlock()

					if reconectionCounter > 14 {
						log.Error("ERROR failed to find symbol exiting gorutine")
						return
					}

					if !ok {
						log.Error("error: symbol not found, atempting to refind in 1 second")
						reconectionCounter++
						time.Sleep(time.Second)
						continue
					}

					break
				}

				s.mut.RLock()
				priceChan := s.symbols[symb]
				s.mut.RUnlock()

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
						tmpMod.CLosePrice = tmpPrice.Bid
					} else {
						tmpMod.CLosePrice = tmpPrice.Ask
					}

					break
				}

				s.mut.Lock()
				s.serv.Add(ctxx, tmpMod)
				s.mut.Unlock()
				// TODO add balance relation
			}(ctx, mod, recv.Symbol, recv.Open, recv.Buy)

			continue
		}

		openCh <- recv.Open
	}

	return nil
}
