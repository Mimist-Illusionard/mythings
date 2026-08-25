package postgres

import (
	"database/sql"
)

type PostgresItemsRepository struct {
	sql *sql.DB
}

func NewItemsRepo(db *sql.DB) *PostgresItemsRepository {
	return &PostgresItemsRepository{
		sql: db,
	}
}
