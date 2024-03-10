package app

import (
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/service"
	"sync"
)

type ServiceContainer struct {
	Auth    *service.Auth
	Order   *service.Order
	User    *service.User
	Accrual *service.Accrual
}

var (
	instance *ServiceContainer
	once     sync.Once
)

func getServiceContainerInstance(auth *service.Auth, order *service.Order, user *service.User, accrual *service.Accrual) *ServiceContainer {
	once.Do(func() {
		instance = &ServiceContainer{
			Auth:    auth,
			Order:   order,
			User:    user,
			Accrual: accrual,
		}
	})
	return instance
}
