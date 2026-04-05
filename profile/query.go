package profile

import (
	"fmt"
	"sort"
	"strings"
)

type resolvedFrame struct {
	Function string
	Module   string
}

type functionCount struct {
	self   int
	total  int
	module string
}

func (p *Profile) ThreadViews() []ThreadView {
	views := make([]ThreadView, 0, len(p.Threads))
	for idx, thread := range p.Threads {
		views = append(views, ThreadView{
			Index:  idx,
			Name:   thread.Name,
			PID:    thread.PID,
			TID:    fmt.Sprint(thread.TID),
			Thread: thread,
		})
	}
	return views
}

func (p *Profile) Summary(threads []ThreadView) Summary {
	stats := p.TopFunctions(threads)
	stableSortFunctions(stats)
	threadStats := p.ThreadStats(threads)
	hottest := ThreadStat{}
	if len(threadStats) > 0 {
		hottest = threadStats[0]
	}
	top := FunctionStat{}
	if len(stats) > 0 {
		top = stats[0]
	}
	return Summary{
		ProfileName:     p.profileName(),
		DurationSeconds: p.durationSeconds(threads),
		TotalSamples:    totalSamplesForThreads(threads),
		ThreadCount:     len(threads),
		HottestThread:   hottest,
		TopFunction:     top,
	}
}

func (p *Profile) TopFunctions(threads []ThreadView) []FunctionStat {
	totalSamples := totalSamplesForThreads(threads)
	if totalSamples == 0 {
		return nil
	}
	funcCounts := map[string]*functionCount{}
	for _, tv := range threads {
		ctx := p.lookupContext(tv.Thread)
		for idx, stackIdx := range tv.Thread.Samples.Stack {
			weight := sampleWeight(tv.Thread.Samples, idx)
			if stackIdx == nil || weight == 0 {
				continue
			}
			frames := p.resolveStack(ctx, *stackIdx)
			if len(frames) == 0 {
				continue
			}
			leaf := frames[len(frames)-1]
			leafCounts := ensureCounts(funcCounts, leaf.Function, leaf.Module)
			leafCounts.self += weight
			seen := map[string]struct{}{}
			for _, frame := range frames {
				if _, ok := seen[frame.Function]; ok {
					continue
				}
				seen[frame.Function] = struct{}{}
				c := ensureCounts(funcCounts, frame.Function, frame.Module)
				c.total += weight
			}
		}
	}
	stats := make([]FunctionStat, 0, len(funcCounts))
	for name, c := range funcCounts {
		stats = append(stats, FunctionStat{
			Name:         name,
			Module:       c.module,
			SelfSamples:  c.self,
			TotalSamples: c.total,
			SelfPercent:  percent(c.self, totalSamples),
			TotalPercent: percent(c.total, totalSamples),
		})
	}
	stableSortFunctions(stats)
	return stats
}

func (p *Profile) CallersOf(function string, threads []ThreadView, limit int) []EdgeStat {
	return p.edgeStats(function, threads, limit, true)
}

func (p *Profile) CalleesOf(function string, threads []ThreadView, limit int) []EdgeStat {
	return p.edgeStats(function, threads, limit, false)
}

func (p *Profile) edgeStats(function string, threads []ThreadView, limit int, callers bool) []EdgeStat {
	totalSamples := totalSamplesForThreads(threads)
	counts := map[string]*EdgeStat{}
	for _, tv := range threads {
		ctx := p.lookupContext(tv.Thread)
		for idx, stackIdx := range tv.Thread.Samples.Stack {
			weight := sampleWeight(tv.Thread.Samples, idx)
			if stackIdx == nil || weight == 0 {
				continue
			}
			frames := p.resolveStack(ctx, *stackIdx)
			for i, frame := range frames {
				if frame.Function != function {
					continue
				}
				var path []string
				if callers {
					if i == 0 {
						path = []string{"[root]", function}
					} else {
						path = []string{frames[i-1].Function, function}
					}
				} else {
					if i == len(frames)-1 {
						continue
					}
					path = []string{function, frames[i+1].Function}
				}
				key := strings.Join(path, "\x00")
				stat := counts[key]
				if stat == nil {
					stat = &EdgeStat{Path: path}
					counts[key] = stat
				}
				stat.Samples += weight
			}
		}
	}
	stats := make([]EdgeStat, 0, len(counts))
	for _, stat := range counts {
		stat.Percent = percent(stat.Samples, totalSamples)
		stats = append(stats, *stat)
	}
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].Samples == stats[j].Samples {
			return strings.Join(stats[i].Path, " -> ") < strings.Join(stats[j].Path, " -> ")
		}
		return stats[i].Samples > stats[j].Samples
	})
	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}
	return stats
}

func (p *Profile) ThreadStats(threads []ThreadView) []ThreadStat {
	total := totalSamplesForThreads(threads)
	stats := make([]ThreadStat, 0, len(threads))
	for _, tv := range threads {
		samples := threadSampleCount(tv.Thread)
		stats = append(stats, ThreadStat{TID: tv.TID, Name: tv.Name, Samples: samples, Percent: percent(samples, total)})
	}
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].Samples == stats[j].Samples {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Samples > stats[j].Samples
	})
	return stats
}

func (p *Profile) HotPaths(threads []ThreadView, limit int) []HotPath {
	total := totalSamplesForThreads(threads)
	counts := map[string]*HotPath{}
	for _, tv := range threads {
		ctx := p.lookupContext(tv.Thread)
		for idx, stackIdx := range tv.Thread.Samples.Stack {
			weight := sampleWeight(tv.Thread.Samples, idx)
			if stackIdx == nil || weight == 0 {
				continue
			}
			frames := p.resolveStack(ctx, *stackIdx)
			functions := make([]string, 0, len(frames))
			for _, frame := range frames {
				functions = append(functions, frame.Function)
			}
			key := strings.Join(functions, "\x00")
			hp := counts[key]
			if hp == nil {
				hp = &HotPath{Functions: functions}
				counts[key] = hp
			}
			hp.Samples += weight
		}
	}
	paths := make([]HotPath, 0, len(counts))
	for _, path := range counts {
		path.Percent = percent(path.Samples, total)
		paths = append(paths, *path)
	}
	sort.SliceStable(paths, func(i, j int) bool {
		if paths[i].Samples == paths[j].Samples {
			return strings.Join(paths[i].Functions, " -> ") < strings.Join(paths[j].Functions, " -> ")
		}
		return paths[i].Samples > paths[j].Samples
	})
	if limit > 0 && len(paths) > limit {
		paths = paths[:limit]
	}
	return paths
}

type lookupContext struct {
	stackTable    StackTable
	frameTable    FrameTable
	funcTable     FuncTable
	resourceTable ResourceTable
	nativeSymbols NativeSymbolTable
	stringArray   []string
	resolver      *SidecarResolver
}

func (p *Profile) lookupContext(thread Thread) lookupContext {
	ctx := lookupContext{
		stackTable:    p.Shared.StackTable,
		frameTable:    p.Shared.FrameTable,
		funcTable:     p.Shared.FuncTable,
		resourceTable: p.Shared.ResourceTable,
		nativeSymbols: p.Shared.NativeSymbols,
		stringArray:   p.Shared.StringArray,
	}
	if len(thread.StackTable.Frame) > 0 {
		ctx.stackTable = thread.StackTable
	}
	if len(thread.FrameTable.Func) > 0 {
		ctx.frameTable = thread.FrameTable
	}
	if len(thread.FuncTable.Name) > 0 {
		ctx.funcTable = thread.FuncTable
	}
	if len(thread.ResourceTable.Name) > 0 || len(thread.ResourceTable.Lib) > 0 {
		ctx.resourceTable = thread.ResourceTable
	}
	if len(thread.NativeSymbols.LibIndex) > 0 {
		ctx.nativeSymbols = thread.NativeSymbols
	}
	if len(thread.StringArray) > 0 {
		ctx.stringArray = thread.StringArray
	}
	ctx.resolver = p.Resolver
	return ctx
}

func (p *Profile) resolveStack(ctx lookupContext, stackIdx int) []resolvedFrame {
	indices := make([]int, 0, 16)
	for stackIdx >= 0 && stackIdx < len(ctx.stackTable.Frame) {
		indices = append(indices, stackIdx)
		if stackIdx >= len(ctx.stackTable.Prefix) {
			break
		}
		prefix := ctx.stackTable.Prefix[stackIdx]
		if prefix == nil {
			break
		}
		stackIdx = *prefix
	}
	frames := make([]resolvedFrame, 0, len(indices))
	for i := len(indices) - 1; i >= 0; i-- {
		frameIdx := ctx.stackTable.Frame[indices[i]]
		frames = append(frames, p.resolveFrame(ctx, frameIdx))
	}
	return frames
}

func (p *Profile) resolveFrame(ctx lookupContext, frameIdx int) resolvedFrame {
	if frameIdx < 0 || frameIdx >= len(ctx.frameTable.Func) {
		return resolvedFrame{Function: "[unknown]", Module: "[unknown]"}
	}
	funcIdx := ctx.frameTable.Func[frameIdx]
	var nativeSymbol *int
	if frameIdx < len(ctx.frameTable.NativeSymbol) {
		nativeSymbol = ctx.frameTable.NativeSymbol[frameIdx]
	}
	module := p.lookupModule(ctx, funcIdx, nativeSymbol)
	name := p.lookupFuncName(ctx, funcIdx)
	if looksLikeAddress(name) {
		frameAddress := -1
		if frameIdx < len(ctx.frameTable.Address) {
			frameAddress = ctx.frameTable.Address[frameIdx]
		}
		if resolved := p.lookupSidecarName(ctx, module, nativeSymbol, frameAddress, name); resolved != "" {
			name = resolved
		}
	}
	return resolvedFrame{Function: name, Module: module}
}

func (p *Profile) lookupFuncName(ctx lookupContext, funcIdx int) string {
	if funcIdx < 0 || funcIdx >= len(ctx.funcTable.Name) {
		return "[unknown]"
	}
	nameIdx := ctx.funcTable.Name[funcIdx]
	if nameIdx < 0 || nameIdx >= len(ctx.stringArray) {
		return "[unknown]"
	}
	return ctx.stringArray[nameIdx]
}

func (p *Profile) lookupModule(ctx lookupContext, funcIdx int, nativeSymbol *int) string {
	if nativeSymbol != nil {
		idx := *nativeSymbol
		if idx >= 0 && idx < len(ctx.nativeSymbols.LibIndex) {
			libIdx := ctx.nativeSymbols.LibIndex[idx]
			if libIdx >= 0 && libIdx < len(p.Libs) {
				return p.Libs[libIdx].Name
			}
		}
	}
	if funcIdx >= 0 && funcIdx < len(ctx.funcTable.Resource) {
		resIdx := ctx.funcTable.Resource[funcIdx]
		if resIdx >= 0 && resIdx < len(ctx.resourceTable.Lib) {
			libPtr := ctx.resourceTable.Lib[resIdx]
			if libPtr != nil && *libPtr >= 0 && *libPtr < len(p.Libs) {
				return p.Libs[*libPtr].Name
			}
			if resIdx < len(ctx.resourceTable.Name) {
				nameIdx := ctx.resourceTable.Name[resIdx]
				if nameIdx >= 0 && nameIdx < len(ctx.stringArray) {
					return ctx.stringArray[nameIdx]
				}
			}
		}
	}
	return "[unknown]"
}

func attachSidecarSymbols(p *Profile, sidecar *SymbolSidecar) {
	resolver := &SidecarResolver{ByLib: map[string]map[int]string{}}
	for _, lib := range sidecar.Data {
		name := normalizeLibName(lib.DebugName)
		if _, ok := resolver.ByLib[name]; !ok {
			resolver.ByLib[name] = map[int]string{}
		}
		for _, entry := range lib.SymbolTable {
			funcName := lookupSidecarFunctionName(sidecar.StringTable, entry)
			if funcName == "" {
				continue
			}
			resolver.ByLib[name][entry.RVA] = funcName
		}
		for _, known := range lib.KnownAddresses {
			if len(known) != 2 {
				continue
			}
			addr := known[0]
			index := known[1]
			if index < 0 || index >= len(lib.SymbolTable) {
				continue
			}
			funcName := lookupSidecarFunctionName(sidecar.StringTable, lib.SymbolTable[index])
			if funcName == "" {
				continue
			}
			resolver.ByLib[name][addr] = funcName
		}
	}
	p.Resolver = resolver
}

func lookupSidecarFunctionName(strings []string, symbol SidecarSymbol) string {
	if len(symbol.Frames) > 0 {
		idx := symbol.Frames[0].Function
		if idx >= 0 && idx < len(strings) {
			return strings[idx]
		}
	}
	if symbol.Symbol >= 0 && symbol.Symbol < len(strings) {
		return strings[symbol.Symbol]
	}
	return ""
}

func normalizeLibName(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".pdb") {
		return strings.TrimSuffix(name, ".pdb")
	}
	return name
}

func (p *Profile) lookupSidecarName(ctx lookupContext, module string, nativeSymbol *int, frameAddress int, fallback string) string {
	if ctx.resolver == nil {
		return ""
	}
	if module == "[unknown]" && nativeSymbol != nil {
		idx := *nativeSymbol
		if idx >= 0 && idx < len(ctx.nativeSymbols.LibIndex) {
			libIdx := ctx.nativeSymbols.LibIndex[idx]
			if libIdx >= 0 && libIdx < len(p.Libs) {
				module = p.Libs[libIdx].Name
			}
		}
	}
	addresses := ctx.resolver.ByLib[normalizeLibName(module)]
	if len(addresses) == 0 {
		return ""
	}
	if frameAddress >= 0 {
		if symbol, ok := addresses[frameAddress]; ok {
			return symbol
		}
	}
	addr, ok := parseHexAddress(fallback)
	if !ok {
		return ""
	}
	if symbol, ok := addresses[addr]; ok {
		return symbol
	}
	return ""
}

func looksLikeAddress(name string) bool {
	return strings.HasPrefix(name, "0x")
}

func parseHexAddress(value string) (int, bool) {
	var parsed int
	_, err := fmt.Sscanf(value, "0x%x", &parsed)
	return parsed, err == nil
}

func (p *Profile) profileName() string {
	if strings.TrimSpace(p.Meta.Arguments) != "" {
		return p.Meta.Arguments
	}
	if strings.TrimSpace(p.Meta.Product) != "" {
		return p.Meta.Product
	}
	return "profile"
}

func (p *Profile) durationSeconds(threads []ThreadView) float64 {
	if p.Meta.EndTime > p.Meta.StartTime {
		return (p.Meta.EndTime - p.Meta.StartTime) / 1000.0
	}
	var min float64
	var max float64
	set := false
	for _, tv := range threads {
		times := sampleTimes(tv.Thread.Samples)
		for _, t := range times {
			if !set || t < min {
				min = t
			}
			if !set || t > max {
				max = t
			}
			set = true
		}
	}
	if !set {
		return 0
	}
	return (max - min) / 1000.0
}

func sampleTimes(samples SamplesTable) []float64 {
	if len(samples.Time) > 0 {
		return samples.Time
	}
	if len(samples.TimeDeltas) == 0 {
		return nil
	}
	times := make([]float64, len(samples.TimeDeltas))
	var current float64
	for i, delta := range samples.TimeDeltas {
		current += delta
		times[i] = current
	}
	return times
}

func totalSamplesForThreads(threads []ThreadView) int {
	total := 0
	for _, tv := range threads {
		total += threadSampleCount(tv.Thread)
	}
	return total
}

func threadSampleCount(thread Thread) int {
	total := 0
	for i := range thread.Samples.Stack {
		total += sampleWeight(thread.Samples, i)
	}
	return total
}

func sampleWeight(samples SamplesTable, index int) int {
	if index < len(samples.Weight) && samples.Weight[index] != 0 {
		return samples.Weight[index]
	}
	return 1
}

func percent(value int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) * 100 / float64(total)
}

func ensureCounts(counts map[string]*functionCount, name string, module string) *functionCount {
	count := counts[name]
	if count == nil {
		count = &functionCount{module: module}
		counts[name] = count
	}
	if count.module == "[unknown]" && module != "[unknown]" {
		count.module = module
	}
	return count
}

func stableSortFunctions(stats []FunctionStat) {
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].SelfSamples == stats[j].SelfSamples {
			if stats[i].TotalSamples == stats[j].TotalSamples {
				return stats[i].Name < stats[j].Name
			}
			return stats[i].TotalSamples > stats[j].TotalSamples
		}
		return stats[i].SelfSamples > stats[j].SelfSamples
	})
}
