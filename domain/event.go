package domain

type EventStatus int
type EventID string

const (
	EVENT_STABDBY EventStatus = iota
	EVENT_OPEN                = iota
	EVENT_CLOSED              = iota
)

type Event struct {
	ID        EventID
	OwnerID   OwnerID
	Status    EventStatus
	CreatedAt int
	UpdatedAt int
}
