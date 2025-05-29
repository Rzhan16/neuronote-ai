package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateSchedule(t *testing.T) {
	// Setup
	app := fiber.New()
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Test data
	now := time.Now()
	req := CreateScheduleRequest{
		Notes: []struct {
			ID      string    `json:"id"`
			DueDate time.Time `json:"due_date"`
			Weight  float64   `json:"weight"`
		}{
			{
				ID:      "note1",
				DueDate: now.Add(24 * time.Hour),
				Weight:  1.0,
			},
		},
		Calendar: []struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
			Busy  bool      `json:"busy"`
		}{
			{
				Start: now,
				End:   now.Add(2 * time.Hour),
				Busy:  false,
			},
		},
	}

	// Expect transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO study_blocks").
		WithArgs(sqlmock.AnyArg(), "test-user", "note1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create request
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	request := httptest.NewRequest("POST", "/api/schedule", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", "test-user")

	// Test
	app.Post("/api/schedule", createSchedule)
	resp, err := app.Test(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response
	var blocks []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&blocks)
	assert.NoError(t, err)
	assert.NotEmpty(t, blocks)
}

func TestGetStudySchedule(t *testing.T) {
	// Setup
	app := fiber.New()
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Test data
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "note_id", "start_time", "end_time"}).
		AddRow("block1", "note1", now, now.Add(30*time.Minute))

	// Expect query
	mock.ExpectQuery("SELECT id, note_id, start_time, end_time FROM study_blocks").
		WithArgs("test-user").
		WillReturnRows(rows)

	// Create request
	request := httptest.NewRequest("GET", "/api/schedule", nil)
	request.Header.Set("X-User-ID", "test-user")

	// Test
	app.Get("/api/schedule", getStudySchedule)
	resp, err := app.Test(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response
	var blocks []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&blocks)
	assert.NoError(t, err)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "block1", blocks[0]["id"])
	assert.Equal(t, "note1", blocks[0]["note_id"])
	assert.Equal(t, "test-user", blocks[0]["user_id"])
}

func TestCreateSchedule_Validation(t *testing.T) {
	app := fiber.New()
	app.Post("/api/schedule", createSchedule)

	tests := []struct {
		name    string
		request CreateScheduleRequest
		status  int
		error   string
	}{
		{
			name: "empty notes",
			request: CreateScheduleRequest{
				Notes: []struct {
					ID      string    `json:"id"`
					DueDate time.Time `json:"due_date"`
					Weight  float64   `json:"weight"`
				}{},
				Calendar: []struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
					Busy  bool      `json:"busy"`
				}{
					{
						Start: time.Now(),
						End:   time.Now().Add(time.Hour),
						Busy:  false,
					},
				},
			},
			status: http.StatusBadRequest,
			error:  "At least one note is required",
		},
		{
			name: "empty calendar",
			request: CreateScheduleRequest{
				Notes: []struct {
					ID      string    `json:"id"`
					DueDate time.Time `json:"due_date"`
					Weight  float64   `json:"weight"`
				}{
					{
						ID:      "note1",
						DueDate: time.Now().Add(time.Hour),
						Weight:  1.0,
					},
				},
				Calendar: []struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
					Busy  bool      `json:"busy"`
				}{},
			},
			status: http.StatusBadRequest,
			error:  "Calendar slots are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/schedule", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "test-user")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.status, resp.StatusCode)

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, tt.error, result["error"])
		})
	}
}
