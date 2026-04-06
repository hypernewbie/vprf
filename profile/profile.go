package profile

type SymbolSidecar struct {
	StringTable []string            `json:"string_table"`
	Data        []SidecarLibrarySet `json:"data"`
}

type SidecarLibrarySet struct {
	DebugName      string          `json:"debug_name"`
	DebugID        string          `json:"debug_id"`
	CodeID         *string         `json:"code_id"`
	SymbolTable    []SidecarSymbol `json:"symbol_table"`
	KnownAddresses [][2]int        `json:"known_addresses"`
}

type SidecarSymbol struct {
	RVA    int                  `json:"rva"`
	Size   int                  `json:"size"`
	Symbol int                  `json:"symbol"`
	Frames []SidecarInlineFrame `json:"frames"`
}

type SidecarInlineFrame struct {
	Function int  `json:"function"`
	File     *int `json:"file"`
	Line     *int `json:"line"`
}

type SidecarResolver struct {
	ByLib       map[string]map[int]string       // exact address → name lookup
	ByLibSorted map[string][]SidecarSymbolEntry // sorted by RVA for range search
}

type SidecarSymbolEntry struct {
	RVA  int
	Size int
	Name string
}

type Profile struct {
	Meta          Meta             `json:"meta"`
	Libs          []Lib            `json:"libs"`
	Shared        SharedData       `json:"shared"`
	Threads       []Thread         `json:"threads"`
	Resolver      *SidecarResolver `json:"-"`
	FunctionNames map[string]bool  `json:"-"`
}

type Meta struct {
	Interval  float64 `json:"interval"`
	Product   string  `json:"product"`
	Arguments string  `json:"arguments"`
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
}

type Lib struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Arch string `json:"arch"`
}

type SharedData struct {
	StackTable    StackTable        `json:"stackTable"`
	FrameTable    FrameTable        `json:"frameTable"`
	FuncTable     FuncTable         `json:"funcTable"`
	ResourceTable ResourceTable     `json:"resourceTable"`
	NativeSymbols NativeSymbolTable `json:"nativeSymbols"`
	StringArray   []string          `json:"stringArray"`
}

type StackTable struct {
	Frame  []int  `json:"frame"`
	Prefix []*int `json:"prefix"`
	Length int    `json:"length"`
}

type FrameTable struct {
	Address      []int  `json:"address"`
	Func         []int  `json:"func"`
	NativeSymbol []*int `json:"nativeSymbol"`
	Line         []*int `json:"line"`
	Column       []*int `json:"column"`
	Length       int    `json:"length"`
}

type FuncTable struct {
	Name     []int `json:"name"`
	Resource []int `json:"resource"`
	Length   int   `json:"length"`
}

type ResourceTable struct {
	Lib    []*int `json:"lib"`
	Name   []int  `json:"name"`
	Type   []int  `json:"type"`
	Length int    `json:"length"`
}

type NativeSymbolTable struct {
	LibIndex []int `json:"libIndex"`
	Name     []int `json:"name"`
	Length   int   `json:"length"`
}

type Thread struct {
	Name          string            `json:"name"`
	PID           string            `json:"pid"`
	TID           any               `json:"tid"`
	Samples       SamplesTable      `json:"samples"`
	StackTable    StackTable        `json:"stackTable"`
	FrameTable    FrameTable        `json:"frameTable"`
	FuncTable     FuncTable         `json:"funcTable"`
	ResourceTable ResourceTable     `json:"resourceTable"`
	NativeSymbols NativeSymbolTable `json:"nativeSymbols"`
	StringArray   []string          `json:"stringArray"`
}

type SamplesTable struct {
	Stack      []*int    `json:"stack"`
	Time       []float64 `json:"time"`
	Weight     []int     `json:"weight"`
	Length     int       `json:"length"`
	WeightType string    `json:"weightType"`
	TimeDeltas []float64 `json:"timeDeltas"`
}

type ThreadView struct {
	Index  int
	Name   string
	PID    string
	TID    string
	Thread *Thread
}

type FunctionStat struct {
	Name         string  `json:"name"`
	Module       string  `json:"module"`
	SelfSamples  int     `json:"self_samples"`
	TotalSamples int     `json:"total_samples"`
	SelfPercent  float64 `json:"self_percent"`
	TotalPercent float64 `json:"total_percent"`
}

type EdgeStat struct {
	Path    []string `json:"path"`
	Samples int      `json:"samples"`
	Percent float64  `json:"percent"`
}

type ThreadStat struct {
	TID     string  `json:"tid"`
	Name    string  `json:"name"`
	Samples int     `json:"samples"`
	Percent float64 `json:"percent"`
}

type HotPath struct {
	Functions []string `json:"functions"`
	Samples   int      `json:"samples"`
	Percent   float64  `json:"percent"`
}

type Summary struct {
	ProfileName     string       `json:"profile_name"`
	DurationSeconds float64      `json:"duration_seconds"`
	TotalSamples    int          `json:"total_samples"`
	ThreadCount     int          `json:"thread_count"`
	HottestThread   ThreadStat   `json:"hottest_thread"`
	TopFunction     FunctionStat `json:"top_function"`
}

type DiffStat struct {
	Name           string  `json:"name"`
	Module         string  `json:"module"`
	SelfA          int     `json:"self_a"`
	SelfB          int     `json:"self_b"`
	DeltaSelf      int     `json:"delta_self"`
	TotalA         int     `json:"total_a"`
	TotalB         int     `json:"total_b"`
	DeltaTotal     int     `json:"delta_total"`
	PctChangeSelf  float64 `json:"pct_change_self"`
	PctChangeTotal float64 `json:"pct_change_total"`
}

type CollapsedStack struct {
	Stack string `json:"stack"`
	Count int    `json:"count"`
}
