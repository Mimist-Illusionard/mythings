package repository

import "context"

type TagsRepository interface {
	List(ctx context.Context) ([]string, error)
	Save(ctx context.Context, tag string) error
	Delete(ctx context.Context, tag string) error
}
