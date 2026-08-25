package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
	"github.com/Mimist-Illusionard/mythings/internal/domain/ports/repository"
)

type Items interface {
	GetByID(ctx context.Context, id int64) (*models.Item, error)
	List(ctx context.Context, filter repository.ItemFilter, limit, offset int) ([]*models.Item, error)
	Save(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id int64) error
	AddTag(ctx context.Context, itemID, tagID int64) error
	RemoveTag(ctx context.Context, itemID, tagID int64) error
}

type Tags interface {
	List(ctx context.Context, name string, limit, offset int) ([]*models.Tag, error)
	Save(ctx context.Context, tag *models.Tag) error
	Delete(ctx context.Context, name string) error
}

type Handler struct {
	items     Items
	tags      Tags
	uploadDir string
}

func New(items Items, tags Tags, uploadDir string) *Handler {
	return &Handler{
		items:     items,
		tags:      tags,
		uploadDir: uploadDir,
	}
}

func (h *Handler) Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/items", h.listItems).Methods(http.MethodGet)
	router.HandleFunc("/items/{id:[0-9]+}", h.getItem).Methods(http.MethodGet)
	router.HandleFunc("/items", h.createItem).Methods(http.MethodPost)
	router.HandleFunc("/items/{id:[0-9]+}", h.updateItem).Methods(http.MethodPut)
	router.HandleFunc("/items/{id:[0-9]+}", h.deleteItem).Methods(http.MethodDelete)

	router.HandleFunc("/items/{id:[0-9]+}/tags/{tagID:[0-9]+}", h.addTagToItem).Methods(http.MethodPost)
	router.HandleFunc("/items/{id:[0-9]+}/tags/{tagID:[0-9]+}", h.removeTagFromItem).Methods(http.MethodDelete)

	router.HandleFunc("/tags", h.listTags).Methods(http.MethodGet)
	router.HandleFunc("/tags", h.createTag).Methods(http.MethodPost)
	router.HandleFunc("/tags/{id:[0-9]+}", h.updateTag).Methods(http.MethodPut)
	router.HandleFunc("/tags/{name}", h.deleteTag).Methods(http.MethodDelete)

	router.HandleFunc("/uploads", h.uploadImage).Methods(http.MethodPost)

	return router
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func parsePagination(r *http.Request) (limit, offset int, err error) {
	limit = 50
	offset = 0

	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 100 {
			return 0, 0, errors.New("limit must be between 1 and 100")
		}
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			return 0, 0, errors.New("offset must be greater than or equal to 0")
		}
	}

	return limit, offset, nil
}
