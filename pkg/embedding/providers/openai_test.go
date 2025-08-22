package providers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *embedding.ProviderConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &embedding.ProviderConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
				Timeout: 30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: &embedding.ProviderConfig{
				BaseURL: "https://api.openai.com/v1",
				Timeout: 30 * time.Second,
			},
			expectError: true,
		},
		{
			name:        "empty config",
			config:      &embedding.ProviderConfig{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, embedding.ProviderTypeOpenAI, provider.Type())
				assert.Equal(t, "OpenAI", provider.Name())
			}
		})
	}
}

func TestOpenAIProvider_Type(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	assert.Equal(t, embedding.ProviderTypeOpenAI, provider.Type())
}

func TestOpenAIProvider_Name(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	assert.Equal(t, "OpenAI", provider.Name())
}

func TestOpenAIProvider_GetModels(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	models, err := provider.GetModels(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, models)

	// Check for expected models
	modelNames := make(map[string]bool)
	for _, model := range models {
		modelNames[model.ID] = true
		assert.Equal(t, embedding.ProviderTypeOpenAI, model.Provider)
		assert.True(t, model.Supported)
		assert.Greater(t, model.Dimensions, 0)
		assert.Greater(t, model.MaxTokens, 0)
		assert.GreaterOrEqual(t, model.CostPer1K, 0.0)
	}

	// Verify specific models exist
	expectedModels := []string{
		"text-embedding-ada-002",
		"text-embedding-3-small",
		"text-embedding-3-large",
	}
	for _, expected := range expectedModels {
		assert.True(t, modelNames[expected], "Expected model %s not found", expected)
	}
}

func TestOpenAIProvider_GetCapabilities(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	capabilities := provider.GetCapabilities()
	assert.Equal(t, 100, capabilities.MaxBatchSize)
	assert.Equal(t, 8191, capabilities.MaxTextLength)
	assert.False(t, capabilities.SupportsAsync)
	assert.False(t, capabilities.SupportsStreaming)
	assert.NotEmpty(t, capabilities.Features)

	// Check rate limits
	rateLimit := capabilities.RateLimit
	assert.Greater(t, rateLimit.RequestsPerMinute, 0)
	assert.Greater(t, rateLimit.TokensPerMinute, 0)
	assert.Greater(t, rateLimit.RequestsPerDay, 0)
	assert.Greater(t, rateLimit.TokensPerDay, 0)
}

func TestOpenAIProvider_CalculateCost(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Test cost calculation for different models
	testCases := []struct {
		model     string
		tokens    int
		expected  float64
		tolerance float64
	}{
		{"text-embedding-ada-002", 1000, 0.0001, 0.00001},
		{"text-embedding-3-small", 1000, 0.00002, 0.00001},
		{"text-embedding-3-large", 1000, 0.00013, 0.00001},
		{"unknown-model", 1000, 0.0001, 0.00001}, // Should default to ada-002
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			cost := provider.calculateCost(tc.tokens, tc.model)
			assert.InDelta(t, tc.expected, cost, tc.tolerance)
		})
	}
}

func TestOpenAIProvider_GenerateEmbeddings_Validation(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with empty texts
	req := &embedding.EmbeddingRequest{
		Texts:    []string{},
		Model:    "text-embedding-ada-002",
		Provider: embedding.ProviderTypeOpenAI,
	}

	_, err = provider.GenerateEmbeddings(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no texts provided")

	// Test with nil texts
	req.Texts = nil
	_, err = provider.GenerateEmbeddings(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no texts provided")
}

func TestOpenAIProvider_Close(t *testing.T) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Close should not error
	err = provider.Close()
	assert.NoError(t, err)
}

// Benchmark tests for performance
func BenchmarkOpenAIProvider_CalculateCost(b *testing.B) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.calculateCost(1000, "text-embedding-ada-002")
	}
}

func BenchmarkOpenAIProvider_GetCapabilities(b *testing.B) {
	config := &embedding.ProviderConfig{
		APIKey: "test-key",
	}
	provider, err := NewOpenAIProvider(config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.GetCapabilities()
	}
}
