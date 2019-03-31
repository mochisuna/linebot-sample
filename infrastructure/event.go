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

func (r *eventRepository) Create(event *domain.Event, tx *sql.Tx) error {
	log.Println("called infrastructure.event Create")
	_, err := squirrel.Insert(EVENTS).
		Columns("event_id", "created_at", "updated_at").
		Values(event.ID, event.CreatedAt, event.UpdatedAt).
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}
	_, err = squirrel.Insert(EVENT_STATUSES).
		Columns("event_id", "owner_id", "status", "created_at", "updated_at").
		Values(event.ID, event.OwnerID, event.Status, event.CreatedAt, event.UpdatedAt).
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *eventRepository) Update(event *domain.Event, tx *sql.Tx) error {
	log.Println("called infrastructure.event Update")
	_, err := squirrel.Update(EVENT_STATUSES).
		SetMap(squirrel.Eq{
			"status":     event.Status,
			"updated_at": event.UpdatedAt,
		}).
		Where(squirrel.Eq{
			"owner_id": event.OwnerID,
		}).
		Where(squirrel.NotEq{
			"status": domain.EVENT_CLOSED,
		}).
		RunWith(tx).
		Exec()
	return err
}
func (r *eventRepository) SelectByOwnerID(ownerID domain.OwnerID, status *domain.EventStatus) (*domain.Event, error) {
	log.Println("called infrastructure.event SelectByOwnerID")
	var col eventStatusColumns
	param := squirrel.Eq{
		"owner_id": ownerID,
	}
	if status != nil {
		param = squirrel.Eq{
			"owner_id": ownerID,
			"status":   *status,
		}
	}
	err := squirrel.Select("event_id", "owner_id", "status", "created_at", "updated_at").
		From(EVENT_STATUSES).
		Where(param).
		Where(squirrel.NotEq{
			"status": domain.EVENT_CLOSED,
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

// TODO Select関数として統合
func (r *eventRepository) SelectByEventID(eventID domain.EventID) (*domain.Event, error) {
	log.Println("called infrastructure.event SelectByEventID")
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

func (r *eventRepository) SelectList(status *domain.EventStatus) ([]domain.Event, error) {
	log.Println("called infrastructure.event SelectList")
	var ret []domain.Event
	rows, err := squirrel.Select("event_id", "owner_id", "status", "created_at", "updated_at").
		From(EVENT_STATUSES).
		Where(squirrel.Eq{
			"status": *status,
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
