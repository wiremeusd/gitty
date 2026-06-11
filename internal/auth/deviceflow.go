// Package auth implements the GitHub OAuth Device Flow and token storage in Keychain.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultDeviceCodeURL = "https://github.com/login/device/code"
	defaultTokenURL      = "https://github.com/login/oauth/access_token"
)

type DeviceFlow struct {
	ClientID      string
	HTTP          *http.Client
	DeviceCodeURL string
	TokenURL      string
}

func NewDeviceFlow(clientID string) *DeviceFlow {
	return &DeviceFlow{
		ClientID:      clientID,
		HTTP:          http.DefaultClient,
		DeviceCodeURL: defaultDeviceCodeURL,
		TokenURL:      defaultTokenURL,
	}
}

type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (f *DeviceFlow) RequestCode(ctx context.Context) (*DeviceCode, error) {
	form := url.Values{"client_id": {f.ClientID}, "scope": {"repo"}}
	var dc DeviceCode
	if err := f.postForm(ctx, f.DeviceCodeURL, form, &dc); err != nil {
		return nil, err
	}
	if dc.DeviceCode == "" {
		return nil, errors.New("GitHub did not return a device_code — check your Client ID")
	}
	return &dc, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

// PollToken polls GitHub until the user confirms the code on the website.
func (f *DeviceFlow) PollToken(ctx context.Context, dc *DeviceCode) (string, error) {
	interval := time.Duration(dc.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	for {
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
		if time.Now().After(deadline) {
			return "", errors.New("code expired — run gitty auth login again")
		}
		form := url.Values{
			"client_id":   {f.ClientID},
			"device_code": {dc.DeviceCode},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		}
		var tr tokenResponse
		if err := f.postForm(ctx, f.TokenURL, form, &tr); err != nil {
			return "", err
		}
		switch tr.Error {
		case "":
			if tr.AccessToken == "" {
				return "", errors.New("GitHub returned an empty token")
			}
			return tr.AccessToken, nil
		case "authorization_pending":
			// user has not confirmed yet — keep waiting
		case "slow_down":
			interval += 5 * time.Second
		case "expired_token":
			return "", errors.New("code expired — run gitty auth login again")
		default:
			return "", fmt.Errorf("authorization failed: %s", tr.Error)
		}
	}
}

func (f *DeviceFlow) postForm(ctx context.Context, u string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub responded with %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
