package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/hypernewbie/vprf/profile"
)

func runCallers(args []string, stdout io.Writer, stderr io.Writer) error {
	return runEdges("callers", args, stdout, stderr, func(p *profile.Profile, threads []profile.ThreadView, function string, limit int) []profile.EdgeStat {
		return p.CallersOf(function, threads, limit)
	})
}

func runCallees(args []string, stdout io.Writer, stderr io.Writer) error {
	return runEdges("callees", args, stdout, stderr, func(p *profile.Profile, threads []profile.ThreadView, function string, limit int) []profile.EdgeStat {
		return p.CalleesOf(function, threads, limit)
	})
}

func runEdges(name string, args []string, stdout io.Writer, stderr io.Writer, query func(*profile.Profile, []profile.ThreadView, string, int) []profile.EdgeStat) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts profileOptions
	addProfileFlags(fs, &opts)
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
	stats := query(p, selectedThreads(p, opts.thread), opts.function, opts.limit)
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

var _ = profile.EdgeStat{}
