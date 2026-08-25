package models

import "time"

const (
	CurrencyRUB = "RUB"
	CurrencyUSD = "USD"
)

type Item struct {
	ID int64 `json:"id"`

	Name             string         `json:"name"`
	ShortDescription string         `json:"short_description"`
	Description      string         `json:"description"`
	ImageURL         string         `json:"image_url"`
	Price            float64        `json:"price"`
	PriceCurrency    string         `json:"price_currency"`
	USDExchangeRate  float64        `json:"usd_exchange_rate"`
	PurchasedAt      *time.Time     `json:"purchased_at,omitempty"`
	Attributes       map[string]any `json:"attributes"`
	Tags             []Tag          `json:"tags"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemParams struct {
	ShortDescription string
	Description      string
	ImageURL         string
	Price            float64
	PriceCurrency    string
	USDExchangeRate  float64
	PurchasedAt      *time.Time
	Attributes       map[string]any
}

func NewItem(name string, params ItemParams) *Item {
	currency := params.PriceCurrency
	if currency == "" {
		currency = CurrencyRUB
	}

	attributes := params.Attributes
	if attributes == nil {
		attributes = map[string]any{}
	}

	return &Item{
		Name:             name,
		ShortDescription: params.ShortDescription,
		Description:      params.Description,
		ImageURL:         params.ImageURL,
		Price:            params.Price,
		PriceCurrency:    currency,
		USDExchangeRate:  params.USDExchangeRate,
		PurchasedAt:      params.PurchasedAt,
		Attributes:       attributes,
		Tags:             make([]Tag, 0),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
