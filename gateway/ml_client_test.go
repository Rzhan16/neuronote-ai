package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*MLClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	serverURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	client := NewMLClient()
	client.client.Transport = &http.Transport{
		Proxy: http.ProxyURL(serverURL),
	}
	client.baseURL = server.URL

	t.Cleanup(func() {
		server.Close()
	})

	return client, server
}

func TestMLClient_Pipeline(t *testing.T) {
	client, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Check method and path
		assert.Equal(t, "/pipeline", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Check headers
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		assert.Equal(t, "test-user", r.Header.Get("X-User-ID"))

		// Check file
		file, header, err := r.FormFile("file")
		assert.NoError(t, err)
		assert.Equal(t, "test.txt", header.Filename)
		content, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, "test content", string(content))

		// Send response
		json.NewEncoder(w).Encode(PipelineResponse{NoteID: "test-note-id"})
	})

	// Test request
	noteID, err := client.Pipeline(
		bytes.NewBufferString("test content"),
		"test.txt",
		"test-user",
	)

	// Check results
	assert.NoError(t, err)
	assert.Equal(t, "test-note-id", noteID)
}

func TestMLClient_OCR(t *testing.T) {
	client, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ocr", r.URL.Path)
		json.NewEncoder(w).Encode(OCRResponse{
			Blocks: []Block{
				{
					Text:       "test text",
					Confidence: 0.95,
					BBox:       []float64{1, 2, 3, 4},
				},
			},
		})
	})

	blocks, err := client.OCR(
		bytes.NewBufferString("test image"),
		"test.jpg",
		"test-user",
	)

	assert.NoError(t, err)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "test text", blocks[0].Text)
	assert.Equal(t, 0.95, blocks[0].Confidence)
	assert.Equal(t, []float64{1, 2, 3, 4}, blocks[0].BBox)
}

func TestMLClient_ASR(t *testing.T) {
	client, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/asr", r.URL.Path)
		json.NewEncoder(w).Encode(ASRResponse{
			Transcript: "test transcript",
		})
	})

	transcript, err := client.ASR(
		bytes.NewBufferString("test audio"),
		"test.wav",
		"test-user",
	)

	assert.NoError(t, err)
	assert.Equal(t, "test transcript", transcript)
}

func TestMLClient_ErrorHandling(t *testing.T) {
	client, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
	})

	_, err := client.Pipeline(
		bytes.NewBufferString("test content"),
		"test.txt",
		"test-user",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestMLClient_Timeout(t *testing.T) {
	client, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second)
	})

	_, err := client.Pipeline(
		bytes.NewBufferString("test content"),
		"test.txt",
		"test-user",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}
