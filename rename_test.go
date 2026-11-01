package main

import (
	"strings"
	"testing"
)

func TestLabelPrefersNickname(t *testing.T) {
	a := Account{Name: "Nathan", Email: "n@example.com", Plan: "Max"}
	if got := a.label(); got != "Nathan (Max)" {
		t.Errorf("label = %q, want Claude's name", got)
	}
	a.Nickname = "Work"
	if got := a.label(); got != "Work (Max)" {
		t.Errorf("label = %q, want the nickname", got)
	}
	a.Name = ""
	if got := a.claudeName(); got != "n@example.com" {
		t.Errorf("claudeName = %q, want the email without a display name", got)
	}
}

func TestCleanNickname(t *testing.T) {
	for in, want := range map[string]string{
		"  Work  ":       "Work",
		"Alex\tmain\r\n": "Alex main",
		"":               "",
		"   ":            "",
	} {
		if got := cleanNickname(in); got != want {
			t.Errorf("cleanNickname(%q) = %q, want %q", in, got, want)
		}
	}
	long := strings.Repeat("é", maxNickname+10)
	if got := []rune(cleanNickname(long)); len(got) != maxNickname {
		t.Errorf("long name kept %d characters, want %d", len(got), maxNickname)
	}
}
