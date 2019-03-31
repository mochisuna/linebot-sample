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

type userRepository struct {
	dbm *db.Client
	dbs *db.Client
}

func NewUserRepository(dbmClient *db.Client, dbsClient *db.Client) repository.UserRepository {
	return &userRepository{
		dbm: dbmClient,
		dbs: dbsClient,
	}
}

// TODO 共通化
func (r *userRepository) WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error {
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

func (r *userRepository) Select(userID *domain.UserID, eventID *domain.EventID) (*domain.User, error) {
	log.Println("called infrastructure.user Select")
	var col eventParticipantsColumns
	err := squirrel.Select("user_id", "event_id", "is_participated", "created_at", "updated_at").
		From(EVENT_PARTICIPANTS).
		Where(squirrel.Eq{
			"user_id":  *userID,
			"event_id": *eventID,
		}).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&col.UserID,
			&col.EventID,
			&col.IsParticipated,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
	return &domain.User{
		ID:             col.UserID,
		EventID:        col.EventID,
		IsParticipated: col.IsParticipated,
		CreatedAt:      col.CreatedAt,
		UpdatedAt:      col.UpdatedAt,
	}, err
}

func (r *userRepository) SelectByIDAndStatus(userID *domain.UserID, isParticipated bool) (*domain.User, error) {
	log.Println("called infrastructure.user SelectByIDAndStatus")
	var col eventParticipantsColumns
	err := squirrel.Select("user_id", "event_id", "is_participated", "created_at", "updated_at").
		From(EVENT_PARTICIPANTS).
		Where(squirrel.Eq{
			"user_id":         *userID,
			"is_participated": isParticipated,
		}).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&col.UserID,
			&col.EventID,
			&col.IsParticipated,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
	return &domain.User{
		ID:             col.UserID,
		EventID:        col.EventID,
		IsParticipated: col.IsParticipated,
		CreatedAt:      col.CreatedAt,
		UpdatedAt:      col.UpdatedAt,
	}, err
}

func (r *userRepository) Update(user *domain.User, tx *sql.Tx) error {
	log.Println("called infrastructure.user Update")
	_, err := squirrel.Update(EVENT_PARTICIPANTS).
		SetMap(squirrel.Eq{
			"is_participated": user.IsParticipated,
			"updated_at":      user.UpdatedAt,
		}).
		Where(squirrel.Eq{
			"user_id":  user.ID,
			"event_id": user.EventID,
		}).
		RunWith(tx).
		Exec()
	return err
}

func (r *userRepository) Participate(user *domain.User, tx *sql.Tx) error {
	log.Println("called infrastructure.user Participate")
	_, err := squirrel.Insert(EVENT_PARTICIPANTS).
		Columns("event_id", "user_id", "is_participated", "created_at", "updated_at").
		Values(user.EventID, user.ID, user.IsParticipated, user.CreatedAt, user.UpdatedAt).
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}
	_, err = squirrel.Insert(EVENT_VOTES).
		Columns("user_id", "event_id", "vote", "created_at", "updated_at").
		Values(user.ID, user.EventID, domain.NOT_VOTED, user.CreatedAt, user.UpdatedAt).
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Vote(user *domain.User, tx *sql.Tx) error {
	log.Println("called infrastructure.user Vote")
	_, err := squirrel.Update(EVENT_VOTES).
		SetMap(squirrel.Eq{
			"vote":       user.Vote,
			"updated_at": user.UpdatedAt,
		}).
		Where(squirrel.Eq{
			"user_id":  user.ID,
			"event_id": user.EventID,
		}).
		RunWith(tx).
		Exec()
	return err
}
