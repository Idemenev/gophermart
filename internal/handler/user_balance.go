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

func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := getUserFromContext(ctx)
	balance, err := h.user.GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Logger.Error().Err(err)
		return
	}

	w.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		log.Logger.Error().Err(err)
	}
}
