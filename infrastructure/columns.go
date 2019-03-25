package infrastructure

import (
	"github.com/mochisuna/linebot-sample/domain"
)

const (
	OWNERS             = "owners"
	EVENT_STATUSES     = "event_statuses"
	EVENTS             = "events"
	EVENT_PARTICIPANTS = "event_participants"
	EVENT_VOTES        = "event_votes"
)

// gorpを使わないので厳密には不要だがカラムと同じ構造体を持たせておいた方が取り回しがしやすい
type ownerColumns struct {
	OwnerID   domain.OwnerID `db:"owner_id"`
	CreatedAt int            `db:"created_at"`
	UpdatedAt int            `db:"updated_at"`
}
type eventColumns struct {
	ID        int            `db:"id"`
	EventID   domain.EventID `db:"event_id"`
	CreatedAt int            `db:"created_at"`
	UpdatedAt int            `db:"updated_at"`
}

type eventStatusColumns struct {
	EventID   domain.EventID     `db:"event_id"`
	OwnerID   domain.OwnerID     `db:"owner_id"`
	Status    domain.EventStatus `db:"status"`
	CreatedAt int                `db:"created_at"`
	UpdatedAt int                `db:"updated_at"`
}

type eventParticipantsColumns struct {
	EventID        domain.EventID `db:"event_id"`
	UserID         domain.UserID  `db:"user_id"`
	IsParticipated bool           `db:"is_participated"`
	CreatedAt      int            `db:"created_at"`
	UpdatedAt      int            `db:"updated_at"`
}

type eventVotesColumns struct {
	EventID   domain.EventID     `db:"event_id"`
	OwnerID   domain.OwnerID     `db:"owner_id"`
	Vote      domain.VOTE_STATUS `db:"vote"`
	CreatedAt int                `db:"created_at"`
	UpdatedAt int                `db:"updated_at"`
}
