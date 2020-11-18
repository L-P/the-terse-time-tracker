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

	"tt/internal/tt"
)

// Version holds the compile-time version number.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	out := flag.CommandLine.Output()
	flag.Usage = func() {
		fmt.Fprint(out, "Please run `tt help` to obtain the documentation.\n")
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

	tt, err := tt.New()
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

func dispatch(tt tt.TT, args []string, out io.Writer) (err error) {
	switch args[0] {
	case "stop":
		return ErrNotImplemented
	case "amend":
		return ErrNotImplemented
	case "help":
		return ErrNotImplemented
	case "start":
		args = args[1:]
		fallthrough
	default:
		_, err = tt.Start(strings.Join(args, " "))
		return ErrNotImplemented
	}

	return err
}
