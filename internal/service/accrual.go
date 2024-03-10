package service

import (
	"context"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Accrual struct {
	pool                 *pgxpool.Pool
	accrualSystemAddress string
}

func NewAccrual(pool *pgxpool.Pool, accrualSystemAddress string) *Accrual {
	return &Accrual{pool: pool, accrualSystemAddress: accrualSystemAddress}
}

// долбимся во внешний чёрный ящик, пока он там в себе не посчитает
func (a Accrual) getAccrualForOrder(ctx context.Context, number domain.OrderNumber) (accrual domain.OrderAccrualSystem, err error) {
	resp, err := resty.New().R().
		EnableTrace().
		SetContext(ctx).
		SetResult(&accrual).
		Get(a.accrualSystemAddress + "/api/orders/" + string(number))
	if err != nil {
		log.Logger.Err(err)
		return
	}
	if resp.StatusCode() != http.StatusOK {
		//log.Logger.Debug().Msgf("accrual service response StatusCode: %d; order num: %s", resp.StatusCode(), number)
		return accrual, postgres.ErrAccrual
	}
	return
}
