package prompt

import (
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	t.Parallel()
	sys, usr := Build("print column two", ",", false, false, "")
	if !strings.Contains(usr, "print column two") {
		t.Fatalf("user prompt missing query: %q", usr)
	}
	if !strings.Contains(usr, `","`) {
		t.Fatalf("user prompt missing field sep: %q", usr)
	}
	if sys == "" {
		t.Fatal("empty system prompt")
	}
	sys2, usr2 := Build("x", "", false, true, "previous error")
	if sys2 == sys {
		t.Fatal("explain should change system prompt")
	}
	if !strings.Contains(usr2, "previous error") {
		t.Fatalf("correction not forwarded: %q", usr2)
	}

	_, usr3 := Build("print first field", "", true, false, "")
	if !strings.Contains(usr3, "CSV-aware parser enabled") {
		t.Fatalf("csv mode not forwarded: %q", usr3)
	}
}
