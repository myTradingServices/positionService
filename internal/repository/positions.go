package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/model"
)

type postgres struct {
	conn *pgxpool.Pool
}

type DbInterface interface {
	Add(ctx context.Context, position model.Position) error
	Deleete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) ([]model.Position, error)
}

func NewPostgresRepository(conn *pgxpool.Pool) DbInterface {
	return &postgres{
		conn: conn,
	}
}

func (p *postgres) Add(ctx context.Context, position model.Position) error {
	_, err := p.conn.Exec(ctx, "INSERT INTO trading.positions (operation_id, user_id, symbol, open_price, close_price, buy) VALUES ($1, $2, $3, $4, $5, $6)",
		position.OperationID,
		position.UserID, position.Symbol,
		position.OpenPrice,
		position.ClosePrice,
		position.Buy,
	)
	return err
}

func (p *postgres) Deleete(ctx context.Context, id uuid.UUID) error {
	_, err := p.conn.Exec(ctx, "DELETE FROM trading.positions WHERE operation_id = $1", id)
	return err
}

func (p *postgres) Get(ctx context.Context, id uuid.UUID) (res []model.Position, err error) {
	rows, err := p.conn.Query(ctx, "SELECT (symbol, operation_id, open_price, close_price) FROM trading.positions WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		tempPosition := model.Position{UserID: id}
		if err := rows.Scan(
			&tempPosition.Symbol,
			&tempPosition.OperationID,
			&tempPosition.OpenPrice,
			&tempPosition.ClosePrice,
			&tempPosition.Buy,
		); err != nil {
			return nil, err
		}

		res = append(res, tempPosition)
	}

	return res, nil
}
