package infrastructure

import (
	"context"
	"database/sql"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/mochisuna/linebot-sample/domain"
	"github.com/mochisuna/linebot-sample/domain/repository"
	"github.com/mochisuna/linebot-sample/infrastructure/db"
)

type eventRepository struct {
	dbm *db.Client
	dbs *db.Client
}

func NewEventRepository(dbmClient *db.Client, dbsClient *db.Client) repository.EventRepository {
	return &eventRepository{
		dbm: dbmClient,
		dbs: dbsClient,
	}
}

// TODO 共通化
func (r *eventRepository) WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error {
	tx, err := r.dbm.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}

func (r *eventRepository) Create(event *domain.Event) error {
	log.Println("called infrastructure.event Create")
	_, err := squirrel.Insert(EVENTS).
		Columns("event_id", "created_at", "updated_at").
		Values(event.ID, event.CreatedAt, event.UpdatedAt).
		RunWith(r.dbm.DB).
		Exec()
	if err != nil {
		return err
	}
	_, err = squirrel.Insert(EVENT_STATUSES).
		Columns("event_id", "owner_id", "status", "created_at", "updated_at").
		Values(event.ID, event.OwnerID, event.Status, event.CreatedAt, event.UpdatedAt).
		RunWith(r.dbm.DB).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *eventRepository) Update(event *domain.Event) error {
	log.Println("called infrastructure.event Update")
	_, err := squirrel.Update(EVENT_STATUSES).
		SetMap(squirrel.Eq{
			"status":     event.Status,
			"updated_at": event.UpdatedAt,
		}).
		Where(squirrel.Eq{
			"owner_id": event.OwnerID,
		}).
		RunWith(r.dbm.DB).
		Exec()
	return err
}
func (r *eventRepository) GetByOwnerID(ownerID domain.OwnerID, status *domain.EventStatus) (*domain.Event, error) {
	log.Println("called infrastructure.event GetByOwnerID")
	var col eventStatusColumns
	param := squirrel.Eq{
		"owner_id": ownerID,
	}
	if status != nil {
		param = squirrel.Eq{
			"owner_id": ownerID,
			"status":   status,
		}
	}
	err := squirrel.Select("event_id", "owner_id", "status", "created_at", "updated_at").
		From(EVENT_STATUSES).
		Where(param).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&col.EventID,
			&col.OwnerID,
			&col.Status,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
	return &domain.Event{
		ID:        col.EventID,
		OwnerID:   col.OwnerID,
		Status:    col.Status,
		CreatedAt: col.CreatedAt,
		UpdatedAt: col.UpdatedAt,
	}, err
}

func (r *eventRepository) GetByEventID(eventID domain.EventID) (*domain.Event, error) {
	log.Println("called infrastructure.event GetByEventID")
	var col eventStatusColumns
	err := squirrel.Select("event_id", "owner_id", "status", "created_at", "updated_at").
		From(EVENT_STATUSES).
		Where(squirrel.Eq{
			"event_id": eventID,
		}).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&col.EventID,
			&col.OwnerID,
			&col.Status,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
	return &domain.Event{
		ID:        col.EventID,
		OwnerID:   col.OwnerID,
		Status:    col.Status,
		CreatedAt: col.CreatedAt,
		UpdatedAt: col.UpdatedAt,
	}, err
}

func (r *eventRepository) GetList(status *domain.EventStatus) ([]domain.Event, error) {
	log.Println("called infrastructure.event Get")
	var ret []domain.Event
	rows, err := squirrel.Select("event_id", "owner_id", "status", "created_at", "updated_at").
		From(EVENT_STATUSES).
		Where(squirrel.Eq{
			"status": status,
		}).
		RunWith(r.dbs.DB).
		Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var eventStatus eventStatusColumns
		err = rows.Scan(
			&eventStatus.EventID,
			&eventStatus.OwnerID,
			&eventStatus.Status,
			&eventStatus.CreatedAt,
			&eventStatus.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ret = append(ret, domain.Event{
			ID:        eventStatus.EventID,
			OwnerID:   eventStatus.OwnerID,
			Status:    eventStatus.Status,
			CreatedAt: eventStatus.CreatedAt,
			UpdatedAt: eventStatus.UpdatedAt,
		})
	}

	return ret, err
}
