package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

func runHotpath(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("hotpath", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts profileOptions
	addBaseProfileFlags(fs, &opts)
	fs.IntVar(&opts.limit, "limit", 10, "Maximum rows to return")
	fs.StringVar(&opts.function, "fn", "", "Regex pattern to filter paths by function")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	printWarnings(p, stderr)
	paths := p.HotPaths(selectedThreads(p, opts.thread), 0)
	if opts.function != "" {
		matched, err := p.MatchFunctions(opts.function)
		if err != nil {
			return err
		}
		if len(matched) == 0 {
			return fmt.Errorf("no functions matching %q found", opts.function)
		}
		matchedSet := make(map[string]bool, len(matched))
		for _, name := range matched {
			matchedSet[name] = true
		}
		filtered := paths[:0]
		for _, path := range paths {
			for _, name := range path.Functions {
				if matchedSet[name] {
					filtered = append(filtered, path)
					break
				}
			}
		}
		paths = filtered
	}
	if opts.limit > 0 && len(paths) > opts.limit {
		paths = paths[:opts.limit]
	}
	rows := make([][]string, 0, len(paths))
	for _, path := range paths {
		rows = append(rows, []string{fmt.Sprintf("%d", path.Samples), fmt.Sprintf("%.2f", path.Percent), strings.Join(path.Functions, " -> ")})
	}
	return writeRows(stdout, opts.format, []string{"samples", "percent", "path"}, rows, paths)
}
