package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type UserRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	Select(*domain.UserID, *domain.EventID) (*domain.User, error)
	SelectByIDAndStatus(*domain.UserID, bool) (*domain.User, error)
	Update(*domain.User, *sql.Tx) error
	Participate(*domain.User, *sql.Tx) error
	Vote(*domain.User, *sql.Tx) error
}
