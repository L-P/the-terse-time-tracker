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
	"tt/internal/util"
)

func dispatch(app *tt.TT, args []string, out io.Writer) error {
	fset := flag.NewFlagSet("main", flag.ExitOnError)
	fset.Usage = func() {
		fmt.Fprint(out, t("Please run `man tt` to obtain the documentation.\n"))
		os.Exit(0)
	}

	showVersion := fset.Bool("v", false, t("displays tt version and exits"))
	showUI := fset.Bool("ui", false, t("displays the TUI"))
	startTask := fset.Bool("start", false, t("starts a new task or updates the current one"))
	stopTask := fset.Bool("stop", false, t("stops the current task"))
	loadFixtures := fset.Bool("fixture", false, t("clears the database and fills it with dev data"))

	if err := fset.Parse(args); err != nil {
		return err
	}

	switch {
	case *showVersion:
		fmt.Fprintf(out, "tt version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return nil
	case *showUI:
		ui := ui.New(app)
		return ui.Run()
	case *stopTask:
		return stop(app, out)
	case *loadFixtures:
		return app.Fixture()
	case *startTask:
		fallthrough
	default:
		if len(fset.Args()) == 0 {
			return showCurrent(app, out)
		}
		return start(app, fset.Args(), out)
	}
}

func showCurrent(app *tt.TT, out io.Writer) error {
	cur, err := app.CurrentTask()
	if err != nil {
		return err
	}

	if cur == nil {
		fmt.Fprint(out, t("There is no task running.\n"))
		return nil
	}

	fmt.Fprintf(
		out,
		t("Current task: %s %s, running for %s\n"),
		cur.Description,
		strings.Join(cur.Tags, " "),
		util.FormatDuration(time.Since(cur.StartedAt)),
	)

	return nil
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
