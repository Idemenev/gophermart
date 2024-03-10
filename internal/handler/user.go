package handler

import "github.com/aleksey-kombainov/gophermart-sp.git/internal/service"

type UserHandler struct {
	order *service.Order
	user  service.User
}

func NewUserHandler(order *service.Order, user service.User) *UserHandler {
	return &UserHandler{order: order, user: user}
}
