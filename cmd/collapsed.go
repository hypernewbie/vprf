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
	addBaseProfileFlags(fs, &opts)
	fs.IntVar(&opts.limit, "limit", 0, "Maximum stacks to return (0 = all)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	printWarnings(p, stderr)
	stacks := p.CollapsedStacks(selectedThreads(p, opts.thread))
	if opts.limit > 0 && len(stacks) > opts.limit {
		stacks = stacks[:opts.limit]
	}
	if opts.format == "json" {
		return writeRows(stdout, "json", nil, nil, stacks)
	}
	for _, s := range stacks {
		fmt.Fprintf(stdout, "%s %d\n", s.Stack, s.Count)
	}
	return nil
}
