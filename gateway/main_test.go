package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() (*fiber.App, sqlmock.Sqlmock) {
	// Initialize ML client
	mlClient = NewMLClient()
	mlClient.baseURL = "http://localhost:8000" // Use local URL for testing

	// Create mock database
	var err error
	var mock sqlmock.Sqlmock
	db, mock, err = sqlmock.New()
	if err != nil {
		panic(err)
	}

	// Create Fiber app
	app := fiber.New()
	return app, mock
}

func TestHealthCheck(t *testing.T) {
	app, _ := setupTestApp()
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
	app, mock := setupTestApp()
	app.Post("/api/notes", uploadNote)

	// Mock the database insert
	mock.ExpectQuery(`INSERT INTO notes`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test-note-id"))

	// Create a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)

	content := []byte("This is a test note content.")
	_, err = part.Write(content)
	assert.NoError(t, err)
	writer.Close()

	// Create a test request
	req := httptest.NewRequest("POST", "/api/notes", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "test-user-id")

	// Test the endpoint
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Contains(t, result, "id")
}

func TestGetNotes(t *testing.T) {
	app, mock := setupTestApp()
	app.Get("/api/notes", getNotes)

	// Mock the database query
	rows := sqlmock.NewRows([]string{"id", "title", "content", "summary", "created_at", "updated_at", "quiz_cards"}).
		AddRow("test-id", "Test Note", "content", "summary", time.Now(), time.Now(), "[]")

	mock.ExpectQuery(`SELECT n.id, n.title, n.content, n.summary, n.created_at, n.updated_at`).
		WithArgs("test-user-id").
		WillReturnRows(rows)

	// Create a test request
	req := httptest.NewRequest("GET", "/api/notes", nil)
	req.Header.Set("X-User-ID", "test-user-id")

	// Test the endpoint
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var notes []Note
	err = json.NewDecoder(resp.Body).Decode(&notes)
	assert.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.Equal(t, "test-id", notes[0].ID)
}

func TestGetStudyBlocks(t *testing.T) {
	app, mock := setupTestApp()
	app.Get("/api/schedule", getStudyBlocks)

	// Mock the database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "note_id", "start_time", "end_time", "status"}).
		AddRow("block-1", "note-1", now, now.Add(time.Hour), "scheduled")

	mock.ExpectQuery(`SELECT id, start_time, end_time, note_id, status FROM study_blocks`).
		WithArgs("test-user-id").
		WillReturnRows(rows)

	// Create a test request
	req := httptest.NewRequest("GET", "/api/schedule", nil)
	req.Header.Set("X-User-ID", "test-user-id")

	// Test the endpoint
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var blocks []StudyBlock
	err = json.NewDecoder(resp.Body).Decode(&blocks)
	assert.NoError(t, err)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "block-1", blocks[0].ID)
}
