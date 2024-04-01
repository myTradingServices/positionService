package consumer

// Note: current test stgae is invalid

import (
	//"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	mocks "github.com/mmfshirokan/positionService/internal/rpc/mock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

func TestConsumePrice(t *testing.T) {
	waitTime := time.Second*2 + 10*time.Millisecond
	dbMock := mocks.NewDBInterface(t)
	mapMock := mocks.NewMapInterface(t)
	//posCons := NewPositionConsumer(dbMock, mapMock)

	type T struct {
		name         string
		posForMapAdd []model.Position
	}

	testTable := []T{
		{
			name: "Positions obtained with GetAllOpened",
			posForMapAdd: []model.Position{
				{
					OperationID: uuid.New(),
					Symbol:      "symb1",
					OpenPrice:   decimal.New(10, 0),
					Long:        true,
				},
				{
					OperationID: uuid.New(),
					Symbol:      "symb2",
					OpenPrice:   decimal.New(9, 0),
					Long:        false,
				},
				{
					OperationID: uuid.New(),
					Symbol:      "symb3",
					OpenPrice:   decimal.New(8, 0),
					Long:        false,
				},
			},
		},
		{
			name: "Positions obtained with GetLaterThen-1",
			posForMapAdd: []model.Position{
				{
					OperationID: uuid.New(),
					Symbol:      "symb1",
					OpenPrice:   decimal.New(7, 0),
					Long:        true,
				},
				{
					OperationID: uuid.New(),
					Symbol:      "symb2",
					OpenPrice:   decimal.New(6, 0),
					Long:        false,
				},
				{
					OperationID: uuid.New(),
					Symbol:      "symb3",
					OpenPrice:   decimal.New(5, 0),
					Long:        false,
				},
			},
		},
	}

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	for i, test := range testTable {
		if i == 0 {
			dbMock.EXPECT().GetAllOpend(mock.Anything).Return(test.posForMapAdd, nil)
		} else {
			dbMock.EXPECT().GetLaterThen(mock.Anything, mock.Anything).Return(test.posForMapAdd, nil)
		}

		for _, pos := range test.posForMapAdd {
			mapMock.EXPECT().Add(model.SymbOperDTO{Symbol: pos.Symbol, UserID: pos.OperationID.String()}, mock.Anything).Return(nil)
		}
	}
	//go posCons.ConsumePrice(ctx)

	time.Sleep(waitTime * time.Duration(len(testTable)))

	//cancel()

	dbMock.AssertExpectations(t)
	mapMock.AssertExpectations(t)
}
