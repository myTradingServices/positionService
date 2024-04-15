package repository

import (
	"testing"
	"time"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	openCh  chan model.Position
	closeCh chan model.Position

	numberOfDbTests = 3
)

func TestListen(t *testing.T) {
	calls := 0
	res := []model.Position{}

loop:
	for {
		select {
		case t := <-openCh:
			{
				calls++
				log.Info(t)
				res = append(res, t)
			}
		case t := <-closeCh:
			{
				calls++
				log.Info(t)
				res = append(res, t)
			}
		default:
			time.Sleep(time.Millisecond * 10)
			log.Info("Sleeping")
			if calls >= numberOfDbTests*2 {
				break loop
			}
		}
	}

	assert.ElementsMatch(t, res,
		[]model.Position{
			{
				UserID:    input1.UserID,
				Symbol:    input1.Symbol,
				OpenPrice: input1.OpenPrice,
				Long:      input1.Long,
			},
			{
				UserID:    input2.UserID,
				Symbol:    input2.Symbol,
				OpenPrice: input2.OpenPrice,
				Long:      input2.Long,
			},
			{
				UserID:    input3.UserID,
				Symbol:    input3.Symbol,
				OpenPrice: input3.OpenPrice,
				Long:      input3.Long,
			},
			{
				UserID:     input1.UserID,
				Symbol:     input1.Symbol,
				ClosePrice: closePrice1,
			},
			{
				UserID:     input2.UserID,
				Symbol:     input2.Symbol,
				ClosePrice: closePrice2,
			},
			{
				UserID:     input3.UserID,
				Symbol:     input3.Symbol,
				ClosePrice: closePrice3,
			},
		},
	)
}
