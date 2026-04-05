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
	addProfileFlags(fs, &opts)
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := loadProfile(opts)
	if err != nil {
		return err
	}
	paths := p.HotPaths(selectedThreads(p, opts.thread), opts.limit)
	rows := make([][]string, 0, len(paths))
	for _, path := range paths {
		rows = append(rows, []string{fmt.Sprintf("%d", path.Samples), fmt.Sprintf("%.2f", path.Percent), strings.Join(path.Functions, " -> ")})
	}
	return writeRows(stdout, opts.format, []string{"samples", "percent", "path"}, rows, paths)
}
