package awk

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExtractProgram(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"{ print $1 }", "{ print $1 }"},
		{"```awk\n{ print $1 }\n```", "{ print $1 }"},
		{"```\n{ print $1 }\n```", "{ print $1 }"},
		{"  \n```awk\nBEGIN { x=1 }\n```\n", "BEGIN { x=1 }"},
	}
	for _, tc := range cases {
		got := ExtractProgram(tc.in)
		if got != tc.want {
			t.Fatalf("ExtractProgram(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidateAndRun(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	prog := "{ print $2 }"
	path, err := WriteTemp(t.TempDir(), prog)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(path) }()
	if err := Validate("awk", path); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	var out bytes.Buffer
	in := strings.NewReader("a b\nc d\n")
	if err := Run("awk", "", path, in, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := out.String(); got != "b\nd\n" {
		t.Fatalf("output = %q; want b\\nd\\n", got)
	}
}

func TestRunCSV_QuotedComma(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	prog := "{ print $2 }"
	path, err := WriteTemp(t.TempDir(), prog)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(path) }()
	var out bytes.Buffer
	in := strings.NewReader("name,city\njohn,\"new york, ny\"\n")
	if err := RunCSV("awk", ",", path, in, &out); err != nil {
		t.Fatalf("RunCSV: %v", err)
	}
	if got := out.String(); got != "city\nnew york, ny\n" {
		t.Fatalf("output = %q; want city\\nnew york, ny\\n", got)
	}
}
