// nolint:wrapcheck
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"tt/internal/tt"
)

// Version holds the compile-time version number.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		var e tt.ErrInvalidInput
		if errors.As(err, &e) {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			flag.Usage()
			os.Exit(1)
		}

		panic(err)
	}
}

func run(args []string, out io.Writer) error {
	path, err := getDBPath()
	if err != nil {
		return err
	}

	tt, err := tt.New(fmt.Sprintf(`file:%s?mode=rwc`, path))
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

var errConfigDir = errors.New(t("unable to fetch config dir"))

func getDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", errConfigDir
	}

	return filepath.Join(dir, "the-terse-time-tracker.db"), nil
}
