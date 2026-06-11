// Package github is a minimal GitHub REST API client: paginated repository listing
// and token owner login lookup.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"
)

var ErrUnauthorized = errors.New("GitHub rejected the token (401)")

type Repo struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Private     bool   `json:"private"`
	CloneURL    string `json:"clone_url"`
}

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func NewClient(token string) *Client {
	return &Client{BaseURL: "https://api.github.com", Token: token, HTTP: http.DefaultClient}
}

// Repos returns all repositories accessible to the account: owned, collaborator,
// and organization repositories. Follows all pages via the Link header.
func (c *Client) Repos(ctx context.Context) ([]Repo, error) {
	pageURL := c.BaseURL + "/user/repos?affiliation=owner,collaborator,organization_member&per_page=100&sort=updated"
	var all []Repo
	for pageURL != "" {
		var page []Repo
		next, err := c.get(ctx, pageURL, &page)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		pageURL = ""
		if next != "" && sameHost(c.BaseURL, next) {
			pageURL = next
		}
	}
	return all, nil
}

func (c *Client) Login(ctx context.Context) (string, error) {
	var u struct {
		Login string `json:"login"`
	}
	if _, err := c.get(ctx, c.BaseURL+"/user", &u); err != nil {
		return "", err
	}
	if u.Login == "" {
		return "", errors.New("GitHub returned an empty login")
	}
	return u.Login, nil
}

func (c *Client) get(ctx context.Context, url string, out any) (next string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gitty")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub responded with %s for %s", resp.Status, url)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return "", err
	}
	return nextLink(resp.Header.Get("Link")), nil
}

// nextLink extracts the URL with rel="next" from a GitHub Link header.
// rel may not be the first parameter, so all sections are checked.
func nextLink(header string) string {
	for _, part := range strings.Split(header, ",") {
		sections := strings.Split(part, ";")
		if len(sections) < 2 {
			continue
		}
		for _, s := range sections[1:] {
			if strings.TrimSpace(s) == `rel="next"` {
				return strings.Trim(strings.TrimSpace(sections[0]), "<>")
			}
		}
	}
	return ""
}

// sameHost verifies that a pagination URL points to the same host as
// BaseURL — so the token cannot be sent to a foreign server via a forged Link.
func sameHost(base, next string) bool {
	b, err1 := neturl.Parse(base)
	n, err2 := neturl.Parse(next)
	return err1 == nil && err2 == nil && b.Host == n.Host
}
