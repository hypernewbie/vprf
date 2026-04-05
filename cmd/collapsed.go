package cmd

import (
	"flag"
	"fmt"
	"io"
)

func runCollapsed(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("collapsed", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts profileOptions
	addProfileFlags(fs, &opts)
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	stacks := p.CollapsedStacks(selectedThreads(p, opts.thread))
	for _, s := range stacks {
		fmt.Fprintf(stdout, "%s %d\n", s.Stack, s.Count)
	}
	return nil
}
