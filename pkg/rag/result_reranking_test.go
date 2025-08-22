package rag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vijaynallagatla/vjvector/pkg/core"
)

func TestNewResultRerankingManager(t *testing.T) {
	tests := []struct {
		name   string
		config *RerankingConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "custom config",
			config: &RerankingConfig{
				EnableSemanticReranking: false,
				EnableContextReranking:  true,
				EnableHybridScoring:     false,
				MaxRerankedResults:      50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewResultRerankingManager(tt.config)
			require.NotNil(t, manager)
			assert.NotNil(t, manager.config)
			assert.NotNil(t, manager.rerankers)
		})
	}
}

func TestResultRerankingManager_RerankResults(t *testing.T) {
	manager := NewResultRerankingManager(nil)
	require.NotNil(t, manager)

	// Create test results
	results := []*QueryResult{
		{
			Vector: &core.Vector{
				ID:        "vec1",
				Embedding: []float64{1.0, 0.0, 0.0},
			},
			Score:     0.8,
			Distance:  0.2,
			Relevance: 0.8,
		},
		{
			Vector: &core.Vector{
				ID:        "vec2",
				Embedding: []float64{0.0, 1.0, 0.0},
			},
			Score:     0.6,
			Distance:  0.4,
			Relevance: 0.6,
		},
	}

	query := &Query{
		Text: "test query",
		Context: map[string]interface{}{
			"domain": "technical",
		},
	}

	reranked, err := manager.RerankResults(context.Background(), results, query, nil)
	require.NoError(t, err)
	assert.NotNil(t, reranked)
	assert.Len(t, reranked, 2)
}

func TestResultRerankingManager_shouldUseReranker(t *testing.T) {
	manager := NewResultRerankingManager(&RerankingConfig{
		EnableSemanticReranking: false,
		EnableContextReranking:  true,
	})

	tests := []struct {
		name     string
		reranker ResultReranker
		query    *Query
		expected bool
	}{
		{
			name:     "semantic reranker disabled",
			reranker: &SemanticReranker{},
			query:    &Query{Text: "test"},
			expected: false,
		},
		{
			name:     "context-aware reranker with context",
			reranker: &ContextAwareReranker{},
			query: &Query{
				Text:    "test",
				Context: map[string]interface{}{"key": "value"},
			},
			expected: true,
		},
		{
			name:     "context-aware reranker without context",
			reranker: &ContextAwareReranker{},
			query:    &Query{Text: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.shouldUseReranker(tt.reranker, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResultRerankingManager_applyHybridScoring(t *testing.T) {
	manager := NewResultRerankingManager(&RerankingConfig{
		VectorWeight:   0.4,
		ContextWeight:  0.3,
		SemanticWeight: 0.3,
	})

	results := []*QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
		},
		{
			Vector: &core.Vector{ID: "vec2"},
			Score:  0.6,
		},
	}

	query := &Query{
		Text:    "test",
		Context: map[string]interface{}{"key": "value"},
	}

	scored := manager.applyHybridScoring(results, query, nil)
	require.NotNil(t, scored)
	assert.Len(t, scored, 2)

	// Scores should be updated with hybrid scoring
	assert.NotEqual(t, 0.8, scored[0].Score)
	assert.NotEqual(t, 0.6, scored[1].Score)
}

func TestResultRerankingManager_calculateContextScore(t *testing.T) {
	manager := NewResultRerankingManager(nil)

	query := &Query{
		Context: map[string]interface{}{
			"domain":  "technical",
			"user_id": "user123",
		},
	}

	result := &QueryResult{
		Context: map[string]interface{}{
			"domain":  "technical",
			"user_id": "user123",
		},
	}

	score := manager.calculateContextScore(result, query)
	assert.Greater(t, score, 0.0)
}

func TestResultRerankingManager_calculateSemanticScore(t *testing.T) {
	manager := NewResultRerankingManager(nil)

	query := &Query{Text: "programming language"}
	result := &QueryResult{
		Metadata: map[string]interface{}{
			"text": "Go programming language tutorial",
		},
	}

	score := manager.calculateSemanticScore(result, query)
	assert.Greater(t, score, 0.0)
}

func TestSemanticReranker(t *testing.T) {
	reranker := NewSemanticReranker(nil, nil)
	require.NotNil(t, reranker)

	results := []*QueryResult{
		{
			Vector: &core.Vector{
				ID:        "vec1",
				Embedding: []float64{1.0, 0.0, 0.0},
			},
			Score: 0.8,
		},
		{
			Vector: &core.Vector{
				ID:        "vec2",
				Embedding: []float64{0.0, 1.0, 0.0},
			},
			Score: 0.6,
		},
	}

	query := &Query{Text: "test query"}

	reranked, err := reranker.Rerank(context.Background(), results, query)
	require.NoError(t, err)
	assert.NotNil(t, reranked)
}

func TestSemanticReranker_cosineSimilarity(t *testing.T) {
	reranker := NewSemanticReranker(nil, nil)

	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{0.0, 1.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{-1.0, 0.0, 0.0},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := reranker.cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, similarity, 0.001)
		})
	}
}

func TestContextAwareReranker(t *testing.T) {
	reranker := NewContextAwareReranker(nil)
	require.NotNil(t, reranker)

	results := []*QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
		},
		{
			Vector: &core.Vector{ID: "vec2"},
			Score:  0.6,
		},
	}

	query := &Query{
		Text: "test",
		Context: map[string]interface{}{
			"domain": "technical",
		},
	}

	reranked, err := reranker.Rerank(context.Background(), results, query)
	require.NoError(t, err)
	assert.NotNil(t, reranked)
	assert.Len(t, reranked, 2)
}

func TestContextAwareReranker_calculateContextScore(t *testing.T) {
	reranker := NewContextAwareReranker(nil)

	query := &Query{
		Context: map[string]interface{}{
			"user_history": "recent",
			"domain":       "technical",
			"time_context": "current",
			"location":     "San Francisco",
		},
	}

	result := &QueryResult{
		Vector: &core.Vector{ID: "vec1"},
		Score:  0.8,
	}

	score := reranker.calculateContextScore(result, query)
	assert.Greater(t, score, 0.0)
}

func TestHybridReranker(t *testing.T) {
	reranker := NewHybridReranker(nil)
	require.NotNil(t, reranker)

	results := []*QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
		},
		{
			Vector: &core.Vector{ID: "vec2"},
			Score:  0.6,
		},
	}

	query := &Query{Text: "test query"}

	reranked, err := reranker.Rerank(context.Background(), results, query)
	require.NoError(t, err)
	assert.NotNil(t, reranked)
	assert.Len(t, reranked, 2)

	// Scores should be updated
	assert.NotEqual(t, 0.8, reranked[0].Score)
	assert.NotEqual(t, 0.6, reranked[1].Score)
}

func TestRerankerInterfaces(t *testing.T) {
	tests := []struct {
		name     string
		reranker ResultReranker
		expected string
	}{
		{
			name:     "semantic reranker",
			reranker: &SemanticReranker{},
			expected: "semantic",
		},
		{
			name:     "context-aware reranker",
			reranker: &ContextAwareReranker{},
			expected: "context-aware",
		},
		{
			name:     "hybrid reranker",
			reranker: &HybridReranker{},
			expected: "hybrid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.reranker.Type())
			assert.Greater(t, tt.reranker.Confidence(), 0.0)
			assert.LessOrEqual(t, tt.reranker.Confidence(), 1.0)
		})
	}
}

func TestRerankingConfigDefaults(t *testing.T) {
	manager := NewResultRerankingManager(nil)
	require.NotNil(t, manager)

	config := manager.config
	assert.True(t, config.EnableSemanticReranking)
	assert.True(t, config.EnableContextReranking)
	assert.True(t, config.EnableHybridScoring)
	assert.Equal(t, 0.4, config.SemanticWeight)
	assert.Equal(t, 0.3, config.ContextWeight)
	assert.Equal(t, 0.3, config.VectorWeight)
	assert.Equal(t, 0.5, config.MinRerankingConfidence)
	assert.Equal(t, 100, config.MaxRerankedResults)
}

func TestRerankingManagerWithCustomConfig(t *testing.T) {
	config := &RerankingConfig{
		EnableSemanticReranking: false,
		EnableContextReranking:  true,
		EnableHybridScoring:     false,
		MaxRerankedResults:      25,
	}

	manager := NewResultRerankingManager(config)
	require.NotNil(t, manager)

	results := []*QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
		},
	}

	query := &Query{
		Text:    "test",
		Context: map[string]interface{}{"key": "value"},
	}

	reranked, err := manager.RerankResults(context.Background(), results, query, nil)
	require.NoError(t, err)
	assert.NotNil(t, reranked)
}

func TestRerankingWithEmptyResults(t *testing.T) {
	manager := NewResultRerankingManager(nil)
	require.NotNil(t, manager)

	// Empty results
	results := []*QueryResult{}
	query := &Query{Text: "test"}

	reranked, err := manager.RerankResults(context.Background(), results, query, nil)
	require.NoError(t, err)
	assert.Empty(t, reranked)

	// Single result
	results = []*QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
		},
	}

	reranked, err = manager.RerankResults(context.Background(), results, query, nil)
	require.NoError(t, err)
	assert.Len(t, reranked, 1)
}
