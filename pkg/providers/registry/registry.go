// Package registry provides a centralized registry for configuration providers.
package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Gosayram/go-envsync/pkg/client"
)

// Constants for registry
const (
	// MaxProviders defines the maximum number of providers that can be registered.
	MaxProviders = 100

	// DefaultProviderPriority is the default priority for providers.
	DefaultProviderPriority = 50

	// HighPriority for high-priority providers.
	HighPriority = 10

	// LowPriority for low-priority providers.
	LowPriority = 90
)

// ProviderFactory is a function that creates a new provider instance.
type ProviderFactory func(config map[string]interface{}) (client.Provider, error)

// ProviderInfo contains information about a registered provider.
type ProviderInfo struct {
	// Name is the provider name.
	Name string

	// Aliases are alternative names for the provider.
	Aliases []string

	// Factory is the function to create provider instances.
	Factory ProviderFactory

	// Priority defines the provider priority (lower = higher priority).
	Priority int

	// Description is a human-readable description of the provider.
	Description string

	// SupportedSources lists the supported source formats.
	SupportedSources []string

	// RequiredConfig lists the required configuration keys.
	RequiredConfig []string

	// OptionalConfig lists the optional configuration keys.
	OptionalConfig []string
}

// Registry manages provider registration and creation.
type Registry struct {
	providers map[string]*ProviderInfo
	aliases   map[string]string
	mutex     sync.RWMutex
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]*ProviderInfo),
		aliases:   make(map[string]string),
	}
}

// Register registers a new provider with the registry.
func (r *Registry) Register(info *ProviderInfo) error {
	if info == nil {
		return fmt.Errorf("provider info cannot be nil")
	}

	if strings.TrimSpace(info.Name) == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if info.Factory == nil {
		return fmt.Errorf("provider factory cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if registry is full
	if len(r.providers) >= MaxProviders {
		return fmt.Errorf("registry is full (max %d providers)", MaxProviders)
	}

	// Check if provider already exists
	if _, exists := r.providers[info.Name]; exists {
		return fmt.Errorf("provider %s already registered", info.Name)
	}

	// Set default priority if not specified
	if info.Priority == 0 {
		info.Priority = DefaultProviderPriority
	}

	// Register provider
	r.providers[info.Name] = info

	// Register aliases
	for _, alias := range info.Aliases {
		alias = strings.TrimSpace(alias)
		if alias != "" {
			// Check if alias conflicts with existing provider names
			if _, exists := r.providers[alias]; exists {
				return fmt.Errorf("alias %s conflicts with existing provider", alias)
			}

			// Check if alias already exists
			if _, exists := r.aliases[alias]; exists {
				return fmt.Errorf("alias %s already registered", alias)
			}

			r.aliases[alias] = info.Name
		}
	}

	return nil
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if provider exists
	info, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	// Remove aliases
	for _, alias := range info.Aliases {
		delete(r.aliases, alias)
	}

	// Remove provider
	delete(r.providers, name)

	return nil
}

// CreateProvider creates a new provider instance.
func (r *Registry) CreateProvider(name string, config map[string]interface{}) (client.Provider, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Resolve provider name (handle aliases)
	providerName := r.resolveProviderName(name)

	// Get provider info
	info, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	// Validate required configuration
	if err := r.validateConfig(info, config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create provider instance
	provider, err := info.Factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", name, err)
	}

	return provider, nil
}

// GetProvider returns information about a registered provider.
func (r *Registry) GetProvider(name string) (*ProviderInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Resolve provider name
	providerName := r.resolveProviderName(name)

	// Get provider info
	info, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	// Return a copy to prevent modification
	return &ProviderInfo{
		Name:             info.Name,
		Aliases:          append([]string{}, info.Aliases...),
		Factory:          info.Factory,
		Priority:         info.Priority,
		Description:      info.Description,
		SupportedSources: append([]string{}, info.SupportedSources...),
		RequiredConfig:   append([]string{}, info.RequiredConfig...),
		OptionalConfig:   append([]string{}, info.OptionalConfig...),
	}, nil
}

// ListProviders returns a list of all registered providers.
func (r *Registry) ListProviders() []*ProviderInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providers := make([]*ProviderInfo, 0, len(r.providers))
	for _, info := range r.providers {
		// Return a copy to prevent modification
		providers = append(providers, &ProviderInfo{
			Name:             info.Name,
			Aliases:          append([]string{}, info.Aliases...),
			Factory:          info.Factory,
			Priority:         info.Priority,
			Description:      info.Description,
			SupportedSources: append([]string{}, info.SupportedSources...),
			RequiredConfig:   append([]string{}, info.RequiredConfig...),
			OptionalConfig:   append([]string{}, info.OptionalConfig...),
		})
	}

	return providers
}

// GetProviderNames returns a list of all registered provider names and aliases.
func (r *Registry) GetProviderNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.providers)+len(r.aliases))

	// Add provider names
	for name := range r.providers {
		names = append(names, name)
	}

	// Add aliases
	for alias := range r.aliases {
		names = append(names, alias)
	}

	return names
}

// IsProviderRegistered checks if a provider is registered.
func (r *Registry) IsProviderRegistered(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providerName := r.resolveProviderName(name)
	_, exists := r.providers[providerName]
	return exists
}

// resolveProviderName resolves an alias to the actual provider name.
func (r *Registry) resolveProviderName(name string) string {
	// Check if it's an alias
	if actualName, exists := r.aliases[name]; exists {
		return actualName
	}

	// Return the name as-is (might be a provider name or unknown)
	return name
}

// validateConfig validates the provider configuration.
func (r *Registry) validateConfig(info *ProviderInfo, config map[string]interface{}) error {
	// Check required configuration
	for _, key := range info.RequiredConfig {
		if _, exists := config[key]; !exists {
			return fmt.Errorf("required configuration key missing: %s", key)
		}
	}

	return nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// Register registers a provider with the global registry.
func Register(info *ProviderInfo) error {
	return globalRegistry.Register(info)
}

// CreateProvider creates a provider using the global registry.
func CreateProvider(name string, config map[string]interface{}) (client.Provider, error) {
	return globalRegistry.CreateProvider(name, config)
}

// GetProvider gets provider information from the global registry.
func GetProvider(name string) (*ProviderInfo, error) {
	return globalRegistry.GetProvider(name)
}

// ListProviders lists all providers in the global registry.
func ListProviders() []*ProviderInfo {
	return globalRegistry.ListProviders()
}

// IsProviderRegistered checks if a provider is registered in the global registry.
func IsProviderRegistered(name string) bool {
	return globalRegistry.IsProviderRegistered(name)
}

// GetProviderNames returns all provider names from the global registry.
func GetProviderNames() []string {
	return globalRegistry.GetProviderNames()
}
