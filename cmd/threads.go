package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/hypernewbie/vprf/profile"
)

func runThreads(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("threads", flag.ContinueOnError)
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
	stats := p.ThreadStats(selectedThreads(p, opts.thread))
	rows := make([][]string, 0, len(stats))
	for _, stat := range stats {
		rows = append(rows, []string{stat.TID, stat.Name, fmt.Sprintf("%d", stat.Samples), fmt.Sprintf("%.2f", stat.Percent)})
	}
	return writeRows(stdout, opts.format, []string{"tid", "name", "samples", "%total"}, rows, stats)
}

var _ = profile.ThreadStat{}
