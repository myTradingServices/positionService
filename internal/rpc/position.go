package rpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type positionServer struct {
	db    service.DBInterface
	price Reciver
	pb.UnimplementedPositionServer
}

type Positioner interface {
	ClosePosition(ctx context.Context, recv *pb.RequestClosePosition) (*emptypb.Empty, error)
	OpenPosition(ctx context.Context, recv *pb.RequestOpenPosition) (*emptypb.Empty, error)
}

func NewPositionServer(db service.DBInterface, price Reciver) pb.PositionServer {
	return &positionServer{
		db:    db,
		price: price,
	}
}

func (p *positionServer) OpenPosition(ctx context.Context, recv *pb.RequestOpenPosition) (*emptypb.Empty, error) {

	price, err := p.price.ReciveLast(ctx, recv.Symbol)
	if err != nil {
		log.Error("GetLastPrice-rpc error while opening in position rpc, exiting stream: ", err)
		return &emptypb.Empty{}, err
	}

	if recv.Buy {
		err = p.db.Add(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			UserID:      uuid.MustParse(recv.UserID),
			Symbol:      recv.Symbol,
			OpenPrice:   price.Bid,
			Buy:         recv.Buy,
			Open:        true,
		})
	} else {
		err = p.db.Add(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			UserID:      uuid.MustParse(recv.UserID),
			Symbol:      recv.Symbol,
			OpenPrice:   price.Ask,
			Buy:         recv.Buy,
			Open:        true,
		})
	}

	if err != nil {
		log.Error("Postgres ADD error while opening position in position rpc: ", err)
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (p *positionServer) ClosePosition(ctx context.Context, recv *pb.RequestClosePosition) (*emptypb.Empty, error) {
	price, err := p.price.ReciveLast(ctx, recv.Symbol)
	if err != nil {
		log.Error("GetLastPrice-rpc error while opening in position rpc, exiting stream: ", err)
		return &emptypb.Empty{}, err
	}

	if recv.Buy {
		err = p.db.Update(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			ClosePrice:  price.Ask,
			Open:        false,
		})
	} else {
		err = p.db.Update(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			ClosePrice:  price.Bid,
			Open:        false,
		})
	}

	if err != nil {
		log.Error("Can't update position in position rpc: ", err)
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
