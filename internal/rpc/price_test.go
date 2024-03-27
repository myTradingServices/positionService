package rpc

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	pricePB "github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/proto/pb"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	priceTarget    string = "localhost:7073"
	positionTarget string = "localhost:7074"
)

var (
	priceCh   chan model.Price
	priceRecv Reciver
	posWr     pb.PositionClient

	testLastPrice1 = model.Price{
		Symbol: "symb1",
		Ask:    decimal.New(11, -1),
		Bid:    decimal.New(1, 0),
		Date:   time.Now(),
	}
	testLastPrice2 model.Price = model.Price{
		Symbol: "symb2",
		Ask:    decimal.New(11, -1),
		Bid:    decimal.New(1, 0),
		Date:   time.Now(),
	}
	testLastPrice3 model.Price = model.Price{
		Symbol: "symb3",
		Ask:    decimal.New(11, -1),
		Bid:    decimal.New(1, 0),
		Date:   time.Now(),
	}
)

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", priceTarget)
	if err != nil {
		log.Errorf("failed to listen port %v: %v", priceTarget, err)
	}

	rpcPriceServer := grpc.NewServer()
	priceServ := NewTestPriceConsumer()
	pricePB.RegisterConsumerServer(rpcPriceServer, priceServ)

	go func() {
		err = rpcPriceServer.Serve(lis)
		if err != nil {
			log.Error("rpc fatal error: Server can't start")
			return
		}
	}()

	option := grpc.WithTransportCredentials(insecure.NewCredentials())
	priceConn, err := grpc.Dial(priceTarget, option)
	if err != nil {
		log.Errorf("grpc connection error on %v: %v", priceTarget, err)
		return
	}
	defer priceConn.Close()

	priceCh = make(chan model.Price)
	priceRecv = NewPriceServer(priceConn, priceCh)

	posConn, err := grpc.Dial(positionTarget, option)
	if err != nil {
		log.Errorf("grpc connection error on %v: %v", positionTarget, err)
		return
	}
	defer priceConn.Close()
	posWr = pb.NewPositionClient(posConn)

	code := m.Run()

	os.Exit(code)
}

func TestReciveLast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	testTable := []struct {
		name     string
		input    string
		expected model.Price
	}{
		{
			name:     "standart input-1",
			input:    "symb1",
			expected: testLastPrice1,
		},
		{
			name:     "standart input-2",
			input:    "symb2",
			expected: testLastPrice2,
		},
		{
			name:     "standart input-3",
			input:    "symb3",
			expected: testLastPrice3,
		},
	}

	for _, test := range testTable {
		actual, err := priceRecv.ReciveLast(ctx, test.input)

		ok := assert.Nil(t, err, test.name)
		if !ok {
			continue
		}

		assert.Equal(t, test.expected.Ask, actual.Ask, test.name)
		assert.Equal(t, test.expected.Bid, actual.Bid, test.name)
		assert.Equal(t, test.expected.Symbol, actual.Symbol, test.name)
		if !test.expected.Date.Equal(actual.Date) {
			t.Error("dates are not equal\nexpected: ", test.expected.Date, "\nactual: ", actual.Date)
		}
	}

}

func TestResiveStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*61)
	defer cancel()

	go priceRecv.ReciveStream(ctx)

	for i := 0; i < 60; i++ {
		tmpPrice := <-priceCh
		assert.Equal(t, decimal.New(int64(i), 0), tmpPrice.Bid, "Bid is not equal")
		assert.Equal(t, decimal.New(int64(i+1), 0), tmpPrice.Ask, "Ask is not equal")
		assert.Equal(t, "symb"+strconv.Itoa(i), tmpPrice.Symbol, "Symbol is not equal")
	}
}
