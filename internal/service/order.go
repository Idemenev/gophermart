package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/aleksey-kombainov/gophermart-sp.git/pkg/defmoney"
	"github.com/aleksey-kombainov/gophermart-sp.git/pkg/helper"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"time"
)

type Order struct {
	pool    *pgxpool.Pool
	accrual Accrual
}

func NewOrder(pool *pgxpool.Pool, accrual Accrual) *Order {
	return &Order{pool: pool, accrual: accrual}
}

func (o *Order) GetListByUser(ctx context.Context, id domain.UserID) (orders []domain.Order, err error) {
	sql := `SELECT
		order_number, status, accrual, created_at
	FROM "order"
	WHERE user_id = $1;`

	rows, err := o.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("getting orders: %w", postgres.ErrorHandler(err))
	}
	defer rows.Close()

	var dummyAccrual int64

	for rows.Next() {
		order := domain.Order{UserID: id}

		err = rows.Scan(
			&order.Number,
			&order.Status,
			&dummyAccrual,
			&order.CreatedAt,
		)
		order.Accrual = *defmoney.New(dummyAccrual)
		if err != nil {
			return nil, fmt.Errorf("scanning fields: %w", postgres.ErrorHandler(err))
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		return nil, postgres.ErrorHandler(err)
	}
	if len(orders) == 0 {
		return nil, postgres.ErrNotFound
	}
	return
}

func (o *Order) CreateOrder(ctx context.Context, order domain.Order) (err error) {

	savedOrder, err := o.getOrderByNumber(ctx, order.Number)
	if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return
	}
	if err == nil {
		if savedOrder.UserID != order.UserID {
			return postgres.ErrDuplicateAnotherUser
		} else {
			return postgres.ErrDuplicate
		}
	}

	var orderID uuid.UUID
	sql := `INSERT INTO "order" (order_number, user_id) VALUES ($1, $2) RETURNING id`
	err = o.pool.QueryRow(ctx, sql, order.Number, order.UserID).Scan(&orderID)

	if err != nil {
		log.Logger.Error().Msgf("unable to create order with number: %s", order.Number)
		return postgres.ErrorHandler(err)
	}
	log.Logger.Error().Msgf("order CREATED with number: %s", order.Number)
	// @todo сильно жирно по рутине на запрос. нужна очередь с воркерами
	go o.updateAccrual(order.Number)

	return
}

func (o Order) updateAccrual(number domain.OrderNumber) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*45*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Logger.Error().Msgf("ai: accrual calc error %s", number)
			return
		case <-ticker.C:
			ac, err := o.accrual.getAccrualForOrder(ctx, number)
			if err != nil {
				//log.Logger.Error().Msg(err.Error())
				continue
			}
			if !helper.InArray(string(ac.Status), domain.OrderFinalStatuses) {
				continue
			}
			err = o.updateOrderAccrualData(ac)
			if err == nil {
				return
			}
		}
	}
}

func (o Order) updateOrderAccrualData(accrual domain.OrderAccrualSystem) (err error) {

	ctx := context.TODO()
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	var userID domain.UserID

	sql := `UPDATE "order" SET status = $1, accrual = $2 WHERE order_number = $3 RETURNING user_id`
	err = o.pool.QueryRow(context.TODO(), sql, accrual.Status, accrual.Accrual.Amount(), accrual.Order).Scan(&userID)
	if err != nil {
		return
	}
	sql = `UPDATE "user" SET current_balance = current_balance + $1 WHERE id = $2`
	_, err = o.pool.Exec(context.TODO(), sql, accrual.Accrual.Amount(), userID)
	if err != nil {
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
	}
	return
}

func (o *Order) getOrderByNumber(ctx context.Context, orderNumber domain.OrderNumber) (order domain.Order, err error) {

	sql := `SELECT id, order_number, status, accrual, user_id, created_at, updated_at 
		FROM "order" WHERE order_number = $1`

	var dummyAccrual int64

	err = o.pool.QueryRow(ctx, sql, orderNumber).Scan(
		&order.ID,
		&order.Number,
		&order.Status,
		&dummyAccrual,
		&order.UserID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	order.Accrual = *defmoney.New(dummyAccrual)
	if err != nil {
		return order, postgres.ErrorHandler(err)
	}
	return
}
