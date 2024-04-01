package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Position struct {
	OperationID uuid.UUID
	UserID      uuid.UUID
	Symbol      string
	OpenPrice   decimal.Decimal
	ClosePrice  decimal.Decimal
	CreatedAt   time.Time
	Long        bool
}
