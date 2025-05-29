package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const timeout = 5 * time.Second

type MLClient struct {
	client  *http.Client
	baseURL string
}

type Block struct {
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	BBox       []float64 `json:"bbox"`
}

type OCRResponse struct {
	Blocks []Block `json:"blocks"`
}

type ASRResponse struct {
	Transcript string `json:"transcript"`
}

type PipelineResponse struct {
	NoteID string `json:"note_id"`
}

func NewMLClient() *MLClient {
	return &MLClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: "http://ml:8000",
	}
}

func (c *MLClient) Pipeline(file io.Reader, filename string, userID string) (string, error) {
	var response PipelineResponse
	err := c.sendFileRequest("/pipeline", file, filename, userID, &response)
	if err != nil {
		return "", fmt.Errorf("pipeline request failed: %w", err)
	}
	return response.NoteID, nil
}

func (c *MLClient) OCR(file io.Reader, filename string, userID string) ([]Block, error) {
	var response OCRResponse
	err := c.sendFileRequest("/ocr", file, filename, userID, &response)
	if err != nil {
		return nil, fmt.Errorf("OCR request failed: %w", err)
	}
	return response.Blocks, nil
}

func (c *MLClient) ASR(file io.Reader, filename string, userID string) (string, error) {
	var response ASRResponse
	err := c.sendFileRequest("/asr", file, filename, userID, &response)
	if err != nil {
		return "", fmt.Errorf("ASR request failed: %w", err)
	}
	return response.Transcript, nil
}

func (c *MLClient) sendFileRequest(endpoint string, file io.Reader, filename string, userID string, response interface{}) error {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
