package awk

import (
	"bufio"
	"bytes"
	"encoding/csv"
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

// RunCSV parses stdin as CSV and executes awk over normalized rows.
// It preserves quoted fields (including commas) by parsing first, then joining
// fields using an internal separator that awk receives via -F.
func RunCSV(awkBin string, csvDelimiter string, programPath string, stdin io.Reader, stdout io.Writer) error {
	if awkBin == "" {
		awkBin = "awk"
	}
	comma, err := parseCSVDelimiter(csvDelimiter)
	if err != nil {
		return err
	}
	const internalFS = "\x1f"
	cmd := exec.Command(awkBin, "-F", internalFS, "-f", programPath)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	awkIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	writeErr := streamCSVRows(stdin, awkIn, comma, internalFS)
	closeErr := awkIn.Close()
	waitErr := cmd.Wait()

	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return waitErr
}

func parseCSVDelimiter(s string) (rune, error) {
	if s == "" {
		return ',', nil
	}
	r := []rune(s)
	if len(r) != 1 {
		return 0, fmt.Errorf("csv mode requires a single-character delimiter for -F; got %q", s)
	}
	return r[0], nil
}

func streamCSVRows(stdin io.Reader, dst io.Writer, comma rune, fs string) error {
	r := csv.NewReader(bufio.NewReader(stdin))
	r.Comma = comma
	w := bufio.NewWriter(dst)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("csv parse: %w", err)
		}
		for i, field := range rec {
			if i > 0 {
				if _, err := w.WriteString(fs); err != nil {
					return err
				}
			}
			escaped := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(field)
			if _, err := w.WriteString(escaped); err != nil {
				return err
			}
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
	}
	return w.Flush()
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
