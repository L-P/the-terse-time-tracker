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
		return tt.ErrExitCode(1)
	}

	fmt.Fprintf(
		out,
		t("Current task: %s %s, running for %s.\n"),
		cur.Description,
		strings.Join(cur.Tags, " "),
		util.FormatDuration(time.Since(cur.StartedAt)),
	)

	if dur, err := app.GetDailyDurationLeft(); err == nil {
		switch {
		case dur > 0:
			fmt.Fprintf(out, t("Time left before overtime: %s.\n"), util.FormatDuration(dur))
		case dur < 0:
			fmt.Fprintf(out, t("Current overtime: %s.\n"), util.FormatDuration(-dur))
		default:
			fmt.Fprint(out, t("Time to leave.\n"))
		}
	} else if !errors.Is(err, tt.ErrNotConfigured) {
		return err
	}

	return nil
}

func start(app *tt.TT, args []string, out io.Writer) error {
	prev, next, err := app.Start(strings.Join(args, " "))
	if err != nil {
		if errors.Is(err, tt.ErrContinue) {
			fmt.Fprintf(
				out,
				t("Task has already been running for %s, not doing anything.\n"),
				util.FormatDuration(time.Since(prev.StartedAt)),
			)
			return nil
		}

		return err
	}

	if prev != nil && prev.ID != next.ID {
		writeStoppedTaskMessage(out, *prev)
	}

	if prev == nil || prev.ID != next.ID {
		fmt.Fprintf(out, t("Created task: \"%s\"\n"), next.Description)
		if len(next.Tags) > 0 {
			fmt.Fprintf(out, t("With tags: \"%s\"\n"), strings.Join(next.Tags, " "))
		}
	} else if prev != nil && prev.ID == next.ID {
		if len(next.Tags) == 0 {
			fmt.Fprint(out, "Removed tags from current task.\n")
		} else {
			fmt.Fprintf(out, "Replaced tags from current task: %s\n", strings.Join(next.Tags, " "))
		}
	}

	return nil
}

func writeStoppedTaskMessage(out io.Writer, task tt.Task) {
	fmt.Fprintf(
		out,
		t("Stopped task that had been running for %s: \"%s\".\n"),
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
