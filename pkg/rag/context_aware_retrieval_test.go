package rag

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextAwareRetrievalManager(t *testing.T) {
	tests := []struct {
		name   string
		config *ContextRetrievalConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "custom config",
			config: &ContextRetrievalConfig{
				EnableUserContext:     false,
				EnableDomainContext:   true,
				EnableTemporalContext: false,
				EnableLocationContext: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewContextAwareRetrievalManager(tt.config)
			require.NotNil(t, manager)
			assert.NotNil(t, manager.config)
			assert.NotNil(t, manager.strategies)
		})
	}
}

func TestContextAwareRetrievalManager_ProcessContextAwareQuery(t *testing.T) {
	manager := NewContextAwareRetrievalManager(nil)
	require.NotNil(t, manager)

	query := &Query{
		Text: "test query",
		Context: map[string]interface{}{
			"user_id":  "user123",
			"domain":   "technical",
			"location": "San Francisco",
		},
	}

	enhanced, err := manager.ProcessContextAwareQuery(context.Background(), query, nil)
	require.NoError(t, err)
	assert.NotNil(t, enhanced)
	assert.Equal(t, query, enhanced.OriginalQuery)
	assert.NotEmpty(t, enhanced.Enhancements)
	assert.Greater(t, enhanced.Confidence, 0.0)
}

func TestContextAwareRetrievalManager_shouldUseStrategy(t *testing.T) {
	manager := NewContextAwareRetrievalManager(&ContextRetrievalConfig{
		EnableUserContext:     false,
		EnableDomainContext:   true,
		EnableTemporalContext: false,
		EnableLocationContext: true,
	})

	tests := []struct {
		name     string
		strategy ContextRetrievalStrategy
		query    *Query
		expected bool
	}{
		{
			name:     "user context strategy disabled",
			strategy: &UserContextStrategy{},
			query:    &Query{Text: "test"},
			expected: false,
		},
		{
			name:     "domain context strategy enabled",
			strategy: &DomainContextStrategy{},
			query:    &Query{Text: "test"},
			expected: true,
		},
		{
			name:     "temporal context strategy disabled",
			strategy: &TemporalContextStrategy{},
			query:    &Query{Text: "test"},
			expected: false,
		},
		{
			name:     "location context strategy with location",
			strategy: &LocationContextStrategy{},
			query: &Query{
				Text:    "test",
				Context: map[string]interface{}{"location": "SF"},
			},
			expected: true,
		},
		{
			name:     "location context strategy without location",
			strategy: &LocationContextStrategy{},
			query:    &Query{Text: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.shouldUseStrategy(tt.strategy, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextAwareRetrievalManager_applyContextDecay(t *testing.T) {
	manager := NewContextAwareRetrievalManager(&ContextRetrievalConfig{
		ContextDecayRate: 0.2,
	})

	tests := []struct {
		name       string
		confidence float64
		expected   float64
	}{
		{
			name:       "high confidence",
			confidence: 1.0,
			expected:   0.8,
		},
		{
			name:       "medium confidence",
			confidence: 0.5,
			expected:   0.4,
		},
		{
			name:       "low confidence",
			confidence: 0.1,
			expected:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.applyContextDecay(tt.confidence)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestUserContextStrategy(t *testing.T) {
	strategy := NewUserContextStrategy(nil)
	require.NotNil(t, strategy)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "with user context",
			query: &Query{
				Text: "test",
				Context: map[string]interface{}{
					"user_id": "user123",
				},
			},
			expected: []string{"user-history", "user-preferences", "user-behavior"},
		},
		{
			name: "without user context",
			query: &Query{
				Text: "test",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced, err := strategy.EnhanceQuery(context.Background(), tt.query)
			require.NoError(t, err)
			assert.NotNil(t, enhanced)

			for _, expected := range tt.expected {
				assert.Contains(t, enhanced.Enhancements, expected)
			}
		})
	}
}

func TestUserContextStrategy_ConfigDefaults(t *testing.T) {
	strategy := NewUserContextStrategy(nil)
	require.NotNil(t, strategy)

	config := strategy.config
	assert.True(t, config.EnableUserHistory)
	assert.True(t, config.EnableUserPreferences)
	assert.True(t, config.EnableUserBehavior)
	assert.Equal(t, 0.4, config.HistoryWeight)
	assert.Equal(t, 0.35, config.PreferencesWeight)
	assert.Equal(t, 0.25, config.BehaviorWeight)
}

func TestDomainContextStrategy(t *testing.T) {
	strategy := NewDomainContextStrategy(nil)
	require.NotNil(t, strategy)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "medical domain",
			query: &Query{
				Text: "health checkup",
			},
			expected: []string{"domain-detection", "domain-rules", "domain-synonyms"},
		},
		{
			name: "legal domain",
			query: &Query{
				Text: "attorney consultation",
			},
			expected: []string{"domain-detection", "domain-rules", "domain-synonyms"},
		},
		{
			name: "financial domain",
			query: &Query{
				Text: "bank account",
			},
			expected: []string{"domain-detection", "domain-rules", "domain-synonyms"},
		},
		{
			name: "technical domain",
			query: &Query{
				Text: "programming code",
			},
			expected: []string{"domain-detection", "domain-rules", "domain-synonyms"},
		},
		{
			name: "educational domain",
			query: &Query{
				Text: "study materials",
			},
			expected: []string{"domain-detection", "domain-rules", "domain-synonyms"},
		},
		{
			name: "no domain detected",
			query: &Query{
				Text: "random text",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced, err := strategy.EnhanceQuery(context.Background(), tt.query)
			require.NoError(t, err)
			assert.NotNil(t, enhanced)

			if len(tt.expected) > 0 {
				assert.NotEmpty(t, enhanced.Context["detected_domain"])
				for _, expected := range tt.expected {
					assert.Contains(t, enhanced.Enhancements, expected)
				}
			} else {
				assert.Empty(t, enhanced.Enhancements)
			}
		})
	}
}

func TestDomainContextStrategy_ConfigDefaults(t *testing.T) {
	strategy := NewDomainContextStrategy(nil)
	require.NotNil(t, strategy)

	config := strategy.config
	assert.True(t, config.EnableDomainDetection)
	assert.True(t, config.EnableDomainRules)
	assert.True(t, config.EnableDomainSynonyms)
	assert.Equal(t, 0.8, config.DomainWeight)
}

func TestTemporalContextStrategy(t *testing.T) {
	strategy := NewTemporalContextStrategy(nil)
	require.NotNil(t, strategy)

	query := &Query{Text: "test query"}

	enhanced, err := strategy.EnhanceQuery(context.Background(), query)
	require.NoError(t, err)
	assert.NotNil(t, enhanced)

	// Should include time context
	assert.NotNil(t, enhanced.Context["current_time"])
	assert.NotNil(t, enhanced.Context["hour_of_day"])
	assert.NotNil(t, enhanced.Context["day_of_week"])
	assert.Contains(t, enhanced.Enhancements, "time-context")

	// Should include seasonal context
	assert.NotNil(t, enhanced.Context["season"])
	assert.Contains(t, enhanced.Enhancements, "seasonal-context")

	// Should include trend context
	assert.NotNil(t, enhanced.Context["trend_period"])
	assert.Contains(t, enhanced.Enhancements, "trend-context")
}

func TestTemporalContextStrategy_getSeason(t *testing.T) {
	strategy := NewTemporalContextStrategy(nil)

	tests := []struct {
		name     string
		month    time.Month
		expected string
	}{
		{"winter months", time.December, "winter"},
		{"winter months", time.January, "winter"},
		{"winter months", time.February, "winter"},
		{"spring months", time.March, "spring"},
		{"spring months", time.April, "spring"},
		{"spring months", time.May, "spring"},
		{"summer months", time.June, "summer"},
		{"summer months", time.July, "summer"},
		{"summer months", time.August, "summer"},
		{"fall months", time.September, "fall"},
		{"fall months", time.October, "fall"},
		{"fall months", time.November, "fall"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.getSeason(tt.month)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemporalContextStrategy_ConfigDefaults(t *testing.T) {
	strategy := NewTemporalContextStrategy(nil)
	require.NotNil(t, strategy)

	config := strategy.config
	assert.True(t, config.EnableTimeContext)
	assert.True(t, config.EnableSeasonalContext)
	assert.True(t, config.EnableTrendContext)
	assert.Equal(t, 0.7, config.TemporalWeight)
}

func TestLocationContextStrategy(t *testing.T) {
	strategy := NewLocationContextStrategy(nil)
	require.NotNil(t, strategy)

	tests := []struct {
		name     string
		query    *Query
		expected []string
	}{
		{
			name: "with location context",
			query: &Query{
				Text: "restaurant",
				Context: map[string]interface{}{
					"location": "San Francisco",
				},
			},
			expected: []string{"geolocation", "regional-context", "cultural-context"},
		},
		{
			name: "without location context",
			query: &Query{
				Text: "restaurant",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced, err := strategy.EnhanceQuery(context.Background(), tt.query)
			require.NoError(t, err)
			assert.NotNil(t, enhanced)

			if len(tt.expected) > 0 {
				assert.Equal(t, "San Francisco", enhanced.Context["user_location"])
				for _, expected := range tt.expected {
					assert.Contains(t, enhanced.Enhancements, expected)
				}
			} else {
				assert.Empty(t, enhanced.Enhancements)
			}
		})
	}
}

func TestLocationContextStrategy_ConfigDefaults(t *testing.T) {
	strategy := NewLocationContextStrategy(nil)
	require.NotNil(t, strategy)

	config := strategy.config
	assert.True(t, config.EnableGeolocation)
	assert.True(t, config.EnableRegionalContext)
	assert.True(t, config.EnableCulturalContext)
	assert.Equal(t, 0.6, config.LocationWeight)
}

func TestContextRetrievalStrategyInterfaces(t *testing.T) {
	tests := []struct {
		name     string
		strategy ContextRetrievalStrategy
		expected string
	}{
		{
			name:     "user context strategy",
			strategy: &UserContextStrategy{},
			expected: "user-context",
		},
		{
			name:     "domain context strategy",
			strategy: &DomainContextStrategy{},
			expected: "domain-context",
		},
		{
			name:     "temporal context strategy",
			strategy: &TemporalContextStrategy{},
			expected: "temporal-context",
		},
		{
			name:     "location context strategy",
			strategy: &LocationContextStrategy{},
			expected: "location-context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.strategy.Type())
			assert.Greater(t, tt.strategy.Priority(), 0)
		})
	}
}

func TestContextRetrievalConfigDefaults(t *testing.T) {
	manager := NewContextAwareRetrievalManager(nil)
	require.NotNil(t, manager)

	config := manager.config
	assert.True(t, config.EnableUserContext)
	assert.True(t, config.EnableDomainContext)
	assert.True(t, config.EnableTemporalContext)
	assert.True(t, config.EnableLocationContext)
	assert.Equal(t, 0.3, config.UserContextWeight)
	assert.Equal(t, 0.25, config.DomainContextWeight)
	assert.Equal(t, 0.25, config.TemporalContextWeight)
	assert.Equal(t, 0.2, config.LocationContextWeight)
	assert.Equal(t, 0.1, config.ContextDecayRate)
	assert.Equal(t, 3, config.MaxContextDepth)
	assert.True(t, config.EnableContextCaching)
	assert.Equal(t, 5*time.Minute, config.ContextCacheTTL)
}

func TestContextAwareRetrievalManagerWithCustomConfig(t *testing.T) {
	config := &ContextRetrievalConfig{
		EnableUserContext:     false,
		EnableDomainContext:   true,
		EnableTemporalContext: false,
		EnableLocationContext: true,
		UserContextWeight:     0.0,
		DomainContextWeight:   0.6,
		TemporalContextWeight: 0.0,
		LocationContextWeight: 0.4,
	}

	manager := NewContextAwareRetrievalManager(config)
	require.NotNil(t, manager)

	query := &Query{
		Text: "test",
		Context: map[string]interface{}{
			"domain":   "technical",
			"location": "San Francisco",
		},
	}

	enhanced, err := manager.ProcessContextAwareQuery(context.Background(), query, nil)
	require.NoError(t, err)
	assert.NotNil(t, enhanced)

	// Should include enabled strategies
	assert.Contains(t, enhanced.Enhancements, "domain-detection")
	assert.Contains(t, enhanced.Enhancements, "geolocation")

	// Verify that disabled strategies are not included
	assert.NotContains(t, enhanced.Enhancements, "user-history")
	assert.NotContains(t, enhanced.Enhancements, "time-context")

	// Verify context is properly set
	assert.Equal(t, "technical", enhanced.Context["detected_domain"])
	assert.Equal(t, "San Francisco", enhanced.Context["user_location"])
}
