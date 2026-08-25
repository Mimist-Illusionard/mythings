package repository

import (
	"context"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
)

type ItemsRepository interface {
	List(ctx context.Context, tags []string) ([]models.Item, error)
	Save(ctx context.Context, item *models.Item) error
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id int) error
}
