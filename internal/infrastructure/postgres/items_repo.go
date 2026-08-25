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
	return &PostgresItemsRepository{sql: db}
}

const itemSelect = `
	SELECT
		i.id,
		i.name,
		i.short_description,
		i.description,
		i.image_url,
		i.price,
		i.price_currency,
		i.usd_exchange_rate,
		i.purchased_at,
		i.attributes,
		i.created_at,
		i.updated_at,
		COALESCE((
			SELECT json_agg(
				json_build_object('id', t.id, 'name', t.name)
				ORDER BY t.name
			)
			FROM items_tags it
			JOIN tags t ON t.id = it.tag_id
			WHERE it.item_id = i.id
		), '[]'::json) AS tags
	FROM items i
`

func (r *PostgresItemsRepository) GetByID(ctx context.Context, id int64) (*models.Item, error) {
	row := r.sql.QueryRowContext(ctx, itemSelect+" WHERE i.id = $1", id)

	item, err := scanItem(row.Scan)
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}

	return item, nil
}

func (r *PostgresItemsRepository) List(
	ctx context.Context,
	filter repository.ItemFilter,
	limit, offset int,
) ([]*models.Item, error) {
	query := itemSelect + " WHERE 1 = 1"
	args := make([]any, 0)
	arg := 1

	if filter.Name != "" {
		query += fmt.Sprintf(" AND i.name ILIKE '%%' || $%d || '%%'", arg)
		args = append(args, filter.Name)
		arg++
	}

	priceInRUB := `CASE
		WHEN i.price_currency = 'USD' THEN i.price * i.usd_exchange_rate
		ELSE i.price
	END`

	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND (%s) >= $%d", priceInRUB, arg)
		args = append(args, *filter.MinPrice)
		arg++
	}

	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND (%s) <= $%d", priceInRUB, arg)
		args = append(args, *filter.MaxPrice)
		arg++
	}

	for _, tag := range filter.Tags {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM items_tags it_filter
				JOIN tags t_filter ON t_filter.id = it_filter.tag_id
				WHERE it_filter.item_id = i.id
				  AND t_filter.name = $%d
			)
		`, arg)
		args = append(args, tag)
		arg++
	}

	query += fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", arg, arg+1)
	args = append(args, limit, offset)

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]*models.Item, 0)
	for rows.Next() {
		item, err := scanItem(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

type scanner func(dest ...any) error

func scanItem(scan scanner) (*models.Item, error) {
	item := &models.Item{}
	var attributes []byte
	var tags []byte
	var purchasedAt sql.NullTime

	if err := scan(
		&item.ID,
		&item.Name,
		&item.ShortDescription,
		&item.Description,
		&item.ImageURL,
		&item.Price,
		&item.PriceCurrency,
		&item.USDExchangeRate,
		&purchasedAt,
		&attributes,
		&item.CreatedAt,
		&item.UpdatedAt,
		&tags,
	); err != nil {
		return nil, err
	}

	if purchasedAt.Valid {
		item.PurchasedAt = &purchasedAt.Time
	}

	if len(attributes) != 0 {
		if err := json.Unmarshal(attributes, &item.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}
	if item.Attributes == nil {
		item.Attributes = map[string]any{}
	}

	if len(tags) != 0 {
		if err := json.Unmarshal(tags, &item.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
	}
	if item.Tags == nil {
		item.Tags = make([]models.Tag, 0)
	}

	return item, nil
}

func (r *PostgresItemsRepository) Save(ctx context.Context, item *models.Item) error {
	attributes, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal item attributes: %w", err)
	}

	if item.ID == 0 {
		const query = `
			INSERT INTO items (
				name,
				short_description,
				description,
				image_url,
				price,
				price_currency,
				usd_exchange_rate,
				purchased_at,
				attributes
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at
		`

		if err := r.sql.QueryRowContext(
			ctx,
			query,
			item.Name,
			item.ShortDescription,
			item.Description,
			item.ImageURL,
			item.Price,
			item.PriceCurrency,
			item.USDExchangeRate,
			item.PurchasedAt,
			attributes,
		).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return fmt.Errorf("insert item: %w", err)
		}

		return nil
	}

	const query = `
		UPDATE items
		SET name = $1,
		    short_description = $2,
		    description = $3,
		    image_url = $4,
		    price = $5,
		    price_currency = $6,
		    usd_exchange_rate = $7,
		    purchased_at = $8,
		    attributes = $9,
		    updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`

	if err := r.sql.QueryRowContext(
		ctx,
		query,
		item.Name,
		item.ShortDescription,
		item.Description,
		item.ImageURL,
		item.Price,
		item.PriceCurrency,
		item.USDExchangeRate,
		item.PurchasedAt,
		attributes,
		item.ID,
	).Scan(&item.UpdatedAt); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func (r *PostgresItemsRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM items WHERE id = $1`

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

func (r *PostgresItemsRepository) AddTag(ctx context.Context, itemID, tagID int64) error {
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

func (r *PostgresItemsRepository) RemoveTag(ctx context.Context, itemID, tagID int64) error {
	const query = `
		DELETE FROM items_tags
		WHERE item_id = $1 AND tag_id = $2
	`

	if _, err := r.sql.ExecContext(ctx, query, itemID, tagID); err != nil {
		return fmt.Errorf("remove tag from item: %w", err)
	}
	return nil
}
