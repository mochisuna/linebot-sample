package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
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
	_, err = squirrel.Insert(STATUSES).
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
	fmt.Println(event)
	_, err := squirrel.Update(STATUSES).
		SetMap(squirrel.Eq{
			"status":     event.Status,
			"updated_at": event.UpdatedAt,
		}).
		Where(squirrel.Eq{
			"owner_id": event.OwnerID,
		}).
		RunWith(r.dbm.DB).
		Exec()
	fmt.Println(err)
	return err
}

func (r *eventRepository) Get(ownerID domain.OwnerID, status *domain.EventStatus) (*domain.Event, error) {
	log.Println("called infrastructure.event Get")
	var ret domain.Event
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
		From(STATUSES).
		Where(param).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&ret.ID,
			&ret.OwnerID,
			&ret.Status,
			&ret.CreatedAt,
			&ret.UpdatedAt,
		)
	return &ret, err
}
