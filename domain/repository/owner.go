package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type OwnerRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	Select(domain.OwnerID) (*domain.Owner, error)
	Create(*domain.Owner, *sql.Tx) error
}
