package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
	"github.com/Mimist-Illusionard/mythings/internal/domain/ports/repository"
)

type PostgresItemsRepository struct {
	sql *sql.DB
}

func NewItemsRepo(db *sql.DB) *PostgresItemsRepository {
	return &PostgresItemsRepository{
		sql: db,
	}
}

func (r *PostgresItemsRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.Item, error) {
	const query = `
		SELECT id, name, description, image_url, price, attributes, created_at, updated_at
		FROM items
		WHERE id = $1
	`

	item := &models.Item{}
	var attributes []byte

	err := r.sql.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.ImageURL,
		&item.Price,
		&attributes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}

	if err := json.Unmarshal(attributes, &item.Attributes); err != nil {
		return nil, fmt.Errorf("unmarshal item attributes: %w", err)
	}

	return item, nil
}

func (r *PostgresItemsRepository) List(
	ctx context.Context,
	filter repository.ItemFilter,
	limit, offset int,
) ([]*models.Item, error) {
	query := `
		SELECT i.id, i.name, i.description, i.image_url,
		       i.price, i.attributes, i.created_at, i.updated_at
		FROM items i
		WHERE 1 = 1
	`

	args := make([]any, 0)
	arg := 1

	if filter.Name != "" {
		query += fmt.Sprintf(
			" AND i.name ILIKE '%%' || $%d || '%%'",
			arg,
		)

		args = append(args, filter.Name)
		arg++
	}

	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND i.price >= $%d", arg)
		args = append(args, *filter.MinPrice)
		arg++
	}

	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND i.price <= $%d", arg)
		args = append(args, *filter.MaxPrice)
		arg++
	}

	for _, tag := range filter.Tags {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM items_tags it
				JOIN tags t ON t.id = it.tag_id
				WHERE it.item_id = i.id
				  AND t.name = $%d
			)
		`, arg)

		args = append(args, tag)
		arg++
	}

	query += fmt.Sprintf(
		" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d",
		arg,
		arg+1,
	)

	args = append(args, limit, offset)

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]*models.Item, 0)

	for rows.Next() {
		item := &models.Item{}
		var attributes []byte

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.ImageURL,
			&item.Price,
			&attributes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		if err := json.Unmarshal(attributes, &item.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal item attributes: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *PostgresItemsRepository) Save(
	ctx context.Context,
	item *models.Item,
) error {
	attributes, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal item attributes: %w", err)
	}

	if item.ID == 0 {
		const query = `
			INSERT INTO items (
				name,
				description,
				image_url,
				price,
				attributes
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`

		err := r.sql.QueryRowContext(
			ctx,
			query,
			item.Name,
			item.Description,
			item.ImageURL,
			item.Price,
			attributes,
		).Scan(
			&item.ID,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}

		return nil
	}

	const query = `
		UPDATE items
		SET name = $1,
		    description = $2,
		    image_url = $3,
		    price = $4,
		    attributes = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err = r.sql.QueryRowContext(
		ctx,
		query,
		item.Name,
		item.Description,
		item.ImageURL,
		item.Price,
		attributes,
		item.ID,
	).Scan(&item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func (r *PostgresItemsRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM items
		WHERE id = $1
	`

	result, err := r.sql.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete item rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresItemsRepository) AddTag(
	ctx context.Context,
	itemID, tagID int64,
) error {
	const query = `
		INSERT INTO items_tags (item_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT (item_id, tag_id) DO NOTHING
	`

	if _, err := r.sql.ExecContext(ctx, query, itemID, tagID); err != nil {
		return fmt.Errorf("add tag to item: %w", err)
	}

	return nil
}

func (r *PostgresItemsRepository) RemoveTag(
	ctx context.Context,
	itemID, tagID int64,
) error {
	const query = `
		DELETE FROM items_tags
		WHERE item_id = $1
		  AND tag_id = $2
	`

	if _, err := r.sql.ExecContext(ctx, query, itemID, tagID); err != nil {
		return fmt.Errorf("remove tag from item: %w", err)
	}

	return nil
}
