package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"
)

func (h *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	userID := getUserFromContext(ctx)
	//userID := uuid.MustParse("7622eb8e-2a03-40ab-89c7-e3f5dafa2bf0")
	//log.Logger.Debug().Msgf("GetOrders. start. userID %s", userID)

	orders, err := h.order.GetListByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
	}
	//log.Logger.Debug().Msgf("GetOrders. end. userID %s; orders len: %d", userID, len(orders))
}

func (h *UserHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := getUserFromContext(ctx)

	//log.Logger.Debug().Msgf("CreateOrder. start. userID %s", userID)

	if !strings.Contains(r.Header.Get(headers.ContentType), mimetype.TextPlain) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber, err := domain.NewOrderNumber(string(b))
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// @todo почему не работает embedding? при компиляции: unknown field Number in struct literal of type domain.Order
	//order := domain.Order{
	//	UserID: userID,
	//	Number: orderNumber,
	//	Status: domain.OrderStatusNew,
	//}
	// @todo чем чреват new vs literal?
	order := new(domain.Order)
	order.UserID = userID
	order.Number = orderNumber
	order.Status = domain.OrderStatusNew

	err = h.order.CreateOrder(ctx, *order)
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicate) {
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, postgres.ErrDuplicateAnotherUser) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
