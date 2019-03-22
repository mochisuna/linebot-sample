package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type OwnerRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	Get(domain.OwnerID) (*domain.Owner, error)
	Create(*domain.Owner) error
}
