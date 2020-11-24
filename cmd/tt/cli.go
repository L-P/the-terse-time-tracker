// nolint:wrapcheck
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
	"tt/internal/tt"
	"tt/internal/ui"
)

func dispatch(app *tt.TT, args []string, out io.Writer) error {
	fset := flag.NewFlagSet("main", flag.ExitOnError)
	fset.Usage = func() {
		fmt.Fprint(out, t("Please run `man tt` to obtain the documentation.\n"))
		os.Exit(0)
	}

	v := fset.Bool("v", false, t("displays tt version and exits"))
	showUI := fset.Bool("ui", false, t("displays the TUI"))
	startTask := fset.Bool("start", false, t("starts a new task or updates the current one"))
	stopTask := fset.Bool("s", false, t("stops the current task"))

	if err := fset.Parse(args); err != nil {
		return err
	}

	switch {
	case *v:
		fmt.Fprintf(out, "tt version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return nil
	case *showUI:
		ui := ui.New(app)
		return ui.Run()
	case *stopTask:
		return stop(app, out)
	case *startTask:
		fallthrough
	default:
		return start(app, fset.Args(), out)
	}
}

func start(app *tt.TT, args []string, out io.Writer) error {
	started, updated, err := app.Start(strings.Join(args, " "))
	if err != nil {
		if errors.Is(err, tt.ErrContinue) {
			fmt.Fprint(out, t("Task is already running, not doing anything.\n"))
			return nil
		}

		return err
	}

	if started != nil {
		fmt.Fprintf(out, t("Created task: \"%s\"\n"), started.Description)
		if len(started.Tags) > 0 {
			fmt.Fprintf(out, t("With tags: \"%s\"\n"), strings.Join(started.Tags, " "))
		}
	}

	if updated != nil { // nolint:nestif
		if updated.StoppedAt.IsZero() {
			if len(updated.Tags) == 0 {
				fmt.Fprint(out, "Removed tags from current task.\n")
			} else if started == nil {
				fmt.Fprintf(out, "Updated task with new tags: %s\n", strings.Join(updated.Tags, " "))
			}
		} else {
			writeStoppedTaskMessage(out, *updated)
		}
	}

	return nil
}

func writeStoppedTaskMessage(out io.Writer, task tt.Task) {
	fmt.Fprintf(
		out,
		t("Stopped task that had been running for %s: \"%s\"\n"),
		task.Duration().Round(time.Second),
		task.Description,
	)
}

func stop(app *tt.TT, out io.Writer) error {
	stopped, err := app.Stop()
	if err != nil {
		if errors.Is(err, tt.ErrNoCurrentTask) {
			fmt.Fprint(out, t("There is no running task.\n"))
			return nil
		}

		return err
	}

	if stopped != nil {
		writeStoppedTaskMessage(out, *stopped)
	}

	return nil
}

func t(msg string) string {
	return msg // TODO, handle locale
}
