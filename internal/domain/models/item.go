package models

import "time"

type Item struct {
	ID int64 `json:"id"`

	Name        string         `json:"name"`
	Description string         `json:"description"`
	ImageURL    string         `json:"image_url"`
	Price       int64          `json:"price"`
	Attributes  map[string]any `json:"attributes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
