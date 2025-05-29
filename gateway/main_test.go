package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/health", healthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.True(t, result["ok"])
}

func TestUploadNote(t *testing.T) {
	app := fiber.New()
	app.Post("/api/notes/upload", uploadNote)

	// Create a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)

	content := []byte("This is a test note")
	_, err = part.Write(content)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/notes/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetNote(t *testing.T) {
	app := fiber.New()
	app.Get("/api/notes/:id", getNote)

	req := httptest.NewRequest("GET", "/api/notes/test-id", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Should return 404 for non-existent note
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetSchedule(t *testing.T) {
	app := fiber.New()
	app.Get("/api/schedule", getSchedule)

	req := httptest.NewRequest("GET", "/api/schedule", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result []StudyBlock
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
}
