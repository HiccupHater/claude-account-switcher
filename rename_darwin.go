package main

// Fork: the box that asks for an account's new name.

import (
	"os/exec"
	"strings"
	"sync"
)

// The prompt and current name come in as arguments, so they need no
// AppleScript quoting.
const nameScript = `on run argv
	set r to display dialog (item 2 of argv) with title (item 1 of argv) default answer (item 3 of argv) buttons {"Cancel", "Rename"} default button "Rename" cancel button "Cancel"
	return text returned of r
end run`

var nameMu sync.Mutex

// askName shows a box with prompt and an edit field holding current, and
// returns what was typed and whether Rename was pressed.
func askName(title, prompt, current string) (string, bool) {
	if !nameMu.TryLock() {
		return "", false
	}
	defer nameMu.Unlock()
	out, err := exec.Command("osascript", "-e", nameScript, title, prompt, current).Output()
	if err != nil {
		// Cancel makes osascript fail with "User canceled".
		return "", false
	}
	return strings.TrimRight(string(out), "\r\n"), true
}
