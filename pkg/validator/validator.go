// Package validator provides configuration validation using JSON Schema and custom rules.
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Constants for validation
const (
	// DefaultSchemaFile is the default schema file name.
	DefaultSchemaFile = ".envschema.json"

	// MaxConfigKeys defines the maximum number of configuration keys allowed.
	MaxConfigKeys = 1000

	// MaxKeyLength defines the maximum length of a configuration key.
	MaxKeyLength = 256

	// MaxValueLength defines the maximum length of a configuration value.
	MaxValueLength = 4096
)

// SchemaValidator implements configuration validation using JSON Schema.
type SchemaValidator struct {
	schemaPath string
	schema     *gojsonschema.Schema
}

// NewSchemaValidator creates a new JSON Schema validator.
func NewSchemaValidator(schemaPath string) (*SchemaValidator, error) {
	if schemaPath == "" {
		schemaPath = DefaultSchemaFile
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve schema path: %w", err)
	}

	// Check if schema file exists
	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("schema file not found: %s", absPath)
	}

	// Load schema using absolute file path
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + absPath)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &SchemaValidator{
		schemaPath: absPath,
		schema:     schema,
	}, nil
}

// Validate validates configuration against the JSON schema.
func (v *SchemaValidator) Validate(_ context.Context, config map[string]string) error {
	// Check if schema file exists
	absPath, err := filepath.Abs(v.schemaPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return fmt.Errorf("schema file not found: %s", absPath)
	}

	// Load schema
	// #nosec G304 - absPath is validated and resolved from a configured schema path
	schemaData, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return fmt.Errorf("failed to read schema file: %w", readErr)
	}

	// Parse schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)

	// Convert config to JSON for validation
	configJSON, marshalErr := json.Marshal(config)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", marshalErr)
	}

	// Create document loader
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Validate
	result, validateErr := gojsonschema.Validate(schemaLoader, documentLoader)
	if validateErr != nil {
		return fmt.Errorf("schema validation failed: %w", validateErr)
	}

	// Check validation result
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CustomValidator implements custom validation rules.
type CustomValidator struct {
	rules []ValidationRule
}

// ValidationRule defines a custom validation rule.
type ValidationRule interface {
	// Name returns the rule name.
	Name() string

	// Validate validates a configuration key-value pair.
	Validate(key, value string) error
}

// NewCustomValidator creates a new custom validator.
func NewCustomValidator(rules ...ValidationRule) *CustomValidator {
	return &CustomValidator{
		rules: rules,
	}
}

// Validate validates configuration using custom rules.
func (v *CustomValidator) Validate(_ context.Context, config map[string]string) error {
	// Check maximum number of keys
	if len(config) > MaxConfigKeys {
		return fmt.Errorf("too many configuration keys: %d > %d", len(config), MaxConfigKeys)
	}

	// Validate each key-value pair
	for key, value := range config {
		// Validate key
		if err := validateKey(key); err != nil {
			return fmt.Errorf("invalid key %s: %w", key, err)
		}

		// Validate value
		if err := validateValue(key, value); err != nil {
			return fmt.Errorf("invalid value for key %s: %w", key, err)
		}
	}

	return nil
}

// validateKey validates a configuration key.
func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if len(key) > MaxKeyLength {
		return fmt.Errorf("key too long: %d > %d", len(key), MaxKeyLength)
	}

	// Check for invalid characters
	if strings.ContainsAny(key, " \t\n\r=") {
		return fmt.Errorf("key contains invalid characters")
	}

	return nil
}

// validateValue validates a configuration value.
func validateValue(_, value string) error {
	if len(value) > MaxValueLength {
		return fmt.Errorf("value too long: %d > %d", len(value), MaxValueLength)
	}

	// Additional validation rules can be added here
	// For example, check for specific patterns, required prefixes, etc.

	return nil
}

// CompositeValidator combines multiple validators.
type CompositeValidator struct {
	validators []Validator
}

// Validator defines the interface for configuration validation.
type Validator interface {
	Validate(ctx context.Context, config map[string]string) error
}

// NewCompositeValidator creates a new composite validator.
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{
		validators: validators,
	}
}

// Validate validates configuration using all configured validators.
func (v *CompositeValidator) Validate(ctx context.Context, config map[string]string) error {
	for _, validator := range v.validators {
		if err := validator.Validate(ctx, config); err != nil {
			return err
		}
	}

	return nil
}
