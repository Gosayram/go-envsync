// Package main contains CLI command implementations for go-envsync.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gosayram/go-envsync/pkg/providers/registry"
)

// Constants for providers command
const (
	// MinProviderNameLength defines the minimum length for provider name column.
	MinProviderNameLength = 12

	// MaxDescriptionLength defines the maximum length for description column.
	MaxDescriptionLength = 60

	// AliasesColumnLength defines the length for aliases column display.
	AliasesColumnLength = 15

	// MaxAliasesDisplay defines the maximum number of aliases to display.
	MaxAliasesDisplay = 15
)

// ProvidersCommand flags
var (
	providersShowDetails bool
	providersFilter      string
)

// providersCmd represents the providers command
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available configuration providers",
	Long: `List all available configuration providers with their information.

This command shows all registered providers that can be used to load configuration
from different sources. Each provider has a name, aliases, and supported source formats.

Examples:
  go-envsync providers                    # List all providers
  go-envsync providers --details          # Show detailed information
  go-envsync providers --filter=local     # Filter by provider name`,
	RunE: runProvidersCommand,
}

func init() {
	// Add providers command to root
	rootCmd.AddCommand(providersCmd)

	// Define flags
	providersCmd.Flags().BoolVar(&providersShowDetails, "details", false, "Show detailed provider information")
	providersCmd.Flags().StringVar(&providersFilter, "filter", "", "Filter providers by name or alias")
}

// runProvidersCommand executes the providers command.
func runProvidersCommand(_ *cobra.Command, _ []string) error {
	// Get all registered providers
	providerNames := registry.GetProviderNames()

	if len(providerNames) == 0 {
		fmt.Println("No providers registered")
		return nil
	}

	// Filter providers if requested
	if providersFilter != "" {
		providerNames = filterProviders(providerNames, providersFilter)
	}

	// Sort providers
	sort.Strings(providerNames)

	if providersShowDetails {
		return showDetailedProviders(providerNames)
	}

	return showProviderList(providerNames)
}

// filterProviders filters providers by name, alias, or description.
func filterProviders(allProviders []string, filter string) []string {
	var filtered []string
	filterLower := strings.ToLower(filter)

	for _, providerName := range allProviders {
		// Get provider info
		providerInfo, err := registry.GetProvider(providerName)
		if err != nil {
			continue // Skip if provider not found
		}

		// Check name
		if strings.Contains(strings.ToLower(providerInfo.Name), filterLower) {
			filtered = append(filtered, providerName)
			continue
		}

		// Check aliases
		for _, alias := range providerInfo.Aliases {
			if strings.Contains(strings.ToLower(alias), filterLower) {
				filtered = append(filtered, providerName)
				break
			}
		}

		// Check description
		if strings.Contains(strings.ToLower(providerInfo.Description), filterLower) {
			filtered = append(filtered, providerName)
		}
	}

	return filtered
}

// showProviderList displays a simple list of providers.
func showProviderList(providerNames []string) error {
	fmt.Printf("Available providers (%d):\n\n", len(providerNames))

	// Table header
	fmt.Printf("%-*s %-*s %s\n",
		MinProviderNameLength, "PROVIDER",
		AliasesColumnLength, "ALIASES",
		"DESCRIPTION")
	fmt.Printf("%s %s %s\n",
		strings.Repeat("-", MinProviderNameLength),
		strings.Repeat("-", AliasesColumnLength),
		strings.Repeat("-", MaxDescriptionLength))

	// Table rows
	for _, name := range providerNames {
		providerInfo, err := registry.GetProvider(name)
		if err != nil {
			continue // Skip if provider not found
		}

		aliasesStr := strings.Join(providerInfo.Aliases, ", ")
		if len(providerInfo.Aliases) > MaxAliasesDisplay {
			aliasesStr = strings.Join(providerInfo.Aliases[:MaxAliasesDisplay], ", ") + "..."
		}

		description := providerInfo.Description
		if len(description) > MaxDescriptionLength {
			description = description[:MaxDescriptionLength-3] + "..."
		}

		fmt.Printf("%-*s %-*s %s\n",
			MinProviderNameLength, name,
			AliasesColumnLength, aliasesStr,
			description)
	}

	return nil
}

// showDetailedProviders displays providers with detailed information.
func showDetailedProviders(providerNames []string) error {
	fmt.Printf("Available Providers (detailed view):\n\n")

	for i, name := range providerNames {
		if i > 0 {
			fmt.Println()
		}

		providerInfo, err := registry.GetProvider(name)
		if err != nil {
			fmt.Printf("Provider %s not found\n", name)
			continue
		}

		fmt.Printf("Provider: %s (priority: %d)\n", providerInfo.Name, providerInfo.Priority)

		if len(providerInfo.Aliases) > 0 {
			fmt.Printf("  Aliases: %s\n", strings.Join(providerInfo.Aliases, ", "))
		}

		fmt.Printf("  Description: %s\n", providerInfo.Description)

		if len(providerInfo.SupportedSources) > 0 {
			fmt.Printf("  Supported Sources:\n")
			for _, source := range providerInfo.SupportedSources {
				fmt.Printf("    - %s\n", source)
			}
		}

		if len(providerInfo.RequiredConfig) > 0 {
			fmt.Printf("  Required Configuration:\n")
			for _, config := range providerInfo.RequiredConfig {
				fmt.Printf("    - %s\n", config)
			}
		}

		if len(providerInfo.OptionalConfig) > 0 {
			fmt.Printf("  Optional Configuration:\n")
			for _, config := range providerInfo.OptionalConfig {
				fmt.Printf("    - %s\n", config)
			}
		}
	}

	fmt.Printf("\nTotal: %d providers\n", len(providerNames))
	return nil
}
