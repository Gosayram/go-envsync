// Package main contains CLI command implementations for go-envsync.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gosayram/go-envsync/pkg/client"
	"github.com/Gosayram/go-envsync/pkg/exporter"
	"github.com/Gosayram/go-envsync/pkg/providers/local"
	"github.com/Gosayram/go-envsync/pkg/validator"
)

// Constants for load command
const (
	// DefaultTimeout for load operations.
	DefaultTimeout = 30 * time.Second

	// DefaultExportFormat when no format is specified.
	DefaultExportFormat = "env"

	// DefaultMergeStrategy when multiple sources are provided.
	DefaultMergeStrategy = "override"

	// MaxSources defines the maximum number of sources that can be loaded.
	MaxSources = 10

	// SourceFormatParts defines the expected number of parts in source format.
	SourceFormatParts = 2
)

// LoadCommand flags
var (
	loadSources       []string
	loadSchema        string
	loadExport        string
	loadMergeStrategy string
	loadTimeout       time.Duration
	loadOutputDir     string
	loadDryRun        bool
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load configuration from specified sources",
	Long: `Load configuration from one or more sources with validation and export capabilities.

Supported sources:
- local:.env (or just .env) - Load from local .env file
- k8s:namespace/secret - Load from Kubernetes Secret (planned)
- vault:path/to/secret - Load from HashiCorp Vault (planned)
- s3:bucket/path - Load from AWS S3 (planned)

Examples:
  go-envsync load --from=.env --validate=./schema.json --export=json:config.json
  go-envsync load --from=.env --from=local:.env.local --export=yaml:config.yaml
  go-envsync load --from=.env --merge-strategy=preserve --dry-run`,
	RunE: runLoadCommand,
}

func init() {
	// Add load command to root
	rootCmd.AddCommand(loadCmd)

	// Define flags
	loadCmd.Flags().StringSliceVar(&loadSources, "from", []string{}, "Configuration sources to load from")
	loadCmd.Flags().StringVar(&loadSchema, "validate", "", "JSON schema file for validation")
	loadCmd.Flags().StringVar(&loadExport, "export", "", "Export format and destination (format:path)")
	loadCmd.Flags().StringVar(&loadMergeStrategy, "merge-strategy", DefaultMergeStrategy,
		"Merge strategy for multiple sources (override, preserve, error)")
	loadCmd.Flags().DurationVar(&loadTimeout, "timeout", DefaultTimeout, "Timeout for load operations")
	loadCmd.Flags().StringVar(&loadOutputDir, "output-dir", ".", "Output directory for exported files")
	loadCmd.Flags().BoolVar(&loadDryRun, "dry-run", false, "Perform a dry run without writing files")

	// Mark required flags
	if err := loadCmd.MarkFlagRequired("from"); err != nil {
		// This should never happen in practice, but we need to handle the error
		panic(fmt.Sprintf("failed to mark 'from' flag as required: %v", err))
	}
}

// runLoadCommand executes the load command.
func runLoadCommand(_ *cobra.Command, _ []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), loadTimeout)
	defer cancel()

	// Validate inputs
	if err := validateLoadInputs(); err != nil {
		return err
	}

	// Create client
	envClient := client.New()

	// Setup providers
	setupProviders(envClient)

	// Setup validator if schema is provided
	if loadSchema != "" {
		if err := setupValidator(envClient); err != nil {
			return fmt.Errorf("failed to setup validator: %w", err)
		}
	}

	// Setup exporter if export is requested
	if loadExport != "" {
		setupExporter(envClient)
	}

	// Parse merge strategy
	mergeStrategy, err := parseMergeStrategy(loadMergeStrategy)
	if err != nil {
		return err
	}

	// Load configuration
	fmt.Printf("Loading configuration from %d sources...\n", len(loadSources))

	loadOptions := client.LoadOptions{
		Sources:       loadSources,
		Schema:        loadSchema,
		MergeStrategy: mergeStrategy,
	}

	env, err := envClient.Load(ctx, loadOptions)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display loaded configuration summary
	fmt.Printf("Successfully loaded %d configuration keys\n", len(env.Data))

	// Export if requested
	if loadExport != "" && !loadDryRun {
		fmt.Printf("Exporting configuration to %s...\n", loadExport)

		if err := exportConfiguration(env, loadExport); err != nil {
			return fmt.Errorf("failed to export configuration: %w", err)
		}

		fmt.Println("Configuration exported successfully")
	}

	// Display dry run information
	if loadDryRun {
		fmt.Println("\nDry run completed - no files were written")
		fmt.Printf("Configuration keys: %v\n", env.Keys())
	}

	return nil
}

// validateLoadInputs validates the load command inputs.
func validateLoadInputs() error {
	// Check number of sources
	if len(loadSources) == 0 {
		return fmt.Errorf("at least one source must be specified")
	}

	if len(loadSources) > MaxSources {
		return fmt.Errorf("too many sources: %d > %d", len(loadSources), MaxSources)
	}

	// Validate merge strategy
	validStrategies := []string{"override", "preserve", "error"}
	valid := false
	for _, strategy := range validStrategies {
		if loadMergeStrategy == strategy {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid merge strategy: %s (valid: %v)", loadMergeStrategy, validStrategies)
	}

	// Validate schema file if provided
	if loadSchema != "" {
		if _, err := os.Stat(loadSchema); os.IsNotExist(err) {
			return fmt.Errorf("schema file not found: %s", loadSchema)
		}
	}

	return nil
}

// setupProviders configures the providers for the client.
func setupProviders(envClient *client.Client) {
	// Setup local provider
	localProvider := local.NewProviderWithBase(".")
	envClient.AddProvider("local", localProvider)

	// Also add as default provider
	envClient.AddProvider(client.DefaultProviderName, localProvider)

	// TODO: Add other providers (K8s, Vault, S3) in future phases
}

// setupValidator configures the validator for the client.
func setupValidator(envClient *client.Client) error {
	schemaValidator, err := validator.NewSchemaValidator(loadSchema)
	if err != nil {
		return err
	}

	envClient.SetValidator(schemaValidator)
	return nil
}

// setupExporter configures the exporter for the client.
func setupExporter(envClient *client.Client) {
	multiExporter := exporter.NewMultiFormatExporter(loadOutputDir)
	envClient.SetExporter(multiExporter)
}

// parseMergeStrategy converts string merge strategy to client enum.
func parseMergeStrategy(strategy string) (client.MergeStrategy, error) {
	switch strategy {
	case "override":
		return client.MergeStrategyOverride, nil
	case "preserve":
		return client.MergeStrategyPreserve, nil
	case "error":
		return client.MergeStrategyError, nil
	default:
		return client.MergeStrategyOverride, fmt.Errorf("unknown merge strategy: %s", strategy)
	}
}

// exportConfiguration exports the loaded configuration.
func exportConfiguration(env *client.Environment, exportSpec string) error {
	// Use the environment's built-in export methods based on format
	if exportSpec == "" {
		return fmt.Errorf("export specification cannot be empty")
	}

	// Parse export format
	parts := []string{exportSpec}
	if len(parts) > 0 && parts[0] != "" {
		// Let the client handle the export through its configured exporter
		return env.ExportEnv(exportSpec) // This will be handled by the exporter based on format
	}

	return fmt.Errorf("invalid export specification: %s", exportSpec)
}
