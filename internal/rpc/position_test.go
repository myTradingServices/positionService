package rpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	mocks "github.com/mmfshirokan/positionService/internal/rpc/mock"
	"github.com/mmfshirokan/positionService/proto/pb"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

var (
	openPosInput1 = &pb.RequestOpenPosition{
		OperationID: uuid.New().String(),
		Long:        true,
		UserID:      uuid.New().String(),
		Symbol:      "symb1",
	}
	openPosInput2 = &pb.RequestOpenPosition{
		OperationID: uuid.New().String(),
		Long:        false,
		UserID:      uuid.New().String(),
		Symbol:      "symb2",
	}
	openPosInput3 = &pb.RequestOpenPosition{
		OperationID: uuid.New().String(),
		Long:        true,
		UserID:      uuid.New().String(),
		Symbol:      "symb3",
	}

	closePosInput1 = &pb.RequestClosePosition{
		UserID: uuid.New().String(),
		Long:   true,
		Symbol: "symb1",
	}
	closePosInput2 = &pb.RequestClosePosition{
		UserID: uuid.New().String(),
		Long:   false,
		Symbol: "symb2",
	}
	closePosInput3 = &pb.RequestClosePosition{
		UserID: uuid.New().String(),
		Long:   true,
		Symbol: "symb3",
	}
)

func TestOpenClosePosition(t *testing.T) {
	lis, err := net.Listen("tcp", positionTarget)
	if err != nil {
		t.Errorf("failed to listen port %v: %v", positionTarget, err)
	}

	dbPosM := mocks.NewPositionManipulator(t)
	lastPriceM := mocks.NewReciver(t)
	posServ := NewPositionServer(dbPosM, lastPriceM)

	rpcPosServer := grpc.NewServer()
	pb.RegisterPositionServer(rpcPosServer, posServ)

	go func() {
		err = rpcPosServer.Serve(lis)
		if err != nil {
			t.Error("rpc fatal error: Server can't start")
			return
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Open position:

	testOpenTable := []struct {
		name     string
		input    *pb.RequestOpenPosition
		expected model.Price
	}{
		{
			name:     "open input-1",
			input:    openPosInput1,
			expected: testLastPrice1,
		},
		{
			name:     "open input-2",
			input:    openPosInput2,
			expected: testLastPrice2,
		},
		{
			name:     "open input-3",
			input:    openPosInput3,
			expected: testLastPrice3,
		},
	}

	for _, test := range testOpenTable {

		bidOrAask := func(long bool) decimal.Decimal {
			if long {
				return test.expected.Bid
			}
			return test.expected.Ask
		}

		posMCall := dbPosM.EXPECT().Add(mock.Anything, model.Position{
			OperationID: uuid.MustParse(test.input.OperationID),
			UserID:      uuid.MustParse(test.input.UserID),
			Symbol:      test.input.Symbol,
			OpenPrice:   bidOrAask(test.input.Long),
			Long:        test.input.Long,
		}).Return(nil)

		prcRPCCall := lastPriceM.EXPECT().ReciveLast(mock.Anything, test.input.Symbol).Return(test.expected, nil)

		if _, err := posWr.OpenPosition(ctx, test.input); err != nil {
			t.Errorf("error occured on %v: %v", test.name, err)
		}

		dbPosM.AssertExpectations(t)
		lastPriceM.AssertExpectations(t)

		posMCall.Unset()
		prcRPCCall.Unset()
	}

	// Close position:

	testCloseTable := []struct {
		name     string
		input    *pb.RequestClosePosition
		expected model.Price
	}{
		{
			name:     "close input-1",
			input:    closePosInput1,
			expected: testLastPrice1,
		},
		{
			name:     "close input-2",
			input:    closePosInput2,
			expected: testLastPrice2,
		},
		{
			name:     "close input-3",
			input:    closePosInput3,
			expected: testLastPrice3,
		},
	}

	for _, test := range testCloseTable {
		bidOrAask := func(long bool) decimal.Decimal {
			if long {
				return test.expected.Ask
			}
			return test.expected.Bid
		}

		posMCall := dbPosM.EXPECT().Update(mock.Anything, model.Position{
			UserID:     uuid.MustParse(test.input.UserID),
			Symbol:     test.input.Symbol,
			ClosePrice: bidOrAask(test.input.Long),
		}).Return(nil)

		prcRPCCall := lastPriceM.EXPECT().ReciveLast(mock.Anything, test.input.Symbol).Return(test.expected, nil)

		if _, err := posWr.ClosePosition(ctx, test.input); err != nil {
			t.Errorf("error occured on %v: %v", test.name, err)
		}

		dbPosM.AssertExpectations(t)
		lastPriceM.AssertExpectations(t)
		posMCall.Unset()
		prcRPCCall.Unset()
	}

}
