package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
	"github.com/Mimist-Illusionard/mythings/internal/domain/ports/repository"
)

type itemRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ImageURL    string         `json:"image_url"`
	Price       int64          `json:"price"`
	Attributes  map[string]any `json:"attributes"`
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
		price, err := strconv.ParseInt(value, 10, 64)
		if err != nil || price < 0 {
			writeError(w, http.StatusBadRequest, "invalid min_price")
			return
		}
		filter.MinPrice = &price
	}

	if value := r.URL.Query().Get("max_price"); value != "" {
		price, err := strconv.ParseInt(value, 10, 64)
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

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be greater than or equal to 0")
		return
	}

	item := models.NewItem(req.Name, models.ItemParams{
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Attributes:  req.Attributes,
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

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be greater than or equal to 0")
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
	item.Description = req.Description
	item.ImageURL = req.ImageURL
	item.Price = req.Price
	item.Attributes = req.Attributes

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
