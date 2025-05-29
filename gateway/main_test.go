package main

import (
	"bytes"
	"database/sql"
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
	// Start a test ML server
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pipeline", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20)
		assert.NoError(t, err)

		// Check file
		file, header, err := r.FormFile("file")
		assert.NoError(t, err)
		assert.Equal(t, "test.txt", header.Filename)
		defer file.Close()

		// Return a mock response
		json.NewEncoder(w).Encode(map[string]string{
			"note_id": "test-note-id",
		})
	}))
	defer mlServer.Close()

	// Set up the app with mock ML service
	app, _ := setupTestApp()
	mlClient.baseURL = mlServer.URL
	app.Post("/api/notes", uploadNote)

	// Create a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)

	content := []byte("This is a test note")
	_, err = part.Write(content)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/notes", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "test-user")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test-note-id", result["note_id"])
}

func TestGetNote(t *testing.T) {
	app, mock := setupTestApp()
	app.Get("/api/notes/:id", getNote)

	// Test non-existent note
	mock.ExpectQuery("SELECT id, content, summary, created_at, updated_at FROM notes WHERE id = \\$1").
		WithArgs("test-id").
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/notes/test-id", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test existing note
	rows := sqlmock.NewRows([]string{"id", "content", "summary", "created_at", "updated_at"}).
		AddRow("test-id", "test content", "test summary", time.Now(), time.Now())
	mock.ExpectQuery("SELECT id, content, summary, created_at, updated_at FROM notes WHERE id = \\$1").
		WithArgs("test-id-2").
		WillReturnRows(rows)

	// Mock quiz cards
	cardRows := sqlmock.NewRows([]string{"id", "question", "answer"}).
		AddRow("card-1", "test question", "test answer")
	mock.ExpectQuery("SELECT id, question, answer FROM quiz_cards WHERE note_id = \\$1").
		WithArgs("test-id-2").
		WillReturnRows(cardRows)

	req = httptest.NewRequest("GET", "/api/notes/test-id-2", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var note Note
	err = json.NewDecoder(resp.Body).Decode(&note)
	assert.NoError(t, err)
	assert.Equal(t, "test-id", note.ID)
	assert.Equal(t, "test content", note.Content)
	assert.Equal(t, "test summary", note.Summary)
	assert.Len(t, note.QuizCards, 1)
	assert.Equal(t, "card-1", note.QuizCards[0].ID)
}

func TestGetSchedule(t *testing.T) {
	app, mock := setupTestApp()
	app.Get("/api/schedule", getSchedule)

	// Mock schedule rows
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "note_id", "start_time", "end_time", "status"}).
		AddRow("block-1", "note-1", now, now.Add(time.Hour), "pending")
	mock.ExpectQuery("SELECT id, note_id, start_time, end_time, status FROM study_blocks").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/api/schedule", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var blocks []StudyBlock
	err = json.NewDecoder(resp.Body).Decode(&blocks)
	assert.NoError(t, err)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "block-1", blocks[0].ID)
	assert.Equal(t, "note-1", blocks[0].NoteID)
	assert.Equal(t, "pending", blocks[0].Status)
}
