package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type UserRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	Get(*domain.User) (*domain.User, error)
	Update(*domain.User) error
	Participate(*domain.User) error
	Vote(*domain.User) error
}
