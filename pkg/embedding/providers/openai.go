package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/utils/logger"
)

// OpenAIProvider implements the OpenAI embedding provider
type OpenAIProvider struct {
	config  *embedding.ProviderConfig
	client  *http.Client
	baseURL string
	apiKey  string
}

// OpenAIEmbeddingRequest represents the OpenAI API request
type OpenAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// OpenAIEmbeddingResponse represents the OpenAI API response
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *embedding.ProviderConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &OpenAIProvider{
		config:  config,
		client:  client,
		baseURL: baseURL,
		apiKey:  config.APIKey,
	}, nil
}

// Type returns the provider type
func (p *OpenAIProvider) Type() embedding.ProviderType {
	return embedding.ProviderTypeOpenAI
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// GenerateEmbeddings generates embeddings using OpenAI API
func (p *OpenAIProvider) GenerateEmbeddings(ctx context.Context, req *embedding.EmbeddingRequest) (*embedding.EmbeddingResponse, error) {
	if len(req.Texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Use default model if not specified
	model := req.Model
	if model == "" {
		model = "text-embedding-ada-002"
	}

	// Process texts in batches
	var allEmbeddings [][]float64
	var totalTokens int
	var totalCost float64

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // OpenAI's recommended batch size
	}

	for i := 0; i < len(req.Texts); i += batchSize {
		end := i + batchSize
		if end > len(req.Texts) {
			end = len(req.Texts)
		}

		batch := req.Texts[i:end]
		batchEmbeddings, batchTokens, batchCost, err := p.processBatch(ctx, batch, model)
		if err != nil {
			return nil, fmt.Errorf("batch processing failed: %w", err)
		}

		allEmbeddings = append(allEmbeddings, batchEmbeddings...)
		totalTokens += batchTokens
		totalCost += batchCost
	}

	response := &embedding.EmbeddingResponse{
		Embeddings: allEmbeddings,
		Model:      model,
		Provider:   embedding.ProviderTypeOpenAI,
		Usage: embedding.UsageStats{
			TotalTokens: totalTokens,
			TotalCost:   totalCost,
			Provider:    "OpenAI",
		},
		CacheHit:       false,
		ProcessingTime: 0, // Will be set by service layer
	}

	return response, nil
}

// processBatch processes a batch of texts
func (p *OpenAIProvider) processBatch(ctx context.Context, texts []string, model string) ([][]float64, int, float64, error) {
	// Create request payload
	payload := OpenAIEmbeddingRequest{
		Input: texts[0], // OpenAI API expects single input for now
		Model: model,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/embeddings", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Make request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("Failed to close response body", "error", err)
		}
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, 0, 0, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var openAIResp OpenAIEmbeddingResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract embeddings
	embeddings := make([][]float64, len(openAIResp.Data))
	for i, data := range openAIResp.Data {
		embeddings[i] = data.Embedding
	}

	// Calculate cost (approximate)
	cost := p.calculateCost(openAIResp.Usage.TotalTokens, model)

	return embeddings, openAIResp.Usage.TotalTokens, cost, nil
}

// calculateCost calculates the cost for the given tokens and model
func (p *OpenAIProvider) calculateCost(tokens int, model string) float64 {
	var costPer1K float64

	switch model {
	case "text-embedding-ada-002":
		costPer1K = 0.0001 // $0.0001 per 1K tokens
	case "text-embedding-3-small":
		costPer1K = 0.00002 // $0.00002 per 1K tokens
	case "text-embedding-3-large":
		costPer1K = 0.00013 // $0.00013 per 1K tokens
	default:
		costPer1K = 0.0001 // Default to ada-002 pricing
	}

	return float64(tokens) * costPer1K / 1000.0
}

// GetModels returns available OpenAI models
func (p *OpenAIProvider) GetModels(ctx context.Context) ([]embedding.Model, error) {
	models := []embedding.Model{
		{
			ID:         "text-embedding-ada-002",
			Name:       "text-embedding-ada-002",
			Provider:   embedding.ProviderTypeOpenAI,
			Dimensions: 1536,
			MaxTokens:  8191,
			CostPer1K:  0.0001,
			Supported:  true,
		},
		{
			ID:         "text-embedding-3-small",
			Name:       "text-embedding-3-small",
			Provider:   embedding.ProviderTypeOpenAI,
			Dimensions: 1536,
			MaxTokens:  8191,
			CostPer1K:  0.00002,
			Supported:  true,
		},
		{
			ID:         "text-embedding-3-large",
			Name:       "text-embedding-3-large",
			Provider:   embedding.ProviderTypeOpenAI,
			Dimensions: 3072,
			MaxTokens:  8191,
			CostPer1K:  0.00013,
			Supported:  true,
		},
	}

	return models, nil
}

// GetCapabilities returns provider capabilities
func (p *OpenAIProvider) GetCapabilities() embedding.Capabilities {
	return embedding.Capabilities{
		MaxBatchSize:      100,
		MaxTextLength:     8191,
		SupportsAsync:     false,
		SupportsStreaming: false,
		RateLimit: embedding.RateLimit{
			RequestsPerMinute: 3500,
			TokensPerMinute:   350000,
			RequestsPerDay:    5000000,
			TokensPerDay:      500000000,
		},
		Features: []string{"embeddings", "text-ada-002", "text-embedding-3"},
	}
}

// HealthCheck checks if the provider is healthy
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	// Simple health check by making a minimal request
	req := &embedding.EmbeddingRequest{
		Texts:    []string{"test"},
		Model:    "text-embedding-ada-002",
		Provider: embedding.ProviderTypeOpenAI,
	}

	_, err := p.GenerateEmbeddings(ctx, req)
	return err
}

// Close closes the provider
func (p *OpenAIProvider) Close() error {
	// Nothing to close for HTTP client
	return nil
}
