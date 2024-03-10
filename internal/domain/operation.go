package domain

import (
	"github.com/Rhymond/go-money"
	"time"
)

type Operation struct {
	UserID      UserID      `json:"-"`
	OrderNumber OrderNumber `json:"order"`
	Sum         money.Money `json:"sum"`
	ProcessedAt time.Time   `json:"processed_at,omitempty"`
}
