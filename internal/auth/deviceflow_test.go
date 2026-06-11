package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestFlow(srv *httptest.Server) *DeviceFlow {
	f := NewDeviceFlow("test-client-id")
	f.DeviceCodeURL = srv.URL + "/device/code"
	f.TokenURL = srv.URL + "/token"
	return f
}

func TestRequestCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/code" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("client_id") != "test-client-id" || r.Form.Get("scope") != "repo" {
			t.Errorf("form = %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"device_code":"dc1","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","expires_in":900,"interval":5}`))
	}))
	defer srv.Close()

	dc, err := newTestFlow(srv).RequestCode(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dc.UserCode != "ABCD-1234" || dc.DeviceCode != "dc1" || dc.Interval != 5 {
		t.Fatalf("dc = %+v", dc)
	}
}

func TestRequestCodeEmptyResponseIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	if _, err := newTestFlow(srv).RequestCode(context.Background()); err == nil {
		t.Fatal("expected error for empty device_code")
	}
}

func TestPollTokenPendingThenSuccess(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.Write([]byte(`{"access_token":"gho_token123"}`))
	}))
	defer srv.Close()

	dc := &DeviceCode{DeviceCode: "dc1", ExpiresIn: 900, Interval: 0}
	token, err := newTestFlow(srv).PollToken(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if token != "gho_token123" {
		t.Fatalf("token = %q", token)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestPollTokenAccessDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":"access_denied"}`))
	}))
	defer srv.Close()

	dc := &DeviceCode{DeviceCode: "dc1", ExpiresIn: 900, Interval: 0}
	if _, err := newTestFlow(srv).PollToken(context.Background(), dc); err == nil {
		t.Fatal("expected error")
	}
}

func TestPollTokenExpiredTokenFromServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":"expired_token"}`))
	}))
	defer srv.Close()

	dc := &DeviceCode{DeviceCode: "dc1", ExpiresIn: 900, Interval: 0}
	_, err := newTestFlow(srv).PollToken(context.Background(), dc)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("err = %v, want friendly expiry message", err)
	}
}

func TestPollTokenSlowDownIncreasesIntervalAndContinues(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Write([]byte(`{"error":"slow_down"}`))
			return
		}
		w.Write([]byte(`{"access_token":"gho_after_slowdown"}`))
	}))
	defer srv.Close()

	dc := &DeviceCode{DeviceCode: "dc1", ExpiresIn: 900, Interval: 0}
	start := time.Now()
	token, err := newTestFlow(srv).PollToken(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if token != "gho_after_slowdown" {
		t.Fatalf("token = %q", token)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	// slow_down adds 5s to the interval — the second poll must not arrive earlier
	if elapsed := time.Since(start); elapsed < 5*time.Second {
		t.Fatalf("second poll came after %v, want >= 5s", elapsed)
	}
}
