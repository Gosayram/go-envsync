// Package main provides the CLI interface for go-envsync.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Gosayram/go-envsync/internal/version"
	"github.com/Gosayram/go-envsync/pkg/providers"
)

// Constants for CLI
const (
	// CLIName is the name of the CLI application.
	CLIName = "go-envsync"

	// CLIDescription is the short description of the CLI.
	CLIDescription = "Unified environment variable and secrets management"

	// ExitCodeSuccess indicates successful execution.
	ExitCodeSuccess = 0

	// ExitCodeError indicates an error during execution.
	ExitCodeError = 1
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   CLIName,
	Short: CLIDescription,
	Long: `go-envsync is a minimal yet extensible Go library and CLI tool for unified
environment variable and secrets management across multiple sources.

Supported providers:
- Local .env files (always available)
- Kubernetes Secrets and ConfigMaps (stub - requires k8s dependencies)
- HashiCorp Vault secrets (stub - requires Vault dependencies)
- AWS S3 (planned for future release)

Examples:
  go-envsync providers                    # List available providers
  go-envsync load --from=.env             # Load from local .env file
  go-envsync load --from=.env --validate=schema.json --export=json:config.json
  go-envsync version                      # Show version information`,
	PersistentPreRunE: initializeApplication,
}

var (
	showVersion bool
)

func init() {
	// Add version flag to root command
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
}

// initializeApplication performs application-wide initialization.
func initializeApplication(_ *cobra.Command, _ []string) error {
	// Initialize providers registry
	if err := providers.InitializeProviders(); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Handle version flag
	if showVersion {
		printVersion()
		os.Exit(ExitCodeSuccess)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeError)
	}
}

// printVersion prints version information
func printVersion() {
	fmt.Print(version.GetFullVersionInfo())
}

func main() {
	Execute()
}
