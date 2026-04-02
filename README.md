# aiwk

> Plain English text processing for the Unix pipeline.

`aiwk` is a command-line tool that lets you describe what you want in plain English and executes it as an `awk` program under the hood. It lives natively in your shell pipeline.

```bash
cat access.log | aiwk "sum the bytes transferred, grouped by IP address"
```

Describe the transformation; `aiwk` generates and runs `awk`. You see the generated program by default. Works with pipes like any other CLI tool.

## Install

**Binaries**

```bash
# macOS
brew tap gehadelrobey/aiwk && brew install aiwk

# Linux
curl -sSL https://github.com/gehadelrobey/aiwk/releases/latest/download/install.sh | sh
```

**Go**

```bash
go install github.com/gehadelrobey/aiwk/cmd/aiwk@latest
```

**Source** (Go 1.21+): clone the repo, then `go build -o aiwk ./cmd/aiwk`.

Needs POSIX `awk` and an LLM API key:

```bash
export AIWK_API_KEY="sk-..."
export AIWK_PROVIDER="anthropic"   # openai | anthropic | ollama (defaults to openai)
```

## Usage

```bash
cat data.txt | aiwk "print the second column"
cat /etc/passwd | aiwk -F: "print the username and home directory"
cat sales.csv | aiwk --csv -F, "total revenue in column 4"

aiwk --dry-run "filter lines where field 1 starts with ERROR"   # print awk only
aiwk --explain "..."   # awk + inline comments
aiwk --to-awk "..." > script.awk && awk -f script.awk data.txt
```

## Flags

| Flag | Meaning |
|------|---------|
| `-F <sep>` | Field separator (same as awk) |
| `--csv` | Parse stdin as CSV, then awk on fields |
| `--dry-run` | Print generated awk, do not run |
| `--explain` | Generated awk with comments |
| `--confirm` | Approve before running |
| `--to-awk` | Raw awk only |
| `--model`, `--provider` | LLM backend |
| `--no-cache`, `--clear-cache` | Cache control |
| `--verbose` | Timing, tokens, cache on stderr |
| `--version` | Version |

## How it works

Your prompt (and flags like `-F`) go to the configured LLM → it returns awk → syntax-checked (`awk --lint`) → retried on failure → executed on stdin → cached in SQLite for reuse.

## Contributing

Issues and PRs welcome; larger changes are easier if discussed first. `go test ./...` and CI (`go vet`, `go test -race`) should pass.

## License

MIT — see [LICENSE](./LICENSE).
