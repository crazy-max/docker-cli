package hub

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	// userURL path to user informations
	userURL = "/v2/user/"
)

//Account represents a user or organization information
type Account struct {
	ID       string
	Name     string
	FullName string
	Location string
	Company  string
	Joined   time.Time
}

type hubUserResponse struct {
	ID            string    `json:"id"`
	UserName      string    `json:"username"`
	FullName      string    `json:"full_name"`
	Location      string    `json:"location"`
	Company       string    `json:"company"`
	GravatarEmail string    `json:"gravatar_email"`
	GravatarURL   string    `json:"gravatar_url"`
	IsStaff       bool      `json:"is_staff"`
	IsAdmin       bool      `json:"is_admin"`
	ProfileURL    string    `json:"profile_url"`
	DateJoined    time.Time `json:"date_joined"`
	Type          string    `json:"type"`
}

// GetUserInfo returns the information on the user retrieved from Hub
func (c *client) GetUserInfo() (*Account, error) {
	u, err := url.Parse(c.domain + userURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var hubResponse hubUserResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, err
	}
	return &Account{
		ID:       hubResponse.ID,
		Name:     hubResponse.UserName,
		FullName: hubResponse.FullName,
		Location: hubResponse.Location,
		Company:  hubResponse.Company,
		Joined:   hubResponse.DateJoined,
	}, nil
}
