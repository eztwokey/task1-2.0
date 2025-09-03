package repo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"wb-order-service/internal/model"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	UpsertOrder(ctx context.Context, o *model.Order) error
	Ping(ctx context.Context) error
	Close()
}

type PGRepo struct { pool *pgxpool.Pool }

func NewPG(pool *pgxpool.Pool) *PGRepo { return &PGRepo{pool: pool} }

func (r *PGRepo) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }
func (r *PGRepo) Close() { r.pool.Close() }

func (r *PGRepo) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	const q = `select order_uid, track_number, created_at, data from orders where order_uid=$1`
	row := r.pool.QueryRow(ctx, q, id)
	var o model.Order
	var raw map[string]any
	err := row.Scan(&o.OrderUID, &o.TrackNumber, &o.CreatedAt, &raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	o.Data = raw
	return &o, nil
}

func (r *PGRepo) UpsertOrder(ctx context.Context, o *model.Order) error {
	if err := o.Validate(); err != nil { return err }
	const q = `insert into orders(order_uid, track_number, created_at, data)
		values ($1,$2,$3,$4)
		on conflict (order_uid) do update set track_number=excluded.track_number, created_at=excluded.created_at, data=excluded.data`
	b, err := json.Marshal(o.Data)
	if err != nil { return err }
	_, err = r.pool.Exec(ctx, q, o.OrderUID, o.TrackNumber, o.CreatedAt, b)
	return err
}

// Optional: batch insert within tx for future extension
func (r *PGRepo) upsertBatch(ctx context.Context, orders []*model.Order) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil { return err }
	defer func(){ _ = tx.Rollback(ctx) }()
	const q = `insert into orders(order_uid, track_number, created_at, data)
		values ($1,$2,$3,$4)
		on conflict (order_uid) do update set track_number=excluded.track_number, created_at=excluded.created_at, data=excluded.data`
	for _, o := range orders {
		if err := o.Validate(); err != nil { return err }
		b, err := json.Marshal(o.Data); if err != nil { return err }
		if _, err = tx.Exec(ctx, q, o.OrderUID, o.TrackNumber, o.CreatedAt, b); err != nil { return err }
	}
	return tx.Commit(ctx)
}

