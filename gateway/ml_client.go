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

// MLClient handles communication with the ML service
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMLClient creates a new ML service client
func NewMLClient() *MLClient {
	return &MLClient{
		baseURL: mlBaseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 300, // 5 minutes timeout for long-running ML tasks
		},
	}
}

// Pipeline processes a file through the ML pipeline
func (c *MLClient) Pipeline(file io.Reader, filename string, userID string) (string, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	// Add user ID
	err = writer.WriteField("user_id", userID)
	if err != nil {
		return "", fmt.Errorf("failed to write user_id field: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/pipeline", c.baseURL), body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ML service error: %s", string(respBody))
	}

	// Parse response
	var result struct {
		NoteID string `json:"note_id"`
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return result.NoteID, nil
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
	resp, err := c.httpClient.Do(req)
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
