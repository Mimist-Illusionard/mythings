package repository

import (
	"context"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
)

type TagsRepository interface {
	List(ctx context.Context, name string, limit, offset int) ([]*models.Tag, error)
	Save(ctx context.Context, tag *models.Tag) error
	Delete(ctx context.Context, name string) error
}
