package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/cli/cli/command"
)

const (
	// registryURL is the regitry Hub URL
	registryURL = "https://index.docker.io/v1/"
	// loginURL path to the Hub login URL
	loginURL = "/v2/users/login"
	// itemsPerPage is the maximum of items per page
	itemsPerPage = 100
)

// Client Docker Hub client
type Client struct {
	domain  string
	account string
	token   string
}

// RequestOp represents an option to customize the request sent to the Hub API
type RequestOp func(r *http.Request) error

// NewClient get a client to request Docker Hub
func NewClient() (*Client, error) {
	return &Client{
		domain: getInstance().APIHubBaseURL,
	}, nil
}

func withHubToken(token string) RequestOp {
	return func(req *http.Request) error {
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}
		return nil
	}
}

// Login tries to authenticate
func (c *Client) Login() error {
	// Retrieve auth config
	authconfig, err := AuthConfig()
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
		c.account = authconfig.Username
		c.token = creds.Token
		return nil
	}

	if ok, err := extractError(buf, resp); ok {
		return err
	}

	return fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, string(buf))
}

func (c *Client) doRequest(req *http.Request, reqOps ...RequestOp) ([]byte, error) {
	resp, err := c.doRawRequest(req, reqOps...)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("operation not permitted")
		}
		buf, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			if ok, err := extractError(buf, resp); ok {
				return nil, err
			}
		}
		return nil, fmt.Errorf("bad status code %q", resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status code %q: %s", resp.Status, string(buf))
	}

	return buf, nil
}

func (c *Client) doRawRequest(req *http.Request, reqOps ...RequestOp) (*http.Response, error) {
	req.Header["Accept"] = []string{"application/json"}
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["User-Agent"] = []string{command.UserAgent()}
	for _, op := range reqOps {
		if err := op(req); err != nil {
			return nil, err
		}
	}
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
