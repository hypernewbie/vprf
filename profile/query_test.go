package profile

import (
	"os"
	"path/filepath"
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

	callers, _ := p.CallersOf("innerLoop", threads, 10)
	if len(callers) == 0 || callers[0].Path[0] != "outer" {
		t.Fatalf("expected outer -> innerLoop caller path, got %#v", callers)
	}

	callees, _ := p.CalleesOf("outer", threads, 10)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mapsKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
