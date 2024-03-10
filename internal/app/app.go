package app

import (
	"context"
	"fmt"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/handler"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
)

const (
	ExitCodeErrorGeneral = 1
)

var connectionTimeout = time.Second * 5
var migrationsTimeout = time.Second * 10
var maxConnections int32 = 30

func Shutdown(exitCode int) {
	log.Info().Msgf("shutting down with exit code %d", exitCode)
	os.Exit(exitCode)
}

func Run(ctx context.Context) (err error) {
	config, err := GetConfig()
	if err != nil {
		return
	}
	if err = setupLogger(config); err != nil {
		return
	}

	connectionContext, connectionContextCancel := context.WithTimeout(ctx, connectionTimeout)
	defer connectionContextCancel()
	dbConPool, err := createDBConnPool(connectionContext, config.DatabaseURI)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
		Shutdown(ExitCodeErrorGeneral)
	}
	defer dbConPool.Close()
	if err != nil {
		return
	}

	migrationsContext, migrationsContextCancel := context.WithTimeout(ctx, migrationsTimeout)
	defer migrationsContextCancel()
	if err = migrationsUP(migrationsContext, dbConPool, config.ProjectRootDir); err != nil {
		return
	}

	mux := getRouter(buildServiceContainer(dbConPool, config))
	if err := http.ListenAndServe(config.RunAddress, mux); err != nil {
		log.Error().Msgf("can't start server: %s", err)
		Shutdown(ExitCodeErrorGeneral)
	}

	return nil
}

func buildServiceContainer(dbConPool *pgxpool.Pool, config Config) *ServiceContainer {
	auth := service.NewAuth(dbConPool)
	accrual := service.NewAccrual(dbConPool, config.AccrualSystemAddress)
	order := service.NewOrder(dbConPool, *accrual)
	user := service.NewUser(dbConPool, auth, order)
	serviceContainer := getServiceContainerInstance(auth, order, user, accrual)
	return serviceContainer
}

func setupLogger(conf Config) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lvl, err := zerolog.ParseLevel(conf.LogLevelName)
	if err != nil {
		return fmt.Errorf("unknown log-level name. Provided: '%s'. %w", conf.LogLevelName, err)
	}
	zerolog.SetGlobalLevel(lvl)
	return nil
}

func getRouter(container *ServiceContainer) *chi.Mux {
	mux := chi.NewRouter()
	mux.Route("/api/user", func(r chi.Router) {
		authHandler := handler.NewAuthHandler(container.Auth)
		r.Group(func(r chi.Router) {
			// Put метод сюда
			r.Post("/register", authHandler.SignUp)
			r.Post("/login", authHandler.SignIn)
		})

		r.Group(func(r chi.Router) {
			r.Use(authHandler.TokenAuthorization)

			userHandler := handler.NewUserHandler(container.Order, *container.User)

			r.Get("/orders", userHandler.GetOrders)
			r.Post("/orders", userHandler.CreateOrder)

			r.Get("/balance", userHandler.GetBalance)

			r.Post("/balance/withdraw", userHandler.PerformWithdraw)
			r.Get("/withdrawals", userHandler.GetWithdrawals)
		})
	})
	return mux
}
