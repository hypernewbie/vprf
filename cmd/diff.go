package cmd

import (
	"flag"
	"fmt"
	"io"
	"sort"

	"github.com/hypernewbie/vprf/profile"
)

func runDiff(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profileA := fs.String("a", "", "First profile (baseline)")
	profileB := fs.String("b", "", "Second profile (comparison)")
	format := fs.String("format", "table", "Output format: table or json")
	limit := fs.Int("limit", 20, "Maximum rows to return")
	sortBy := fs.String("sort", "delta", "Sort by: delta, self_a, self_b, name")
	threadA := fs.String("thread-a", "", "Filter threads in profile A")
	threadB := fs.String("thread-b", "", "Filter threads in profile B")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *profileA == "" || *profileB == "" {
		return fmt.Errorf("-a and -b are required")
	}

	pA, err := profile.Load(*profileA)
	if err != nil {
		return fmt.Errorf("load profile A: %w", err)
	}
	pB, err := profile.Load(*profileB)
	if err != nil {
		return fmt.Errorf("load profile B: %w", err)
	}

	threadsA := selectedThreads(pA, *threadA)
	threadsB := selectedThreads(pB, *threadB)

	diffs := pA.DiffProfiles(pB, threadsA, threadsB)

	switch *sortBy {
	case "self_a":
		diffs = diffStatsBySelfA(diffs)
	case "self_b":
		diffs = diffStatsBySelfB(diffs)
	case "name":
		diffs = diffStatsByName(diffs)
	}

	if *limit > 0 && len(diffs) > *limit {
		diffs = diffs[:*limit]
	}

	rows := make([][]string, 0, len(diffs))
	for _, d := range diffs {
		rows = append(rows, []string{
			fmt.Sprintf("%d", d.DeltaSelf),
			fmt.Sprintf("%.2f", d.PctChangeSelf),
			fmt.Sprintf("%d", d.SelfA),
			fmt.Sprintf("%d", d.SelfB),
			fmt.Sprintf("%d", d.DeltaTotal),
			fmt.Sprintf("%.2f", d.PctChangeTotal),
			d.Name,
			d.Module,
		})
	}
	return writeRows(stdout, *format, []string{"delta_self", "pct_chg", "self_a", "self_b", "delta_total", "pct_chg_total", "function", "module"}, rows, diffs)
}

func diffStatsBySelfA(diffs []profile.DiffStat) []profile.DiffStat {
	sorted := make([]profile.DiffStat, len(diffs))
	copy(sorted, diffs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SelfA == sorted[j].SelfA {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].SelfA > sorted[j].SelfA
	})
	return sorted
}

func diffStatsBySelfB(diffs []profile.DiffStat) []profile.DiffStat {
	sorted := make([]profile.DiffStat, len(diffs))
	copy(sorted, diffs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SelfB == sorted[j].SelfB {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].SelfB > sorted[j].SelfB
	})
	return sorted
}

func diffStatsByName(diffs []profile.DiffStat) []profile.DiffStat {
	sorted := make([]profile.DiffStat, len(diffs))
	copy(sorted, diffs)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}
