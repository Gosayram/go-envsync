// Package client provides the main SDK interface for go-envsync.
package client

import (
	"context"
	"fmt"
	"strings"
)

// Constants for client configuration
const (
	// DefaultProviderName is the name of the default provider.
	DefaultProviderName = "default"

	// MaxEnvironmentKeys defines the maximum number of keys in an environment.
	MaxEnvironmentKeys = 10000

	// MaxKeyLength defines the maximum length of an environment key.
	MaxKeyLength = 256

	// MaxValueLength defines the maximum length of an environment value.
	MaxValueLength = 4096

	// MaxProviders defines the maximum number of providers that can be registered.
	MaxProviders = 50

	// SourceProviderParts defines the expected number of parts when parsing provider:source format.
	SourceProviderParts = 2

	// MinSourceParts defines the minimum number of parts required for provider:source parsing.
	MinSourceParts = 2
)

// MergeStrategy defines how to handle conflicting keys from multiple sources.
type MergeStrategy int

const (
	// MergeStrategyOverride overwrites existing keys with new values.
	MergeStrategyOverride MergeStrategy = iota

	// MergeStrategyPreserve keeps the first value and ignores subsequent ones.
	MergeStrategyPreserve

	// MergeStrategyError returns an error if duplicate keys are found.
	MergeStrategyError
)

// Provider defines the interface for configuration providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Load loads configuration from the provider.
	Load(ctx context.Context, source string) (map[string]string, error)

	// Validate validates the source before loading.
	Validate(source string) error
}

// Validator defines the interface for configuration validation.
type Validator interface {
	// Validate validates the configuration.
	Validate(ctx context.Context, config map[string]string) error
}

// Exporter defines the interface for configuration export.
type Exporter interface {
	// Export exports configuration to the specified destination.
	Export(ctx context.Context, config map[string]string, destination string) error
}

// Client is the main client for go-envsync operations.
type Client struct {
	providers map[string]Provider
	validator Validator
	exporter  Exporter
}

// New creates a new go-envsync client.
func New() *Client {
	return &Client{
		providers: make(map[string]Provider),
	}
}

// AddProvider adds a configuration provider.
func (c *Client) AddProvider(name string, provider Provider) {
	if len(c.providers) >= MaxProviders {
		return // Silently ignore to prevent DoS
	}
	c.providers[name] = provider
}

// SetValidator sets the configuration validator.
func (c *Client) SetValidator(validator Validator) {
	c.validator = validator
}

// SetExporter sets the configuration exporter.
func (c *Client) SetExporter(exporter Exporter) {
	c.exporter = exporter
}

// LoadOptions defines options for loading configuration.
type LoadOptions struct {
	// Sources is the list of sources to load from.
	Sources []string

	// Schema is the path to the JSON schema file for validation.
	Schema string

	// MergeStrategy defines how to handle conflicting keys.
	MergeStrategy MergeStrategy
}

// Environment represents a loaded configuration environment.
type Environment struct {
	// Data contains the configuration key-value pairs.
	Data map[string]string

	// Sources contains information about the sources.
	Sources []SourceInfo

	// client reference for export operations
	client *Client
}

// SourceInfo contains information about a configuration source.
type SourceInfo struct {
	// Name is the source name or path.
	Name string

	// Provider is the provider name used to load this source.
	Provider string

	// KeyCount is the number of keys loaded from this source.
	KeyCount int
}

// Load loads configuration from the specified sources.
func (c *Client) Load(ctx context.Context, options LoadOptions) (*Environment, error) {
	// Validate options
	if len(options.Sources) == 0 {
		return nil, fmt.Errorf("no sources specified")
	}

	env := &Environment{
		Data:    make(map[string]string),
		Sources: make([]SourceInfo, 0, len(options.Sources)),
		client:  c,
	}

	// Load from each source
	for _, source := range options.Sources {
		if err := c.loadFromSource(ctx, source, env, options.MergeStrategy); err != nil {
			return nil, fmt.Errorf("failed to load from source %s: %w", source, err)
		}
	}

	// Validate if validator is set
	if c.validator != nil {
		if err := c.validator.Validate(ctx, env.Data); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Check environment size
	if len(env.Data) > MaxEnvironmentKeys {
		return nil, fmt.Errorf("too many environment keys: %d > %d", len(env.Data), MaxEnvironmentKeys)
	}

	return env, nil
}

// loadFromSource loads configuration from a single source.
func (c *Client) loadFromSource(ctx context.Context, source string, env *Environment, strategy MergeStrategy) error {
	// Parse source to determine provider
	providerName, actualSource := c.parseSource(source)

	// Get provider
	provider, exists := c.providers[providerName]
	if !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	// Validate source
	if validateErr := provider.Validate(actualSource); validateErr != nil {
		return fmt.Errorf("source validation failed for %s: %w", source, validateErr)
	}

	// Load configuration
	config, err := provider.Load(ctx, actualSource)
	if err != nil {
		return fmt.Errorf("failed to load from provider %s: %w", providerName, err)
	}

	// Merge configuration
	originalSize := len(env.Data)
	if err := c.mergeConfiguration(env.Data, config, strategy); err != nil {
		return err
	}

	// Add source info
	env.Sources = append(env.Sources, SourceInfo{
		Name:     source,
		Provider: providerName,
		KeyCount: len(env.Data) - originalSize,
	})

	return nil
}

// parseSource parses a source string and returns provider name and source path.
func (c *Client) parseSource(source string) (providerName, sourcePath string) {
	// Handle sources without provider prefix (use default)
	parts := strings.SplitN(source, ":", SourceProviderParts)
	if len(parts) == MinSourceParts {
		return parts[0], parts[1]
	}

	// Use default provider if no provider specified
	return DefaultProviderName, source
}

// mergeConfiguration merges configuration based on the merge strategy.
func (c *Client) mergeConfiguration(target, source map[string]string, strategy MergeStrategy) error {
	for key, value := range source {
		if existingValue, exists := target[key]; exists {
			switch strategy {
			case MergeStrategyError:
				return fmt.Errorf("duplicate key found: %s (existing: %s, new: %s)", key, existingValue, value)
			case MergeStrategyPreserve:
				// Keep existing value, skip new one
				continue
			case MergeStrategyOverride:
				// Override with new value (default behavior)
			}
		}

		target[key] = value
	}

	return nil
}

// Keys returns the list of configuration keys.
func (e *Environment) Keys() []string {
	keys := make([]string, 0, len(e.Data))
	for key := range e.Data {
		keys = append(keys, key)
	}
	return keys
}

// Get returns the value for the specified key.
func (e *Environment) Get(key string) (string, bool) {
	value, exists := e.Data[key]
	return value, exists
}

// Set sets a value for the specified key.
func (e *Environment) Set(key, value string) {
	e.Data[key] = value
}

// Export exports the environment using the configured exporter.
func (e *Environment) Export(ctx context.Context, destination string) error {
	if e.client.exporter == nil {
		return fmt.Errorf("no exporter configured")
	}

	return e.client.exporter.Export(ctx, e.Data, destination)
}

// ExportEnv exports the environment to the specified destination.
// This method is kept for backward compatibility.
func (e *Environment) ExportEnv(destination string) error {
	return e.Export(context.Background(), destination)
}

// Size returns the number of configuration keys.
func (e *Environment) Size() int {
	return len(e.Data)
}

// IsEmpty returns true if the environment has no configuration keys.
func (e *Environment) IsEmpty() bool {
	return len(e.Data) == 0
}
