package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	mocks "github.com/mmfshirokan/positionService/internal/consumer/mock"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

var (
	sameUUID  = uuid.New()
	allOpened = []model.Position{
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb1",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      sameUUID,
			Symbol:      "symb2",
			OpenPrice:   decimal.New(14, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb3",
			OpenPrice:   decimal.New(14, 0),
			Long:        false,
		},
	}

	additionalTestModels = []model.Position{
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb4",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      sameUUID,
			Symbol:      "symb5",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb6",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb7",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
	}
)

func TestOpen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20000)
	defer cancel()
	lPos := service.NewLocalPositions(
		repository.NewLocalPosition(
			make(map[string]chan model.Position),
		),
	)

	prsM := make(map[string]map[string]chan model.Price)
	prs := service.NewPrices(
		repository.NewPrices(
			prsM,
		),
	)
	pcMock := mocks.NewPositionController(t)
	lisCh := make(chan model.Position)

	consumer := NewOpener(lPos, prs, pcMock, lisCh)

	pcMock.EXPECT().GetAllOpened(mock.Anything).Return(allOpened, nil)
	pcMock.EXPECT().Update(mock.Anything, model.Position{
		UserID:     allOpened[2].UserID,
		Symbol:     allOpened[2].Symbol,
		ClosePrice: decimal.New(15, 0),
	}).Return(nil)

	go consumer.Open(ctx)

	for _, val := range additionalTestModels {
		lisCh <- val
	}

	for symb, val := range prsM {
		for _, prsCh := range val {
			prsCh <- model.Price{
				Bid:    decimal.New(15, 0),
				Ask:    decimal.New(15, 0),
				Date:   time.Now(),
				Symbol: symb,
			}
		}
	}

	time.Sleep(time.Second)
	cancel()
	pcMock.AssertExpectations(t)
}

func TestMockOpen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	pcMock := mocks.NewPositionController(t)
	lpcMock := mocks.NewLPositionController(t)
	pmMock := mocks.NewPriceManipulator(t)
	lisCh := make(chan model.Position)

	consumer := NewOpener(lpcMock, pmMock, pcMock, lisCh)

	allOpened := []model.Position{
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb1",
			OpenPrice:   decimal.New(13, 0),
			Long:        true,
		},
		{
			OperationID: uuid.New(),
			UserID:      uuid.New(),
			Symbol:      "symb2",
			OpenPrice:   decimal.New(14, 0),
			Long:        false,
		},
	}

	pcMock.EXPECT().GetAllOpened(mock.Anything).Return(allOpened, nil)

	for _, p := range allOpened {
		lpcMock.EXPECT().Add(p.UserID.String(), mock.Anything)

		pmMock.EXPECT().Add(model.SymbOperDTO{
			Symbol: p.Symbol,
			UserID: p.UserID.String(),
		}, mock.Anything)
	}

	go consumer.Open(ctx)

	for i := range additionalTestModels {
		lpcMock.EXPECT().Get(mock.Anything).Return(nil, false)
		lpcMock.EXPECT().Add(additionalTestModels[i].UserID.String(), mock.Anything)
		pmMock.EXPECT().Add(model.SymbOperDTO{
			Symbol: additionalTestModels[i].Symbol,
			UserID: additionalTestModels[i].UserID.String(),
		}, mock.Anything)

		lisCh <- additionalTestModels[i]
	}
	cancel()

	pcMock.AssertExpectations(t)
	lpcMock.AssertExpectations(t)
	pmMock.AssertExpectations(t)
}
