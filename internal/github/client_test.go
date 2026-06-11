package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReposIgnoresForeignNextLink(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Link", `<https://evil.example.com/steal?page=2>; rel="next"`)
		w.Write([]byte(`[{"full_name":"me/one","name":"one"}]`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "tok", HTTP: srv.Client()}
	repos, err := c.Repos(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || calls != 1 {
		t.Fatalf("repos=%d calls=%d, want 1/1 (foreign link must not be followed)", len(repos), calls)
	}
}

func TestReposPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("auth header = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			w.Write([]byte(`[{"full_name":"me/two","name":"two","stargazers_count":1,"language":"Go","clone_url":"https://github.com/me/two.git"}]`))
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s/user/repos?page=2>; rel="next", <%s/user/repos?page=2>; rel="last"`, srv.URL, srv.URL))
		w.Write([]byte(`[{"full_name":"me/one","name":"one","stargazers_count":5,"language":"Go","clone_url":"https://github.com/me/one.git"}]`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "tok", HTTP: srv.Client()}
	repos, err := c.Repos(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 || repos[0].FullName != "me/one" || repos[1].FullName != "me/two" {
		t.Fatalf("repos = %+v", repos)
	}
}

func TestReposUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "bad", HTTP: srv.Client()}
	if _, err := c.Repos(context.Background()); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"login":"wiremeusd"}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "tok", HTTP: srv.Client()}
	login, err := c.Login(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if login != "wiremeusd" {
		t.Fatalf("login = %q", login)
	}
}

func TestNextLink(t *testing.T) {
	h := `<https://api.github.com/user/repos?page=2>; rel="next", <https://api.github.com/user/repos?page=9>; rel="last"`
	if got := nextLink(h); got != "https://api.github.com/user/repos?page=2" {
		t.Fatalf("nextLink = %q", got)
	}
	if got := nextLink(`<https://x>; rel="last"`); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := nextLink(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	h2 := `<https://api.github.com/user/repos?page=2>; type="application/json"; rel="next"`
	if got := nextLink(h2); got != "https://api.github.com/user/repos?page=2" {
		t.Fatalf("nextLink with rel not first = %q", got)
	}
}
