package awk

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var fenceRE = regexp.MustCompile("(?s)```(?:awk)?\\s*\n(.*?)```")

// ExtractProgram strips markdown fences and surrounding whitespace from model output.
func ExtractProgram(raw string) string {
	s := strings.TrimSpace(raw)
	if m := fenceRE.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// Validate runs awk -f against an empty stdin to catch syntax errors.
func Validate(awkBin, programPath string) error {
	if awkBin == "" {
		awkBin = "awk"
	}
	cmd := exec.Command(awkBin, "-f", programPath)
	cmd.Stdin = strings.NewReader("")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// Run executes awk with optional -F and a program file, streaming stdin to stdout.
func Run(awkBin string, fieldSep string, programPath string, stdin io.Reader, stdout io.Writer) error {
	if awkBin == "" {
		awkBin = "awk"
	}
	args := make([]string, 0, 5)
	if fieldSep != "" {
		args = append(args, "-F", fieldSep)
	}
	args = append(args, "-f", programPath)
	cmd := exec.Command(awkBin, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WriteTemp writes program to a temp file and returns its path.
func WriteTemp(dir, program string) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	}
	f, err := os.CreateTemp(dir, "aiwk-*.awk")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err := f.WriteString(program); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	// Normalize path for clearer errors on some platforms.
	return filepath.Clean(path), nil
}
