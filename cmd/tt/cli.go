// nolint:wrapcheck
package main

import (
	"encoding/json"
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

type output struct {
	json bool
	w    io.Writer
}

func dispatch(app *tt.TT, args []string, w io.Writer) error {
	fset := flag.NewFlagSet("main", flag.ExitOnError)
	fset.Usage = func() {
		fmt.Fprint(w, t("Please run `man tt` to obtain the documentation.\n"))
		os.Exit(0)
	}

	showVersion := fset.Bool("v", false, t("displays tt version and exits"))
	showUI := fset.Bool("ui", false, t("displays the TUI"))
	startTask := fset.Bool("start", false, t("starts a new task or updates the current one"))
	stopTask := fset.Bool("stop", false, t("stops the current task"))
	loadFixtures := fset.Bool("fixture", false, t("clears the database and fills it with dev data"))
	showReport := fset.Bool("report", false, t("weekly report"))
	jsonOutput := fset.Bool("json", false, t("outputs JSON"))

	if err := fset.Parse(args); err != nil {
		return err
	}

	out := output{json: *jsonOutput, w: w}

	switch {
	case *showVersion:
		fmt.Fprintf(out.w, "tt version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return nil
	case *showUI:
		ui := ui.New(app)
		return ui.Run()
	case *stopTask:
		return stop(app, out)
	case *loadFixtures:
		return app.Fixture()
	case *showReport:
		return report(app, out)
	case *startTask:
		fallthrough
	default:
		if len(fset.Args()) == 0 {
			return showCurrent(app, out)
		}
		return start(app, fset.Args(), out)
	}
}

type currentTaskOutput struct {
	Task                *tt.Task
	DailyUntilOvertime  *time.Duration
	WeeklyUntilOvertime *time.Duration
}

func (c currentTaskOutput) String() string {
	var b strings.Builder

	if c.Task != nil {
		fmt.Fprintf(
			&b,
			t("Current task: %s %s, running for %s.\n"),
			c.Task.Description,
			strings.Join(c.Task.Tags, " "),
			util.FormatDuration(time.Since(c.Task.StartedAt)),
		)
	} else {
		fmt.Fprint(&b, t("There is no task running.\n"))
	}

	if c.DailyUntilOvertime == nil || c.WeeklyUntilOvertime == nil {
		return b.String()
	}

	writeTimeLeft(&b, *c.DailyUntilOvertime, *c.WeeklyUntilOvertime)

	return b.String()
}

func writeTimeLeft(w io.Writer, daily, weekly time.Duration) {
	switch {
	case daily > 0 && weekly > 0:
		fmt.Fprintf(
			w,
			t("%s left today, %s left before weekly overtime.\n"),
			util.FormatDuration(daily),
			util.FormatDuration(weekly),
		)
	case daily > 0 && weekly <= 0:
		fmt.Fprintf(
			w,
			t("%s left today, currently %s of weekly overtime.\n"),
			util.FormatDuration(daily),
			util.FormatDuration(-weekly),
		)
	case daily <= 0 && weekly > 0:
		fmt.Fprintf(
			w,
			t("%s overtime, %s left before weekly overtime.\n"),
			util.FormatDuration(-daily),
			util.FormatDuration(weekly),
		)
	case daily <= 0 && weekly <= 0:
		fmt.Fprintf(
			w,
			t("%s overtime, currently %s of weekly overtime.\n"),
			util.FormatDuration(-daily),
			util.FormatDuration(-weekly),
		)
	}
}

func showCurrent(app *tt.TT, out output) error {
	cur, err := app.CurrentTask()
	if err != nil {
		return err
	}

	var (
		ret       error
		formatter = currentTaskOutput{Task: cur}
	)
	if cur == nil {
		ret = tt.ErrExitCode(1)
	}

	if daily, weekly, err := app.GetDurationLeft(); err == nil {
		formatter.DailyUntilOvertime = &daily
		formatter.WeeklyUntilOvertime = &weekly
	} else if !errors.Is(err, tt.ErrNotConfigured) {
		return err
	}

	if out.json {
		enc := json.NewEncoder(out.w)
		if err := enc.Encode(formatter); err != nil {
			return err
		}
		return ret
	}

	fmt.Fprint(out.w, formatter.String())
	return ret
}

func start(app *tt.TT, args []string, out output) error {
	prev, next, err := app.Start(strings.Join(args, " "))
	if err != nil {
		if errors.Is(err, tt.ErrContinue) {
			fmt.Fprintf(
				out.w,
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
		fmt.Fprintf(out.w, t("Created task: \"%s\"\n"), next.Description)
		if len(next.Tags) > 0 {
			fmt.Fprintf(out.w, t("With tags: \"%s\"\n"), strings.Join(next.Tags, " "))
		}
	} else if prev != nil && prev.ID == next.ID {
		if len(next.Tags) == 0 {
			fmt.Fprint(out.w, "Removed tags from current task.\n")
		} else {
			fmt.Fprintf(out.w, "Replaced tags from current task: %s\n", strings.Join(next.Tags, " "))
		}
	}

	return nil
}

func writeStoppedTaskMessage(out output, task tt.Task) {
	fmt.Fprintf(
		out.w,
		t("Stopped task that had been running for %s: \"%s\".\n"),
		task.Duration().Round(time.Second),
		task.Description,
	)
}

func stop(app *tt.TT, out output) error {
	stopped, err := app.Stop()
	if err != nil {
		if errors.Is(err, tt.ErrNoCurrentTask) {
			fmt.Fprint(out.w, t("There is no running task.\n"))
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
