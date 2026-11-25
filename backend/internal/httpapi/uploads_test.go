package httpapi

import (
	"bytes"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSaveUploadAcceptsJPEG(t *testing.T) {
	dir := t.TempDir()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "box.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(part, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	url, status, message := saveUpload(dir, rec, req)
	if status != http.StatusCreated || message != "" || len(url) < len("/uploads/") {
		t.Fatalf("url=%q status=%d message=%q", url, status, message)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("files=%v err=%v", entries, err)
	}
}

func TestSaveUploadRejectsText(t *testing.T) {
	dir := t.TempDir()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("not a photo"))
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_, status, message := saveUpload(dir, httptest.NewRecorder(), req)
	if status != http.StatusUnprocessableEntity || message != "Upload a JPEG, PNG, or WebP photo" {
		t.Fatalf("status=%d message=%q", status, message)
	}
}
