package rpc

import (
	"context"
	"io"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type PriceRPC struct {
	client  pb.ConsumerClient
	chPrice chan model.Price
}

func NewPriceServer(connecion *grpc.ClientConn, chPrice chan model.Price) *PriceRPC {
	client := pb.NewConsumerClient(connecion)
	return &PriceRPC{
		client:  client,
		chPrice: chPrice,
	}
}

func (p *PriceRPC) ReciveStream(ctx context.Context) {
	log.Info("ReciveStream (price) rpc started")
	defer log.Info("ReciveStream (price) rpc exited")

	stream, err := p.client.DataStream(ctx, &pb.RequestDataStream{Start: true})
	if err != nil {
		log.Errorf("Error in DataStream: %v", err)
		return
	}
	defer stream.CloseSend()
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Infof("Exitin stream, because error is %v", err)
			break
		}
		if err != nil {
			log.Errorf("Error occured: %v", err)
			return
		}

		//log.Info("Send price to ch: ", p.chPrice)
		p.chPrice <- model.Price{
			Date:   recv.Date.AsTime(),
			Bid:    decimal.New(recv.Bid.Value, recv.Bid.Exp),
			Ask:    decimal.New(recv.Ask.Value, recv.Ask.Exp),
			Symbol: recv.Symbol,
		}
		//log.Infof("Send price to ch: %v compleet", p.chPrice)
	}
}

func (p *PriceRPC) ReciveLast(ctx context.Context, symb string) (model.Price, error) {
	log.Info("ReciveLast (price) rpc started")
	defer log.Info("ReciveLast (price) rpc exited")

	recv, err := p.client.GetLastPrice(ctx, &pb.RequestGetLastPrice{
		Symbol: symb,
	})
	if err != nil {
		log.Error("Reciving error: ", err)
		return model.Price{}, err
	}

	return model.Price{
		Date:   recv.Data.Date.AsTime(),
		Bid:    decimal.New(recv.Data.Bid.Value, recv.Data.Bid.Exp),
		Ask:    decimal.New(recv.Data.Ask.Value, recv.Data.Ask.Exp),
		Symbol: recv.Data.Symbol,
	}, nil
}
