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
	userRepo  repository.UserRepository
}

// NewCallbackService inject eventRepo
func NewCallbackService(eventRepo repository.EventRepository, ownerRepo repository.OwnerRepository, userRepo repository.UserRepository) service.CallbackService {
	return &CallbackService{
		eventRepo: eventRepo,
		ownerRepo: ownerRepo,
		userRepo:  userRepo,
	}
}

func (s *CallbackService) Follow(ctx context.Context, ownerID domain.OwnerID) (*domain.Owner, error) {
	log.Println("called application.Follow")
	now := int(time.Now().Unix())
	owner := &domain.Owner{
		ID:        ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.ownerRepo.Select(ownerID)
	if err != nil && err == sql.ErrNoRows {
		err := s.ownerRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
			return s.ownerRepo.Create(owner, tx)
		})
		if err != nil {
			return nil, err
		}
	}
	return owner, nil
}

func (s *CallbackService) GetEventByOwnerID(ownerID domain.OwnerID, status domain.EventStatus) (*domain.Event, error) {
	log.Println("called application.GetEventByOwnerID")
	return s.eventRepo.SelectByOwnerID(ownerID, &status)
}

func (s *CallbackService) UpdateEventStatus(ctx context.Context, ownerID domain.OwnerID, status domain.EventStatus) (*domain.Event, error) {
	log.Println("called application.UpdateEventStatus")
	event, err := s.eventRepo.SelectByOwnerID(ownerID, nil)
	if err != nil {
		return nil, err
	}
	event.UpdatedAt = int(time.Now().Unix())
	event.Status = status

	err = s.eventRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		return s.eventRepo.Update(event, tx)
	})
	return event, err
}

func (s *CallbackService) RegisterEvent(ctx context.Context, ownerID domain.OwnerID) (*domain.Event, error) {
	log.Println("called application.RegisterEvent")
	now := int(time.Now().Unix())
	event := &domain.Event{
		ID:        domain.EventID(xid.New().String()),
		OwnerID:   ownerID,
		Status:    domain.EVENT_STABDBY,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.eventRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		return s.eventRepo.Create(event, tx)
	})
	return event, err
}
func (s *CallbackService) GetParticipatedEvent(userID domain.UserID) (*domain.User, error) {
	log.Println("called application.GetParticipatedEvent")
	return s.userRepo.SelectByIDAndStatus(&userID, true)
}

func (s *CallbackService) GetActiveEvents() ([]domain.Event, error) {
	log.Println("called application.GetActiveEvents")
	status := domain.EVENT_OPEN
	return s.eventRepo.SelectList(&status)
}

func (s *CallbackService) ParticipateEvent(ctx context.Context, userID *domain.UserID, eventID *domain.EventID) error {
	log.Println("called application.ParticipateEvent")
	now := int(time.Now().Unix())
	user := &domain.User{
		ID:             *userID,
		EventID:        *eventID,
		IsParticipated: true,
		Vote:           domain.NOT_VOTED,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	// not foundの場合だけUPDATEする
	ret, err := s.userRepo.Select(userID, eventID)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	return s.userRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		if len(ret.ID) > 0 {
			return s.userRepo.Update(user, tx)
		}
		return s.userRepo.Participate(user, tx)
	})
}

func (s *CallbackService) GetEventByEventID(eventID domain.EventID) (*domain.Event, error) {
	log.Println("called application.GetEventByEventID")
	return s.eventRepo.SelectByEventID(eventID)
}

func (s *CallbackService) LeaveEvent(ctx context.Context, userID *domain.UserID, eventID *domain.EventID) error {
	log.Println("called application.LeaveEvent")
	now := int(time.Now().Unix())
	user := &domain.User{
		ID:             *userID,
		EventID:        *eventID,
		IsParticipated: false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return s.userRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		return s.userRepo.Update(user, tx)
	})
}

func (s *CallbackService) VoteEvent(ctx context.Context, userID *domain.UserID, eventID *domain.EventID, vote domain.VOTE_STATUS) error {
	log.Println("called application.VoteEvent")
	now := int(time.Now().Unix())
	user := &domain.User{
		ID:        *userID,
		EventID:   *eventID,
		Vote:      vote,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.userRepo.WithTransaction(ctx, func(tx *sql.Tx) error {
		return s.userRepo.Vote(user, tx)
	})
}
