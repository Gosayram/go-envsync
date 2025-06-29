// Package exporter provides configuration export functionality in multiple formats.
package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Constants for export formats and limits
const (
	// FormatEnv represents .env file format.
	FormatEnv = "env"

	// FormatJSON represents JSON file format.
	FormatJSON = "json"

	// FormatYAML represents YAML file format.
	FormatYAML = "yaml"

	// MaxFileSize defines the maximum export file size in bytes.
	MaxFileSize = 10 * 1024 * 1024 // 10MB

	// DefaultFilePermissions defines the default file permissions for exported files.
	DefaultFilePermissions = 0o644

	// DefaultDirPermissions defines the default directory permissions.
	DefaultDirPermissions = 0o750

	// JSONIndentSpaces defines the number of spaces for JSON indentation.
	JSONIndentSpaces = 2

	// FormatPathParts defines the expected number of parts in format:path.
	FormatPathParts = 2
)

// MultiFormatExporter implements export functionality for multiple formats.
type MultiFormatExporter struct {
	outputDir string
}

// NewMultiFormatExporter creates a new multi-format exporter.
func NewMultiFormatExporter(outputDir string) *MultiFormatExporter {
	if outputDir == "" {
		outputDir = "."
	}

	return &MultiFormatExporter{
		outputDir: outputDir,
	}
}

// Export exports configuration to the specified format and destination.
func (e *MultiFormatExporter) Export(_ context.Context, config map[string]string, destination string) error {
	// Parse destination format and path
	format, filePath, err := e.parseDestination(destination)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := e.ensureOutputDir(filePath); err != nil {
		return err
	}

	// Export based on format
	switch format {
	case FormatEnv:
		return e.exportEnv(config, filePath)
	case FormatJSON:
		return e.exportJSON(config, filePath)
	case FormatYAML:
		return e.exportYAML(config, filePath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// parseDestination parses the destination string to extract format and file path.
func (e *MultiFormatExporter) parseDestination(destination string) (format, filePath string, err error) {
	parts := strings.SplitN(destination, ":", FormatPathParts)
	if len(parts) != FormatPathParts {
		return "", "", fmt.Errorf("invalid destination format, expected 'format:path', got: %s", destination)
	}

	format = strings.ToLower(parts[0])
	filePath = parts[1]

	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(e.outputDir, filePath)
	}

	return format, filePath, nil
}

// ensureOutputDir ensures the output directory exists.
func (e *MultiFormatExporter) ensureOutputDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, DefaultDirPermissions)
}

// exportEnv exports configuration to .env format.
func (e *MultiFormatExporter) exportEnv(config map[string]string, filePath string) error {
	var content strings.Builder

	// Add header comment
	content.WriteString("# Environment configuration exported by go-envsync\n")
	content.WriteString("# Generated automatically - do not edit manually\n\n")

	// Write key-value pairs
	for key, value := range config {
		// Escape value if necessary
		escapedValue := e.escapeEnvValue(value)
		content.WriteString(fmt.Sprintf("%s=%s\n", key, escapedValue))
	}

	return e.writeFile(filePath, content.String())
}

// exportJSON exports configuration to JSON format.
func (e *MultiFormatExporter) exportJSON(config map[string]string, filePath string) error {
	// Create output structure
	output := struct {
		Metadata map[string]string `json:"metadata"`
		Config   map[string]string `json:"config"`
	}{
		Metadata: map[string]string{
			"exported_by": "go-envsync",
			"format":      "json",
		},
		Config: config,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(output, "", strings.Repeat(" ", JSONIndentSpaces))
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return e.writeFile(filePath, string(data))
}

// exportYAML exports configuration to YAML format.
func (e *MultiFormatExporter) exportYAML(config map[string]string, filePath string) error {
	// Create output structure
	output := struct {
		Metadata map[string]string `yaml:"metadata"`
		Config   map[string]string `yaml:"config"`
	}{
		Metadata: map[string]string{
			"exported_by": "go-envsync",
			"format":      "yaml",
		},
		Config: config,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return e.writeFile(filePath, string(data))
}

// escapeEnvValue escapes a value for .env format.
func (e *MultiFormatExporter) escapeEnvValue(value string) string {
	// If value contains spaces or special characters, quote it
	if strings.ContainsAny(value, " \t\n\r\"'\\") {
		// Escape quotes and backslashes
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("%q", escaped)
	}

	return value
}

// writeFile writes content to a file with size validation.
func (e *MultiFormatExporter) writeFile(filePath, content string) error {
	// Check file size
	if len(content) > MaxFileSize {
		return fmt.Errorf("export content too large: %d bytes > %d bytes", len(content), MaxFileSize)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// GetSupportedFormats returns a list of supported export formats.
func GetSupportedFormats() []string {
	return []string{FormatEnv, FormatJSON, FormatYAML}
}
