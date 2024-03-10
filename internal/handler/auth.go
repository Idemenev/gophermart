package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/service"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

const authHeaderFieldName = "Authorization"
const tokenTypePrefix = "Bearer "

var ctxKeyUserID struct{}

type AuthHandler struct {
	auth *service.Auth
}

func NewAuthHandler(auth *service.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	auth := domain.Authentication{}

	err := json.NewDecoder(r.Body).Decode(&auth)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Logger.Error().Msg(err.Error())
		return
	}

	err = auth.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Logger.Error().Msg(err.Error())
		return
	}

	ctx := r.Context()

	userID, err := h.auth.SignUp(ctx, auth)
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicate) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Logger.Error().Msg(err.Error())
		return
	}
	tokenStr, err := h.auth.GetTokenStringForUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Error().Msg(err.Error())
		return
	}
	h.setAuthToken(tokenStr, w)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var auth domain.Authentication

	err := json.NewDecoder(r.Body).Decode(&auth)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Logger.Error().Msg(err.Error())
		return
	}

	err = auth.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Logger.Error().Msg(err.Error())
		return
	}

	ctx := r.Context()

	userID, err := h.auth.SignIn(ctx, auth)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Logger.Error().Msg(err.Error())
		return
	}

	tokenStr, err := h.auth.GetTokenStringForUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Error().Msg(err.Error())
		return
	}

	h.setAuthToken(tokenStr, w)
}

func (h *AuthHandler) TokenAuthorization(next http.Handler) http.Handler {
	auth := func(w http.ResponseWriter, r *http.Request) {
		tokenStr := h.getAuthToken(r)

		userID, err := h.auth.GetUserFromTokenString(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			log.Logger.Error().Msg(err.Error())
			return
		}

		//log.Logger.Debug().Msgf("TokenStr: %s; userID: %s", tokenStr, userID)

		ctx := r.Context()

		err = h.auth.CheckUserExists(ctx, userID)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			log.Logger.Error().Msg(err.Error())
			return
		}

		*r = *r.WithContext(context.WithValue(ctx, ctxKeyUserID, userID))

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(auth)
}

func (h AuthHandler) setAuthToken(tokenStr string, w http.ResponseWriter) {
	w.Header().Set(authHeaderFieldName, tokenTypePrefix+tokenStr)
}

func (h AuthHandler) getAuthToken(r *http.Request) (token string) {
	s := r.Header.Get(authHeaderFieldName)
	token, _ = strings.CutPrefix(s, tokenTypePrefix)
	return
}

// =================

func getUserFromContext(ctx context.Context) (userID domain.UserID) {
	userID, ok := ctx.Value(ctxKeyUserID).(domain.UserID)
	if !ok {
		return domain.EmptyUserID
	}
	return
}
