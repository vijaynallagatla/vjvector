package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// EncryptionConfig represents encryption service configuration
type EncryptionConfig struct {
	DefaultAlgorithm string        `json:"default_algorithm"`
	DefaultKeySize   int           `json:"default_key_size"`
	DefaultMode      string        `json:"default_mode"`
	KeyRotationDays  int           `json:"key_rotation_days"`
	KeyExpiryDays    int           `json:"key_expiry_days"`
	MaxKeyUsage      int64         `json:"max_key_usage"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
}

// DefaultEncryptionConfig returns the default encryption configuration
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		DefaultAlgorithm: "AES-256",
		DefaultKeySize:   256,
		DefaultMode:      "GCM",
		KeyRotationDays:  90,
		KeyExpiryDays:    365,
		MaxKeyUsage:      1000000,
		CleanupInterval:  24 * time.Hour,
	}
}

// DefaultEncryptionService implements the encryption service
type DefaultEncryptionService struct {
	config        *EncryptionConfig
	policies      map[string]*EncryptionPolicy
	keys          map[string]*EncryptionKey
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	done          chan bool
}

// NewDefaultEncryptionService creates a new default encryption service
func NewDefaultEncryptionService(config *EncryptionConfig) *DefaultEncryptionService {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	service := &DefaultEncryptionService{
		config:   config,
		policies: make(map[string]*EncryptionPolicy),
		keys:     make(map[string]*EncryptionKey),
		done:     make(chan bool),
	}

	// Start cleanup routine
	service.startCleanup()

	return service
}

// Encrypt encrypts data using the specified policy
func (s *DefaultEncryptionService) Encrypt(ctx context.Context, data []byte, policyID string) ([]byte, error) {
	s.mu.RLock()
	policy, exists := s.policies[policyID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("encryption policy not found: %s", policyID)
	}

	if policy.Status != EncryptionPolicyStatusActive {
		return nil, fmt.Errorf("encryption policy is not active: %s", policyID)
	}

	// Get or create encryption key
	key, err := s.getOrCreateKey(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt data based on algorithm and mode
	switch policy.Algorithm {
	case "AES-256":
		return s.encryptAES(data, key, policy.Mode)
	case "ChaCha20-Poly1305":
		return s.encryptChaCha20(data, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", policy.Algorithm)
	}
}

// Decrypt decrypts data using the specified key
func (s *DefaultEncryptionService) Decrypt(ctx context.Context, encryptedData []byte, keyID string) ([]byte, error) {
	s.mu.RLock()
	key, exists := s.keys[keyID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("encryption key not found: %s", keyID)
	}

	if key.Status != EncryptionKeyStatusActive {
		return nil, fmt.Errorf("encryption key is not active: %s", keyID)
	}

	// Check if key has expired
	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("encryption key has expired: %s", keyID)
	}

	// Get policy for key
	s.mu.RLock()
	policy, exists := s.policies[key.TenantID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("encryption policy not found for key: %s", keyID)
	}

	// Decrypt data based on algorithm and mode
	switch policy.Algorithm {
	case "AES-256":
		return s.decryptAES(encryptedData, key, policy.Mode)
	case "ChaCha20-Poly1305":
		return s.decryptChaCha20(encryptedData, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", policy.Algorithm)
	}
}

// GenerateKey generates a new encryption key for a policy
func (s *DefaultEncryptionService) GenerateKey(ctx context.Context, policyID string) (*EncryptionKey, error) {
	s.mu.RLock()
	policy, exists := s.policies[policyID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("encryption policy not found: %s", policyID)
	}

	// Generate random key material
	keyMaterial := make([]byte, policy.KeySize/8)
	if _, err := rand.Read(keyMaterial); err != nil {
		return nil, fmt.Errorf("failed to generate random key material: %w", err)
	}

	// Create encryption key
	now := time.Now()
	key := &EncryptionKey{
		ID:         s.generateKeyID(),
		TenantID:   policy.TenantID,
		KeyID:      s.generateKeyID(),
		Algorithm:  policy.Algorithm,
		KeySize:    policy.KeySize,
		Status:     EncryptionKeyStatusActive,
		CreatedAt:  now,
		ExpiresAt:  now.AddDate(0, 0, s.config.KeyExpiryDays),
		LastUsed:   now,
		UsageCount: 0,
		Metadata:   make(map[string]string),
	}

	// Store key material securely (in production, use HSM or secure key storage)
	key.Metadata["key_material"] = hex.EncodeToString(keyMaterial)

	s.mu.Lock()
	s.keys[key.ID] = key
	s.mu.Unlock()

	return key, nil
}

// RotateKey rotates an existing encryption key
func (s *DefaultEncryptionService) RotateKey(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[keyID]
	if !exists {
		return fmt.Errorf("encryption key not found: %s", keyID)
	}

	// Mark old key as inactive
	key.Status = EncryptionKeyStatusInactive

	// Generate new key
	newKey, err := s.GenerateKey(ctx, key.TenantID)
	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Update key metadata
	newKey.Metadata["rotated_from"] = keyID
	newKey.Metadata["rotation_date"] = time.Now().Format(time.RFC3339)

	return nil
}

// GetKey retrieves an encryption key by ID
func (s *DefaultEncryptionService) GetKey(ctx context.Context, keyID string) (*EncryptionKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("encryption key not found: %s", keyID)
	}

	return key, nil
}

// CreatePolicy creates a new encryption policy
func (s *DefaultEncryptionService) CreatePolicy(ctx context.Context, policy *EncryptionPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if policy.ID == "" {
		policy.ID = s.generatePolicyID()
	}

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid encryption policy: %w", err)
	}

	s.policies[policy.ID] = policy
	return nil
}

// UpdatePolicy updates an existing encryption policy
func (s *DefaultEncryptionService) UpdatePolicy(ctx context.Context, policy *EncryptionPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.ID]; !exists {
		return fmt.Errorf("encryption policy not found: %s", policy.ID)
	}

	policy.UpdatedAt = time.Now()

	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid encryption policy: %w", err)
	}

	s.policies[policy.ID] = policy
	return nil
}

// GetPolicy retrieves an encryption policy by ID
func (s *DefaultEncryptionService) GetPolicy(ctx context.Context, policyID string) (*EncryptionPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("encryption policy not found: %s", policyID)
	}

	return policy, nil
}

// HealthCheck performs a health check on the service
func (s *DefaultEncryptionService) HealthCheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.policies == nil || s.keys == nil {
		return fmt.Errorf("encryption service not properly initialized")
	}

	return nil
}

// Close performs cleanup operations
func (s *DefaultEncryptionService) Close() error {
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	close(s.done)
	return nil
}

// Helper methods

// getOrCreateKey gets an existing key or creates a new one
func (s *DefaultEncryptionService) getOrCreateKey(ctx context.Context, policyID string) (*EncryptionKey, error) {
	s.mu.RLock()
	policy, exists := s.policies[policyID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("encryption policy not found: %s", policyID)
	}

	// Look for existing active key
	s.mu.RLock()
	for _, key := range s.keys {
		if key.TenantID == policy.TenantID && key.Status == EncryptionKeyStatusActive {
			// Check if key needs rotation
			// if time.Since(key.CreatedAt) > time.Duration(s.config.KeyRotationDays)*24*time.Hour {
			// 	// Key needs rotation, but continue using it for now
			// 	// In production, implement proper key rotation
			// }
			return key, nil
		}
	}
	s.mu.RUnlock()

	// Create new key if none exists
	return s.GenerateKey(ctx, policyID)
}

// encryptAES encrypts data using AES
func (s *DefaultEncryptionService) encryptAES(data []byte, key *EncryptionKey, mode string) ([]byte, error) {
	// Get key material
	keyMaterialHex, exists := key.Metadata["key_material"]
	if !exists {
		return nil, fmt.Errorf("key material not found for key: %s", key.ID)
	}

	keyMaterial, err := hex.DecodeString(keyMaterialHex)
	if err != nil {
		return nil, fmt.Errorf("invalid key material: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(keyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	switch mode {
	case "GCM":
		return s.encryptAESGCM(block, data)
	case "CBC":
		return s.encryptAESCBC(block, data)
	default:
		return nil, fmt.Errorf("unsupported AES mode: %s", mode)
	}
}

// encryptAESGCM encrypts data using AES-GCM
func (s *DefaultEncryptionService) encryptAESGCM(block cipher.Block, data []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// encryptAESCBC encrypts data using AES-CBC
func (s *DefaultEncryptionService) encryptAESCBC(block cipher.Block, data []byte) ([]byte, error) {
	// Pad data to block size
	blockSize := block.BlockSize()
	padding := blockSize - len(data)%blockSize
	paddedData := make([]byte, len(data)+padding)
	copy(paddedData, data)
	for i := len(data); i < len(paddedData); i++ {
		paddedData[i] = byte(padding)
	}

	// Generate IV
	iv := make([]byte, blockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// Prepend IV
	result := make([]byte, len(iv)+len(ciphertext))
	copy(result, iv)
	copy(result[len(iv):], ciphertext)

	return result, nil
}

// encryptChaCha20 encrypts data using ChaCha20-Poly1305
func (s *DefaultEncryptionService) encryptChaCha20(data []byte, key *EncryptionKey) ([]byte, error) {
	// For now, return error as ChaCha20 is not implemented
	// In production, implement ChaCha20-Poly1305 encryption
	return nil, fmt.Errorf("ChaCha20-Poly1305 encryption not implemented")
}

// decryptAES decrypts data using AES
func (s *DefaultEncryptionService) decryptAES(encryptedData []byte, key *EncryptionKey, mode string) ([]byte, error) {
	// Get key material
	keyMaterialHex, exists := key.Metadata["key_material"]
	if !exists {
		return nil, fmt.Errorf("key material not found for key: %s", key.ID)
	}

	keyMaterial, err := hex.DecodeString(keyMaterialHex)
	if err != nil {
		return nil, fmt.Errorf("invalid key material: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(keyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	switch mode {
	case "GCM":
		return s.decryptAESGCM(block, encryptedData)
	case "CBC":
		return s.decryptAESCBC(block, encryptedData)
	default:
		return nil, fmt.Errorf("unsupported AES mode: %s", mode)
	}
}

// decryptAESGCM decrypts data using AES-GCM
func (s *DefaultEncryptionService) decryptAESGCM(block cipher.Block, encryptedData []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// decryptAESCBC decrypts data using AES-CBC
func (s *DefaultEncryptionService) decryptAESCBC(block cipher.Block, encryptedData []byte) ([]byte, error) {
	blockSize := block.BlockSize()
	if len(encryptedData) < blockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	iv := encryptedData[:blockSize]
	ciphertext := encryptedData[blockSize:]

	if len(ciphertext)%blockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Decrypt
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	return plaintext[:len(plaintext)-padding], nil
}

// decryptChaCha20 decrypts data using ChaCha20-Poly1305
func (s *DefaultEncryptionService) decryptChaCha20(encryptedData []byte, key *EncryptionKey) ([]byte, error) {
	// For now, return error as ChaCha20 is not implemented
	return nil, fmt.Errorf("ChaCha20-Poly1305 decryption not implemented")
}

// validatePolicy validates an encryption policy
func (s *DefaultEncryptionService) validatePolicy(policy *EncryptionPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Algorithm == "" {
		return fmt.Errorf("algorithm is required")
	}

	if policy.KeySize <= 0 {
		return fmt.Errorf("key size must be positive")
	}

	if policy.Mode == "" {
		return fmt.Errorf("mode is required")
	}

	if policy.KeyRotation <= 0 {
		return fmt.Errorf("key rotation days must be positive")
	}

	return nil
}

// generateKeyID generates a unique key ID
func (s *DefaultEncryptionService) generateKeyID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("key_%s", hex.EncodeToString(bytes))
}

// generatePolicyID generates a unique policy ID
func (s *DefaultEncryptionService) generatePolicyID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("policy_%s", hex.EncodeToString(bytes))
}

// startCleanup starts the cleanup routine for expired keys
func (s *DefaultEncryptionService) startCleanup() {
	s.cleanupTicker = time.NewTicker(s.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanup()
			case <-s.done:
				return
			}
		}
	}()
}

// cleanup removes expired and compromised keys
func (s *DefaultEncryptionService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, key := range s.keys {
		// Mark expired keys
		if now.After(key.ExpiresAt) {
			key.Status = EncryptionKeyStatusExpired
		}

		// Remove keys that have exceeded usage limit
		if key.UsageCount > s.config.MaxKeyUsage {
			key.Status = EncryptionKeyStatusInactive
		}
	}
}
