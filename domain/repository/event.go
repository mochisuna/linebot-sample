package repository

import (
	"context"
	"database/sql"

	"github.com/mochisuna/linebot-sample/domain"
)

type EventRepository interface {
	WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error
	SelectByOwnerID(domain.OwnerID, *domain.EventStatus) (*domain.Event, error)
	SelectByEventID(domain.EventID) (*domain.Event, error)
	SelectList(*domain.EventStatus) ([]domain.Event, error)
	Update(*domain.Event, *sql.Tx) error
	Create(*domain.Event, *sql.Tx) error
}
