package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// ResultRerankingManager manages multiple reranking strategies
type ResultRerankingManager struct {
	rerankers []ResultReranker
	config    *RerankingConfig
	logger    *slog.Logger
	mu        sync.RWMutex
}

// RerankingConfig configures result reranking behavior
type RerankingConfig struct {
	EnableSemanticReranking bool    `json:"enable_semantic_reranking"`
	EnableContextReranking  bool    `json:"enable_context_reranking"`
	EnableHybridScoring     bool    `json:"enable_hybrid_scoring"`
	SemanticWeight          float64 `json:"semantic_weight"`
	ContextWeight           float64 `json:"context_weight"`
	VectorWeight            float64 `json:"vector_weight"`
	MinRerankingConfidence  float64 `json:"min_reranking_confidence"`
	MaxRerankedResults      int     `json:"max_reranked_results"`
}

// NewResultRerankingManager creates a new result reranking manager
func NewResultRerankingManager(config *RerankingConfig) *ResultRerankingManager {
	if config == nil {
		config = &RerankingConfig{
			EnableSemanticReranking: true,
			EnableContextReranking:  true,
			EnableHybridScoring:     true,
			SemanticWeight:          0.4,
			ContextWeight:           0.3,
			VectorWeight:            0.3,
			MinRerankingConfidence:  0.5,
			MaxRerankedResults:      100,
		}
	}

	manager := &ResultRerankingManager{
		config: config,
		logger: slog.With("component", "result-reranking-manager"),
	}

	// Register default rerankers
	manager.registerDefaultRerankers()

	return manager
}

// RerankResults performs intelligent result reranking using multiple strategies
func (m *ResultRerankingManager) RerankResults(ctx context.Context, results []*QueryResult, query *Query, embeddingService embedding.Service) ([]*QueryResult, error) {
	if len(results) <= 1 {
		return results, nil
	}

	var rerankedResults []*QueryResult
	rerankingScores := make(map[string]float64)

	m.mu.RLock()
	rerankers := make([]ResultReranker, len(m.rerankers))
	copy(rerankers, m.rerankers)
	m.mu.RUnlock()

	// Apply each reranking strategy
	for _, reranker := range rerankers {
		if !m.shouldUseReranker(reranker, query) {
			continue
		}

		reranked, err := reranker.Rerank(ctx, results, query)
		if err != nil {
			m.logger.Warn("Reranker failed", "reranker", reranker.Type(), "error", err)
			continue
		}

		// Score reranking based on confidence
		for _, result := range reranked {
			score := reranker.Confidence()
			if existingScore, exists := rerankingScores[result.Vector.ID]; exists {
				score = (score + existingScore) / 2 // Average scores
			}
			rerankingScores[result.Vector.ID] = score
		}

		rerankedResults = reranked
	}

	// If no reranking was applied, use original results
	if len(rerankedResults) == 0 {
		rerankedResults = results
	}

	// Apply hybrid scoring if enabled
	if m.config.EnableHybridScoring {
		rerankedResults = m.applyHybridScoring(rerankedResults, query, embeddingService)
	}

	// Sort by final score
	sort.Slice(rerankedResults, func(i, j int) bool {
		return rerankedResults[i].Score > rerankedResults[j].Score
	})

	// Limit results
	if len(rerankedResults) > m.config.MaxRerankedResults {
		rerankedResults = rerankedResults[:m.config.MaxRerankedResults]
	}

	return rerankedResults, nil
}

// shouldUseReranker determines if a reranker should be used for a query
func (m *ResultRerankingManager) shouldUseReranker(reranker ResultReranker, query *Query) bool {
	switch reranker.Type() {
	case "semantic":
		return m.config.EnableSemanticReranking
	case "context-aware":
		return m.config.EnableContextReranking && len(query.Context) > 0
	default:
		return true
	}
}

// applyHybridScoring applies hybrid scoring combining multiple factors
func (m *ResultRerankingManager) applyHybridScoring(results []*QueryResult, query *Query, embeddingService embedding.Service) []*QueryResult {
	for _, result := range results {
		// Calculate hybrid score
		vectorScore := result.Score * m.config.VectorWeight
		contextScore := m.calculateContextScore(result, query) * m.config.ContextWeight
		semanticScore := m.calculateSemanticScore(result, query) * m.config.SemanticWeight

		// Combine scores
		result.Score = vectorScore + contextScore + semanticScore
	}

	return results
}

// calculateContextScore calculates context-based relevance score
func (m *ResultRerankingManager) calculateContextScore(result *QueryResult, query *Query) float64 {
	if query.Context == nil || result.Context == nil {
		return 0.0
	}

	var score float64
	var matches int

	// Check for context matches
	for queryKey, queryValue := range query.Context {
		if resultValue, exists := result.Context[queryKey]; exists {
			if queryValue == resultValue {
				score += 1.0
				matches++
			}
		}
	}

	if matches > 0 {
		return score / float64(matches)
	}

	return 0.0
}

// calculateSemanticScore calculates semantic relevance score
func (m *ResultRerankingManager) calculateSemanticScore(result *QueryResult, query *Query) float64 {
	// For now, use a simple heuristic based on text similarity
	// In production, you would use more sophisticated semantic analysis

	queryText := strings.ToLower(query.Text)
	resultText := ""

	if metadata, exists := result.Metadata["text"]; exists {
		if text, ok := metadata.(string); ok {
			resultText = strings.ToLower(text)
		}
	}

	if resultText == "" {
		return 0.0
	}

	// Simple word overlap scoring
	queryWords := strings.Fields(queryText)
	resultWords := strings.Fields(resultText)

	var overlap int
	for _, queryWord := range queryWords {
		for _, resultWord := range resultWords {
			if queryWord == resultWord {
				overlap++
				break
			}
		}
	}

	if len(queryWords) == 0 {
		return 0.0
	}

	return float64(overlap) / float64(len(queryWords))
}

// registerDefaultRerankers registers the default reranking strategies
func (m *ResultRerankingManager) registerDefaultRerankers() {
	m.rerankers = append(m.rerankers,
		NewSemanticReranker(nil, nil),
		NewContextAwareReranker(nil),
		NewHybridReranker(nil),
	)
}

// SemanticReranker reranks results using semantic similarity
type SemanticReranker struct {
	embeddingService embedding.Service
	config           *SemanticRerankingConfig
}

// SemanticRerankingConfig configures semantic reranking
type SemanticRerankingConfig struct {
	SimilarityThreshold float64 `json:"similarity_threshold"`
	UseLocalModels      bool    `json:"use_local_models"`
	MaxSemanticResults  int     `json:"max_semantic_results"`
}

// NewSemanticReranker creates a new semantic reranker
func NewSemanticReranker(embeddingService embedding.Service, config *SemanticRerankingConfig) *SemanticReranker {
	if config == nil {
		config = &SemanticRerankingConfig{
			SimilarityThreshold: 0.7,
			UseLocalModels:      true,
			MaxSemanticResults:  50,
		}
	}

	return &SemanticReranker{
		embeddingService: embeddingService,
		config:           config,
	}
}

// Rerank reranks results using semantic similarity
func (r *SemanticReranker) Rerank(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error) {
	if r.embeddingService == nil {
		return results, nil
	}

	// Generate query embedding
	queryEmbedding, err := r.generateQueryEmbedding(ctx, query)
	if err != nil {
		return results, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate semantic similarity for each result
	var scoredResults []*QueryResult
	for _, result := range results {
		similarity := r.calculateSemanticSimilarity(queryEmbedding, result)
		if similarity >= r.config.SimilarityThreshold {
			// Create a copy with updated score
			scoredResult := *result
			scoredResult.Score = similarity
			scoredResults = append(scoredResults, &scoredResult)
		}
	}

	// Sort by semantic similarity
	sort.Slice(scoredResults, func(i, j int) bool {
		return scoredResults[i].Score > scoredResults[j].Score
	})

	// Limit results
	if len(scoredResults) > r.config.MaxSemanticResults {
		scoredResults = scoredResults[:r.config.MaxSemanticResults]
	}

	return scoredResults, nil
}

// generateQueryEmbedding generates embedding for the query
func (r *SemanticReranker) generateQueryEmbedding(ctx context.Context, query *Query) ([]float64, error) {
	embeddingReq := &embedding.EmbeddingRequest{
		Texts:    []string{query.Text},
		Model:    "text-embedding-ada-002",
		Provider: embedding.ProviderTypeOpenAI,
	}

	embeddingResp, err := r.embeddingService.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, err
	}

	if len(embeddingResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}

	return embeddingResp.Embeddings[0], nil
}

// calculateSemanticSimilarity calculates semantic similarity between query and result
func (r *SemanticReranker) calculateSemanticSimilarity(queryEmbedding []float64, result *QueryResult) float64 {
	// For now, use cosine similarity between embeddings
	// In production, you might use more sophisticated similarity metrics

	if result.Vector == nil || len(result.Vector.Embedding) == 0 {
		return 0.0
	}

	return r.cosineSimilarity(queryEmbedding, result.Vector.Embedding)
}

// cosineSimilarity calculates cosine similarity between two vectors
func (r *SemanticReranker) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Type returns the reranker type
func (r *SemanticReranker) Type() string { return "semantic" }

// Confidence returns confidence in the reranking
func (r *SemanticReranker) Confidence() float64 { return 0.8 }

// ContextAwareReranker reranks results using context information
type ContextAwareReranker struct {
	config *ContextRerankingConfig
}

// ContextRerankingConfig configures context-aware reranking
type ContextRerankingConfig struct {
	ContextWeight     float64 `json:"context_weight"`
	UserHistoryWeight float64 `json:"user_history_weight"`
	DomainWeight      float64 `json:"domain_weight"`
	TimeWeight        float64 `json:"time_weight"`
	LocationWeight    float64 `json:"location_weight"`
}

// NewContextAwareReranker creates a new context-aware reranker
func NewContextAwareReranker(config *ContextRerankingConfig) *ContextAwareReranker {
	if config == nil {
		config = &ContextRerankingConfig{
			ContextWeight:     0.3,
			UserHistoryWeight: 0.2,
			DomainWeight:      0.2,
			TimeWeight:        0.15,
			LocationWeight:    0.15,
		}
	}

	return &ContextAwareReranker{
		config: config,
	}
}

// Rerank reranks results using context information
func (r *ContextAwareReranker) Rerank(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error) {
	contextScoredResults := make([]*QueryResult, 0)

	for _, result := range results {
		contextScore := r.calculateContextScore(result, query)

		// Create a copy with updated score
		scoredResult := *result
		scoredResult.Score = contextScore
		contextScoredResults = append(contextScoredResults, &scoredResult)
	}

	// Sort by context score
	sort.Slice(contextScoredResults, func(i, j int) bool {
		return contextScoredResults[i].Score > contextScoredResults[j].Score
	})

	return contextScoredResults, nil
}

// calculateContextScore calculates context-based relevance score
func (r *ContextAwareReranker) calculateContextScore(result *QueryResult, query *Query) float64 {
	if query.Context == nil {
		return 0.0
	}

	var totalScore float64
	var totalWeight float64

	// User history relevance
	if userHistory, exists := query.Context["user_history"]; exists {
		score := r.calculateUserHistoryScore(result, userHistory)
		totalScore += score * r.config.UserHistoryWeight
		totalWeight += r.config.UserHistoryWeight
	}

	// Domain relevance
	if domain, exists := query.Context["domain"]; exists {
		score := r.calculateDomainScore(result, domain)
		totalScore += score * r.config.DomainWeight
		totalWeight += r.config.DomainWeight
	}

	// Time relevance
	if timeContext, exists := query.Context["time_context"]; exists {
		score := r.calculateTimeScore(result, timeContext)
		totalScore += score * r.config.TimeWeight
		totalWeight += r.config.TimeWeight
	}

	// Location relevance
	if location, exists := query.Context["location"]; exists {
		score := r.calculateLocationScore(result, location)
		totalScore += score * r.config.LocationWeight
		totalWeight += r.config.LocationWeight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateUserHistoryScore calculates user history relevance score
func (r *ContextAwareReranker) calculateUserHistoryScore(result *QueryResult, userHistory interface{}) float64 {
	// Placeholder implementation
	// In production, you would analyze user interaction history
	return 0.5
}

// calculateDomainScore calculates domain relevance score
func (r *ContextAwareReranker) calculateDomainScore(result *QueryResult, domain interface{}) float64 {
	// Placeholder implementation
	// In production, you would analyze domain-specific relevance
	return 0.6
}

// calculateTimeScore calculates time relevance score
func (r *ContextAwareReranker) calculateTimeScore(result *QueryResult, timeContext interface{}) float64 {
	// Placeholder implementation
	// In production, you would analyze temporal relevance
	return 0.4
}

// calculateLocationScore calculates location relevance score
func (r *ContextAwareReranker) calculateLocationScore(result *QueryResult, location interface{}) float64 {
	// Placeholder implementation
	// In production, you would analyze geographical relevance
	return 0.3
}

// Type returns the reranker type
func (r *ContextAwareReranker) Type() string { return "context-aware" }

// Confidence returns confidence in the reranking
func (r *ContextAwareReranker) Confidence() float64 { return 0.7 }

// HybridReranker combines multiple reranking strategies
type HybridReranker struct {
	config *HybridRerankingConfig
}

// HybridRerankingConfig configures hybrid reranking
type HybridRerankingConfig struct {
	VectorWeight   float64 `json:"vector_weight"`
	SemanticWeight float64 `json:"semantic_weight"`
	ContextWeight  float64 `json:"context_weight"`
	RerankingDepth int     `json:"reranking_depth"`
}

// NewHybridReranker creates a new hybrid reranker
func NewHybridReranker(config *HybridRerankingConfig) *HybridReranker {
	if config == nil {
		config = &HybridRerankingConfig{
			VectorWeight:   0.4,
			SemanticWeight: 0.35,
			ContextWeight:  0.25,
			RerankingDepth: 2,
		}
	}

	return &HybridReranker{
		config: config,
	}
}

// Rerank combines multiple reranking strategies
func (r *HybridReranker) Rerank(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error) {
	// Apply hybrid scoring
	for _, result := range results {
		// Combine vector similarity, semantic relevance, and context relevance
		vectorScore := result.Score * r.config.VectorWeight
		semanticScore := r.calculateSemanticRelevance(result, query) * r.config.SemanticWeight
		contextScore := r.calculateContextRelevance(result, query) * r.config.ContextWeight

		result.Score = vectorScore + semanticScore + contextScore
	}

	// Sort by hybrid score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateSemanticRelevance calculates semantic relevance score
func (r *HybridReranker) calculateSemanticRelevance(result *QueryResult, query *Query) float64 {
	// Placeholder implementation
	// In production, you would use more sophisticated semantic analysis
	return 0.6
}

// calculateContextRelevance calculates context relevance score
func (r *HybridReranker) calculateContextRelevance(result *QueryResult, query *Query) float64 {
	// Placeholder implementation
	// In production, you would use more sophisticated context analysis
	return 0.5
}

// Type returns the reranker type
func (r *HybridReranker) Type() string { return "hybrid" }

// Confidence returns confidence in the reranking
func (r *HybridReranker) Confidence() float64 { return 0.9 }
