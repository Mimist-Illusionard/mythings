package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
)

type PostgresTagsRepository struct {
	sql *sql.DB
}

func NewTagsRepo(db *sql.DB) *PostgresTagsRepository {
	return &PostgresTagsRepository{
		sql: db,
	}
}

func (r *PostgresTagsRepository) List(ctx context.Context, name string, limit, offset int) ([]*models.Tag, error) {
	tags := make([]*models.Tag, 0)

	query := `
        SELECT * FROM tags 
        WHERE name ILIKE '%' || $1 || '%'
        ORDER BY name
        LIMIT $2 OFFSET $3
    `

	rows, err := r.sql.QueryContext(ctx, query, name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("tags list error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		tag := &models.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("tags scan: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func (r *PostgresTagsRepository) Save(ctx context.Context, tag *models.Tag) error {
	if tag.ID == 0 {
		const query = `
			INSERT INTO tags (name)
			VALUES ($1)
			RETURNING id
		`

		if err := r.sql.QueryRowContext(
			ctx,
			query,
			tag.Name,
		).Scan(&tag.ID); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}

		return nil
	}

	const query = `
		UPDATE tags
		SET name = $1
		WHERE id = $2
	`

	result, err := r.sql.ExecContext(ctx, query, tag.Name, tag.ID)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag with id %d not found", tag.ID)
	}

	return nil
}

func (r *PostgresTagsRepository) Delete(
	ctx context.Context,
	name string,
) error {
	query := `
		DELETE FROM tags
		WHERE name = $1
	`

	result, err := r.sql.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after tag delete: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag %q not found", name)
	}

	return nil
}
