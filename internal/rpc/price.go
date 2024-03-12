package rpc

import (
	"context"
	"io"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type priceServer struct {
	conn    *grpc.ClientConn
	symbols map[string]chan model.Price
}

type Reciver interface {
	Recive(ctx context.Context)
}

func NewPriceServer(connecion *grpc.ClientConn, symbolMap map[string]chan model.Price) Reciver {
	return &priceServer{
		conn:    connecion,
		symbols: symbolMap,
	}
}

func (p *priceServer) Recive(ctx context.Context) {
	consumer := pb.NewConsumerClient(p.conn)
	stream, err := consumer.DataStream(ctx, &pb.RequestDataStream{Start: true})
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

		ch, ok := p.symbols[recv.Symbol]
		if !ok {
			ch = make(chan model.Price)
			p.symbols[recv.Symbol] = ch
		}

		ch <- model.Price{
			Date:   time.Now(),
			Bid:    decimal.New(recv.Bid.Value, recv.Bid.Exp),
			Ask:    decimal.New(recv.Ask.Value, recv.Ask.Exp),
			Symbol: recv.Symbol,
		}
	}
}
