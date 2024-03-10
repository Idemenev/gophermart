package handler

import (
	"encoding/json"
	"errors"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *UserHandler) PerformWithdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := getUserFromContext(ctx)

	var operation domain.Operation

	err := json.NewDecoder(r.Body).Decode(&operation)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Logger.Error().Err(err)
		return
	}

	err = operation.OrderNumber.Validate()
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Logger.Error().Err(err)
		return
	}

	operation.UserID = userID

	err = h.user.PerformOperation(ctx, userID, operation)
	if err != nil {
		if errors.Is(err, postgres.ErrBalanceBelowZero) {
			w.WriteHeader(http.StatusPaymentRequired)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Logger.Error().Err(err)
	}
}
