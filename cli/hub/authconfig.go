package hub

import (
	cliconfig "github.com/docker/cli/cli/config"
	clicreds "github.com/docker/cli/cli/config/credentials"
	clitypes "github.com/docker/cli/cli/config/types"
)

// AuthConfig retrieves auth config from the daemon
func AuthConfig() (*clitypes.AuthConfig, error) {
	config, err := cliconfig.Load(cliconfig.Dir())
	if err != nil {
		return nil, err
	}
	if !config.ContainsAuth() {
		config.CredentialsStore = clicreds.DetectDefaultStore(config.CredentialsStore)
	}
	authconfig, err := config.GetAuthConfig(registryURL)
	if err != nil {
		return nil, err
	}
	return &authconfig, nil
}
