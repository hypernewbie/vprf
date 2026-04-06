package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/hypernewbie/vprf/output"
	"github.com/hypernewbie/vprf/profile"
)

type command struct {
	name string
	run  func(args []string, stdout io.Writer, stderr io.Writer) error
	desc string
}

var commands = []command{
	{name: "record", run: runRecord, desc: "Record a profile with samply"},
	{name: "summary", run: runSummary, desc: "Show a compact profile summary"},
	{name: "top", run: runTop, desc: "Show top functions by self or total time"},
	{name: "callers", run: runCallers, desc: "Show callers of a function"},
	{name: "callees", run: runCallees, desc: "Show callees of a function"},
	{name: "threads", run: runThreads, desc: "Show per-thread sample counts"},
	{name: "hotpath", run: runHotpath, desc: "Show hottest call paths"},
	{name: "diff", run: runDiff, desc: "Compare two profiles"},
	{name: "collapsed", run: runCollapsed, desc: "Output collapsed stacks for flamegraph generation"},
}

func Execute(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	if args[0] == "help" {
		printUsage(stdout)
		return nil
	}

	for _, cmd := range commands {
		if cmd.name == args[0] {
			return cmd.run(args[1:], stdout, stderr)
		}
	}

	printUsage(stderr)
	return fmt.Errorf("unknown command %q", args[0])
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "vprf: queryable CPU profile CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  vprf <command> [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	for _, cmd := range commands {
		fmt.Fprintf(w, "  %-10s %s\n", cmd.name, cmd.desc)
	}
}

type profileOptions struct {
	profilePath string
	format      string
	thread      string
	limit       int
	sortBy      string
	function    string
}

func addBaseProfileFlags(fs *flag.FlagSet, opts *profileOptions) {
	fs.StringVar(&opts.profilePath, "profile", "", "Path to samply profile (.json or .json.gz)")
	fs.StringVar(&opts.profilePath, "p", "", "Path to samply profile (.json or .json.gz)")
	fs.StringVar(&opts.format, "format", "table", "Output format: table or json")
	fs.StringVar(&opts.thread, "thread", "", "Filter by thread name or tid")
}

func addProfileFlags(fs *flag.FlagSet, opts *profileOptions) {
	addBaseProfileFlags(fs, opts)
	fs.IntVar(&opts.limit, "limit", 10, "Maximum rows to return")
	fs.StringVar(&opts.sortBy, "sort", "self", "Sort field: self or total")
	fs.StringVar(&opts.function, "fn", "", "Regex pattern to match function names")
}

func loadProfile(opts profileOptions) (*profile.Profile, error) {
	if opts.profilePath == "" {
		return nil, errors.New("-profile is required")
	}
	p, err := profile.Load(opts.profilePath)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func selectedThreads(p *profile.Profile, threadFilter string) []profile.ThreadView {
	threads := p.ThreadViews()
	if threadFilter == "" {
		return threads
	}
	filtered := make([]profile.ThreadView, 0, len(threads))
	lower := strings.ToLower(threadFilter)
	for _, tv := range threads {
		if strings.Contains(strings.ToLower(tv.Name), lower) || tv.TID == threadFilter {
			filtered = append(filtered, tv)
		}
	}
	return filtered
}

func writeRows(stdout io.Writer, format string, headers []string, rows [][]string, payload any) error {
	switch format {
	case "table":
		output.WriteTable(stdout, headers, rows)
		return nil
	case "json":
		return output.WriteJSON(stdout, payload)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}
