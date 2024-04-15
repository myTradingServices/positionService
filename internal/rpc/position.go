package rpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type positionServer struct {
	db    PositionManipulator
	price Reciver
	pb.UnimplementedPositionServer
}

type PositionManipulator interface {
	Add(ctx context.Context, position model.Position) error
	Update(ctx context.Context, position model.Position) error
}

type Reciver interface {
	ReciveStream(ctx context.Context)
	ReciveLast(ctx context.Context, symb string) (model.Price, error)
}

func NewPositionServer(db PositionManipulator, price Reciver) pb.PositionServer {
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

	if recv.Long {
		err = p.db.Add(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			UserID:      uuid.MustParse(recv.UserID),
			Symbol:      recv.Symbol,
			OpenPrice:   price.Bid,
			Long:        recv.Long,
		})
	} else {
		err = p.db.Add(ctx, model.Position{
			OperationID: uuid.MustParse(recv.OperationID),
			UserID:      uuid.MustParse(recv.UserID),
			Symbol:      recv.Symbol,
			OpenPrice:   price.Ask,
			Long:        recv.Long,
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

	if recv.Long {
		err = p.db.Update(ctx, model.Position{
			UserID:     uuid.MustParse(recv.UserID),
			Symbol:     recv.Symbol,
			ClosePrice: price.Ask,
		})
	} else {
		err = p.db.Update(ctx, model.Position{
			UserID:     uuid.MustParse(recv.UserID),
			Symbol:     recv.Symbol,
			ClosePrice: price.Bid,
		})
	}

	if err != nil {
		log.Error("Can't update position in position rpc: ", err)
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
