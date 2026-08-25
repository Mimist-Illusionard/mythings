package models

import "time"

type Item struct {
	id          int64
	Name        string
	Description string
	ImageURL    string
	Price       int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ItemParams struct {
	Description string
	ImageURL    string
	Price       int64
}

func NewItem(name string, params *ItemParams) *Item {
	return &Item{
		Name:        name,
		Description: params.Description,
		ImageURL:    params.ImageURL,
		Price:       params.Price,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (item *Item) ID() int64 {
	return item.id
}
