package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Mimist-Illusionard/mythings/internal/domain/models"
)

type tagRequest struct {
	Name string `json:"name"`
}

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tags, err := h.tags.List(r.Context(), r.URL.Query().Get("name"), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	writeJSON(w, http.StatusOK, tags)
}

func (h *Handler) createTag(w http.ResponseWriter, r *http.Request) {
	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	tag := models.NewTag(req.Name)
	if err := h.tags.Save(r.Context(), tag); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	writeJSON(w, http.StatusCreated, tag)
}

func (h *Handler) updateTag(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	tag := &models.Tag{
		ID:   id,
		Name: req.Name,
	}

	if err := h.tags.Save(r.Context(), tag); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tag")
		return
	}

	writeJSON(w, http.StatusOK, tag)
}

func (h *Handler) deleteTag(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(mux.Vars(r)["name"])
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.tags.Delete(r.Context(), name); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
