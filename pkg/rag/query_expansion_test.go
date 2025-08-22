package rag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryExpansionManager(t *testing.T) {
	tests := []struct {
		name   string
		config *ExpansionConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "custom config",
			config: &ExpansionConfig{
				MaxExpansionTerms:  10,
				MinConfidence:      0.5,
				EnableSemantic:     false,
				EnableSynonym:      true,
				EnableContextAware: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewQueryExpansionManager(tt.config)
			require.NotNil(t, manager)
			assert.NotNil(t, manager.config)
			assert.NotNil(t, manager.expanders)
		})
	}
}

func TestQueryExpansionManager_ExpandQuery(t *testing.T) {
	manager := NewQueryExpansionManager(nil)
	require.NotNil(t, manager)

	query := &Query{
		Text: "fast car",
		Context: map[string]interface{}{
			"domain": "automotive",
		},
	}

	// Mock embedding service
	expansions, err := manager.ExpandQuery(context.Background(), query, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, expansions)

	// Should include synonym expansions
	assert.Contains(t, expansions, "quick")
	assert.Contains(t, expansions, "rapid")
}

func TestSynonymExpander(t *testing.T) {
	expander := NewSynonymExpander()
	require.NotNil(t, expander)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "basic synonyms",
			query: &Query{
				Text: "fast car",
			},
			expected: []string{"quick", "rapid", "swift", "speedy"},
		},
		{
			name: "no synonyms",
			query: &Query{
				Text: "xyz abc",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expansions, err := expander.Expand(context.Background(), tt.query)
			require.NoError(t, err)

			for _, expected := range tt.expected {
				assert.Contains(t, expansions, expected)
			}
		})
	}
}

func TestSemanticExpander(t *testing.T) {
	expander := NewSemanticExpander(nil, nil)
	require.NotNil(t, expander)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "how to query",
			query: &Query{
				Text: "how to build a car",
			},
			expected: []string{"tutorial", "guide", "instructions", "steps"},
		},
		{
			name: "what is query",
			query: &Query{
				Text: "what is a vector database",
			},
			expected: []string{"definition", "explanation", "description", "meaning"},
		},
		{
			name: "best query",
			query: &Query{
				Text: "best programming language",
			},
			expected: []string{"top", "excellent", "superior", "optimal"},
		},
		{
			name: "compare query",
			query: &Query{
				Text: "compare Python vs Go",
			},
			expected: []string{"versus", "difference", "similarity", "analysis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expansions, err := expander.Expand(context.Background(), tt.query)
			require.NoError(t, err)

			for _, expected := range tt.expected {
				assert.Contains(t, expansions, expected)
			}
		})
	}
}

func TestContextAwareExpander(t *testing.T) {
	expander := NewContextAwareExpander(nil)
	require.NotNil(t, expander)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "user preferences context",
			query: &Query{
				Text: "programming",
				Context: map[string]interface{}{
					"user_preferences": map[string]interface{}{
						"interests": []interface{}{"AI", "machine learning"},
					},
				},
			},
			expected: []string{"AI", "machine learning"},
		},
		{
			name: "domain context",
			query: &Query{
				Text: "query",
				Context: map[string]interface{}{
					"domain": "technical",
				},
			},
			expected: []string{"technical", "domain-specific"},
		},
		{
			name: "time context",
			query: &Query{
				Text: "news",
				Context: map[string]interface{}{
					"time_context": "recent",
				},
			},
			expected: []string{"recent", "current"},
		},
		{
			name: "location context",
			query: &Query{
				Text: "restaurant",
				Context: map[string]interface{}{
					"location": "San Francisco",
				},
			},
			expected: []string{"San Francisco", "local", "regional"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expansions, err := expander.Expand(context.Background(), tt.query)
			require.NoError(t, err)

			for _, expected := range tt.expected {
				assert.Contains(t, expansions, expected)
			}
		})
	}
}

func TestQueryExpansionManager_shouldUseExpander(t *testing.T) {
	manager := NewQueryExpansionManager(nil)
	require.NotNil(t, manager)

	tests := []struct {
		name     string
		expander QueryExpander
		query    *Query
		expected bool
	}{
		{
			name:     "semantic expander enabled",
			expander: &SemanticExpander{},
			query:    &Query{Text: "test"},
			expected: true,
		},
		{
			name:     "context-aware expander with context",
			expander: &ContextAwareExpander{},
			query: &Query{
				Text:    "test",
				Context: map[string]interface{}{"key": "value"},
			},
			expected: true,
		},
		{
			name:     "context-aware expander without context",
			expander: &ContextAwareExpander{},
			query:    &Query{Text: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.shouldUseExpander(tt.expander, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryExpansionManager_filterAndRankExpansions(t *testing.T) {
	manager := NewQueryExpansionManager(&ExpansionConfig{
		MinConfidence: 0.5,
	})

	expansions := []string{"term1", "term2", "term3", "term1"} // Duplicate
	scores := map[string]float64{
		"term1": 0.8,
		"term2": 0.3, // Below threshold
		"term3": 0.7,
	}

	filtered := manager.filterAndRankExpansions(expansions, scores)

	// Should remove duplicates and filter by confidence
	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, "term1")
	assert.Contains(t, filtered, "term3")
	assert.NotContains(t, filtered, "term2")
}

func TestExpanderInterfaces(t *testing.T) {
	tests := []struct {
		name     string
		expander QueryExpander
		expected string
	}{
		{
			name:     "synonym expander",
			expander: &SynonymExpander{},
			expected: "synonym",
		},
		{
			name:     "semantic expander",
			expander: &SemanticExpander{},
			expected: "semantic",
		},
		{
			name:     "context-aware expander",
			expander: &ContextAwareExpander{},
			expected: "context-aware",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expander.Type())
			assert.Greater(t, tt.expander.Confidence(), 0.0)
			assert.LessOrEqual(t, tt.expander.Confidence(), 1.0)
		})
	}
}

func TestExpansionConfigDefaults(t *testing.T) {
	manager := NewQueryExpansionManager(nil)
	require.NotNil(t, manager)

	config := manager.config
	assert.Equal(t, 5, config.MaxExpansionTerms)
	assert.Equal(t, 0.3, config.MinConfidence)
	assert.True(t, config.EnableSemantic)
	assert.True(t, config.EnableSynonym)
	assert.True(t, config.EnableContextAware)
	assert.Equal(t, 2, config.MaxExpansionDepth)
	assert.Equal(t, 0.7, config.SemanticThreshold)
	assert.Equal(t, 0.6, config.ContextWeight)
}

func TestExpansionManagerWithCustomConfig(t *testing.T) {
	config := &ExpansionConfig{
		MaxExpansionTerms:  3,
		MinConfidence:      0.8,
		EnableSemantic:     false,
		EnableSynonym:      true,
		EnableContextAware: false,
	}

	manager := NewQueryExpansionManager(config)
	require.NotNil(t, manager)

	query := &Query{Text: "fast car"}
	expansions, err := manager.ExpandQuery(context.Background(), query, nil)
	require.NoError(t, err)

	// Should respect custom config
	assert.Len(t, expansions, 3)
}
