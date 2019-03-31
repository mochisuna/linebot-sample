package service

import (
	"context"

	"github.com/mochisuna/linebot-sample/domain"
)

type CallbackService interface {
	Follow(context.Context, domain.OwnerID) (*domain.Owner, error)
	GetEventByOwnerID(domain.OwnerID, domain.EventStatus) (*domain.Event, error)
	GetActiveEvents() ([]domain.Event, error)
	UpdateEventStatus(context.Context, domain.OwnerID, domain.EventStatus) (*domain.Event, error)
	RegisterEvent(context.Context, domain.OwnerID) (*domain.Event, error)
	GetEventByEventID(domain.EventID) (*domain.Event, error)
	GetParticipatedEvent(domain.UserID) (*domain.User, error)
	ParticipateEvent(context.Context, *domain.UserID, *domain.EventID) error
	LeaveEvent(context.Context, *domain.UserID, *domain.EventID) error
}
