package rpc

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type priceServer struct {
	conn    *grpc.ClientConn
	symbols service.MapInterface[string]
	mut     sync.RWMutex
}

type Reciver interface {
	Recive(ctx context.Context)
}

func NewPriceServer(connecion *grpc.ClientConn, symbolMap service.MapInterface[string]) Reciver {
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

		ok := p.symbols.Contains(recv.Symbol)
		var ch chan model.Price

		if !ok {
			ch := make(chan model.Price)
			p.symbols.Add(recv.Symbol, ch)
		} else {
			ch, _ = p.symbols.Get(recv.Symbol)
		}

		ch <- model.Price{
			Date:   time.Now(),
			Bid:    decimal.New(recv.Bid.Value, recv.Bid.Exp),
			Ask:    decimal.New(recv.Ask.Value, recv.Ask.Exp),
			Symbol: recv.Symbol,
		}
	}
}
