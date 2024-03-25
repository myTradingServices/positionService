package rpc

import (
	"io"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
)

type positionServer struct {
	db    service.DBInterface
	price priceServer
	pb.UnimplementedPositionServer
}

type Positioner interface {
	Positioner(stream pb.Position_PositionerServer) error
}

func NewPositionServer(db service.DBInterface, price priceServer) pb.PositionServer {
	return &positionServer{
		db:    db,
		price: price,
	}
}

func (s *positionServer) Positioner(stream pb.Position_PositionerServer) error {
	ctx := stream.Context()
	operMap := make(map[string]struct{})

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

		if _, ok := operMap[recv.OperationID]; !ok {
			operMap[recv.OperationID] = struct{}{}

			price, err := s.price.ReciveLast(ctx, recv.Symbol)
			if err != nil {
				log.Error("GetLastPrice-rpc error while opening in position rpc, exiting stream: ", err)
				return err
			}

			if recv.Buy {
				err = s.db.Add(ctx, model.Position{
					OperationID: uuid.MustParse(recv.OperationID),
					UserID:      uuid.MustParse(recv.UserID),
					Symbol:      recv.Symbol,
					OpenPrice:   price.Bid,
					Buy:         recv.Buy,
					Open:        recv.Open,
				})
			} else {
				err = s.db.Add(ctx, model.Position{
					OperationID: uuid.MustParse(recv.OperationID),
					UserID:      uuid.MustParse(recv.UserID),
					Symbol:      recv.Symbol,
					OpenPrice:   price.Ask,
					Buy:         recv.Buy,
					Open:        recv.Open,
				})
			}

			if err != nil {
				log.Error("Postgres ADD error while opening position in position rpc: ", err)
				return err
			}
		}

		if !recv.Open {
			price, err := s.price.ReciveLast(ctx, recv.Symbol)
			if err != nil {
				log.Error("Get last price error while closing position in position rpc, exiting stream: ", err)
				return err
			}

			if recv.Buy {
				err = s.db.Update(ctx, model.Position{
					OperationID: uuid.MustParse(recv.OperationID),
					ClosePrice:  price.Bid,
					Open:        recv.Open,
				})
			} else {
				err = s.db.Update(ctx, model.Position{
					OperationID: uuid.MustParse(recv.OperationID),
					ClosePrice:  price.Ask,
					Open:        recv.Open,
				})
			}

			if err != nil {
				log.Error("Can't update position in position rpc: ", err)
				return err
			}
		}
	}

	return nil
}
