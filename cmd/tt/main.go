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
	if err := run(os.Args[1:], os.Stdout); err != nil {
		var inputError tt.InvalidInputError
		if errors.As(err, &inputError) {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			flag.Usage()
			os.Exit(1)
		}

		var exitCode tt.ExitCodeError
		if errors.As(err, &exitCode) {
			os.Exit(exitCode.Code())
		}

		panic(err)
	}
}

func run(args []string, out io.Writer) error {
	path, err := getDBPath()
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf(`file:%s?mode=rwc`, path)
	tt, err := tt.New(dsn)
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
	if path := os.Getenv("TT_DB_PATH"); path != "" {
		return path, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", errConfigDir
	}

	return filepath.Join(dir, "the-terse-time-tracker.db"), nil
}
