package application

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mochisuna/linebot-sample/domain"
	"github.com/mochisuna/linebot-sample/domain/repository"
	"github.com/mochisuna/linebot-sample/domain/service"

	"github.com/rs/xid"
)

type CallbackService struct {
	eventRepo repository.EventRepository
	ownerRepo repository.OwnerRepository
}

// NewCallbackService inject eventRepo
func NewCallbackService(eventRepo repository.EventRepository, ownerRepo repository.OwnerRepository) service.CallbackService {
	return &CallbackService{
		eventRepo: eventRepo,
		ownerRepo: ownerRepo,
	}
}

func (s *CallbackService) Follow(ctx context.Context, ownerID domain.OwnerID) (*domain.Owner, error) {
	log.Println("application.Follow")
	now := int(time.Now().Unix())
	owner := &domain.Owner{
		ID:        ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.ownerRepo.Get(ownerID)
	if err != nil && err == sql.ErrNoRows {
		err := s.ownerRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
			if err := s.ownerRepo.Create(owner); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return owner, nil
}

func (s *CallbackService) ReferEventStatus(ownerID domain.OwnerID, status domain.EventStatus) (*domain.Event, error) {
	log.Println("application.ReferEventStatus")
	return s.eventRepo.Get(ownerID, &status)
}

func (s *CallbackService) UpdateEventStatus(ctx context.Context, ownerID domain.OwnerID, status domain.EventStatus) (*domain.Event, error) {
	log.Println("application.UpdateEventStatus")
	event, err := s.eventRepo.Get(ownerID, nil)
	if err != nil {
		return nil, err
	}
	event.UpdatedAt = int(time.Now().Unix())
	event.Status = status

	err = s.eventRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		if err := s.eventRepo.Update(event); err != nil {
			return err
		}
		return nil
	})
	return event, err
}

func (s *CallbackService) RegisterEvent(ctx context.Context, ownerID domain.OwnerID) (*domain.Event, error) {
	log.Println("application.RegisterEvent")
	now := int(time.Now().Unix())
	event := &domain.Event{
		ID:        domain.EventID(xid.New().String()),
		OwnerID:   ownerID,
		Status:    domain.EVENT_STABDBY,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.eventRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		if err := s.eventRepo.Create(event); err != nil {
			return err
		}
		return nil
	})
	return event, err
}
