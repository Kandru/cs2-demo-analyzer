package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/update"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
)

const (
	updateCheckTimeout = 30 * time.Second
	updateApplyTimeout = 5 * time.Minute
)

// runUpdate implements the explicit update command.
func runUpdate(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var checkOnly, force bool
	fs.BoolVar(&checkOnly, "check", false, "only report whether a newer release is available")
	fs.BoolVar(&force, "force", false, "reinstall the latest release even if the version matches")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "run 'cs2-demo-analyzer help' for usage")
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "update accepts no arguments")
		fmt.Fprintln(os.Stderr, "run 'cs2-demo-analyzer help' for usage")
		return 2
	}

	current := version.Short()
	st, err := checkUpdate(ctx, current)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if checkOnly {
		printUpdateStatus(os.Stdout, st)
		return 0
	}

	_, currentErr := update.ParseVersion(current)
	if currentErr != nil && !force {
		fmt.Fprintf(os.Stderr, "error: current build %q is not a release version; rerun with --force to install %s\n", current, st.Latest)
		return 1
	}
	if !st.Newer && !force {
		printUpdateStatus(os.Stdout, st)
		return 0
	}

	if err := applyUpdate(ctx, st); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if st.Newer {
		fmt.Fprintf(os.Stdout, "updated %s -> %s\n", current, st.Latest)
	} else {
		fmt.Fprintf(os.Stdout, "reinstalled %s\n", st.Latest)
	}
	return 0
}

// maybeUpdate performs a best-effort update before analyze/watch. Failures are
// logged and never block the requested work.
func maybeUpdate(ctx context.Context, enabled bool, log *slog.Logger) {
	if !enabled {
		return
	}
	current := version.Short()
	st, err := checkUpdate(ctx, current)
	if err != nil {
		log.Warn("update check failed", "error", err)
		return
	}
	if _, err := update.ParseVersion(current); err != nil {
		log.Debug("skipping update: development build", "version", current, "latest", st.Latest)
		return
	}
	if !st.Newer {
		log.Debug("already up to date", "version", current)
		return
	}
	log.Info("updating before run", "from", current, "to", st.Latest)
	if err := applyUpdate(ctx, st); err != nil {
		log.Warn("update failed", "error", err)
		return
	}
	log.Info("updated binary; restart to use the new version", "from", current, "to", st.Latest)
}

func checkUpdate(ctx context.Context, current string) (update.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
	defer cancel()
	return update.NewClient().Check(ctx, current)
}

func applyUpdate(ctx context.Context, st update.Status) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, updateApplyTimeout)
	defer cancel()
	return update.NewClient().Apply(ctx, st, exe)
}

func printUpdateStatus(w io.Writer, st update.Status) {
	fmt.Fprintf(w, "current: %s\nlatest:  %s\n", st.Current, st.Latest)
	if st.Newer {
		fmt.Fprintf(w, "update available: %s\n", st.Latest)
		return
	}
	current, err := update.ParseVersion(st.Current)
	if err != nil {
		fmt.Fprintf(w, "current build is not a release version; use --force to install %s\n", st.Latest)
		return
	}
	latest, err := update.ParseVersion(st.Latest)
	if err == nil && update.Compare(current, latest) > 0 {
		fmt.Fprintf(w, "current build is newer than the latest release (%s)\n", st.Latest)
		return
	}
	fmt.Fprintln(w, "already up to date")
}
