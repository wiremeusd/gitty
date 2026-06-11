package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPickAccountValidChoice(t *testing.T) {
	var out bytes.Buffer
	acc, err := pickAccount(strings.NewReader("2\n"), &out, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if acc != "b" {
		t.Fatalf("acc = %q", acc)
	}
}

func TestPickAccountRejectsOutOfRange(t *testing.T) {
	var out bytes.Buffer
	if _, err := pickAccount(strings.NewReader("9\n"), &out, []string{"a", "b"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestAskYesNo(t *testing.T) {
	cases := map[string]bool{"y\n": true, "Y\n": true, "yes\n": true, "n\n": false, "\n": false}
	for input, want := range cases {
		var out bytes.Buffer
		if got := askYesNo(strings.NewReader(input), &out, "?"); got != want {
			t.Fatalf("askYesNo(%q) = %v, want %v", input, got, want)
		}
	}
}
