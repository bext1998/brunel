package tools

// Result is a structured terminal result for one of the eight frozen tools.
// Exactly one payload field is non-nil for a successful call.
type Result struct {
	Tool     string          `json:"tool"`
	List     *ListResult     `json:"list,omitempty"`
	Search   *SearchResult   `json:"search,omitempty"`
	ReadFile *ReadFileResult `json:"read_file,omitempty"`
	Hash     *HashResult     `json:"hash,omitempty"`
	Run      *RunResult      `json:"run,omitempty"`
	Diff     *DiffResult     `json:"diff,omitempty"`
}

type ListResult struct {
	Entries []FileEntry `json:"entries"`
}

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type SearchResult struct {
	Matches []TextMatch `json:"matches"`
}

type TextMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

type ReadFileResult struct {
	Lines []FileLine `json:"lines"`
	Hash  string     `json:"hash"`
}

type FileLine struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

type HashResult struct {
	Hash string `json:"hash"`
}

type RunResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	Truncated bool   `json:"truncated"`
}

type DiffResult struct {
	Diff string `json:"diff"`
}
