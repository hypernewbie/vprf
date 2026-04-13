package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/hypernewbie/vprf/profile"
)

func runCallers(args []string, stdout io.Writer, stderr io.Writer) error {
	return runEdges("callers", args, stdout, stderr, func(p *profile.Profile, threads []profile.ThreadView, function string, limit int) ([]profile.EdgeStat, []string, error) {
		return p.CallersOf(function, threads, limit)
	})
}

func runCallees(args []string, stdout io.Writer, stderr io.Writer) error {
	return runEdges("callees", args, stdout, stderr, func(p *profile.Profile, threads []profile.ThreadView, function string, limit int) ([]profile.EdgeStat, []string, error) {
		return p.CalleesOf(function, threads, limit)
	})
}

func runEdges(name string, args []string, stdout io.Writer, stderr io.Writer, query func(*profile.Profile, []profile.ThreadView, string, int) ([]profile.EdgeStat, []string, error)) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts profileOptions
	addBaseProfileFlags(fs, &opts)
	fs.IntVar(&opts.limit, "limit", 10, "Maximum rows to return")
	fs.StringVar(&opts.function, "fn", "", "Regex pattern to match function names")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if opts.function == "" {
		return fmt.Errorf("--fn is required")
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	printWarnings(p, stderr)

	matchedFns, err := p.MatchFunctions(opts.function)
	if err != nil {
		return err
	}
	if len(matchedFns) == 0 {
		return fmt.Errorf("no functions matching %q found", opts.function)
	}
	if len(matchedFns) > 1 {
		fmt.Fprintf(stderr, "note: matched functions: %s\n", strings.Join(matchedFns, ", "))
	}

	stats, _, err := query(p, selectedThreads(p, opts.thread), opts.function, opts.limit)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(stats))
	for _, stat := range stats {
		rows = append(rows, []string{
			fmt.Sprintf("%d", stat.Samples),
			fmt.Sprintf("%.2f", stat.Percent),
			strings.Join(stat.Path, " -> "),
		})
	}
	return writeRows(stdout, opts.format, []string{"samples", "percent", name}, rows, stats)
}
