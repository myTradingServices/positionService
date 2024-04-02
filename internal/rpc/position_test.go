package rpc

// import (
// 	"context"
// 	"net"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/mmfshirokan/positionService/internal/model"
// 	mocks "github.com/mmfshirokan/positionService/internal/rpc/mock"
// 	"github.com/mmfshirokan/positionService/proto/pb"
// 	"github.com/shopspring/decimal"
// 	"github.com/stretchr/testify/mock"
// 	"google.golang.org/grpc"
// )

// var (
// 	openPosInput1 = &pb.RequestOpenPosition{
// 		OperationID: uuid.New().String(),
// 		Long:        true,
// 		UserID:      uuid.New().String(),
// 		Symbol:      "symb1",
// 	}
// 	openPosInput2 = &pb.RequestOpenPosition{
// 		OperationID: uuid.New().String(),
// 		Long:        false,
// 		UserID:      uuid.New().String(),
// 		Symbol:      "symb2",
// 	}
// 	openPosInput3 = &pb.RequestOpenPosition{
// 		OperationID: uuid.New().String(),
// 		Long:        true,
// 		UserID:      uuid.New().String(),
// 		Symbol:      "symb3",
// 	}

// 	closePosInput1 = &pb.RequestClosePosition{
// 		OperationID: uuid.New().String(),
// 		Long:        true,
// 		Symbol:      "symb1",
// 	}
// 	closePosInput2 = &pb.RequestClosePosition{
// 		OperationID: uuid.New().String(),
// 		Long:        false,
// 		Symbol:      "symb2",
// 	}
// 	closePosInput3 = &pb.RequestClosePosition{
// 		OperationID: uuid.New().String(),
// 		Long:        true,
// 		Symbol:      "symb3",
// 	}
// )

// func TestOpenClosePosition(t *testing.T) {
// 	lis, err := net.Listen("tcp", positionTarget)
// 	if err != nil {
// 		t.Errorf("failed to listen port %v: %v", positionTarget, err)
// 	}

// 	dbMocks := mocks.NewDBInterface(t)
// 	posServ := NewPositionServer(dbMocks, priceRecv)

// 	rpcPosServer := grpc.NewServer()
// 	pb.RegisterPositionServer(rpcPosServer, posServ)

// 	go func() {
// 		err = rpcPosServer.Serve(lis)
// 		if err != nil {
// 			t.Error("rpc fatal error: Server can't start")
// 			return
// 		}
// 	}()

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	testOpenTable := []struct {
// 		name     string
// 		input    *pb.RequestOpenPosition
// 		lpOutput model.Price
// 	}{
// 		{
// 			name:     "open input-1",
// 			input:    openPosInput1,
// 			lpOutput: testLastPrice1,
// 		},
// 		{
// 			name:     "open input-2",
// 			input:    openPosInput2,
// 			lpOutput: testLastPrice2,
// 		},
// 		{
// 			name:     "open input-3",
// 			input:    openPosInput3,
// 			lpOutput: testLastPrice3,
// 		},
// 	}

// 	for _, test := range testOpenTable {

// 		bidOrAask := func(buy bool) decimal.Decimal {
// 			if buy {
// 				return test.lpOutput.Bid
// 			}
// 			return test.lpOutput.Ask
// 		}

// 		mCall := dbMocks.EXPECT().Add(mock.Anything, model.Position{
// 			OperationID: uuid.MustParse(test.input.OperationID),
// 			UserID:      uuid.MustParse(test.input.UserID),
// 			Symbol:      test.input.Symbol,
// 			OpenPrice:   bidOrAask(test.input.Long),
// 			Long:        test.input.Long,
// 		}).Return(nil)

// 		if _, err := posWr.OpenPosition(ctx, test.input); err != nil {
// 			t.Errorf("error occured on %v: %v", test.name, err)
// 		}

// 		dbMocks.AssertExpectations(t)
// 		mCall.Unset()
// 	}

// 	testCloseTable := []struct {
// 		name     string
// 		input    *pb.RequestClosePosition
// 		lpOutput model.Price
// 	}{
// 		{
// 			name:     "close input-1",
// 			input:    closePosInput1,
// 			lpOutput: testLastPrice1,
// 		},
// 		{
// 			name:     "close input-2",
// 			input:    closePosInput2,
// 			lpOutput: testLastPrice2,
// 		},
// 		{
// 			name:     "close input-3",
// 			input:    closePosInput3,
// 			lpOutput: testLastPrice3,
// 		},
// 	}

// 	for _, test := range testCloseTable {

// 		bidOrAask := func(buy bool) decimal.Decimal {
// 			if buy {
// 				return test.lpOutput.Ask
// 			}
// 			return test.lpOutput.Bid
// 		}

// 		mCall := dbMocks.EXPECT().Update(mock.Anything, model.Position{
// 			OperationID: uuid.MustParse(test.input.OperationID),
// 			ClosePrice:  bidOrAask(test.input.Long),
// 		}).Return(nil)

// 		if _, err := posWr.ClosePosition(ctx, test.input); err != nil {
// 			t.Errorf("error occured on %v: %v", test.name, err)
// 		}

// 		dbMocks.AssertExpectations(t)
// 		mCall.Unset()
// 	}

// }

// func ClosePosition(t *testing.T) {
// 	lis, err := net.Listen("tcp", positionTarget)
// 	if err != nil {
// 		t.Errorf("failed to listen port %v: %v", positionTarget, err)
// 	}

// 	dbMocks := mocks.NewDBInterface(t)
// 	posServ := NewPositionServer(dbMocks, priceRecv)

// 	rpcPosServer := grpc.NewServer()
// 	pb.RegisterPositionServer(rpcPosServer, posServ)

// 	go func() {
// 		err = rpcPosServer.Serve(lis)
// 		if err != nil {
// 			t.Error("rpc fatal error: Server can't start")
// 			return
// 		}
// 	}()

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	type T struct {
// 		name     string
// 		input    *pb.RequestOpenPosition
// 		lpOutput model.Price
// 	}

// 	testTable := []T{
// 		{
// 			name:     "standart input-1",
// 			input:    openPosInput1,
// 			lpOutput: testLastPrice1,
// 		},
// 		{
// 			name:     "standart input-2",
// 			input:    openPosInput2,
// 			lpOutput: testLastPrice2,
// 		},
// 		{
// 			name:     "standart input-3",
// 			input:    openPosInput3,
// 			lpOutput: testLastPrice3,
// 		},
// 	}

// 	for _, test := range testTable {

// 		bidOrAask := func(buy bool) decimal.Decimal {
// 			if buy {
// 				return test.lpOutput.Bid
// 			}
// 			return test.lpOutput.Ask
// 		}

// 		mCall := dbMocks.EXPECT().Add(mock.Anything, model.Position{
// 			OperationID: uuid.MustParse(test.input.OperationID),
// 			UserID:      uuid.MustParse(test.input.UserID),
// 			Symbol:      test.input.Symbol,
// 			OpenPrice:   bidOrAask(test.input.Long),
// 			Long:        test.input.Long,
// 		}).Return(nil)

// 		if _, err := posWr.OpenPosition(ctx, test.input); err != nil {
// 			t.Errorf("error occured on %v: %v", test.name, err)
// 		}

// 		dbMocks.AssertExpectations(t)
// 		mCall.Unset()
// 	}
// }
