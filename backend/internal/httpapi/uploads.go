package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadBytes = 5 << 20

var uploadTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func saveUpload(dir string, w http.ResponseWriter, r *http.Request) (string, int, string) {
	if dir == "" {
		return "", http.StatusServiceUnavailable, "Photo upload is not configured"
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return "", http.StatusRequestEntityTooLarge, "Photo must be 5 MB or smaller"
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return "", http.StatusUnprocessableEntity, "Choose a photo"
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(file, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", http.StatusBadRequest, "Photo could not be read"
	}
	kind := http.DetectContentType(header[:n])
	ext, ok := uploadTypes[kind]
	if !ok {
		return "", http.StatusUnprocessableEntity, "Upload a JPEG, PNG, or WebP photo"
	}
	name, err := randomName(ext)
	if err != nil {
		return "", http.StatusInternalServerError, "Something went wrong"
	}
	target := filepath.Join(dir, name)
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", http.StatusInternalServerError, "Something went wrong"
	}
	defer out.Close()
	if _, err := out.Write(header[:n]); err != nil {
		os.Remove(target)
		return "", http.StatusInternalServerError, "Something went wrong"
	}
	if _, err := io.Copy(out, file); err != nil {
		os.Remove(target)
		return "", http.StatusInternalServerError, "Something went wrong"
	}
	return "/uploads/" + name, http.StatusCreated, ""
}

func randomName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}

func serveUpload(dir string, w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if dir == "" || name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(dir, name))
}
