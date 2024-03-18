package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Position struct {
	OperationID uuid.UUID
	UserID      uuid.UUID
	Symbol      string
	OpenPrice   decimal.Decimal
	ClosePrice  decimal.Decimal
	Buy         bool
}
