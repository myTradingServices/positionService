package rpc

import (
	"context"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testPrServer struct {
	pb.UnimplementedConsumerServer
}

func NewTestPriceConsumer() pb.ConsumerServer {
	return &testPrServer{}
}

func (t *testPrServer) DataStream(req *pb.RequestDataStream, stream pb.Consumer_DataStreamServer) error {
	if !req.Start {
		return errors.New("start is not initiated")
	}
	for i := 0; i < 60; i++ {
		err := stream.Send(&pb.ResponseDataStream{
			Date: timestamppb.New(time.Now()),
			Bid: &pb.ResponseDataStreamDecimal{
				Value: int64(i),
				Exp:   0,
			},
			Ask: &pb.ResponseDataStreamDecimal{
				Value: int64(i + 1),
				Exp:   0,
			},
			Symbol: "symb" + strconv.Itoa(i),
		})
		if err == io.EOF {
			log.Infof("Stream exited, because error is: %v", err)
			break
		}
		if err != nil {
			log.Errorf("Error sending message: %v.", err)
		}

		log.Infof("Sent %v", i) // delete
		time.Sleep(time.Second)
	}

	return nil
}

func (t *testPrServer) GetLastPrice(ctx context.Context, req *pb.RequestGetLastPrice) (*pb.ResponseGetLastPrice, error) {
	symb := req.Symbol
	switch symb {
	case "symb1":
		return &pb.ResponseGetLastPrice{
			Data: &pb.ResponseDataStream{
				Symbol: "symb1",
				Bid: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice1.Bid.CoefficientInt64(),
					Exp:   testLastPrice1.Bid.Exponent(),
				},
				Ask: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice1.Ask.CoefficientInt64(),
					Exp:   testLastPrice1.Ask.Exponent(),
				},
				Date: timestamppb.New(testLastPrice1.Date),
			},
		}, nil

	case "symb2":
		return &pb.ResponseGetLastPrice{
			Data: &pb.ResponseDataStream{
				Symbol: "symb2",
				Bid: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice2.Bid.CoefficientInt64(),
					Exp:   testLastPrice2.Bid.Exponent(),
				},
				Ask: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice2.Ask.CoefficientInt64(),
					Exp:   testLastPrice2.Ask.Exponent(),
				},
				Date: timestamppb.New(testLastPrice2.Date),
			},
		}, nil

	case "symb3":
		return &pb.ResponseGetLastPrice{
			Data: &pb.ResponseDataStream{
				Symbol: "symb3",
				Bid: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice3.Bid.CoefficientInt64(),
					Exp:   testLastPrice3.Bid.Exponent(),
				},
				Ask: &pb.ResponseDataStreamDecimal{
					Value: testLastPrice3.Ask.CoefficientInt64(),
					Exp:   testLastPrice3.Ask.Exponent(),
				},
				Date: timestamppb.New(testLastPrice3.Date),
			},
		}, nil

	default:
		return nil, errors.New("invalid symbol")
	}
}
