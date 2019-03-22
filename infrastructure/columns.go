package infrastructure

import (
	"github.com/mochisuna/linebot-sample/domain"
)

const (
	OWNERS   = "owners"
	STATUSES = "statuses"
	EVENTS   = "events"
)

type ownerColumns struct {
	OwnerID   string `db:"owner_id"`
	CreatedAt int    `db:"created_at"`
	UpdatedAt int    `db:"updated_at"`
}
type eventColumns struct {
	ID        int            `db:"id"`
	EventID   domain.EventID `db:"event_id"`
	CreatedAt int            `db:"created_at"`
	UpdatedAt int            `db:"updated_at"`
}
type statusColumns struct {
	EventID   domain.EventID     `db:"event_id"`
	OwnerID   string             `db:"owner_id"`
	Status    domain.EventStatus `db:"status"`
	CreatedAt int                `db:"created_at"`
	UpdatedAt int                `db:"updated_at"`
}
