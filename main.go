package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/krichprollsch/mblog/gen"
)

const (
	exitOK   = 0
	exitFail = 1
)

// main starts interruptable context and runs the program.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(exitFail)
	}

	os.Exit(exitOK)
}

// run configures the flags and generates static web pages.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {

	// declare runtime flag parameters.
	// TODO use env var by default is set.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		in  = flags.String("in", "in", "input markdown directory")
		out = flags.String("out", "out", "output html directory")
	)
	// usage func declaration.
	flags.Usage = func() {
		fmt.Fprintf(stdout, "%s is micro blogging static generator\n", args[0])
		fmt.Fprintf(stdout, "usage:\n")
		fmt.Fprintf(stdout, "\t%s [-in <input dir>] [-out <output dir>]\tgenerates the html.\n", args[0])
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	g := gen.New(os.DirFS(*in), *out)
	return g.Run(ctx)
}
