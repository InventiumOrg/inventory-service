package processors

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type SchemaRegistryClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

type SchemaResponse struct {
	Schema string `json:"schema"`
}

func NewSchemaRegistryClient(baseURL, username, password string) *SchemaRegistryClient {
	return &SchemaRegistryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		username: username,
		password: password,
	}
}

func (src *SchemaRegistryClient) GetSchema(schemaID int) (string, error) {
	url := fmt.Sprintf("%s/schemas/ids/%d", src.baseURL, schemaID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if provided
	if src.username != "" && src.password != "" {
		req.SetBasicAuth(src.username, src.password)
	}

	req.Header.Set("Accept", "application/vnd.schemaregistry.v1+json")

	resp, err := src.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch schema from registry: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("schema registry returned status %d: %s", resp.StatusCode, string(body))
	}

	var schemaResponse SchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&schemaResponse); err != nil {
		return "", fmt.Errorf("failed to decode schema response: %w", err)
	}

	return schemaResponse.Schema, nil
}
