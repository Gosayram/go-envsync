// Package providers initializes and registers all available providers.
package providers

import (
	"fmt"

	"github.com/Gosayram/go-envsync/pkg/client"
	"github.com/Gosayram/go-envsync/pkg/providers/kubernetes"
	"github.com/Gosayram/go-envsync/pkg/providers/local"
	"github.com/Gosayram/go-envsync/pkg/providers/registry"
	"github.com/Gosayram/go-envsync/pkg/providers/vault"
)

// Constants for provider initialization
const (
	// LocalProviderDescription describes the local file provider.
	LocalProviderDescription = "Load configuration from local .env files"

	// KubernetesProviderDescription describes the Kubernetes provider.
	KubernetesProviderDescription = "Load configuration from Kubernetes Secrets and ConfigMaps (requires k8s dependencies)"

	// VaultProviderDescription describes the Vault provider.
	VaultProviderDescription = "Load configuration from HashiCorp Vault secrets (requires Vault dependencies)"
)

// InitializeProviders registers all available providers in the global registry.
func InitializeProviders() error {
	// Initialize local provider
	if err := initializeLocalProvider(); err != nil {
		return fmt.Errorf("failed to initialize local provider: %w", err)
	}

	// Initialize Kubernetes provider
	if err := initializeKubernetesProvider(); err != nil {
		return fmt.Errorf("failed to initialize kubernetes provider: %w", err)
	}

	// Initialize Vault provider
	if err := initializeVaultProvider(); err != nil {
		return fmt.Errorf("failed to initialize vault provider: %w", err)
	}

	return nil
}

// initializeLocalProvider registers the local file system provider.
func initializeLocalProvider() error {
	localInfo := &registry.ProviderInfo{
		Name:        "local",
		Description: "Load configuration from local files (.env, JSON, YAML)",
		Aliases:     []string{"file", "fs", "filesystem"},
		Priority:    registry.HighPriority,
		Factory: func(config map[string]interface{}) (client.Provider, error) {
			basePath := "."
			if path, exists := config["base_path"]; exists {
				if pathStr, ok := path.(string); ok {
					basePath = pathStr
				}
			}
			return local.NewProviderWithBase(basePath), nil
		},
		SupportedSources: []string{
			".env",
			"path/to/.env",
			"config.json",
			"config.yaml",
		},
		OptionalConfig: []string{"base_path"},
	}

	return registry.Register(localInfo)
}

// initializeKubernetesProvider registers the Kubernetes provider.
func initializeKubernetesProvider() error {
	k8sInfo := &registry.ProviderInfo{
		Name:        "kubernetes",
		Description: "Load configuration from Kubernetes Secrets and ConfigMaps",
		Aliases:     []string{"k8s", "kube"},
		Priority:    registry.DefaultProviderPriority,
		Factory: func(config map[string]interface{}) (client.Provider, error) {
			var kubeconfig, namespace string

			if kc, exists := config["kubeconfig"]; exists {
				if kcStr, ok := kc.(string); ok {
					kubeconfig = kcStr
				}
			}

			if ns, exists := config["namespace"]; exists {
				if nsStr, ok := ns.(string); ok {
					namespace = nsStr
				}
			}

			return kubernetes.NewProviderWithConfig(kubeconfig, namespace)
		},
		SupportedSources: []string{
			"namespace/secret/secret-name",
			"namespace/configmap/config-name",
			"default/secret/app-secrets",
		},
		OptionalConfig: []string{"kubeconfig", "context", "namespace"},
	}

	return registry.Register(k8sInfo)
}

// initializeVaultProvider registers the HashiCorp Vault provider.
func initializeVaultProvider() error {
	vaultInfo := &registry.ProviderInfo{
		Name:        "vault",
		Description: "Load secrets from HashiCorp Vault",
		Aliases:     []string{"hcvault", "hashicorp-vault"},
		Priority:    registry.DefaultProviderPriority,
		Factory: func(config map[string]interface{}) (client.Provider, error) {
			var addr, token, mountPath string

			if a, exists := config["address"]; exists {
				if aStr, ok := a.(string); ok {
					addr = aStr
				}
			}

			if t, exists := config["token"]; exists {
				if tStr, ok := t.(string); ok {
					token = tStr
				}
			}

			if mp, exists := config["mount_path"]; exists {
				if mpStr, ok := mp.(string); ok {
					mountPath = mpStr
				}
			}

			return vault.NewProviderWithConfig(addr, token, mountPath)
		},
		SupportedSources: []string{
			"secret/data/app-config",
			"kv/production/database",
			"auth/token/secrets",
		},
		RequiredConfig: []string{"token"},
		OptionalConfig: []string{"address", "mount_path", "version"},
	}

	return registry.Register(vaultInfo)
}

// GetAvailableProviders returns information about all available providers.
func GetAvailableProviders() []*registry.ProviderInfo {
	return registry.ListProviders()
}

// IsProviderAvailable checks if a provider is available and registered.
func IsProviderAvailable(name string) bool {
	return registry.IsProviderRegistered(name)
}

// CreateProviderInstance creates a new provider instance with the given configuration.
func CreateProviderInstance(name string, config map[string]interface{}) (client.Provider, error) {
	return registry.CreateProvider(name, config)
}
