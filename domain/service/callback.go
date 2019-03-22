package service

import (
	"context"

	"github.com/mochisuna/linebot-sample/domain"
)

type CallbackService interface {
	Follow(context.Context, domain.OwnerID) (*domain.Owner, error)
	ReferEventStatus(domain.OwnerID, domain.EventStatus) (*domain.Event, error)
	UpdateEventStatus(context.Context, domain.OwnerID, domain.EventStatus) (*domain.Event, error)
	RegisterEvent(context.Context, domain.OwnerID) (*domain.Event, error)
}
