package handler

import (
	"encoding/json"
	"errors"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *UserHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := getUserFromContext(ctx)

	operations, err := h.user.GetOperations(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Logger.Error().Err(err)
		return
	}

	w.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(operations)
	if err != nil {
		log.Logger.Error().Err(err)
	}
}
