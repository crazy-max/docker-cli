package hub

import (
	"os"

	"github.com/docker/docker/api/types/registry"
)

// Instance stores all the specific pieces needed to dialog with Hub
type Instance struct {
	APIHubBaseURL string
	RegistryInfo  *registry.IndexInfo
}

var (
	hub = Instance{
		APIHubBaseURL: "https://hub.docker.com",
		RegistryInfo: &registry.IndexInfo{
			Name:     "registry-1.docker.io",
			Mirrors:  nil,
			Secure:   true,
			Official: true,
		},
	}
)

// getInstance returns the current hub instance, which can be overridden by
// DOCKER_REGISTRY_URL and DOCKER_REGISTRY_URL env var
func getInstance() *Instance {
	apiBaseURL := os.Getenv("DOCKER_HUB_API_URL")
	reg := os.Getenv("DOCKER_REGISTRY_URL")

	if apiBaseURL != "" && reg != "" {
		return &Instance{
			APIHubBaseURL: apiBaseURL,
			RegistryInfo: &registry.IndexInfo{
				Name:     reg,
				Mirrors:  nil,
				Secure:   true,
				Official: false,
			},
		}
	}

	return &hub
}
