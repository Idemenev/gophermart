package postgres

import (
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicate            = errors.New("duplicate value")
	ErrDuplicateAnotherUser = errors.New("another user value duplication")
	ErrNotFound             = errors.New("value not found")
	ErrOther                = errors.New("other storage error")
	ErrAccessViolation      = errors.New("access violation")
	ErrBalanceBelowZero     = errors.New("balance can't be below zero")
	ErrAccrual              = errors.New("error accrual")
)

func ErrorHandler(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("%w: %s", ErrDuplicate, pgErr.Message)
		}
		return fmt.Errorf("%w: %s", ErrOther, pgErr.Message)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s", ErrNotFound, err)
	}

	return fmt.Errorf("%s: %s", ErrOther, err)
}
