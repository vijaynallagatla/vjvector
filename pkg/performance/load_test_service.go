package performance

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultLoadTestService implements the load testing service
type DefaultLoadTestService struct {
	tests      map[string]*LoadTestResult
	testStatus map[string]*TestStatus
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewDefaultLoadTestService creates a new default load test service
func NewDefaultLoadTestService() *DefaultLoadTestService {
	// Create HTTP client with reasonable timeouts
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	return &DefaultLoadTestService{
		tests:      make(map[string]*LoadTestResult),
		testStatus: make(map[string]*TestStatus),
		httpClient: httpClient,
	}
}

// RunTest executes a load test with the given configuration
func (s *DefaultLoadTestService) RunTest(ctx context.Context, config *LoadTestConfig) (*LoadTestResult, error) {
	testID := s.generateTestID()

	// Create test status
	status := &TestStatus{
		TestID:            testID,
		Status:            "running",
		Progress:          0.0,
		StartTime:         time.Now(),
		EstimatedEnd:      time.Now().Add(config.Duration),
		CurrentLoad:       0,
		CurrentLatency:    0,
		CurrentThroughput: 0.0,
		ErrorCount:        0,
	}

	s.mu.Lock()
	s.testStatus[testID] = status
	s.mu.Unlock()

	// Start test execution in background
	go s.executeTest(ctx, testID, config, status)

	// Return initial result
	result := &LoadTestResult{
		TestID:             testID,
		Config:             config,
		StartTime:          status.StartTime,
		Duration:           0,
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		ErrorRate:          0.0,
		AverageLatency:     0,
		P95Latency:         0,
		P99Latency:         0,
		Throughput:         0.0,
		ResourceUsage:      &ResourceUsage{},
		ScenarioResults:    []*ScenarioResult{},
		Recommendations:    []string{},
	}

	s.mu.Lock()
	s.tests[testID] = result
	s.mu.Unlock()

	return result, nil
}

// StopTest stops a running load test
func (s *DefaultLoadTestService) StopTest(ctx context.Context, testID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.testStatus[testID]
	if !exists {
		return fmt.Errorf("test not found: %s", testID)
	}

	if status.Status == "running" {
		status.Status = "stopped"
		status.Progress = 1.0
		now := time.Now()
		status.EndTime = &now
	}

	return nil
}

// GetTestStatus retrieves the current status of a test
func (s *DefaultLoadTestService) GetTestStatus(ctx context.Context, testID string) (*TestStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.testStatus[testID]
	if !exists {
		return nil, fmt.Errorf("test not found: %s", testID)
	}

	return status, nil
}

// GetTestResults retrieves the results of a completed test
func (s *DefaultLoadTestService) GetTestResults(ctx context.Context, testID string) (*LoadTestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.tests[testID]
	if !exists {
		return nil, fmt.Errorf("test not found: %s", testID)
	}

	return result, nil
}

// ListTests lists all load tests
func (s *DefaultLoadTestService) ListTests(ctx context.Context) ([]*TestInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	testInfos := make([]*TestInfo, 0, len(s.tests))
	for testID, result := range s.tests {
		status, exists := s.testStatus[testID]
		if !exists {
			continue
		}

		testInfo := &TestInfo{
			TestID:         testID,
			Name:           fmt.Sprintf("Load Test %s", testID),
			Status:         status.Status,
			StartTime:      result.StartTime,
			EndTime:        status.EndTime,
			Duration:       result.Duration,
			TotalRequests:  result.TotalRequests,
			ErrorRate:      result.ErrorRate,
			AverageLatency: result.AverageLatency,
		}

		testInfos = append(testInfos, testInfo)
	}

	return testInfos, nil
}

// DeleteTest removes a test and its results
func (s *DefaultLoadTestService) DeleteTest(ctx context.Context, testID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tests, testID)
	delete(s.testStatus, testID)

	return nil
}

// ExportResults exports test results in the specified format
func (s *DefaultLoadTestService) ExportResults(ctx context.Context, testID string, format string) ([]byte, error) {
	result, err := s.GetTestResults(ctx, testID)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.Marshal(result)
	case "csv":
		return s.exportToCSV(result)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// HealthCheck performs a health check on the service
func (s *DefaultLoadTestService) HealthCheck(ctx context.Context) error {
	if s.tests == nil || s.testStatus == nil || s.httpClient == nil {
		return fmt.Errorf("load test service not properly initialized")
	}

	return nil
}

// Close performs cleanup operations
func (s *DefaultLoadTestService) Close() error {
	// Clean up any resources if needed
	return nil
}

// Helper methods

// executeTest runs the actual load test
//
//nolint:gocognit,funlen,gocyclo,gocritic
func (s *DefaultLoadTestService) executeTest(ctx context.Context, testID string, config *LoadTestConfig, status *TestStatus) {
	defer func() {
		status.Status = "completed"
		status.Progress = 1.0
		now := time.Now()
		status.EndTime = &now
	}()

	// Initialize metrics
	var totalRequests int64
	var successfulRequests int64
	var failedRequests int64
	var totalLatency time.Duration
	var latencies []time.Duration
	var scenarioResults = make(map[string]*ScenarioResult)

	// Initialize scenario results
	for _, scenario := range config.TestScenarios {
		scenarioResults[scenario.Name] = &ScenarioResult{
			ScenarioName:       scenario.Name,
			TotalRequests:      0,
			SuccessfulRequests: 0,
			FailedRequests:     0,
			ErrorRate:          0.0,
			AverageLatency:     0,
			P95Latency:         0,
			P99Latency:         0,
			Throughput:         0.0,
		}
	}

	// Calculate ramp-up and ramp-down steps
	rampUpSteps := int(config.RampUpTime.Seconds())
	rampDownSteps := int(config.RampDownTime.Seconds())
	steadyStateSteps := int(config.Duration.Seconds()) - rampUpSteps - rampDownSteps

	// Start time tracking
	startTime := time.Now()
	testDuration := config.Duration
	stepDuration := time.Second

	// Execute test steps
	for step := 0; step < int(testDuration.Seconds()); step++ {
		select {
		case <-ctx.Done():
			status.Status = "stopped"
			return
		default:
			// Calculate current concurrency based on ramp-up/steady-state/ramp-down
			var currentConcurrency int
			if step < rampUpSteps {
				// Ramp up
				currentConcurrency = int(float64(config.Concurrency) * float64(step+1) / float64(rampUpSteps))
			} else if step >= rampUpSteps+steadyStateSteps {
				// Ramp down
				rampDownStep := step - rampUpSteps - steadyStateSteps
				currentConcurrency = int(float64(config.Concurrency) * (1.0 - float64(rampDownStep)/float64(rampDownSteps)))
			} else {
				// Steady state
				currentConcurrency = config.Concurrency
			}

			// Update status
			status.CurrentLoad = currentConcurrency
			status.Progress = float64(step) / float64(int(testDuration.Seconds()))

			// Execute requests for this step
			stepResults := s.executeStep(ctx, config, currentConcurrency, stepDuration)

			// Aggregate results
			totalRequests += stepResults.TotalRequests
			successfulRequests += stepResults.SuccessfulRequests
			failedRequests += stepResults.FailedRequests
			totalLatency += stepResults.TotalLatency
			latencies = append(latencies, stepResults.Latencies...)

			// Update scenario results
			for scenarioName, scenarioResult := range stepResults.ScenarioResults {
				if existing, exists := scenarioResults[scenarioName]; exists {
					existing.TotalRequests += scenarioResult.TotalRequests
					existing.SuccessfulRequests += scenarioResult.SuccessfulRequests
					existing.FailedRequests += scenarioResult.FailedRequests
				}
			}

			// Update current metrics
			if len(stepResults.Latencies) > 0 {
				status.CurrentLatency = stepResults.AverageLatency
				status.CurrentThroughput = stepResults.Throughput
			}

			time.Sleep(stepDuration)
		}
	}

	// Calculate final metrics
	duration := time.Since(startTime)
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(failedRequests) / float64(totalRequests)
	}

	// Calculate latency percentiles
	var p95Latency, p99Latency time.Duration
	if len(latencies) > 0 {
		sortLatencies(latencies)
		p95Index := int(float64(len(latencies)) * 0.95)
		p99Index := int(float64(len(latencies)) * 0.99)

		if p95Index < len(latencies) {
			p95Latency = latencies[p95Index]
		}
		if p99Index < len(latencies) {
			p99Latency = latencies[p99Index]
		}
	}

	// Calculate throughput
	throughput := 0.0
	if duration.Seconds() > 0 {
		throughput = float64(totalRequests) / duration.Seconds()
	}

	// Update final results
	s.mu.Lock()
	if result, exists := s.tests[testID]; exists {
		result.EndTime = time.Now()
		result.Duration = duration
		result.TotalRequests = totalRequests
		result.SuccessfulRequests = successfulRequests
		result.FailedRequests = failedRequests
		result.ErrorRate = errorRate
		result.AverageLatency = totalLatency / time.Duration(totalRequests)
		result.P95Latency = p95Latency
		result.P99Latency = p99Latency
		result.Throughput = throughput
		result.ScenarioResults = s.convertScenarioResults(scenarioResults)
		result.Recommendations = s.generateRecommendations(config, errorRate, throughput, p95Latency)
	}
	s.mu.Unlock()

	// Update final status
	status.Status = "completed"
	status.Progress = 1.0
	now := time.Now()
	status.EndTime = &now
}

// executeStep executes requests for a single time step
func (s *DefaultLoadTestService) executeStep(ctx context.Context, config *LoadTestConfig, concurrency int, stepDuration time.Duration) *stepResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	result := &stepResult{
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		TotalLatency:       0,
		Latencies:          make([]time.Duration, 0),
		ScenarioResults:    make(map[string]*scenarioStepResult),
	}

	// Initialize scenario results
	for _, scenario := range config.TestScenarios {
		result.ScenarioResults[scenario.Name] = &scenarioStepResult{
			TotalRequests:      0,
			SuccessfulRequests: 0,
			FailedRequests:     0,
		}
	}

	// Start concurrent requests
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Select scenario based on weights
			scenario := s.selectScenario(config.TestScenarios)
			if scenario == nil {
				return
			}

			// Execute request
			startTime := time.Now()
			err := s.executeRequest(ctx, scenario)
			latency := time.Since(startTime)

			mu.Lock()
			result.TotalRequests++
			result.TotalLatency += latency
			result.Latencies = append(result.Latencies, latency)

			scenarioResult := result.ScenarioResults[scenario.Name]
			scenarioResult.TotalRequests++

			if err != nil {
				result.FailedRequests++
				scenarioResult.FailedRequests++
			} else {
				result.SuccessfulRequests++
				scenarioResult.SuccessfulRequests++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Calculate step metrics
	if result.TotalRequests > 0 {
		result.AverageLatency = result.TotalLatency / time.Duration(result.TotalRequests)
		result.Throughput = float64(result.TotalRequests) / stepDuration.Seconds()
	}

	return result
}

// executeRequest executes a single HTTP request
func (s *DefaultLoadTestService) executeRequest(ctx context.Context, scenario *TestScenario) error {
	// Create request
	var req *http.Request
	var err error

	if scenario.Body != nil {
		_, err := json.Marshal(scenario.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %v", err)
		}
		req, err = http.NewRequestWithContext(ctx, scenario.Method, scenario.Endpoint, io.NopCloser(nil))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.Body = io.NopCloser(nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, scenario.Method, scenario.Endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
	}

	// Add headers
	for key, value := range scenario.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != scenario.ExpectedStatus {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, scenario.ExpectedStatus)
	}

	return nil
}

// selectScenario selects a test scenario based on weights
func (s *DefaultLoadTestService) selectScenario(scenarios []TestScenario) *TestScenario {
	if len(scenarios) == 0 {
		return nil
	}

	// Simple weighted selection (in production, use proper weighted random)
	totalWeight := 0.0
	for _, scenario := range scenarios {
		totalWeight += scenario.Weight
	}

	if totalWeight == 0 {
		return &scenarios[0]
	}

	// For now, just return the first scenario
	// In production, implement proper weighted random selection
	return &scenarios[0]
}

// generateTestID generates a unique test ID
func (s *DefaultLoadTestService) generateTestID() string {
	return fmt.Sprintf("load_test_%d", time.Now().UnixNano())
}

// generateRecommendations generates performance recommendations
func (s *DefaultLoadTestService) generateRecommendations(config *LoadTestConfig, errorRate, throughput float64, p95Latency time.Duration) []string {
	var recommendations []string

	if errorRate > config.ErrorThreshold {
		recommendations = append(recommendations, "Error rate exceeds threshold - investigate system stability")
	}

	if p95Latency > config.TargetLatency {
		recommendations = append(recommendations, "P95 latency exceeds target - optimize response times")
	}

	if throughput < float64(config.RequestRate) {
		recommendations = append(recommendations, "Throughput below target - increase system capacity")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Performance meets targets")
	}

	return recommendations
}

// convertScenarioResults converts step results to final scenario results
func (s *DefaultLoadTestService) convertScenarioResults(stepResults map[string]*ScenarioResult) []*ScenarioResult {
	var results []*ScenarioResult
	for _, result := range stepResults {
		if result.TotalRequests > 0 {
			result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests)
			results = append(results, result)
		}
	}
	return results
}

// exportToCSV exports results to CSV format
func (s *DefaultLoadTestService) exportToCSV(result *LoadTestResult) ([]byte, error) {
	// Simple CSV export (in production, use proper CSV library)
	csv := "Test ID,Start Time,Duration,Total Requests,Successful Requests,Failed Requests,Error Rate,Average Latency,P95 Latency,P99 Latency,Throughput\n"
	csv += fmt.Sprintf("%s,%s,%v,%d,%d,%d,%.2f,%v,%v,%v,%.2f\n",
		result.TestID,
		result.StartTime.Format(time.RFC3339),
		result.Duration,
		result.TotalRequests,
		result.SuccessfulRequests,
		result.FailedRequests,
		result.ErrorRate,
		result.AverageLatency,
		result.P95Latency,
		result.P99Latency,
		result.Throughput)

	return []byte(csv), nil
}

// sortLatencies sorts latencies for percentile calculation
func sortLatencies(latencies []time.Duration) {
	// Simple bubble sort (in production, use proper sorting)
	for i := 0; i < len(latencies)-1; i++ {
		for j := 0; j < len(latencies)-i-1; j++ {
			if latencies[j] > latencies[j+1] {
				latencies[j], latencies[j+1] = latencies[j+1], latencies[j]
			}
		}
	}
}

// Helper structs

// stepResult represents results for a single test step
type stepResult struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalLatency       time.Duration
	AverageLatency     time.Duration
	Throughput         float64
	Latencies          []time.Duration
	ScenarioResults    map[string]*scenarioStepResult
}

// scenarioStepResult represents results for a scenario in a single step
type scenarioStepResult struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
}
