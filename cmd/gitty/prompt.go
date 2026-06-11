package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func pickAccount(in io.Reader, out io.Writer, accounts []string) (string, error) {
	fmt.Fprintln(out, "This folder is not bound to an account. Pick one:")
	for i, a := range accounts {
		fmt.Fprintf(out, "  %d) %s\n", i+1, a)
	}
	fmt.Fprint(out, "Number: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(accounts) {
		return "", errors.New("invalid selection")
	}
	return accounts[n-1], nil
}

func askYesNo(in io.Reader, out io.Writer, prompt string) bool {
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	line, _ := bufio.NewReader(in).ReadString('\n')
	s := strings.ToLower(strings.TrimSpace(line))
	return s == "y" || s == "yes"
}
