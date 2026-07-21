package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// KeyStore manages API keys by provider, saving them to a JSON file securely.
type KeyStore struct {
	mu   sync.RWMutex
	path string
	Keys map[string][]string `json:"keys"` // ProviderID -> []API Keys
}

func NewKeyStore(path string) (*KeyStore, error) {
	ks := &KeyStore{
		path: path,
		Keys: make(map[string][]string),
	}
	err := ks.load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ks, nil
}

func (ks *KeyStore) load() error {
	data, err := os.ReadFile(ks.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &ks.Keys)
}

func (ks *KeyStore) save() error {
	dir := filepath.Dir(ks.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ks.Keys, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ks.path, data, 0600)
}

func (ks *KeyStore) AddKey(providerID, key string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.Keys == nil {
		ks.Keys = make(map[string][]string)
	}

	// Check if key already exists
	for _, existing := range ks.Keys[providerID] {
		if existing == key {
			return nil
		}
	}

	ks.Keys[providerID] = append(ks.Keys[providerID], key)
	return ks.save()
}

func (ks *KeyStore) GetKeys(providerID string) []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys := ks.Keys[providerID]
	result := make([]string, len(keys))
	copy(result, keys)
	return result
}

// GetConfigSummary returns the number of keys registered per provider.
func (ks *KeyStore) GetConfigSummary() map[string]int {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	summary := make(map[string]int)
	for provider, keys := range ks.Keys {
		summary[provider] = len(keys)
	}
	return summary
}
