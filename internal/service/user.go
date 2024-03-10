package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/aleksey-kombainov/gophermart-sp.git/pkg/defmoney"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type User struct {
	pool  *pgxpool.Pool
	auth  *Auth
	order *Order
}

func NewUser(pool *pgxpool.Pool, auth *Auth, order *Order) *User {
	return &User{pool: pool, auth: auth, order: order}
}

func (u User) GetBalance(ctx context.Context, id domain.UserID) (balance domain.UserBalance, err error) {
	user, err := u.auth.GetUserByID(ctx, id)
	if err != nil {
		return
	}
	return user.Balance, nil
}

func (u User) GetOperations(ctx context.Context, id domain.UserID) (operations []domain.Operation, err error) {
	sql := `SELECT order_number, op.amount, op.processed_at
		FROM operation op
		WHERE user_id = $1
		ORDER BY processed_at;`

	rows, err := u.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("operations error occured: %w", postgres.ErrorHandler(err))
	}
	defer rows.Close()

	for rows.Next() {
		operation := domain.Operation{UserID: id}

		var dummySum int64

		err = rows.Scan(
			&operation.OrderNumber,
			&dummySum,
			&operation.ProcessedAt,
		)
		operation.Sum = *defmoney.New(dummySum)
		if err != nil {
			return nil, fmt.Errorf("copying operation fields: %w", postgres.ErrorHandler(err))
		}
		operations = append(operations, operation)
	}
	err = rows.Err()
	if err != nil {
		return nil, postgres.ErrorHandler(err)
	}
	if len(operations) == 0 {
		return nil, postgres.ErrNotFound
	}
	return
}

func (u User) PerformOperation(ctx context.Context, id domain.UserID, operation domain.Operation) (err error) {

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	user, err := u.auth.GetUserByID(ctx, id)
	if err != nil {
		log.Logger.Error().Err(fmt.Errorf("unable to get user: %w", err))
		return
	}

	// @todo ACHTUNG, ATTENCION, ALARM!!!! вполне можем уйти в минус при параллельных транзакциях
	subResult, err := user.Balance.Current.Subtract(&operation.Sum)
	if err != nil {
		log.Logger.Error().Err(fmt.Errorf("unable to substract operation amount from current balance: %w", err))
		return
	}
	lessThanZero, err := subResult.LessThan(defmoney.New(0))
	if err != nil {
		log.Logger.Error().Err(fmt.Errorf("unable to compare with zero balance: %w", err))
		return
	}

	if lessThanZero {
		log.Logger.Error().Msgf("balance is going below zero. user balance: %v; operation.sum: %d", user.Balance, operation.Sum.Amount())
		return postgres.ErrBalanceBelowZero
	}
	// achtung

	sql := `INSERT INTO operation (order_number, user_id, amount) VALUES ($1, $2, $3)`
	tag, err := tx.Exec(ctx, sql, operation.OrderNumber, id, operation.Sum.Amount())
	if err != nil {
		return
	}
	if tag.RowsAffected() != 1 {
		log.Logger.Error().Msgf(`unknown error while inserting operation. rows affected: %d`, tag.RowsAffected())
		return errors.New("unknown")
	}

	sql = `UPDATE "user"
		SET current_balance = current_balance - $1, withdrawn_balance = withdrawn_balance + $1
		WHERE id = $2`
	_, err = tx.Exec(ctx, sql, operation.Sum.Amount(), id)
	if err != nil {
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Logger.Error().Err(fmt.Errorf("unable to commit transaction: %w", err))
	}
	return
}
