package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 10 << 20

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1<<20)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "image is too large or multipart body is invalid")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read image")
		return
	}
	if len(data) > maxUploadSize {
		writeError(w, http.StatusRequestEntityTooLarge, "image must be smaller than 10 MB")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "image is empty")
		return
	}

	contentType := http.DetectContentType(data)
	ext, ok := imageExtensions[contentType]
	if !ok {
		writeError(w, http.StatusBadRequest, "supported image formats: JPEG, PNG, WEBP, GIF")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare upload directory")
		return
	}

	name, err := randomFileName(ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate image name")
		return
	}

	path := filepath.Join(h.uploadDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save image")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": "/uploads/" + name,
	})
}

func randomFileName(ext string) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random bytes: %w", err)
	}
	return hex.EncodeToString(buf) + ext, nil
}
