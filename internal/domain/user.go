package domain

import (
	"errors"
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type UserID = uuid.UUID

var EmptyUserID = uuid.Nil

type User struct {
	ID           UserID
	Login        string
	PasswordHash string
	Balance      UserBalance
}

type UserBalance struct {
	Current   money.Money `json:"current"`
	Withdrawn money.Money `json:"withdrawn"`
}

type Authentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a Authentication) Validate() error {
	if a.Login == "" {
		return errors.New("login must be not empty")
	}
	if a.Password == "" {
		return errors.New("password must be not empty")
	}

	return nil
}
