package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/model"
)

type postgres struct {
	dbpool *pgxpool.Pool
}

type DBInterface interface {
	Add(ctx context.Context, position model.Position) error
	Deleete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
	Update(ctx context.Context, position model.Position) error
}

func NewPostgresRepository(conn *pgxpool.Pool) DBInterface {
	return &postgres{
		dbpool: conn,
	}
}

// NOTE: Add without close price
func (p *postgres) Add(ctx context.Context, position model.Position) error {
	_, err := p.dbpool.Exec(ctx, "INSERT INTO trading.positions (operation_id, user_id, symbol, open_price, long) VALUES ($1, $2, $3, $4, $5)",
		position.OperationID,
		position.UserID,
		position.Symbol,
		position.OpenPrice,
		position.Long,
	)
	return err
}

func (p *postgres) Deleete(ctx context.Context, operID uuid.UUID) error {
	_, err := p.dbpool.Exec(ctx, "DELETE FROM trading.positions WHERE operation_id = $1", operID)
	return err
}

func (p *postgres) Get(ctx context.Context, userID uuid.UUID) (res []model.Position, err error) {
	rows, err := p.dbpool.Query(ctx, "SELECT (operation_id, user_id, symbol, open_price, close_price, created_at, long) FROM trading.positions WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		tmpPos := model.Position{}

		if err := rows.Scan(
			&tmpPos,
		); err != nil {
			return nil, err
		}

		res = append(res, tmpPos)
	}

	return res, nil
}

func (p *postgres) Update(ctx context.Context, pos model.Position) error {
	_, err := p.dbpool.Exec(ctx, "UPDATE trading.positions SET close_price = $1 WHERE user_id = $2 AND symbol = $3", pos.ClosePrice, pos.UserID, pos.Symbol)
	return err
}
