package embedding

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// // MockProvider is a mock implementation of the Provider interface for testing
// type MockProvider struct {
// 	name string
// }

// func (m *MockProvider) Type() ProviderType {
// 	return ProviderTypeLocal
// }

// func (m *MockProvider) Name() string {
// 	return m.name
// }

// func (m *MockProvider) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
// 	return nil, nil
// }

// func (m *MockProvider) GetModels(ctx context.Context) ([]Model, error) {
// 	return nil, nil
// }

// func (m *MockProvider) GetCapabilities() Capabilities {
// 	return Capabilities{}
// }

// func (m *MockProvider) HealthCheck(ctx context.Context) error {
// 	return nil
// }

// func (m *MockProvider) Close() error {
// 	return nil
// }

func TestNewModelManager(t *testing.T) {
	tests := []struct {
		name   string
		config ModelManagerConfig
	}{
		{
			name: "default config",
			config: ModelManagerConfig{
				UpdateInterval:    24 * time.Hour,
				MaxModels:         100,
				CleanupInterval:   1 * time.Hour,
				PerformanceWindow: 1 * time.Hour,
			},
		},
		{
			name: "custom config",
			config: ModelManagerConfig{
				AutoUpdate:        true,
				UpdateInterval:    2 * time.Hour,
				MaxModels:         50,
				CleanupInterval:   30 * time.Minute,
				PerformanceWindow: 30 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewModelManager(tt.config)
			require.NotNil(t, manager)
			assert.NotNil(t, manager.logger)
			assert.Equal(t, tt.config, manager.config)
		})
	}
}

func TestModelManager_RegisterModel(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Test successful registration
	model := &ManagedModel{
		ID:       "test-model-1",
		Name:     "Test Model 1",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}

	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Test duplicate registration
	err = manager.RegisterModel(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test empty ID
	emptyModel := &ManagedModel{
		Name: "Empty Model",
	}
	err = manager.RegisterModel(emptyModel)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestModelManager_GetModel(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Register a model
	model := &ManagedModel{
		ID:       "test-model-2",
		Name:     "Test Model 2",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}
	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Test successful retrieval
	retrievedModel, exists := manager.GetModel("test-model-2")
	assert.True(t, exists)
	assert.Equal(t, model, retrievedModel)

	// Test non-existent model
	_, exists = manager.GetModel("non-existent")
	assert.False(t, exists)
}

func TestModelManager_ListModels(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Initially no models
	models := manager.ListModels()
	assert.Len(t, models, 0)

	// Register models
	model1 := &ManagedModel{
		ID:       "model-1",
		Name:     "Model 1",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}
	model2 := &ManagedModel{
		ID:       "model-2",
		Name:     "Model 2",
		Version:  "2.0.0",
		Type:     "openai",
		Provider: "openai",
	}

	err := manager.RegisterModel(model1)
	require.NoError(t, err)
	err = manager.RegisterModel(model2)
	require.NoError(t, err)

	// List all models
	models = manager.ListModels()
	assert.Len(t, models, 2)

	// Verify models are present
	modelIDs := make(map[string]bool)
	for _, m := range models {
		modelIDs[m.ID] = true
	}
	assert.True(t, modelIDs["model-1"])
	assert.True(t, modelIDs["model-2"])
}

func TestModelManager_UpdateModel(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Register a model
	model := &ManagedModel{
		ID:       "update-test-model",
		Name:     "Update Test Model",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}
	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Update model
	updates := map[string]interface{}{
		"name":    "Updated Model Name",
		"version": "2.0.0",
		"status":  ModelStatusDeprecated,
	}

	err = manager.UpdateModel("update-test-model", updates)
	require.NoError(t, err)

	// Verify updates
	updatedModel, exists := manager.GetModel("update-test-model")
	assert.True(t, exists)
	assert.Equal(t, "Updated Model Name", updatedModel.Name)
	assert.Equal(t, "2.0.0", updatedModel.Version)
	assert.Equal(t, ModelStatusDeprecated, updatedModel.Status)
	assert.False(t, updatedModel.UpdatedAt.IsZero())

	// Test updating non-existent model
	err = manager.UpdateModel("non-existent", updates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelManager_DeleteModel(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Register a model
	model := &ManagedModel{
		ID:       "delete-test-model",
		Name:     "Delete Test Model",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}
	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Verify model exists
	_, exists := manager.GetModel("delete-test-model")
	assert.True(t, exists)

	// Delete model
	err = manager.DeleteModel("delete-test-model")
	require.NoError(t, err)

	// Verify model is deleted
	_, exists = manager.GetModel("delete-test-model")
	assert.False(t, exists)

	// Test deleting non-existent model
	err = manager.DeleteModel("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelManager_RegisterProvider(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Create a mock provider
	mockProvider := &MockProvider{
		name: "mock-provider",
	}

	// Test successful registration
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Test duplicate registration
	err = manager.RegisterProvider(mockProvider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestModelManager_GetProvider(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Register a provider
	mockProvider := &MockProvider{
		name: "test-provider",
	}
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Test successful retrieval
	retrievedProvider, exists := manager.GetProvider("test-provider")
	assert.True(t, exists)
	assert.Equal(t, mockProvider, retrievedProvider)

	// Test non-existent provider
	_, exists = manager.GetProvider("non-existent")
	assert.False(t, exists)
}

func TestModelManager_ListProviders(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Initially no providers
	providers := manager.ListProviders()
	assert.Len(t, providers, 0)

	// Register providers
	provider1 := &MockProvider{name: "provider-1"}
	provider2 := &MockProvider{name: "provider-2"}

	err := manager.RegisterProvider(provider1)
	require.NoError(t, err)
	err = manager.RegisterProvider(provider2)
	require.NoError(t, err)

	// List all providers
	providers = manager.ListProviders()
	assert.Len(t, providers, 2)

	// Verify providers are present
	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerNames[p.Name()] = true
	}
	assert.True(t, providerNames["provider-1"])
	assert.True(t, providerNames["provider-2"])
}

func TestModelManager_UpdateModelPerformance(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Register a model
	model := &ManagedModel{
		ID:       "performance-test-model",
		Name:     "Performance Test Model",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}
	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Update performance
	performance := ModelPerformance{
		AverageLatency: 100.5,
		Throughput:     50.0,
		Accuracy:       0.95,
		MemoryUsage:    1024 * 1024, // 1MB
		GPUUtilization: 75.0,
		ErrorRate:      0.01,
	}

	err = manager.UpdateModelPerformance("performance-test-model", performance)
	require.NoError(t, err)

	// Verify performance update
	updatedModel, exists := manager.GetModel("performance-test-model")
	assert.True(t, exists)
	assert.Equal(t, performance.AverageLatency, updatedModel.Performance.AverageLatency)
	assert.Equal(t, performance.Throughput, updatedModel.Performance.Throughput)
	assert.Equal(t, performance.Accuracy, updatedModel.Performance.Accuracy)
	assert.Equal(t, performance.MemoryUsage, updatedModel.Performance.MemoryUsage)
	assert.Equal(t, performance.GPUUtilization, updatedModel.Performance.GPUUtilization)
	assert.Equal(t, performance.ErrorRate, updatedModel.Performance.ErrorRate)
	assert.False(t, updatedModel.Performance.LastUpdated.IsZero())
	assert.False(t, updatedModel.LastUsed.IsZero())
	assert.Equal(t, int64(1), updatedModel.UsageCount)

	// Test updating non-existent model
	err = manager.UpdateModelPerformance("non-existent", performance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelManager_GetModelStats(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Initially empty stats
	stats := manager.GetModelStats()
	assert.Equal(t, 0, stats["total_models"])
	assert.Equal(t, 0, stats["total_providers"])
	assert.Equal(t, int64(0), stats["total_usage"])

	// Register models and providers
	model1 := &ManagedModel{
		ID:       "stats-model-1",
		Name:     "Stats Model 1",
		Type:     "sentence-transformers",
		Provider: "local",
		Status:   ModelStatusReady,
	}
	model2 := &ManagedModel{
		ID:       "stats-model-2",
		Name:     "Stats Model 2",
		Type:     "openai",
		Provider: "openai",
		Status:   ModelStatusReady,
	}

	err := manager.RegisterModel(model1)
	require.NoError(t, err)
	err = manager.RegisterModel(model2)
	require.NoError(t, err)

	provider := &MockProvider{name: "stats-provider"}
	err = manager.RegisterProvider(provider)
	require.NoError(t, err)

	// Update usage count
	_ = manager.UpdateModelPerformance("stats-model-1", ModelPerformance{})
	_ = manager.UpdateModelPerformance("stats-model-2", ModelPerformance{})

	// Get updated stats
	stats = manager.GetModelStats()
	assert.Equal(t, 2, stats["total_models"])
	assert.Equal(t, 1, stats["total_providers"])
	assert.Equal(t, int64(2), stats["total_usage"])

	// Check status breakdown
	statusBreakdown := stats["status_breakdown"].(map[ModelStatus]int)
	assert.Equal(t, 2, statusBreakdown[ModelStatusReady])

	// Check type breakdown
	typeBreakdown := stats["type_breakdown"].(map[string]int)
	assert.Equal(t, 1, typeBreakdown["sentence-transformers"])
	assert.Equal(t, 1, typeBreakdown["openai"])
}

func TestModelManager_DefaultValues(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Test that default values are set correctly
	assert.Equal(t, 24*time.Hour, manager.config.UpdateInterval)
	assert.Equal(t, 100, manager.config.MaxModels)
	assert.Equal(t, 1*time.Hour, manager.config.CleanupInterval)
	assert.Equal(t, 1*time.Hour, manager.config.PerformanceWindow)
}

func TestModelManager_ModelLifecycle(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Test model lifecycle: create -> update -> delete
	model := &ManagedModel{
		ID:       "lifecycle-model",
		Name:     "Lifecycle Model",
		Version:  "1.0.0",
		Type:     "sentence-transformers",
		Provider: "local",
	}

	// Create
	err := manager.RegisterModel(model)
	require.NoError(t, err)

	// Verify creation
	createdModel, exists := manager.GetModel("lifecycle-model")
	assert.True(t, exists)
	assert.False(t, createdModel.CreatedAt.IsZero())
	assert.False(t, createdModel.UpdatedAt.IsZero())
	assert.Equal(t, ModelStatusReady, createdModel.Status)
	assert.NotNil(t, createdModel.Metadata)

	// Update
	updates := map[string]interface{}{
		"status": ModelStatusUpdating,
	}
	err = manager.UpdateModel("lifecycle-model", updates)
	require.NoError(t, err)

	// Verify update
	updatedModel, exists := manager.GetModel("lifecycle-model")
	assert.True(t, exists)
	assert.Equal(t, ModelStatusUpdating, updatedModel.Status)
	assert.True(t, updatedModel.UpdatedAt.After(updatedModel.CreatedAt))

	// Delete
	err = manager.DeleteModel("lifecycle-model")
	require.NoError(t, err)

	// Verify deletion
	_, exists = manager.GetModel("lifecycle-model")
	assert.False(t, exists)
}

func TestModelManager_Concurrency(t *testing.T) {
	manager := NewModelManager(ModelManagerConfig{})

	// Test concurrent model registration
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			model := &ManagedModel{
				ID:       fmt.Sprintf("concurrent-model-%d", id),
				Name:     fmt.Sprintf("Concurrent Model %d", id),
				Version:  "1.0.0",
				Type:     "sentence-transformers",
				Provider: "local",
			}
			err := manager.RegisterModel(model)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify all models were registered
	models := manager.ListModels()
	assert.Len(t, models, numGoroutines)
}
