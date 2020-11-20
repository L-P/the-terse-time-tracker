// nolint:wrapcheck
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"tt/internal/tt"
)

// Version holds the compile-time version number.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	out := flag.CommandLine.Output()
	flag.Usage = func() {
		fmt.Fprint(out, "Please run `man tt` to obtain the documentation.\n")
		os.Exit(0)
	}

	v := flag.Bool("v", false, "displays tt version and exits")
	flag.Parse()
	if *v {
		fmt.Fprintf(out, "tt version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if err := run(os.Args[1:], out); err != nil {
		var e tt.ErrInvalidInput
		if errors.As(err, &e) {
			fmt.Fprintf(out, "error: %s\n", err)
			flag.Usage()
			os.Exit(1)
		}

		panic(err)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return tt.ErrInvalidInput("no task or command provided")
	}

	path, err := getDBPath()
	if err != nil {
		return err
	}

	tt, err := tt.New(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := tt.Close(); err != nil {
			fmt.Fprintf(out, "error: unable to close TT: %s", err)
		}
	}()

	return dispatch(tt, args, out)
}

var ErrNotImplemented = errors.New("not implemented")

var errConfigDir = errors.New("unable to fetch config dir")

func getDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", errConfigDir
	}

	return filepath.Join(dir, "the-terse-time-tracker.db"), nil
}
