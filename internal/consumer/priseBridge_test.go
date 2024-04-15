package consumer

import (
	"context"
	"testing"
	"time"

	mocks "github.com/mmfshirokan/positionService/internal/consumer/mock"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPriceBridge(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	ch := make(chan model.Price)
	pg := mocks.NewPriceGeter(t)
	pb := NewPriceBridge(ch, pg)

	type T struct {
		name       string
		inputPrice model.Price
		outputCHs  []chan model.Price
		outputOK   bool
	}
	testTable := []T{
		{
			name: "test case for non-existed symb",
			inputPrice: model.Price{
				Symbol: "symb1231",
				Bid:    decimal.New(13, -1),
				Ask:    decimal.New(13, 0),
				Date:   time.Now(),
			},
			outputCHs: nil,
			outputOK:  false,
		},
		{
			name: "test case for symb1",
			inputPrice: model.Price{
				Symbol: "symb1",
				Bid:    decimal.New(11, -1),
				Ask:    decimal.New(11, 0),
				Date:   time.Now(),
			},
			outputCHs: []chan model.Price{
				make(chan model.Price),
				make(chan model.Price),
			},
			outputOK: true,
		},
		{
			name: "test case for symb2",
			inputPrice: model.Price{
				Symbol: "symb2",
				Bid:    decimal.New(12, -1),
				Ask:    decimal.New(12, 0),
				Date:   time.Now(),
			},
			outputCHs: []chan model.Price{
				make(chan model.Price),
				make(chan model.Price),
				make(chan model.Price),
			},
			outputOK: true,
		},
	}

	go func(tt []T, writeToCh chan model.Price, mck *mocks.PriceGeter) {
		for _, test := range tt {
			mck.EXPECT().GetAllChanForSymb(test.inputPrice.Symbol).Return(test.outputCHs, test.outputOK)
			writeToCh <- test.inputPrice
		}
	}(testTable, ch, pg)

	go pb.PriceBridge(ctx)

	for _, test := range testTable {
		for _, userCH := range test.outputCHs {
			pr := <-userCH
			assert.Equal(t, test.inputPrice.Ask, pr.Ask)
			assert.Equal(t, test.inputPrice.Bid, pr.Bid)
			assert.Equal(t, test.inputPrice.Symbol, pr.Symbol)
			if test.inputPrice.Date.Compare(pr.Date) != 0 {
				t.Errorf("Time error have actual != expected: %v != %v", pr.Date, test.inputPrice.Date)
			}
		}
	}

	cancel()

	pg.AssertExpectations(t)
	log.Info("TestPositionBridge Finished!")
}
