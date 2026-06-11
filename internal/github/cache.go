package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Cache is an on-disk repository list cache, one JSON file per account.
// It allows gitty to display the list instantly while fetching fresh data
// in the background.
type Cache struct {
	Dir string
}

// DefaultCache places the cache in os.UserCacheDir()/gitty
// (on macOS this is ~/Library/Caches/gitty).
func DefaultCache() (*Cache, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &Cache{Dir: filepath.Join(base, "gitty")}, nil
}

func (c *Cache) file(account string) string {
	return filepath.Join(c.Dir, account+".json")
}

// validAccount rejects names containing path separator characters — protection
// against path traversal via a manually edited config.
func validAccount(account string) error {
	if account == "" {
		return errors.New("empty account name")
	}
	for _, r := range account {
		ok := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-'
		if !ok {
			return fmt.Errorf("invalid account name: %q", account)
		}
	}
	return nil
}

// Load returns (nil, nil) if the cache is missing or corrupt.
func (c *Cache) Load(account string) ([]Repo, error) {
	if err := validAccount(account); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(c.file(account))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		// corrupt cache — treat as a miss, not an error
		return nil, nil
	}
	return repos, nil
}

func (c *Cache) Save(account string, repos []Repo) error {
	if err := validAccount(account); err != nil {
		return err
	}
	if err := os.MkdirAll(c.Dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(repos)
	if err != nil {
		return err
	}
	return os.WriteFile(c.file(account), data, 0o600)
}
