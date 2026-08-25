package repository

import (
	"context"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
)

type ItemFilter struct {
	Name     string
	Tags     []string
	MinPrice *int64
	MaxPrice *int64
}

type ItemsRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Item, error)
	List(ctx context.Context, filter ItemFilter, limit, offset int) ([]*models.Item, error)

	Save(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id int64) error

	AddTag(ctx context.Context, itemID, tagID int64) error
	RemoveTag(ctx context.Context, itemID, tagID int64) error
}
