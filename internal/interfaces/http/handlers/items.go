package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
	"github.com/Mimist-Illusionard/mythings/internal/domain/ports/repository"
)

type itemRequest struct {
	Name             string         `json:"name"`
	ShortDescription string         `json:"short_description"`
	Description      string         `json:"description"`
	ImageURL         string         `json:"image_url"`
	Price            float64        `json:"price"`
	PriceCurrency    string         `json:"price_currency"`
	USDExchangeRate  float64        `json:"usd_exchange_rate"`
	PurchasedAt      string         `json:"purchased_at"`
	Attributes       map[string]any `json:"attributes,omitempty"`
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.items.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := repository.ItemFilter{
		Name: r.URL.Query().Get("name"),
		Tags: r.URL.Query()["tag"],
	}

	if value := r.URL.Query().Get("min_price"); value != "" {
		price, err := strconv.ParseFloat(value, 64)
		if err != nil || price < 0 {
			writeError(w, http.StatusBadRequest, "invalid min_price")
			return
		}
		filter.MinPrice = &price
	}

	if value := r.URL.Query().Get("max_price"); value != "" {
		price, err := strconv.ParseFloat(value, 64)
		if err != nil || price < 0 {
			writeError(w, http.StatusBadRequest, "invalid max_price")
			return
		}
		filter.MaxPrice = &price
	}

	items, err := h.items.List(r.Context(), filter, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	var req itemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	purchasedAt, currency, err := validateItemRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item := models.NewItem(req.Name, models.ItemParams{
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		ImageURL:         req.ImageURL,
		Price:            req.Price,
		PriceCurrency:    currency,
		USDExchangeRate:  req.USDExchangeRate,
		PurchasedAt:      purchasedAt,
		Attributes:       req.Attributes,
	})

	if err := h.items.Save(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req itemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	purchasedAt, currency, err := validateItemRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.items.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	item.Name = req.Name
	item.ShortDescription = req.ShortDescription
	item.Description = req.Description
	item.ImageURL = req.ImageURL
	item.Price = req.Price
	item.PriceCurrency = currency
	item.USDExchangeRate = req.USDExchangeRate
	item.PurchasedAt = purchasedAt
	if req.Attributes != nil {
		item.Attributes = req.Attributes
	}

	if err := h.items.Save(r.Context(), item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func validateItemRequest(req itemRequest) (*time.Time, string, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, "", errors.New("name is required")
	}
	if req.Price < 0 {
		return nil, "", errors.New("price must be greater than or equal to 0")
	}
	if req.USDExchangeRate <= 0 {
		return nil, "", errors.New("usd_exchange_rate must be greater than 0")
	}

	currency := strings.ToUpper(strings.TrimSpace(req.PriceCurrency))
	if currency == "" {
		currency = models.CurrencyRUB
	}
	if currency != models.CurrencyRUB && currency != models.CurrencyUSD {
		return nil, "", errors.New("price_currency must be RUB or USD")
	}

	if strings.TrimSpace(req.PurchasedAt) == "" {
		return nil, currency, nil
	}

	value, err := time.Parse("2006-01-02", req.PurchasedAt)
	if err != nil {
		return nil, "", errors.New("purchased_at must have YYYY-MM-DD format")
	}
	return &value, currency, nil
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.items.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addTagToItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	tagID, err := parseID(mux.Vars(r)["tagID"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tag id")
		return
	}

	if err := h.items.AddTag(r.Context(), itemID, tagID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeTagFromItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	tagID, err := parseID(mux.Vars(r)["tagID"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tag id")
		return
	}

	if err := h.items.RemoveTag(r.Context(), itemID, tagID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
