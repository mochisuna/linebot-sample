package domain

type OwnerID string

type Owner struct {
	ID        OwnerID
	CreatedAt int
	UpdatedAt int
}
