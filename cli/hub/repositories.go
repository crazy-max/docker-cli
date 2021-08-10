package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	// repositoriesURL path to the Hub API listing the repositories
	repositoriesURL = "/v2/repositories/%s/"
)

// Repository represents a Docker Hub repository
type Repository struct {
	Name        string
	Description string
	LastUpdated time.Time
	PullCount   int
	StarCount   int
	IsPrivate   bool
}

type hubRepositoryResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next,omitempty"`
	Previous string                `json:"previous,omitempty"`
	Results  []hubRepositoryResult `json:"results,omitempty"`
}

type hubRepositoryResult struct {
	Name           string    `json:"name"`
	Namespace      string    `json:"namespace"`
	PullCount      int       `json:"pull_count"`
	StarCount      int       `json:"star_count"`
	RepositoryType string    `json:"repository_type"`
	CanEdit        bool      `json:"can_edit"`
	Description    string    `json:"description,omitempty"`
	IsAutomated    bool      `json:"is_automated"`
	IsMigrated     bool      `json:"is_migrated"`
	IsPrivate      bool      `json:"is_private"`
	LastUpdated    time.Time `json:"last_updated"`
	Status         int       `json:"status"`
	User           string    `json:"user"`
}

// GetRepositories lists all the repositories a user can access
func (c *client) GetRepositories(account string) ([]Repository, int, error) {
	if len(account) == 0 {
		account = c.account
	}
	u, err := url.Parse(c.domain + fmt.Sprintf(repositoriesURL, account))
	if err != nil {
		return nil, 0, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	q.Add("ordering", "last_updated")
	u.RawQuery = q.Encode()

	repos, total, next, err := c.getRepositoriesPage(u.String(), account)
	if err != nil {
		return nil, 0, err
	}

	for len(next) > 0 {
		pageRepos, _, n, err := c.getRepositoriesPage(next, account)
		if err != nil {
			return nil, 0, err
		}
		next = n
		repos = append(repos, pageRepos...)
	}

	return repos, total, nil
}

func (c *client) getRepositoriesPage(url, account string) ([]Repository, int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, "", err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, 0, "", err
	}
	var hubResponse hubRepositoryResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, 0, "", err
	}
	var repos []Repository
	for _, result := range hubResponse.Results {
		repo := Repository{
			Name:        fmt.Sprintf("%s/%s", account, result.Name),
			Description: result.Description,
			LastUpdated: result.LastUpdated,
			PullCount:   result.PullCount,
			StarCount:   result.StarCount,
			IsPrivate:   result.IsPrivate,
		}
		repos = append(repos, repo)
	}
	return repos, hubResponse.Count, hubResponse.Next, nil
}
