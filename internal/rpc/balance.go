package rpc

import (
	"io"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/proto/pb"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type balancerServer struct {
	balance <-chan model.Position
	pb.UnimplementedBalanceServer
}

func NewBalancerServer(balance <-chan model.Position) pb.BalanceServer {
	return &balancerServer{
		balance: balance,
	}
}

func (b *balancerServer) Balancer(stream pb.Balance_BalancerServer) error {
	for {
		profit := decimal.Zero
		position := <-b.balance

		if position.Buy {
			profit.Add(position.ClosePrice).Add(position.OpenPrice.Neg())
		} else {
			profit.Add(position.OpenPrice).Add(position.ClosePrice.Neg())
		}

		err := stream.Send(&pb.ResponseBalancer{
			Uuid: position.UserID.String(),
			Add: &pb.ResponseBalancerDecimal{
				Value: profit.CoefficientInt64(),
				Exp:   profit.Exponent(),
			},
		})

		if err == io.EOF {
			log.Info("Breaking stream becasuse of IOF error")
			break
		}
		if err != nil {
			log.Error("Error occured while sending data, exiting stream: ", err)
			return err
		}

		recv, err := stream.Recv()
		if err != nil {
			log.Error("Recive error, exiting stream: ", err)
			return err
		}

		if !recv.Ok {
			log.Info("Transaction denied, it seems you haven't got enough money to close position")
		}
	}

	return nil
}
