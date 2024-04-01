package repository

import (
	"context"
	"encoding/json"

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
	dbpool *pgxpool.Pool
	ch     chan model.Position
}

func NewPgListen(ch chan model.Position, dbpool *pgxpool.Pool) Listener {
	return &postgresListen{
		dbpool: dbpool,
		ch:     ch,
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

		if nontification.Channel == "positionOpen" {
			tmpModel := struct {
				OperationID uuid.UUID
				OpenPrice   decimal.Decimal
				Long        bool
			}{}

			err = json.Unmarshal([]byte(nontification.Payload), &tmpModel)
			if err != nil {
				log.Error("Error unmarshalling notification:", err)
			}

			p.ch <- model.Position{
				OperationID: tmpModel.OperationID,
				OpenPrice:   tmpModel.OpenPrice,
				Long:        tmpModel.Long,
			}
			continue
		}

		tmpModel := struct {
			OperationID uuid.UUID
			OpenPrice   decimal.Decimal
		}{}

		err = json.Unmarshal([]byte(nontification.Payload), &tmpModel)
		if err != nil {
			log.Error("Error unmarshalling notification:", err)
		}

		p.ch <- model.Position{
			OperationID: tmpModel.OperationID,
			OpenPrice:   tmpModel.OpenPrice,
		}
	}

}
