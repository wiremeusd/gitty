// Package config stores the account list and "directory → account" bindings
// in ~/.config/gitty/config.toml. Tokens are not stored here (see internal/auth).
package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Accounts []string          `toml:"accounts"`
	Bindings map[string]string `toml:"bindings"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gitty", "config.toml"), nil
}

// Load reads the config; a missing file is treated as an empty config, not an error.
func Load(path string) (*Config, error) {
	cfg := &Config{Bindings: map[string]string{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Bindings == nil {
		cfg.Bindings = map[string]string{}
	}
	cleaned := make(map[string]string, len(cfg.Bindings))
	for k, v := range cfg.Bindings {
		cleaned[filepath.Clean(k)] = v
	}
	cfg.Bindings = cleaned
	return cfg, nil
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return err
	}
	// Write to a temporary file and rename so that an interrupted write
	// does not leave a corrupt config.
	tmp, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (c *Config) AddAccount(name string) {
	for _, a := range c.Accounts {
		if a == name {
			return
		}
	}
	c.Accounts = append(c.Accounts, name)
}

func (c *Config) RemoveAccount(name string) {
	out := c.Accounts[:0]
	for _, a := range c.Accounts {
		if a != name {
			out = append(out, a)
		}
	}
	c.Accounts = out
	for dir, acc := range c.Bindings {
		if acc == name {
			delete(c.Bindings, dir)
		}
	}
}

func (c *Config) Bind(dir, account string) {
	c.Bindings[filepath.Clean(dir)] = account
}

// AccountFor walks up the directory tree from dir and returns the first
// matching binding (i.e. the deepest one). If no binding is found and
// there is exactly one account, it is returned.
// dir must be an absolute path (e.g. from os.Getwd()).
func (c *Config) AccountFor(dir string) (string, bool) {
	d := filepath.Clean(dir)
	for {
		if acc, ok := c.Bindings[d]; ok {
			return acc, true
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	if len(c.Accounts) == 1 {
		return c.Accounts[0], true
	}
	return "", false
}
