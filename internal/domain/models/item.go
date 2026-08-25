package models

import "time"

type Item struct {
	ID int64

	Name        string
	Description string
	ImageURL    string
	Price       int64
	Attributes  map[string]any

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ItemParams struct {
	Description string
	ImageURL    string
	Price       int64
	Attributes  map[string]any
}

func NewItem(name string, params ItemParams) *Item {
	return &Item{
		Name:        name,
		Description: params.Description,
		ImageURL:    params.ImageURL,
		Price:       params.Price,
		Attributes:  params.Attributes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
