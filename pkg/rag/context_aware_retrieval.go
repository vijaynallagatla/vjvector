package rag

import (
	"context"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// ContextAwareRetrievalManager manages context-aware retrieval strategies
type ContextAwareRetrievalManager struct {
	strategies []ContextRetrievalStrategy
	config     *ContextRetrievalConfig
	logger     *slog.Logger
	mu         sync.RWMutex
}

// ContextRetrievalConfig configures context-aware retrieval behavior
type ContextRetrievalConfig struct {
	EnableUserContext     bool          `json:"enable_user_context"`
	EnableDomainContext   bool          `json:"enable_domain_context"`
	EnableTemporalContext bool          `json:"enable_temporal_context"`
	EnableLocationContext bool          `json:"enable_location_context"`
	UserContextWeight     float64       `json:"user_context_weight"`
	DomainContextWeight   float64       `json:"domain_context_weight"`
	TemporalContextWeight float64       `json:"temporal_context_weight"`
	LocationContextWeight float64       `json:"location_context_weight"`
	ContextDecayRate      float64       `json:"context_decay_rate"`
	MaxContextDepth       int           `json:"max_context_depth"`
	EnableContextCaching  bool          `json:"enable_context_caching"`
	ContextCacheTTL       time.Duration `json:"context_cache_ttl"`
}

// NewContextAwareRetrievalManager creates a new context-aware retrieval manager
func NewContextAwareRetrievalManager(config *ContextRetrievalConfig) *ContextAwareRetrievalManager {
	if config == nil {
		config = &ContextRetrievalConfig{
			EnableUserContext:     true,
			EnableDomainContext:   true,
			EnableTemporalContext: true,
			EnableLocationContext: true,
			UserContextWeight:     0.3,
			DomainContextWeight:   0.25,
			TemporalContextWeight: 0.25,
			LocationContextWeight: 0.2,
			ContextDecayRate:      0.1,
			MaxContextDepth:       3,
			EnableContextCaching:  true,
			ContextCacheTTL:       5 * time.Minute,
		}
	}

	manager := &ContextAwareRetrievalManager{
		config: config,
		logger: slog.With("component", "context-aware-retrieval-manager"),
	}

	// Register default strategies
	manager.registerDefaultStrategies()

	return manager
}

// ProcessContextAwareQuery processes a query with context-aware retrieval
func (m *ContextAwareRetrievalManager) ProcessContextAwareQuery(ctx context.Context, query *Query, embeddingService embedding.Service) (*ContextEnhancedQuery, error) {
	enhancedQuery := &ContextEnhancedQuery{
		OriginalQuery: query,
		Context:       make(map[string]interface{}),
		Enhancements:  make([]string, 0),
		Confidence:    1.0,
	}

	// Apply context enhancement strategies
	for _, strategy := range m.strategies {
		if !m.shouldUseStrategy(strategy, query) {
			continue
		}

		enhanced, err := strategy.EnhanceQuery(ctx, query)
		if err != nil {
			m.logger.Warn("Strategy failed", "strategy", strategy.Type(), "error", err)
			continue
		}

		// Merge enhancements
		enhancedQuery.Enhancements = append(enhancedQuery.Enhancements, enhanced.Enhancements...)
		enhancedQuery.Confidence *= enhanced.Confidence

		// Merge context
		for key, value := range enhanced.Context {
			enhancedQuery.Context[key] = value
		}
	}

	// Apply context decay
	enhancedQuery.Confidence = m.applyContextDecay(enhancedQuery.Confidence)

	return enhancedQuery, nil
}

// shouldUseStrategy determines if a strategy should be used for a query
func (m *ContextAwareRetrievalManager) shouldUseStrategy(strategy ContextRetrievalStrategy, query *Query) bool {
	switch strategy.Type() {
	case "user-context":
		return m.config.EnableUserContext
	case "domain-context":
		return m.config.EnableDomainContext
	case "temporal-context":
		return m.config.EnableTemporalContext
	case "location-context":
		return m.config.EnableLocationContext && query.Context != nil && query.Context["location"] != nil
	default:
		return true
	}
}

// applyContextDecay applies decay to context confidence
func (m *ContextAwareRetrievalManager) applyContextDecay(confidence float64) float64 {
	decay := confidence * m.config.ContextDecayRate
	if decay < 0.1 {
		decay = 0.1
	}
	return confidence - decay
}

// registerDefaultStrategies registers the default context retrieval strategies
func (m *ContextAwareRetrievalManager) registerDefaultStrategies() {
	m.strategies = append(m.strategies,
		NewUserContextStrategy(nil),
		NewDomainContextStrategy(nil),
		NewTemporalContextStrategy(nil),
		NewLocationContextStrategy(nil),
	)
}

// ContextEnhancedQuery represents a query enhanced with context information
type ContextEnhancedQuery struct {
	OriginalQuery *Query                 `json:"original_query"`
	Context       map[string]interface{} `json:"context"`
	Enhancements  []string               `json:"enhancements"`
	Confidence    float64                `json:"confidence"`
}

// ContextRetrievalStrategy represents a context retrieval strategy
type ContextRetrievalStrategy interface {
	// EnhanceQuery enhances a query with context information
	EnhanceQuery(ctx context.Context, query *Query) (*ContextEnhancedQuery, error)

	// Type returns the strategy type
	Type() string

	// Priority returns the strategy priority (lower = higher priority)
	Priority() int
}

// UserContextStrategy enhances queries with user context information
type UserContextStrategy struct {
	config *UserContextConfig
}

// UserContextConfig configures user context strategy
type UserContextConfig struct {
	EnableUserHistory     bool    `json:"enable_user_history"`
	EnableUserPreferences bool    `json:"enable_user_preferences"`
	EnableUserBehavior    bool    `json:"enable_user_behavior"`
	HistoryWeight         float64 `json:"history_weight"`
	PreferencesWeight     float64 `json:"preferences_weight"`
	BehaviorWeight        float64 `json:"behavior_weight"`
}

// NewUserContextStrategy creates a new user context strategy
func NewUserContextStrategy(config *UserContextConfig) *UserContextStrategy {
	if config == nil {
		config = &UserContextConfig{
			EnableUserHistory:     true,
			EnableUserPreferences: true,
			EnableUserBehavior:    true,
			HistoryWeight:         0.4,
			PreferencesWeight:     0.35,
			BehaviorWeight:        0.25,
		}
	}

	return &UserContextStrategy{
		config: config,
	}
}

// EnhanceQuery enhances a query with user context
func (s *UserContextStrategy) EnhanceQuery(ctx context.Context, query *Query) (*ContextEnhancedQuery, error) {
	enhanced := &ContextEnhancedQuery{
		OriginalQuery: query,
		Context:       make(map[string]interface{}),
		Enhancements:  make([]string, 0),
		Confidence:    1.0,
	}

	// Extract user context from query
	if query.Context != nil {
		if userID, exists := query.Context["user_id"]; exists {
			enhanced.Context["user_id"] = userID

			// Add user-specific enhancements
			if s.config.EnableUserHistory {
				enhanced.Enhancements = append(enhanced.Enhancements, "user-history")
			}

			if s.config.EnableUserPreferences {
				enhanced.Enhancements = append(enhanced.Enhancements, "user-preferences")
			}

			if s.config.EnableUserBehavior {
				enhanced.Enhancements = append(enhanced.Enhancements, "user-behavior")
			}
		}
	}

	return enhanced, nil
}

// Type returns the strategy type
func (s *UserContextStrategy) Type() string { return "user-context" }

// Priority returns the strategy priority
func (s *UserContextStrategy) Priority() int { return 100 }

// DomainContextStrategy enhances queries with domain context information
type DomainContextStrategy struct {
	config *DomainContextConfig
}

// DomainContextConfig configures domain context strategy
type DomainContextConfig struct {
	EnableDomainDetection bool    `json:"enable_domain_detection"`
	EnableDomainRules     bool    `json:"enable_domain_rules"`
	EnableDomainSynonyms  bool    `json:"enable_domain_synonyms"`
	DomainWeight          float64 `json:"domain_weight"`
}

// NewDomainContextStrategy creates a new domain context strategy
func NewDomainContextStrategy(config *DomainContextConfig) *DomainContextStrategy {
	if config == nil {
		config = &DomainContextConfig{
			EnableDomainDetection: true,
			EnableDomainRules:     true,
			EnableDomainSynonyms:  true,
			DomainWeight:          0.8,
		}
	}

	return &DomainContextStrategy{
		config: config,
	}
}

// EnhanceQuery enhances a query with domain context
func (s *DomainContextStrategy) EnhanceQuery(ctx context.Context, query *Query) (*ContextEnhancedQuery, error) {
	enhanced := &ContextEnhancedQuery{
		OriginalQuery: query,
		Context:       make(map[string]interface{}),
		Enhancements:  make([]string, 0),
		Confidence:    1.0,
	}

	// First check if domain is specified in query context
	if query.Context != nil {
		if contextDomain, exists := query.Context["domain"]; exists {
			if domainStr, ok := contextDomain.(string); ok && domainStr != "" {
				enhanced.Context["detected_domain"] = domainStr
				enhanced.Enhancements = append(enhanced.Enhancements, "domain-detection")

				// Add domain-specific rules
				if s.config.EnableDomainRules {
					enhanced.Enhancements = append(enhanced.Enhancements, "domain-rules")
				}

				// Add domain synonyms
				if s.config.EnableDomainSynonyms {
					enhanced.Enhancements = append(enhanced.Enhancements, "domain-synonyms")
				}

				return enhanced, nil
			}
		}
	}

	// Fall back to detecting domain from query text
	domain := s.detectDomain(query.Text)
	if domain != "" {
		enhanced.Context["detected_domain"] = domain
		enhanced.Enhancements = append(enhanced.Enhancements, "domain-detection")

		// Add domain-specific rules
		if s.config.EnableDomainRules {
			enhanced.Enhancements = append(enhanced.Enhancements, "domain-rules")
		}

		// Add domain synonyms
		if s.config.EnableDomainSynonyms {
			enhanced.Enhancements = append(enhanced.Enhancements, "domain-synonyms")
		}
	}

	return enhanced, nil
}

// detectDomain detects the domain from query text
func (s *DomainContextStrategy) detectDomain(text string) string {
	// Simple domain detection based on keywords
	// In production, you would use more sophisticated NLP techniques

	text = strings.ToLower(text)

	if strings.Contains(text, "medical") || strings.Contains(text, "health") || strings.Contains(text, "doctor") {
		return "medical"
	}

	if strings.Contains(text, "legal") || strings.Contains(text, "law") || strings.Contains(text, "attorney") {
		return "legal"
	}

	if strings.Contains(text, "financial") || strings.Contains(text, "money") || strings.Contains(text, "bank") {
		return "financial"
	}

	if strings.Contains(text, "technical") || strings.Contains(text, "code") || strings.Contains(text, "programming") {
		return "technical"
	}

	if strings.Contains(text, "educational") || strings.Contains(text, "learn") || strings.Contains(text, "study") {
		return "educational"
	}

	return ""
}

// Type returns the strategy type
func (s *DomainContextStrategy) Type() string { return "domain-context" }

// Priority returns the strategy priority
func (s *DomainContextStrategy) Priority() int { return 200 }

// TemporalContextStrategy enhances queries with temporal context information
type TemporalContextStrategy struct {
	config *TemporalContextConfig
}

// TemporalContextConfig configures temporal context strategy
type TemporalContextConfig struct {
	EnableTimeContext     bool    `json:"enable_time_context"`
	EnableSeasonalContext bool    `json:"enable_seasonal_context"`
	EnableTrendContext    bool    `json:"enable_trend_context"`
	TemporalWeight        float64 `json:"temporal_weight"`
}

// NewTemporalContextStrategy creates a new temporal context strategy
func NewTemporalContextStrategy(config *TemporalContextConfig) *TemporalContextStrategy {
	if config == nil {
		config = &TemporalContextConfig{
			EnableTimeContext:     true,
			EnableSeasonalContext: true,
			EnableTrendContext:    true,
			TemporalWeight:        0.7,
		}
	}

	return &TemporalContextStrategy{
		config: config,
	}
}

// EnhanceQuery enhances a query with temporal context
func (s *TemporalContextStrategy) EnhanceQuery(ctx context.Context, query *Query) (*ContextEnhancedQuery, error) {
	enhanced := &ContextEnhancedQuery{
		OriginalQuery: query,
		Context:       make(map[string]interface{}),
		Enhancements:  make([]string, 0),
		Confidence:    1.0,
	}

	now := time.Now()

	// Add current time context
	if s.config.EnableTimeContext {
		enhanced.Context["current_time"] = now
		enhanced.Context["hour_of_day"] = now.Hour()
		enhanced.Context["day_of_week"] = now.Weekday()
		enhanced.Enhancements = append(enhanced.Enhancements, "time-context")
	}

	// Add seasonal context
	if s.config.EnableSeasonalContext {
		month := now.Month()
		season := s.getSeason(month)
		enhanced.Context["season"] = season
		enhanced.Enhancements = append(enhanced.Enhancements, "seasonal-context")
	}

	// Add trend context
	if s.config.EnableTrendContext {
		enhanced.Context["trend_period"] = "current"
		enhanced.Enhancements = append(enhanced.Enhancements, "trend-context")
	}

	return enhanced, nil
}

// getSeason determines the season based on month
func (s *TemporalContextStrategy) getSeason(month time.Month) string {
	switch month {
	case time.December, time.January, time.February:
		return "winter"
	case time.March, time.April, time.May:
		return "spring"
	case time.June, time.July, time.August:
		return "summer"
	case time.September, time.October, time.November:
		return "fall"
	default:
		return "unknown"
	}
}

// Type returns the strategy type
func (s *TemporalContextStrategy) Type() string { return "temporal-context" }

// Priority returns the strategy priority
func (s *TemporalContextStrategy) Priority() int { return 300 }

// LocationContextStrategy enhances queries with location context information
type LocationContextStrategy struct {
	config *LocationContextConfig
}

// LocationContextConfig configures location context strategy
type LocationContextConfig struct {
	EnableGeolocation     bool    `json:"enable_geolocation"`
	EnableRegionalContext bool    `json:"enable_regional_context"`
	EnableCulturalContext bool    `json:"enable_cultural_context"`
	LocationWeight        float64 `json:"location_weight"`
}

// NewLocationContextStrategy creates a new location context strategy
func NewLocationContextStrategy(config *LocationContextConfig) *LocationContextStrategy {
	if config == nil {
		config = &LocationContextConfig{
			EnableGeolocation:     true,
			EnableRegionalContext: true,
			EnableCulturalContext: true,
			LocationWeight:        0.6,
		}
	}

	return &LocationContextStrategy{
		config: config,
	}
}

// EnhanceQuery enhances a query with location context
func (s *LocationContextStrategy) EnhanceQuery(ctx context.Context, query *Query) (*ContextEnhancedQuery, error) {
	enhanced := &ContextEnhancedQuery{
		OriginalQuery: query,
		Context:       make(map[string]interface{}),
		Enhancements:  make([]string, 0),
		Confidence:    1.0,
	}

	// Extract location from query context
	if query.Context != nil {
		if location, exists := query.Context["location"]; exists {
			enhanced.Context["user_location"] = location

			if s.config.EnableGeolocation {
				enhanced.Enhancements = append(enhanced.Enhancements, "geolocation")
			}

			if s.config.EnableRegionalContext {
				enhanced.Enhancements = append(enhanced.Enhancements, "regional-context")
			}

			if s.config.EnableCulturalContext {
				enhanced.Enhancements = append(enhanced.Enhancements, "cultural-context")
			}
		}
	}

	return enhanced, nil
}

// Type returns the strategy type
func (s *LocationContextStrategy) Type() string { return "location-context" }

// Priority returns the strategy priority
func (s *LocationContextStrategy) Priority() int { return 400 }
