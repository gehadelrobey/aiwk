package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gehadelrobey/aiwk/internal/awk"
	"github.com/gehadelrobey/aiwk/internal/cache"
	"github.com/gehadelrobey/aiwk/internal/llm"
	"github.com/gehadelrobey/aiwk/internal/prompt"
)

// Options configures a single invocation.
type Options struct {
	Query    string
	FieldSep string

	DryRun   bool
	Explain  bool
	Confirm  bool
	ToAwk    bool
	NoCache  bool
	Verbose  bool
	AwkBin   string
	Provider string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Client llm.Client
	Store  *cache.Store
}

// Run generates (or loads) awk and optionally executes it on stdin.
func Run(ctx context.Context, o Options) error {
	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	if o.AwkBin == "" {
		o.AwkBin = "awk"
	}
	q := strings.TrimSpace(o.Query)
	if q == "" {
		return fmt.Errorf("missing natural language query")
	}
	if o.Client == nil {
		return fmt.Errorf("no LLM client configured")
	}

	logv := func(format string, args ...any) {
		if o.Verbose {
			fmt.Fprintf(o.Stderr, "[aiwk] "+format+"\n", args...)
		}
	}

	var (
		program string
		from    string
	)
	start := time.Now()

	if o.Store != nil && !o.NoCache {
		if s, hit, err := o.Store.Get(q, o.FieldSep, o.Explain, o.Provider); err != nil {
			logv("cache read error: %v", err)
		} else if hit {
			program = s
			from = "cache"
			logv("cache hit in %s", time.Since(start))
		}
	}

	if program == "" {
		var correction string
		const maxAttempts = 3
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			sys, usr := prompt.Build(q, o.FieldSep, o.Explain, correction)
			genStart := time.Now()
			raw, err := o.Client.Complete(ctx, sys, usr)
			if err != nil {
				return fmt.Errorf("llm: %w", err)
			}
			logv("llm round %d done in %s", attempt, time.Since(genStart))
			program = awk.ExtractProgram(raw)
			if strings.TrimSpace(program) == "" {
				return fmt.Errorf("model returned empty awk program")
			}
			tmp, err := awk.WriteTemp("", program)
			if err != nil {
				return err
			}
			valErr := awk.Validate(o.AwkBin, tmp)
			_ = os.Remove(tmp)
			if valErr != nil {
				correction = fmt.Sprintf("awk validation failed: %v\nProgram was:\n%s\n", valErr, program)
				logv("validation failed (attempt %d): %v", attempt, valErr)
				if attempt == maxAttempts {
					return fmt.Errorf("awk validation failed after %d attempts: %w", maxAttempts, valErr)
				}
				continue
			}
			from = "llm"
			break
		}
		if o.Store != nil && !o.NoCache && from == "llm" {
			if err := o.Store.Put(q, o.FieldSep, o.Explain, o.Provider, program); err != nil {
				logv("cache write error: %v", err)
			}
		}
	}

	logv("total generate path: %s (source=%s)", time.Since(start), from)

	if o.DryRun && !o.ToAwk {
		fmt.Fprintln(o.Stdout, "Generated awk:")
		fmt.Fprintln(o.Stdout, program)
		return nil
	}
	if o.ToAwk {
		fmt.Fprint(o.Stdout, program)
		if !strings.HasSuffix(program, "\n") {
			fmt.Fprintln(o.Stdout)
		}
		return nil
	}
	if o.Confirm {
		fmt.Fprintf(o.Stderr, "Generated awk:\n%s\n", program)
		if !confirm(o.Stdin, o.Stderr) {
			return fmt.Errorf("cancelled by user")
		}
	}
	tmp, err := awk.WriteTemp("", program)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	if err := awk.Validate(o.AwkBin, tmp); err != nil {
		return fmt.Errorf("awk validation: %w", err)
	}
	execStart := time.Now()
	if err := awk.Run(o.AwkBin, o.FieldSep, tmp, o.Stdin, o.Stdout); err != nil {
		return fmt.Errorf("awk: %w", err)
	}
	logv("awk finished in %s", time.Since(execStart))
	return nil
}

func confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Run this awk on the stream? [y/N]: ")
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// ClearCache removes all cached entries.
func ClearCache(s *cache.Store) error {
	if s == nil {
		return fmt.Errorf("cache not initialized")
	}
	return s.Clear()
}
