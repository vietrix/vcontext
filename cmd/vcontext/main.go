package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"vcontext/internal/common"
	"vcontext/internal/db"
	"vcontext/internal/mcp"
	"vcontext/internal/tools"
	"vcontext/internal/update"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logger := common.NewLogger()

	if handled := handleSubcommand(logger); handled {
		return
	}

	dbPath := resolveDBPath()
	store, err := db.Open(dbPath, logger)
	if err != nil {
		logger.Fatalf("failed to open db: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Printf("failed to close db: %v", err)
		}
	}()

	server := mcp.NewServer(logger)
	server.Register("tools/save_context/invoke", tools.SaveContextHandler(store))
	server.Register("tools/search_context/invoke", tools.SearchContextHandler(store))
	server.Register("tools/get_context/invoke", tools.GetContextHandler(store))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil {
		if err != context.Canceled {
			logger.Printf("server stopped: %v", err)
		}
	}
}

func resolveDBPath() string {
	var dbPath string
	fs := flag.NewFlagSet("vcontext", flag.ExitOnError)
	fs.StringVar(&dbPath, "db", "", "path to sqlite database")
	_ = fs.Parse(os.Args[1:])

	if dbPath != "" {
		return dbPath
	}

	if env := os.Getenv("VCONTEXT_DB_PATH"); env != "" {
		return env
	}

	return defaultDBPath()
}

func defaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return "vcontext.db"
	}

	dir := filepath.Join(configDir, "vcontext")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "vcontext.db")
}

func handleSubcommand(logger *log.Logger) bool {
	args := os.Args[1:]
	if len(args) == 0 {
		return false
	}

	switch strings.ToLower(args[0]) {
	case "version", "--version", "-version":
		fmt.Printf("vcontext %s (%s) %s\n", version, commit, date)
		return true
	case "update":
		runUpdate(logger, args[1:])
		return true
	default:
		return false
	}
}

func runUpdate(logger *log.Logger, args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	repo := fs.String("repo", "", "GitHub repo (org/name)")
	_ = fs.Parse(args)

	if *repo == "" {
		if env := strings.TrimSpace(os.Getenv("VCONTEXT_UPDATE_REPO")); env != "" {
			*repo = env
		} else {
			*repo = "vietrix/vcontext"
		}
	}

	tag, err := update.SelfUpdate(context.Background(), *repo, version)
	if err != nil {
		if errors.Is(err, update.ErrAlreadyLatest) {
			logger.Printf("already up to date (%s)", version)
			return
		}
		logger.Fatalf("update failed: %v", err)
	}

	logger.Printf("updated to %s, please restart the server", tag)
}
