package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/shopspring/decimal"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPosition(conn *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool: conn,
	}
}

// Adds position to database withot close price that is to be added later, and without Time that is auto genrated.
func (p *Postgres) Add(ctx context.Context, position model.Position) error {
	_, err := p.pool.Exec(ctx, "INSERT INTO trading.positions (operation_id, user_id, symbol, open_price, long) VALUES ($1, $2, $3, $4, $5)",
		position.OperationID,
		position.UserID,
		position.Symbol,
		position.OpenPrice,
		position.Long,
	)
	return err
}

func (p *Postgres) Deleete(ctx context.Context, operationID uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM trading.positions WHERE operation_id = $1", operationID)
	return err
}

func (p *Postgres) Get(ctx context.Context, userID uuid.UUID) (userPositions []model.Position, err error) {
	rows, err := p.pool.Query(ctx, "SELECT (operation_id, user_id, symbol, open_price, close_price, created_at, long) FROM trading.positions WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		pst := model.Position{}

		if err := rows.Scan(
			&pst,
		); err != nil {
			return nil, err
		}

		userPositions = append(userPositions, pst)
	}

	return userPositions, nil
}

func (p *Postgres) GetAllOpened(ctx context.Context) ([]model.Position, error) {
	rows, err := p.pool.Query(ctx, "SELECT (user_id, symbol, open_price, long) FROM trading.positions WHERE close_price IS NULL")
	if err != nil {
		return nil, err
	}

	var res []model.Position
	for rows.Next() {
		tmpPos := struct {
			UserID    uuid.UUID
			Symbol    string
			OpenPrice decimal.Decimal
			Buy       bool
		}{}

		if err := rows.Scan(
			&tmpPos,
		); err != nil {
			return nil, err
		}

		res = append(res, model.Position{
			UserID:    tmpPos.UserID,
			Symbol:    tmpPos.Symbol,
			OpenPrice: tmpPos.OpenPrice,
			Long:      tmpPos.Buy,
		})
	}

	return res, nil
}

// Updates position with close price, using UseerID and Symbol for search.
func (p *Postgres) Update(ctx context.Context, pos model.Position) error {
	_, err := p.pool.Exec(ctx, "UPDATE trading.positions SET close_price = $1 WHERE user_id = $2 AND symbol = $3", pos.ClosePrice, pos.UserID, pos.Symbol)
	return err
}
