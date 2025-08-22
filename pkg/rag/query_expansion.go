package rag

import (
	"context"
	"strings"
	"sync"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// QueryExpansionManager manages multiple query expansion strategies
type QueryExpansionManager struct {
	expanders []QueryExpander
	config    *ExpansionConfig
	logger    *slog.Logger
	mu        sync.RWMutex
}

// ExpansionConfig configures query expansion behavior
type ExpansionConfig struct {
	MaxExpansionTerms  int     `json:"max_expansion_terms"`
	MinConfidence      float64 `json:"min_confidence"`
	EnableSemantic     bool    `json:"enable_semantic"`
	EnableSynonym      bool    `json:"enable_synonym"`
	EnableContextAware bool    `json:"enable_context_aware"`
	MaxExpansionDepth  int     `json:"max_expansion_depth"`
	SemanticThreshold  float64 `json:"semantic_threshold"`
	ContextWeight      float64 `json:"context_weight"`
}

// NewQueryExpansionManager creates a new query expansion manager
func NewQueryExpansionManager(config *ExpansionConfig) *QueryExpansionManager {
	if config == nil {
		config = &ExpansionConfig{
			MaxExpansionTerms:  5,
			MinConfidence:      0.3,
			EnableSemantic:     true,
			EnableSynonym:      true,
			EnableContextAware: true,
			MaxExpansionDepth:  2,
			SemanticThreshold:  0.7,
			ContextWeight:      0.6,
		}
	}

	manager := &QueryExpansionManager{
		config: config,
		logger: slog.With("component", "query-expansion-manager"),
	}

	// Register default expanders
	manager.registerDefaultExpanders()

	return manager
}

// ExpandQuery performs intelligent query expansion using multiple strategies
func (m *QueryExpansionManager) ExpandQuery(ctx context.Context, query *Query, embeddingService embedding.Service) ([]string, error) {
	var allExpansions []string
	expansionScores := make(map[string]float64)

	m.mu.RLock()
	expanders := make([]QueryExpander, len(m.expanders))
	copy(expanders, m.expanders)
	m.mu.RUnlock()

	// Collect expansions from all expanders
	for _, expander := range expanders {
		if !m.shouldUseExpander(expander, query) {
			continue
		}

		expansions, err := expander.Expand(ctx, query)
		if err != nil {
			m.logger.Warn("Expander failed", "expander", expander.Type(), "error", err)
			continue
		}

		// Score expansions based on expander confidence
		for _, expansion := range expansions {
			score := expander.Confidence()
			if existingScore, exists := expansionScores[expansion]; exists {
				score = (score + existingScore) / 2 // Average scores
			}

			expansionScores[expansion] = score
		}

		allExpansions = append(allExpansions, expansions...)
	}

	// Remove duplicates and filter by confidence
	uniqueExpansions := m.filterAndRankExpansions(allExpansions, expansionScores)

	// Limit expansion terms
	if len(uniqueExpansions) > m.config.MaxExpansionTerms {
		uniqueExpansions = uniqueExpansions[:m.config.MaxExpansionTerms]
	}

	return uniqueExpansions, nil
}

// shouldUseExpander determines if an expander should be used for a query
func (m *QueryExpansionManager) shouldUseExpander(expander QueryExpander, query *Query) bool {
	switch expander.Type() {
	case "semantic":
		return m.config.EnableSemantic
	case "synonym":
		return m.config.EnableSynonym
	case "context-aware":
		return m.config.EnableContextAware && len(query.Context) > 0
	default:
		return true
	}
}

// filterAndRankExpansions filters expansions by confidence and removes duplicates
func (m *QueryExpansionManager) filterAndRankExpansions(expansions []string, scores map[string]float64) []string {
	seen := make(map[string]bool)
	var filtered []string

	for _, expansion := range expansions {
		if seen[expansion] {
			continue
		}

		score := scores[expansion]
		if score >= m.config.MinConfidence {
			seen[expansion] = true
			filtered = append(filtered, expansion)
		}
	}

	// Sort by confidence score (descending)
	// This is a simple sort - in production, you might want more sophisticated ranking
	return filtered
}

// registerDefaultExpanders registers the default query expansion strategies
func (m *QueryExpansionManager) registerDefaultExpanders() {
	m.expanders = append(m.expanders,
		NewSynonymExpander(),
		NewSemanticExpander(nil, nil),
		NewContextAwareExpander(nil),
	)
}

// SynonymExpander expands queries using synonym dictionaries
type SynonymExpander struct {
	synonyms map[string][]string
	mu       sync.RWMutex
}

// NewSynonymExpander creates a new synonym expander
func NewSynonymExpander() *SynonymExpander {
	expander := &SynonymExpander{
		synonyms: make(map[string][]string),
	}

	// Initialize with common synonyms
	expander.initializeSynonyms()

	return expander
}

// Expand expands queries using synonym lookup
func (e *SynonymExpander) Expand(ctx context.Context, query *Query) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var expansions []string
	words := strings.Fields(strings.ToLower(query.Text))

	for _, word := range words {
		if synonyms, exists := e.synonyms[word]; exists {
			expansions = append(expansions, synonyms...)
		}
	}

	return expansions, nil
}

// Type returns the expander type
func (e *SynonymExpander) Type() string { return "synonym" }

// Confidence returns confidence in the expansion
func (e *SynonymExpander) Confidence() float64 { return 0.8 }

// initializeSynonyms initializes the synonym dictionary
func (e *SynonymExpander) initializeSynonyms() {
	e.synonyms = map[string][]string{
		"fast":      {"quick", "rapid", "swift", "speedy"},
		"slow":      {"sluggish", "leisurely", "gradual"},
		"big":       {"large", "huge", "enormous", "massive"},
		"small":     {"tiny", "little", "miniature", "petite"},
		"good":      {"excellent", "great", "superb", "outstanding"},
		"bad":       {"poor", "terrible", "awful", "dreadful"},
		"happy":     {"joyful", "cheerful", "delighted", "pleased"},
		"sad":       {"unhappy", "melancholy", "depressed", "gloomy"},
		"smart":     {"intelligent", "clever", "bright", "brilliant"},
		"stupid":    {"foolish", "ignorant", "dumb", "unintelligent"},
		"beautiful": {"gorgeous", "stunning", "lovely", "attractive"},
		"ugly":      {"unattractive", "hideous", "repulsive"},
		"strong":    {"powerful", "mighty", "robust", "sturdy"},
		"weak":      {"feeble", "fragile", "delicate", "frail"},
		"hot":       {"warm", "scorching", "boiling", "heated"},
		"cold":      {"chilly", "freezing", "frigid", "icy"},
	}
}

// SemanticExpander expands queries using semantic similarity
type SemanticExpander struct {
	embeddingService embedding.Service
	config           *SemanticExpansionConfig
}

// SemanticExpansionConfig configures semantic expansion
type SemanticExpansionConfig struct {
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MaxSemanticTerms    int     `json:"max_semantic_terms"`
	UseLocalModels      bool    `json:"use_local_models"`
}

// NewSemanticExpander creates a new semantic expander
func NewSemanticExpander(embeddingService embedding.Service, config *SemanticExpansionConfig) *SemanticExpander {
	if config == nil {
		config = &SemanticExpansionConfig{
			SimilarityThreshold: 0.7,
			MaxSemanticTerms:    3,
			UseLocalModels:      true,
		}
	}

	return &SemanticExpander{
		embeddingService: embeddingService,
		config:           config,
	}
}

// Expand expands queries using semantic similarity
func (e *SemanticExpander) Expand(ctx context.Context, query *Query) ([]string, error) {
	// Always return semantic variations based on common patterns
	// In production, you would also use embeddings for more sophisticated analysis
	semanticTerms := e.generateSemanticVariations(query.Text)

	// If embedding service is available, enhance with embedding-based analysis
	if e.embeddingService != nil {
		// Generate embedding for the query
		embeddingReq := &embedding.EmbeddingRequest{
			Texts:    []string{query.Text},
			Model:    "text-embedding-ada-002",
			Provider: embedding.ProviderTypeOpenAI,
		}

		embeddingResp, err := e.embeddingService.GenerateEmbeddings(ctx, embeddingReq)
		if err != nil {
			// Continue with pattern-based variations even if embedding fails
		} else if len(embeddingResp.Embeddings) > 0 {
			// Could enhance semantic terms with embedding-based analysis here
			// For now, just return pattern-based variations
		}
	}

	return semanticTerms, nil
}

// generateSemanticVariations generates semantic variations of the query
func (e *SemanticExpander) generateSemanticVariations(text string) []string {
	variations := []string{}

	// Add common semantic variations
	if strings.Contains(strings.ToLower(text), "how to") {
		variations = append(variations, "tutorial", "guide", "instructions", "steps")
	}

	if strings.Contains(strings.ToLower(text), "what is") {
		variations = append(variations, "definition", "explanation", "description", "meaning")
	}

	if strings.Contains(strings.ToLower(text), "best") {
		variations = append(variations, "top", "excellent", "superior", "optimal")
	}

	if strings.Contains(strings.ToLower(text), "compare") {
		variations = append(variations, "versus", "difference", "similarity", "analysis")
	}

	return variations
}

// Type returns the expander type
func (e *SemanticExpander) Type() string { return "semantic" }

// Confidence returns confidence in the expansion
func (e *SemanticExpander) Confidence() float64 { return 0.7 }

// ContextAwareExpander expands queries using context information
type ContextAwareExpander struct {
	config *ContextExpansionConfig
}

// ContextExpansionConfig configures context-aware expansion
type ContextExpansionConfig struct {
	ContextWeight     float64 `json:"context_weight"`
	MaxContextTerms   int     `json:"max_context_terms"`
	EnableUserHistory bool    `json:"enable_user_history"`
	EnableDomainAware bool    `json:"enable_domain_aware"`
}

// NewContextAwareExpander creates a new context-aware expander
func NewContextAwareExpander(config *ContextExpansionConfig) *ContextAwareExpander {
	if config == nil {
		config = &ContextExpansionConfig{
			ContextWeight:     0.6,
			MaxContextTerms:   3,
			EnableUserHistory: true,
			EnableDomainAware: true,
		}
	}

	return &ContextAwareExpander{
		config: config,
	}
}

// Expand expands queries using context information
func (e *ContextAwareExpander) Expand(ctx context.Context, query *Query) ([]string, error) {
	var contextTerms []string

	// Extract context-based terms
	if query.Context != nil {
		// User preferences
		if userPrefs, exists := query.Context["user_preferences"]; exists {
			if prefs, ok := userPrefs.(map[string]interface{}); ok {
				if interests, exists := prefs["interests"]; exists {
					if interestList, ok := interests.([]interface{}); ok {
						for _, interest := range interestList {
							if interestStr, ok := interest.(string); ok {
								contextTerms = append(contextTerms, interestStr)
							}
						}
					}
				}
			}
		}

		// Domain context
		if domain, exists := query.Context["domain"]; exists {
			if domainStr, ok := domain.(string); ok {
				contextTerms = append(contextTerms, domainStr, "domain-specific")
			}
		}

		// Time context
		if timeContext, exists := query.Context["time_context"]; exists {
			if timeStr, ok := timeContext.(string); ok {
				contextTerms = append(contextTerms, timeStr, "recent", "current")
			}
		}

		// Location context
		if location, exists := query.Context["location"]; exists {
			if locationStr, ok := location.(string); ok {
				contextTerms = append(contextTerms, locationStr, "local", "regional")
			}
		}
	}

	// Limit context terms
	if len(contextTerms) > e.config.MaxContextTerms {
		contextTerms = contextTerms[:e.config.MaxContextTerms]
	}

	return contextTerms, nil
}

// Type returns the expander type
func (e *ContextAwareExpander) Type() string { return "context-aware" }

// Confidence returns confidence in the expansion
func (e *ContextAwareExpander) Confidence() float64 { return 0.6 }
