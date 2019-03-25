package domain

type VOTE_STATUS int
type UserID string

const (
	NOT_VOTED VOTE_STATUS = iota
	GREAT
	GOOD
	NOT_GOOD
	BAD
)

type User struct {
	ID             UserID
	EventID        EventID
	IsParticipated bool
	Vote           VOTE_STATUS
	CreatedAt      int
	UpdatedAt      int
}
