package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gehadelrobey/aiwk/internal/app"
	"github.com/gehadelrobey/aiwk/internal/cache"
	"github.com/gehadelrobey/aiwk/internal/config"
	"github.com/gehadelrobey/aiwk/internal/llm"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"

	fieldSep   string
	dryRun     bool
	explain    bool
	confirm    bool
	toAwk      bool
	noCache    bool
	clearCache bool
	verbose    bool
	provider   string
	model      string
	awkBin     string
	cachePath  string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "aiwk [flags] \"natural language query\"",
	Short: "Plain English text processing for the Unix pipeline",
	Long:  "Describe an awk transformation in plain English; aiwk generates and runs awk on stdin.",
	Args: func(cmd *cobra.Command, args []string) error {
		if clearCache {
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("requires a natural language query (or use --clear-cache)")
		}
		return nil
	},
	Version: fmt.Sprintf("%s (%s)", version, commit),
	RunE:    runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	flags := rootCmd.Flags()
	flags.StringVarP(&fieldSep, "field-separator", "F", "", "field separator passed to awk (-F)")
	flags.BoolVar(&dryRun, "dry-run", false, "print generated awk without executing")
	flags.BoolVar(&explain, "explain", false, "ask the model for commented awk")
	flags.BoolVar(&confirm, "confirm", false, "prompt before running awk on the stream")
	flags.BoolVar(&toAwk, "to-awk", false, "output only the awk program")
	flags.BoolVar(&noCache, "no-cache", false, "skip the local SQLite cache")
	flags.BoolVar(&clearCache, "clear-cache", false, "delete the local prompt cache and exit")
	flags.BoolVarP(&verbose, "verbose", "v", false, "print timing and cache info to stderr")
	flags.StringVar(&provider, "provider", "", "LLM provider: openai, anthropic, ollama (default: $AIWK_PROVIDER or openai)")
	flags.StringVar(&model, "model", "", "model name (default: $AIWK_MODEL or provider default)")
	flags.StringVar(&awkBin, "awk-bin", "", "awk executable (default: awk)")
	flags.StringVar(&cachePath, "cache-path", "", "SQLite cache file (default: $AIWK_CACHE_PATH or XDG cache dir)")

	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unsupported shell %q", args[0])
		}
	},
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cp := strings.TrimSpace(cachePath)
	if cp == "" {
		cp = strings.TrimSpace(os.Getenv("AIWK_CACHE_PATH"))
	}
	var store *cache.Store
	if !clearCache && cp != "none" {
		var path string
		var err error
		if cp != "" {
			path = cp
		} else {
			path, err = cache.DefaultPath()
			if err != nil {
				return err
			}
		}
		st, err := cache.Open(path)
		if err != nil {
			return fmt.Errorf("cache: %w", err)
		}
		defer st.Close()
		store = st
	}

	if clearCache {
		if cp == "none" {
			return fmt.Errorf("--clear-cache cannot be used with cache disabled (--cache-path none)")
		}
		if store == nil {
			path := cp
			if path == "" {
				var err error
				path, err = cache.DefaultPath()
				if err != nil {
					return err
				}
			}
			st, err := cache.Open(path)
			if err != nil {
				return fmt.Errorf("cache: %w", err)
			}
			defer st.Close()
			store = st
		}
		if err := app.ClearCache(store); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "cache cleared")
		return nil
	}

	query := strings.Join(args, " ")

	envProv, apiKey, envModel, ollamaBase := config.FromEnv()
	prov := strings.TrimSpace(provider)
	if prov == "" {
		prov = envProv
	}
	m := strings.TrimSpace(model)
	if m == "" {
		m = envModel
	}

	client, err := llm.NewFromConfig(prov, apiKey, m, ollamaBase)
	if err != nil {
		return err
	}

	return app.Run(ctx, app.Options{
		Query:    query,
		FieldSep: fieldSep,
		DryRun:   dryRun,
		Explain:  explain,
		Confirm:  confirm,
		ToAwk:    toAwk,
		NoCache:  noCache,
		Verbose:  verbose,
		AwkBin:   awkBin,
		Provider: prov,
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   cmd.ErrOrStderr(),
		Client:   client,
		Store:    store,
	})
}
