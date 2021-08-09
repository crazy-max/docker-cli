package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	cliconfig "github.com/docker/cli/cli/config"
	clicreds "github.com/docker/cli/cli/config/credentials"
	clitypes "github.com/docker/cli/cli/config/types"
)

const (
	// registryURL is the regitry Hub URL
	registryURL = "https://index.docker.io/v1/"
	// loginURL path to the Hub login URL
	loginURL = "/v2/users/login"
	// twoFactorLoginURL path to the 2FA
	//twoFactorLoginURL = "/v2/users/2fa-login?refresh_token=true"
	// secondFactorDetailMessage returned by login if 2FA is enabled
	secondFactorDetailMessage = "Require secondary authentication on MFA enabled account"
)

// Client Docker Hub client
type Client struct {
	domain   string
	username string
	token    string
}

// NewClient get a client to request Docker Hub
func NewClient() (*Client, error) {
	hubInstance := getInstance()
	return &Client{
		domain: hubInstance.APIHubBaseURL,
	}, nil
}

// Login tries to authenticate
func (c *Client) Login() error {
	// Retrieve auth config
	authconfig, err := authConfig()
	if err != nil {
		return err
	}
	authdata, err := json.Marshal(authconfig)
	if err != nil {
		return err
	}
	authbody := bytes.NewBuffer(authdata)

	// Login on the Docker Hub
	req, err := http.NewRequest("POST", c.domain+loginURL, ioutil.NopCloser(authbody))
	if err != nil {
		return err
	}
	resp, err := c.doRawRequest(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Retrieve token
	if resp.StatusCode == http.StatusOK {
		creds := struct {
			Token string `json:"token"`
		}{}
		if err = json.Unmarshal(buf, &creds); err != nil {
			return err
		}
		c.username = authconfig.Username
		c.token = creds.Token
		return nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		response2FA := struct {
			Detail        string `json:"detail"`
			Login2FAToken string `json:"login_2fa_token"`
		}{}
		if err = json.Unmarshal(buf, &response2FA); err != nil {
			return err
		}

		// Check if 2FA is enabled and needs a second authentication
		if response2FA.Detail != secondFactorDetailMessage {
			return fmt.Errorf(response2FA.Detail)
		}

		// TODO: Retrieve two factor token
		c.username = authconfig.Username
		//c.token =
		return nil
	}

	if ok, err := extractError(buf, resp); ok {
		return err
	}

	return fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, string(buf))
}

func (c *Client) doRawRequest(req *http.Request) (*http.Response, error) {
	req.Header["Accept"] = []string{"application/json"}
	req.Header["Content-Type"] = []string{"application/json"}
	return http.DefaultClient.Do(req)
}

func extractError(buf []byte, resp *http.Response) (bool, error) {
	var responseBody map[string]string
	if err := json.Unmarshal(buf, &responseBody); err == nil {
		for _, k := range []string{"message", "detail"} {
			if msg, ok := responseBody[k]; ok {
				return true, fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, msg)
			}
		}
	}
	return false, nil
}

// authConfig retrieves auth config from the daemon
func authConfig() (*clitypes.AuthConfig, error) {
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
