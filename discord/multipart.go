package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// buildHTTPRequest constructs an HTTP POST request for Discord, supporting both JSON and multipart form-data for file attachments.
func buildHTTPRequest(ctx context.Context, apiURL string, authHeader string, payload any, filePath string) (*http.Request, error) {
	if filePath == "" {
		// Standard JSON payload
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encode JSON payload: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		return req, nil
	}

	// Multipart form-data payload with file attachment
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment file %q: %w", filePath, err)
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JSON payload: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add payload_json field
	if err := writer.WriteField("payload_json", string(jsonBytes)); err != nil {
		return nil, fmt.Errorf("failed to write payload_json field: %w", err)
	}

	// Add file field
	part, err := writer.CreateFormFile("files[0]", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file content to form: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	return req, nil
}
