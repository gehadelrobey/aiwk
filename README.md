# aiwk

> Plain English text processing for the Unix pipeline.

`aiwk` is a command-line tool that lets you describe what you want in plain English and executes it as an `awk` program under the hood. It lives natively in your shell pipeline — no GUI, no chat interface, no context switching.

```bash
cat access.log | aiwk "sum the bytes transferred, grouped by IP address"
```

---

## Why aiwk?

`awk` is one of the most powerful text processing tools ever written. It is also notoriously hard to remember. Field separators, NR vs FNR, printf formatting, associative arrays — the syntax gets in the way of getting things done.

`aiwk` removes that friction. You describe the transformation in plain English; it generates, validates, and runs the `awk` program for you. The result is a tool that feels natural to newcomers and saves time for experts.

---

## Features

### Core
- **Natural language input** — describe your transformation in plain English, no awk syntax required
- **Runs awk under the hood** — output is deterministic, fast, and portable; no runtime dependency beyond awk itself
- **Full pipe support** — works as a first-class Unix pipe primitive: `cat`, `grep`, `sort`, `aiwk`, `wc` all compose naturally
- **Always shows generated awk** — transparency by default; you always know what is being run

### Usability
- **`--dry-run` mode** — print the generated awk program without executing it; great for auditing or learning
- **`--explain` flag** — annotates the generated awk with inline comments explaining each part
- **`--confirm` flag** — prompts you to approve the generated awk before running it on large or sensitive files
- **Custom field separators** — pass `-F ":"` just like native awk; the context is forwarded to the model
- **Multiline programs** — handles complex logic (multi-pass, associative arrays, BEGIN/END blocks) not just one-liners

### Performance & Reliability
- **Local cache** — prompt/awk pairs are cached in a local SQLite database; repeated queries are instant and free
- **Offline fallback** — optionally routes to a local model via Ollama when no internet connection is available
- **Validation layer** — generated awk is syntax-checked before execution; bad output is retried automatically
- **Streaming input safe** — correctly handles large files and piped streams without loading into memory

### Developer Experience
- **`--to-awk` flag** — outputs only the raw awk program, nothing else; pipe it into a script or save it for later
- **`--model` flag** — choose which LLM backend to use (OpenAI, Anthropic, Ollama, etc.)
- **`--verbose` flag** — prints timing, token usage, and cache hit/miss information
- **Shell completion** — tab completion for flags in bash, zsh, and fish

---

## Installation

**Pre-built binaries** (recommended):

```bash
# macOS (Homebrew)
brew tap gehadelrobey/aiwk
brew install aiwk

# Linux
curl -sSL https://github.com/gehadelrobey/aiwk/releases/latest/download/install.sh | sh
```

**Via `go install`:**

```bash
go install github.com/gehadelrobey/aiwk/cmd/aiwk@latest
```

**From source** (requires Go 1.21+):

```bash
git clone https://github.com/gehadelrobey/aiwk
cd aiwk
go build -o aiwk ./cmd/aiwk
sudo mv aiwk /usr/local/bin/
```

Requires `awk` (any POSIX-compatible version) and an API key for your chosen LLM provider.

```bash
export AIWK_API_KEY="sk-..."        # OpenAI or Anthropic key
export AIWK_PROVIDER="anthropic"    # openai | anthropic | ollama
```

---

## Usage

### Basic

```bash
# Print the second column of every line
cat data.txt | aiwk "print the second column"

# Filter lines where the third field exceeds 500
cat data.txt | aiwk "show lines where field 3 is greater than 500"

# Count the number of lines
cat data.txt | aiwk "count how many lines there are"
```

### With a field separator

```bash
# Parse a colon-separated file
cat /etc/passwd | aiwk -F: "print the username and home directory"

# Parse a CSV
cat sales.csv | aiwk -F, "show the total revenue in column 4"

# Parse CSV safely when fields may contain quoted commas
cat sales.csv | aiwk --csv -F, "show the total revenue in column 4"
```

### Aggregation and grouping

```bash
# Sum bytes per IP from an access log
cat access.log | aiwk "sum the bytes transferred grouped by IP address"

# Count occurrences of each HTTP status code
cat access.log | aiwk "count how many times each status code appears"

# Average response time per endpoint
cat access.log | aiwk "calculate the average response time for each URL path"
```

### Dry run and explanation

```bash
# See the generated awk without running it
cat data.txt | aiwk --dry-run "print every line where the first field starts with ERROR"

# Output:
# Generated awk:
# /^ERROR/ { print }

# See the generated awk with comments
cat data.txt | aiwk --explain "print every line where the first field starts with ERROR"

# Output:
# /^ERROR/ { print }
# ^-- regex anchor: match from start of line
# ERROR   -- literal string to match in field 1
# print   -- print the entire matching line
```

### Save the awk for reuse

```bash
# Extract just the awk program and save it
aiwk --to-awk "print lines where column 2 is a valid email address" > validate_email.awk
awk -f validate_email.awk users.txt
```

### Using in scripts

```bash
#!/bin/bash
# Nightly report: top 10 IPs by bandwidth
cat /var/log/nginx/access.log \
  | aiwk "sum bytes by IP" \
  | sort -rn \
  | head -10
```

---

## Flags

| Flag | Description |
|---|---|
| `-F <sep>` | Field separator, passed directly to awk |
| `--csv` | Parse stdin as CSV first (quoted fields supported), then run awk on parsed fields |
| `--dry-run` | Print the generated awk program without executing |
| `--explain` | Print the generated awk with inline explanatory comments |
| `--confirm` | Prompt for approval before executing |
| `--to-awk` | Output only the raw awk program (no execution) |
| `--model <n>` | LLM model to use (default: configured provider default) |
| `--provider <n>` | LLM provider: `openai`, `anthropic`, `ollama` |
| `--no-cache` | Skip the local cache and always call the LLM |
| `--clear-cache` | Delete the local prompt cache |
| `--verbose` | Print timing, token usage, and cache info to stderr |
| `--version` | Show version and exit |

---

## How it works

1. Your natural language query and any flags (like `-F`) are assembled into a structured prompt
2. The prompt is sent to the configured LLM, asking it to produce a single valid awk program
3. The response is syntax-checked using `awk --lint` before execution
4. If the check fails, the model is asked to self-correct (up to 2 retries)
5. The validated awk program is executed via subprocess with your piped input
6. The (query → awk) pair is stored in a local SQLite cache for future reuse

---

## Roadmap

- [ ] `--to-sed` and `--to-jq` variants for the same workflow
- [ ] Interactive REPL mode for iterative data exploration
- [ ] Named saved queries (`aiwk --save "my-report"`)
- [ ] VS Code extension for in-editor dry runs
- [ ] Support for multi-file inputs with FILENAME awareness
- [ ] Fine-tuned local model specifically trained on awk programs

---

## Contributing

Contributions are welcome. Please open an issue before submitting large changes so we can discuss approach first. All PRs should include tests for any new behavior.

GitHub Actions runs `go vet` and `go test -race ./...` on pushes and pull requests to `main`. Pushing a SemVer tag `v*` (for example `v0.1.0`) triggers a [GoReleaser](https://goreleaser.com/) workflow that uploads cross-compiled binaries and checksums to a GitHub Release.

```bash
git clone https://github.com/gehadelrobey/aiwk
cd aiwk
go mod download
go test ./...
```

---

## License

MIT — see [LICENSE](./LICENSE) for full text.
