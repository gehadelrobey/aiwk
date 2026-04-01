package app

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gehadelrobey/aiwk/internal/cache"
	"github.com/gehadelrobey/aiwk/internal/llm"
)

func TestRun_ToAwk(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := Run(context.Background(), Options{
		Query:    "print the second column",
		ToAwk:    true,
		Stdout:   &buf,
		Stderr:   &bytes.Buffer{},
		Client:   &llm.Mock{Text: "```awk\n{ print $2 }\n```"},
		Provider: "openai",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != "{ print $2 }" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_Pipe(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	t.Parallel()
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		Query:    "print second field",
		Stdin:    strings.NewReader("x y\na b\n"),
		Stdout:   &out,
		Stderr:   &bytes.Buffer{},
		Client:   &llm.Mock{Text: "{ print $2 }"},
		Provider: "openai",
		NoCache:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "y\nb\n" {
		t.Fatalf("got %q", out.String())
	}
}

func TestRun_CacheHitSkipsLLM(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	t.Parallel()
	st, err := cache.Open(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	calls := 0
	cl := llm.Func(func(ctx context.Context, system, user string) (string, error) {
		calls++
		return "{ print $1 }", nil
	})
	var out bytes.Buffer
	opts := Options{
		Query:    "print first column",
		Stdin:    strings.NewReader("hello world\n"),
		Stdout:   &out,
		Stderr:   &bytes.Buffer{},
		Client:   cl,
		Store:    st,
		Provider: "openai",
	}
	if err := Run(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("first run: llm calls = %d", calls)
	}
	out.Reset()
	opts.Stdin = strings.NewReader("hello world\n")
	if err := Run(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("second run should use cache; llm calls = %d", calls)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestRun_ValidationRetry(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	t.Parallel()
	calls := 0
	cl := llm.Func(func(ctx context.Context, system, user string) (string, error) {
		calls++
		if calls == 1 {
			return "{ this is not valid awk !!!", nil
		}
		return "{ print $1 }", nil
	})
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		Query:    "print first field",
		Stdin:    strings.NewReader("ok\n"),
		Stdout:   &out,
		Stderr:   &bytes.Buffer{},
		Client:   cl,
		Provider: "openai",
		NoCache:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 llm calls after retry; got %d", calls)
	}
	if out.String() != "ok\n" {
		t.Fatalf("got %q", out.String())
	}
}

func TestRun_CSVModeQuotedFields(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("awk not in PATH")
	}
	t.Parallel()
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		Query:    "print second field",
		FieldSep: ",",
		CSVMode:  true,
		Stdin:    strings.NewReader("id,name\n1,\"doe, john\"\n"),
		Stdout:   &out,
		Stderr:   &bytes.Buffer{},
		Client:   &llm.Mock{Text: "{ print $2 }"},
		Provider: "openai",
		NoCache:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "name\ndoe, john\n" {
		t.Fatalf("got %q", out.String())
	}
}
