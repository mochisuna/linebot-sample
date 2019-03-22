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

type ownerRepository struct {
	dbm *db.Client
	dbs *db.Client
}

func NewOwnerRepository(dbmClient *db.Client, dbsClient *db.Client) repository.OwnerRepository {
	return &ownerRepository{
		dbm: dbmClient,
		dbs: dbsClient,
	}
}

func (r *ownerRepository) WithTransaction(ctx context.Context, txFunc func(*sql.Tx) error) error {
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

// upsert 処理
func (r *ownerRepository) Create(owner *domain.Owner) error {
	log.Println("called infrastructure.owner Create")
	_, err := squirrel.Insert(OWNERS).
		Columns("owner_id", "created_at", "updated_at").
		Values(owner.ID, owner.CreatedAt, owner.UpdatedAt).
		RunWith(r.dbm.DB).
		Exec()
	return err

}

func (r *ownerRepository) Get(ownerID domain.OwnerID) (*domain.Owner, error) {
	log.Println("called infrastructure.owner Get")
	var ret domain.Owner
	err := squirrel.Select("owner_id", "created_at", "updated_at").
		From(OWNERS).
		Where(squirrel.Eq{
			"owner_id": ownerID,
		}).
		RunWith(r.dbs.DB).
		QueryRow().
		Scan(
			&ret.ID,
			&ret.CreatedAt,
			&ret.UpdatedAt,
		)
	return &ret, err
}
