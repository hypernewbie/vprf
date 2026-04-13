package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndQueryFixture(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	threads := p.ThreadViews()
	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}

	stats := p.TopFunctions(threads)
	if len(stats) == 0 || stats[0].Name != "innerLoop" {
		t.Fatalf("expected innerLoop hottest, got %#v", stats)
	}

	callers, _, err := p.CallersOf("innerLoop", threads, 10)
	if err != nil {
		t.Fatalf("CallersOf error: %v", err)
	}
	if len(callers) == 0 || callers[0].Path[0] != "outer" {
		t.Fatalf("expected outer -> innerLoop caller path, got %#v", callers)
	}

	callees, _, err := p.CalleesOf("outer", threads, 10)
	if err != nil {
		t.Fatalf("CalleesOf error: %v", err)
	}
	if len(callees) == 0 || callees[0].Path[1] != "innerLoop" {
		t.Fatalf("expected outer -> innerLoop callee path, got %#v", callees)
	}

	hotpaths := p.HotPaths(threads, 5)
	if len(hotpaths) == 0 || len(hotpaths[0].Functions) < 2 {
		t.Fatalf("expected hot path, got %#v", hotpaths)
	}
}

func TestLoadAttachesSymbolSidecar(t *testing.T) {
	path := filepath.Join("..", "real-profile-debug.json.gz")
	if _, err := os.Stat(path); err != nil {
		t.Skip("real samply profile not present")
	}
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load real profile: %v", err)
	}
	if p.Resolver == nil {
		t.Fatalf("expected sidecar resolver to be attached")
	}
	threads := p.ThreadViews()
	ctx := p.lookupContext(threads[0].Thread)
	if ctx.resolver == nil {
		t.Fatalf("expected lookup context resolver")
	}
	if len(ctx.resolver.ByLib["burn-go-debug"]) == 0 {
		t.Fatalf("expected burn-go-debug sidecar entries")
	}
	module := p.lookupModule(ctx, ctx.frameTable.Func[4], nil)
	if module != "burn-go-debug" {
		t.Fatalf("expected module burn-go-debug, got %q", module)
	}
	if _, ok := ctx.resolver.ByLib[module]; !ok {
		t.Fatalf("expected exact module key %q in resolver; keys=%v", module, mapsKeys(ctx.resolver.ByLib))
	}
	if _, ok := ctx.resolver.ByLib[module][ctx.frameTable.Address[4]]; !ok {
		t.Fatalf("expected address %d for module %q", ctx.frameTable.Address[4], module)
	}
	if got := p.lookupSidecarName(ctx, module, nil, ctx.frameTable.Address[4], "0x9edfc"); got != "main.innerLoop" {
		t.Fatalf("expected sidecar lookup to return main.innerLoop, got %q", got)
	}
	stats := p.TopFunctions(threads)
	for _, stat := range stats {
		if stat.Name == "main.innerLoop" {
			return
		}
	}
	t.Fatalf("expected main.innerLoop in top functions, got %#v", stats[:min(5, len(stats))])
}

func mapsKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func TestCollapsedStacks(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	threads := p.ThreadViews()
	stacks := p.CollapsedStacks(threads)
	if len(stacks) == 0 {
		t.Fatalf("expected non-empty collapsed stacks")
	}
	found := false
	for _, s := range stacks {
		if s.Stack == "main;outer;innerLoop" && s.Count == 15 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected main;outer;innerLoop with count 15 in stacks, got %v", stacks)
	}
}

func TestThreadStats(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	threads := p.ThreadViews()
	stats := p.ThreadStats(threads)
	if len(stats) != 2 {
		t.Fatalf("expected 2 thread stats, got %d", len(stats))
	}
	mainSamples := 0
	for _, s := range stats {
		if s.Name == "main" {
			mainSamples = s.Samples
		}
	}
	if mainSamples <= 0 {
		t.Fatalf("expected main thread to have samples")
	}
	totalPercent := 0.0
	for _, s := range stats {
		totalPercent += s.Percent
	}
	if totalPercent < 99.0 || totalPercent > 101.0 {
		t.Fatalf("expected percents to sum to ~100, got %.2f", totalPercent)
	}
}

func TestDiffProfilesIdentical(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p1, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	p2, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	threads := p1.ThreadViews()
	diffs := p1.DiffProfiles(p2, threads, threads)
	for _, d := range diffs {
		if d.DeltaSelf != 0 || d.DeltaTotal != 0 {
			t.Fatalf("expected zero deltas for identical profiles, got DeltaSelf=%d DeltaTotal=%d for %s", d.DeltaSelf, d.DeltaTotal, d.Name)
		}
		if d.PctChangeSelf != 0.0 || d.PctChangeTotal != 0.0 {
			t.Fatalf("expected zero percent changes for identical profiles, got PctChangeSelf=%.2f PctChangeTotal=%.2f for %s", d.PctChangeSelf, d.PctChangeTotal, d.Name)
		}
	}
}

func TestMatchFunctions(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	matched, err := p.MatchFunctions("inner.*")
	if err != nil {
		t.Fatalf("MatchFunctions error: %v", err)
	}
	found := false
	for _, name := range matched {
		if name == "innerLoop" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected innerLoop in matched functions, got %v", matched)
	}

	matchedAll, err := p.MatchFunctions(".*")
	if err != nil {
		t.Fatalf("MatchFunctions .* error: %v", err)
	}
	if len(matchedAll) == 0 {
		t.Fatalf("expected at least some matches for .* pattern")
	}

	_, err = p.MatchFunctions("[invalid")
	if err == nil {
		t.Fatalf("expected error for invalid regex [invalid")
	}
}

func TestEdgeStatsInvalidRegex(t *testing.T) {
	path := filepath.Join("..", "tests", "testdata", "fixture.json")
	p, err := Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	threads := p.ThreadViews()
	_, _, err = p.CallersOf("[invalid", threads, 10)
	if err == nil {
		t.Fatalf("expected error for invalid regex in CallersOf")
	}
}

func TestLoadBadPath(t *testing.T) {
	_, err := Load("/nonexistent/path/to/profile.json")
	if err == nil {
		t.Fatalf("expected error for nonexistent path")
	}
}

func TestLoadSidecarWarning(t *testing.T) {
	fixturePath := filepath.Join("..", "tests", "testdata", "fixture.json")
	fixtureData, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	tempDir := t.TempDir()
	profilePath := filepath.Join(tempDir, "fixture.json")
	if err := os.WriteFile(profilePath, fixtureData, 0o644); err != nil {
		t.Fatalf("write fixture copy: %v", err)
	}
	if err := os.WriteFile(profilePath+".syms.json", []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write malformed sidecar: %v", err)
	}

	p, err := Load(profilePath)
	if err != nil {
		t.Fatalf("load profile with malformed sidecar: %v", err)
	}
	if len(p.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(p.Warnings), p.Warnings)
	}
	if !strings.Contains(strings.ToLower(p.Warnings[0]), "sidecar") {
		t.Fatalf("expected sidecar warning, got %q", p.Warnings[0])
	}
}
