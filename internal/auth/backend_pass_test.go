package auth

import (
	"errors"
	"testing"
)

func TestPassSetInvokesInsert(t *testing.T) {
	var gotArgs []string
	var gotStdin string
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		gotArgs, gotStdin = args, stdin
		return "", nil
	}}
	if err := p.set("work", "gho_abc"); err != nil {
		t.Fatal(err)
	}
	want := []string{"insert", "-e", "-f", "gitty/work"}
	if len(gotArgs) != len(want) {
		t.Fatalf("args = %v, want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("args = %v, want %v", gotArgs, want)
		}
	}
	if gotStdin != "gho_abc\n" {
		t.Fatalf("stdin = %q, want %q", gotStdin, "gho_abc\n")
	}
}

func TestPassGetReturnsFirstLine(t *testing.T) {
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		return "gho_abc\nhttps://github.com/work\n", nil
	}}
	got, err := p.get("work")
	if err != nil {
		t.Fatal(err)
	}
	if got != "gho_abc" {
		t.Fatalf("token = %q", got)
	}
}

func TestPassGetMissingMapsToErrNotFound(t *testing.T) {
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		return "", errors.New("pass show: Error: gitty/work is not in the password store.: exit status 1")
	}}
	if _, err := p.get("work"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPassGetOtherErrorPropagates(t *testing.T) {
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		return "", errors.New("pass show: gpg: decryption failed: No secret key: exit status 2")
	}}
	_, err := p.get("work")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("expected a non-not-found error, got %v", err)
	}
}

func TestPassDeleteInvokesRm(t *testing.T) {
	var gotArgs []string
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		gotArgs = args
		return "", nil
	}}
	if err := p.delete("work"); err != nil {
		t.Fatal(err)
	}
	want := []string{"rm", "-f", "gitty/work"}
	if len(gotArgs) != len(want) {
		t.Fatalf("args = %v, want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("args = %v, want %v", gotArgs, want)
		}
	}
}

func TestPassDeleteMissingIsNotAnError(t *testing.T) {
	p := &passStore{run: func(args []string, stdin string) (string, error) {
		return "", errors.New("pass rm: Error: gitty/work is not in the password store.: exit status 1")
	}}
	if err := p.delete("work"); err != nil {
		t.Fatalf("deleting a missing entry must be a no-op, got %v", err)
	}
}
