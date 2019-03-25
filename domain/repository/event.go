package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type EventRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	GetByOwnerID(domain.OwnerID, *domain.EventStatus) (*domain.Event, error)
	GetByEventID(domain.EventID) (*domain.Event, error)
	GetList(*domain.EventStatus) ([]domain.Event, error)
	Update(*domain.Event) error
	Create(*domain.Event) error
}
