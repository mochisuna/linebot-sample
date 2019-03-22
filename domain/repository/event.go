package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type EventRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	Get(domain.OwnerID) (*domain.Event, error)
	GetWithStatus(domain.OwnerID, domain.EventStatus) (*domain.Event, error)
	Update(*domain.Event) error
	Create(*domain.Event) error
}
