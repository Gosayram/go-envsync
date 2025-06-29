// Package local provides a local file system provider for go-envsync.
package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Constants for local provider
const (
	// ProviderName is the name of the local provider.
	ProviderName = "local"

	// MaxFileSize defines the maximum file size in bytes that can be loaded.
	MaxFileSize = 10 * 1024 * 1024 // 10MB

	// MaxLineLength defines the maximum length of a single line in the file.
	MaxLineLength = 8192

	// DefaultEnvFile is the default environment file name.
	DefaultEnvFile = ".env"

	// WorldWritableMask is the mask for world-writable files.
	WorldWritableMask = 0o002
)

// Provider implements the local file system provider.
type Provider struct {
	basePath string
}

// NewProvider creates a new local provider with the current directory as base path.
func NewProvider() *Provider {
	return &Provider{
		basePath: ".",
	}
}

// NewProviderWithBase creates a new local provider with the specified base path.
func NewProviderWithBase(basePath string) *Provider {
	if basePath == "" {
		basePath = "."
	}

	return &Provider{
		basePath: basePath,
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return ProviderName
}

// Load loads configuration from a local file.
func (p *Provider) Load(_ context.Context, source string) (map[string]string, error) {
	// Resolve file path
	filePath := p.resolveFilePath(source)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Check file size
	if err := p.validateFileSize(filePath); err != nil {
		return nil, err
	}

	// Load environment variables
	config, err := godotenv.Read(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %s: %w", filePath, err)
	}

	// Validate loaded configuration
	if err := p.validateConfiguration(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the source before loading.
func (p *Provider) Validate(source string) error {
	// Check if source is empty
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("source cannot be empty")
	}

	// Resolve file path
	filePath := p.resolveFilePath(source)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// Check if it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", filePath)
	}

	// Check file size
	if err := p.validateFileSize(filePath); err != nil {
		return err
	}

	// Check file permissions
	if err := p.validateFilePermissions(filePath); err != nil {
		return err
	}

	return nil
}

// resolveFilePath resolves the file path relative to the base path.
func (p *Provider) resolveFilePath(source string) string {
	// If source is empty, use default
	if strings.TrimSpace(source) == "" {
		source = DefaultEnvFile
	}

	// If source is absolute, use as-is
	if filepath.IsAbs(source) {
		return source
	}

	// Resolve relative to base path
	return filepath.Join(p.basePath, source)
}

// validateFileSize validates that the file size is within acceptable limits.
func (p *Provider) validateFileSize(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	if fileInfo.Size() > MaxFileSize {
		return fmt.Errorf("file too large: %d bytes > %d bytes", fileInfo.Size(), MaxFileSize)
	}

	return nil
}

// validateFilePermissions validates that the file has appropriate permissions.
func (p *Provider) validateFilePermissions(filePath string) error {
	// Check if file is readable
	// #nosec G304 - filePath is validated and resolved from configured sources
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %w", err)
	}
	defer file.Close()

	// Check file permissions (should not be world-writable for security)
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	mode := fileInfo.Mode()
	if mode&WorldWritableMask != 0 { // World-writable
		return fmt.Errorf("file is world-writable, which is insecure: %s", filePath)
	}

	return nil
}

// validateConfiguration validates the loaded configuration.
func (p *Provider) validateConfiguration(config map[string]string) error {
	// Check for empty keys
	for key, value := range config {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty key is not allowed")
		}

		// Check value length
		if len(value) > MaxLineLength {
			return fmt.Errorf("value too long for key %s: %d > %d", key, len(value), MaxLineLength)
		}

		// Check for potentially problematic characters in keys
		if strings.ContainsAny(key, " \t\n\r=") {
			return fmt.Errorf("key contains invalid characters: %s", key)
		}
	}

	return nil
}

// SetBasePath sets the base path for resolving relative file paths.
func (p *Provider) SetBasePath(basePath string) {
	if basePath == "" {
		basePath = "."
	}
	p.basePath = basePath
}

// GetBasePath returns the current base path.
func (p *Provider) GetBasePath() string {
	return p.basePath
}
