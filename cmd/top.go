package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/hypernewbie/vprf/profile"
)

func runTop(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("top", flag.ContinueOnError)
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
	printWarnings(p, stderr)
	stats := p.TopFunctions(selectedThreads(p, opts.thread))
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
		filtered := make([]profile.FunctionStat, 0, len(stats))
		for _, stat := range stats {
			if matchedSet[stat.Name] {
				filtered = append(filtered, stat)
			}
		}
		stats = filtered
	}
	profile.SortFunctionStats(stats, opts.sortBy)
	if opts.limit > 0 && len(stats) > opts.limit {
		stats = stats[:opts.limit]
	}
	rows := make([][]string, 0, len(stats))
	for idx, stat := range stats {
		rows = append(rows, []string{
			fmt.Sprintf("%d", idx+1),
			fmt.Sprintf("%.2f", stat.SelfPercent),
			fmt.Sprintf("%.2f", stat.TotalPercent),
			stat.Name,
			stat.Module,
		})
	}
	return writeRows(stdout, opts.format, []string{"rank", "self%", "total%", "function", "module"}, rows, stats)
}
