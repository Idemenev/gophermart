package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/domain"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/postgres"
	"github.com/aleksey-kombainov/gophermart-sp.git/pkg/defmoney"
	"github.com/aleksey-kombainov/gophermart-sp.git/pkg/password"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// @todo хранить НЕ в коде.
const secretKey = "to-secret-key-for-encoding"

// вынести в конфиг
const expiresDuration = time.Duration(24 * 7 * time.Hour)

var signatureMethod = jwt.SigningMethodHS256

type Auth struct {
	pool *pgxpool.Pool
}

func NewAuth(pool *pgxpool.Pool) *Auth {
	return &Auth{pool: pool}
}

func (a *Auth) SignUp(ctx context.Context, auth domain.Authentication) (userID domain.UserID, err error) {

	passHash, err := password.Hash(auth.Password)
	if err != nil {
		return domain.EmptyUserID, fmt.Errorf("can't calc pass hash during signup: %w", err)
	}
	user := domain.User{
		Login:        auth.Login,
		PasswordHash: passHash,
	}

	userID, err = a.createUser(ctx, user)
	if err != nil {
		return domain.EmptyUserID, fmt.Errorf("creating a new user: %w", err)
	}
	return userID, nil
}

// @todo в го принято экспортируемые как-то внизу или сверху кода размещать?
func (a *Auth) createUser(ctx context.Context, user domain.User) (userID domain.UserID, err error) {

	sql := `INSERT INTO "user" (login, password_hash) VALUES ($1, $2) RETURNING id`
	err = a.pool.QueryRow(ctx, sql, user.Login, user.PasswordHash).Scan(&userID)

	if err != nil {
		return domain.EmptyUserID, fmt.Errorf("creating a new user: %w", postgres.ErrorHandler(err))
	}
	return userID, nil
}

func (a *Auth) SignIn(ctx context.Context, auth domain.Authentication) (domain.UserID, error) {
	user, err := a.getUserByLogin(ctx, auth.Login)
	if err != nil {
		return domain.EmptyUserID, fmt.Errorf("user not found: %w", err)
	}

	if user.Login != auth.Login || !password.CompareHashAndPassword(user.PasswordHash, auth.Password) {
		return domain.EmptyUserID, postgres.ErrNotFound
	}

	return user.ID, nil
}

func (a *Auth) GetUserByID(ctx context.Context, id domain.UserID) (user domain.User, err error) {
	user, err = a.getUserByField(ctx, "id", id.String())
	return
}

func (a *Auth) getUserByLogin(ctx context.Context, userLogin string) (user domain.User, err error) {

	user, err = a.getUserByField(ctx, "login", userLogin)
	return
}

func (a *Auth) getUserByField(ctx context.Context, searchFieldName string, searchValue string) (user domain.User, err error) {
	sql := fmt.Sprintf(`SELECT id, "login", password_hash, current_balance, withdrawn_balance
			FROM "user"
			WHERE "%s" = $1`, searchFieldName)

	// @todo не понятно как реализовать Scan интерфейс для money.Money. Через емединг в свой struct не получается, т.к. поля в money.Money не экспортируемые

	var dummyBalanceCurrent int64
	var dummyBalanceWithdrawn int64

	err = a.pool.QueryRow(ctx, sql, searchValue).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&dummyBalanceCurrent,
		&dummyBalanceWithdrawn,
	)
	user.Balance.Current = *defmoney.New(dummyBalanceCurrent)
	user.Balance.Withdrawn = *defmoney.New(dummyBalanceWithdrawn)
	if err != nil {
		return user, fmt.Errorf("db user search: %w", postgres.ErrorHandler(err))
	}
	return
}

func (a *Auth) CheckUserExists(ctx context.Context, id domain.UserID) (err error) {
	sql := `SELECT 1
			FROM "user"
			WHERE id = $1`
	var dummy int // @todo how to find out was a row returned or not?
	err = a.pool.QueryRow(ctx, sql, id).Scan(&dummy)
	if err != nil {
		return postgres.ErrorHandler(err)
	}
	return
}

func (a *Auth) GetTokenStringForUser(id domain.UserID) (tokenString string, err error) {
	now := time.Now()
	token := jwt.NewWithClaims(signatureMethod, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        id.String(),
	})
	tokenString, err = token.SignedString([]byte(secretKey))
	return

}

func (a Auth) GetUserFromTokenString(t string) (domain.UserID, error) {
	opts := jwt.WithValidMethods([]string{signatureMethod.Alg()})
	token, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}, opts)

	if err == nil && token.Valid {
		claims := token.Claims.(*jwt.RegisteredClaims)
		return uuid.Parse(claims.ID)
	}

	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return domain.EmptyUserID, fmt.Errorf("that's not even a token: %s", err)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return domain.EmptyUserID, fmt.Errorf("invalid signature: %s", err)
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		return domain.EmptyUserID, fmt.Errorf("timing is everything: %s", err)
	default:
		return domain.EmptyUserID, fmt.Errorf("couldn't handle this token: %s", err)
	}
}
