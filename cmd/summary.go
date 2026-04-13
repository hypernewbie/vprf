package cmd

import (
	"flag"
	"fmt"
	"io"
)

func runSummary(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts profileOptions
	addBaseProfileFlags(fs, &opts)
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	printWarnings(p, stderr)
	threads := selectedThreads(p, opts.thread)
	summary := p.Summary(threads)
	if opts.format == "json" {
		return writeRows(stdout, "json", nil, nil, summary)
	}
	fmt.Fprintf(stdout, "Profile: %s\n", summary.ProfileName)
	fmt.Fprintf(stdout, "Duration: %.2fs | Samples: %d | Threads: %d\n", summary.DurationSeconds, summary.TotalSamples, summary.ThreadCount)
	if summary.HottestThread.Name != "" {
		fmt.Fprintf(stdout, "Hottest thread: %s (%d samples)\n", summary.HottestThread.Name, summary.HottestThread.Samples)
	}
	if summary.TopFunction.Name != "" {
		fmt.Fprintf(stdout, "Top function: %s [self %.2f%% | total %.2f%%]\n", summary.TopFunction.Name, summary.TopFunction.SelfPercent, summary.TopFunction.TotalPercent)
	}
	return nil
}
