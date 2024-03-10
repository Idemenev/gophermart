package domain

import (
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/google/uuid"
	"time"
)

const (
	OrderStatusUnknown    = "UNKNOWN"
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
	OrderStatusInvalid    = "INVALID"
)

var OrderFinalStatuses = []string{
	OrderStatusProcessed,
	OrderStatusInvalid,
}

type OrderStatus string

type OrderNumber string

func NewOrderNumber(s string) (OrderNumber, error) {

	number := OrderNumber(s)
	err := number.Validate()
	if err != nil {
		return "", fmt.Errorf("order number validation: %w", err)
	}
	return number, nil
}

func (o OrderNumber) Validate() (err error) {
	err = goluhn.Validate(string(o))
	return
}

type OrderAccrual struct {
	Status  OrderStatus `json:"status"`
	Accrual money.Money `json:"accrual"`
}

type Order struct {
	OrderAccrual
	Number    OrderNumber `json:"number"`
	ID        uuid.UUID   `json:"-"`
	UserID    UserID      `json:"-"`
	CreatedAt time.Time   `json:"uploaded_at"`
	UpdatedAt time.Time   `json:"-"`
}

func (o Order) IsEmpty() bool {
	return o.Equal(Order{})
}

func (o Order) Equal(x Order) bool {
	return o.UserID == x.UserID && o.Number == x.Number && o.Status == x.Status &&
		o.Accrual.Amount() == x.Accrual.Amount() &&
		o.CreatedAt.Equal(x.CreatedAt)
}

type OrderAccrualSystem struct {
	OrderAccrual
	Order OrderNumber `json:"order"`
}
