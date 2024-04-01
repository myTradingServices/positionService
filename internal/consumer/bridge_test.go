package consumer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBridge(t *testing.T) {
	ch := make(chan model.Price)

	mp := make(map[string]map[string]chan model.Price)
	mpRepo := repository.NewSymbOperMap(mp)
	mpServ := service.NewSymbOperMap(mpRepo)

	br := NewBridge(ch, mpServ)

	type T struct {
		name     string
		mapKey   model.SymbOperDTO
		mapValue chan model.Price
	}
	testTable := []T{
		{
			name: "symb1; oper1",
			mapKey: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
		{
			name: "symb1; oper2",
			mapKey: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
		{
			name: "symb1; oper3",
			mapKey: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
		{
			name: "symb2; oper4",
			mapKey: model.SymbOperDTO{
				Symbol: "symb2",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
		{
			name: "symb2; oper5",
			mapKey: model.SymbOperDTO{
				Symbol: "symb2",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
		{
			name: "symb3; oper6",
			mapKey: model.SymbOperDTO{
				Symbol: "symb3",
				UserID: uuid.NewString(),
			},
			mapValue: make(chan model.Price),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(testT []T) {
		defer wg.Done()
		for _, test := range testT {
			mpServ.Add(test.mapKey, test.mapValue)
			log.Info("val add", test.name)
		}
	}(testTable)

	go func() {
		for i := 0; i < 60; i++ {
			ch <- priceGeneration(i)
			log.Info("modelPrice send:", i)
		}
	}()

	wg.Wait()
	go br.Bridge(ctx)

	for _, testCase := range testTable {
		wg.Add(1)
		go func(test T) {
			defer wg.Done()
			for i := 0; i < 60; i = i + 3 {
				select {
				case tmpPrice := <-test.mapValue:
					switch tmpPrice.Symbol {
					case "symb1":
						assert.Equal(t, priceGeneration(i).Bid, tmpPrice.Bid, "Bid is not equal")
						assert.Equal(t, priceGeneration(i).Ask, tmpPrice.Ask, "Ask is not equal")
						assert.Equal(t, priceGeneration(i).Symbol, "symb1", "Symbol is not equal")
					case "symb2":
						assert.Equal(t, priceGeneration(i+1).Bid, tmpPrice.Bid, "Bid is not equal")
						assert.Equal(t, priceGeneration(i+1).Ask, tmpPrice.Ask, "Ask is not equal")
						assert.Equal(t, priceGeneration(i+1).Symbol, "symb2", "Symbol is not equal")
					case "symb3":
						assert.Equal(t, priceGeneration(i+2).Bid, tmpPrice.Bid, "Bid is not equal")
						assert.Equal(t, priceGeneration(i+2).Ask, tmpPrice.Ask, "Ask is not equal")
						assert.Equal(t, priceGeneration(i+2).Symbol, "symb3", "Symbol is not equal")
					default:
						t.Error("Unexpected symbol")
					}
				default:
				}
			}
			log.Info("Exit")
		}(testCase)
	}
	wg.Wait()
}

func priceGeneration(i int) model.Price {
	switch i % 3 {
	case 0:
		return model.Price{
			Symbol: "symb1",
			Bid:    decimal.New(int64(i+1), 0),
			Ask:    decimal.New(int64((i+1)*10), -1),
			Date:   time.Now(),
		}
	case 1:
		return model.Price{
			Symbol: "symb2",
			Bid:    decimal.New(int64(i+2), 0),
			Ask:    decimal.New(int64((i+2)*10), -1),
			Date:   time.Now(),
		}
	case 2:
		return model.Price{
			Symbol: "symb3",
			Bid:    decimal.New(int64(i+3), 0),
			Ask:    decimal.New(int64((i+3)*10), -1),
			Date:   time.Now(),
		}
	default:
		panic("Our world is going to end soon, remainder of the division by 3 is not equal to 0, 1 or 2")
	}
}
