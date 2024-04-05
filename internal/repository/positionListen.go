package repository

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Listener interface {
	Listen(ctx context.Context)
}

type postgresListen struct {
	dbpool  *pgxpool.Pool
	openCh  chan model.Position
	closeCh chan model.Position
}

func NewPgListen(openCh chan model.Position, closeCh chan model.Position, dbpool *pgxpool.Pool) Listener {
	return &postgresListen{
		dbpool:  dbpool,
		openCh:  openCh,
		closeCh: closeCh,
	}
}

func (p *postgresListen) Listen(ctx context.Context) {
	conn, err := p.dbpool.Acquire(ctx)
	if err != nil {
		log.Error("Error listening to chat channel:", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN positionOpen")
	if err != nil {
		log.Error("Error listening to positionOpen channel:", err)
		return
	}

	_, err = conn.Exec(ctx, "LISTEN positionClose")
	if err != nil {
		log.Error("Error listening to positionClose channel:", err)
		return
	}

	for {
		nontification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			log.Error("Error waiting for notification:", err)
			return
		}

		tmpModel := struct {
			Symbol     string    `json:"symbol"`
			UserID     uuid.UUID `json:"user_id"`
			OpenPrice  string    `json:"open_price"`
			ClosePrice string    `json:"close_price"`
			Long       string    `json:"long"`
		}{}

		err = json.Unmarshal([]byte(nontification.Payload), &tmpModel)
		if err != nil {
			log.Error("Error unmarshalling notification:", err)
			return
		}

		if nontification.Channel == "positionopen" {
			tmpOpPrice, err := decimal.NewFromString(tmpModel.OpenPrice)
			if err != nil {
				log.Error("Parsing string into decimal error:", err)
				return
			}

			tmpLong, err := strconv.ParseBool(tmpModel.Long)
			if err != nil {
				log.Error("Parsing string into bool error:", err)
				return
			}

			p.openCh <- model.Position{
				Symbol:    tmpModel.Symbol,
				UserID:    tmpModel.UserID,
				OpenPrice: tmpOpPrice,
				Long:      tmpLong,
			}
			continue
		}

		tmpClPrice, err := decimal.NewFromString(tmpModel.ClosePrice)
		if err != nil {
			log.Error("Parsing string into decimal error:", err)
			return
		}

		p.closeCh <- model.Position{
			Symbol:     tmpModel.Symbol,
			UserID:     tmpModel.UserID,
			ClosePrice: tmpClPrice,
		}
	}

}
