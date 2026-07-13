package session

import (
	"encoding/json"
	"time"
)

const (
	ModeWorkspace = "workspace"
	ModeReadonly  = "readonly"
	ModeBenchmark = "benchmark"

	ExitStatusClean   = "clean"
	ExitStatusAborted = "aborted"
	ExitStatusRunning = "running"
)

type Meta struct {
	ID            string    `json:"id"`
	Name          *string   `json:"name"`
	WorkspaceRoot string    `json:"workspace_root"`
	Mode          string    `json:"mode"`
	ModelID       string    `json:"model_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExitStatus    string    `json:"exit_status"`
	IsGitRepo     bool      `json:"is_git_repo"`
}

type Event struct {
	Seq       int             `json:"seq"`
	Timestamp time.Time       `json:"ts"`
	Kind      EventKind       `json:"kind"`
	Payload   json.RawMessage `json:"payload"`
	Tokens    int             `json:"tokens"`
	Source    string          `json:"source"`
}

type EventKind string

const (
	EvUserInstruction EventKind = "user_instruction"
	EvAssistantText   EventKind = "assistant_text"
	EvToolCall        EventKind = "tool_call"
	EvToolResult      EventKind = "tool_result"
	EvApproval        EventKind = "approval"
	EvSummary         EventKind = "summary"
	EvError           EventKind = "error"
)

type Summary struct {
	Goal          string         `json:"goal"`
	Decisions     []string       `json:"decisions"`
	ModifiedFiles []string       `json:"modified_files"`
	LatestDiff    string         `json:"latest_diff"`
	Verifications []Verification `json:"verifications"`
	OpenErrors    []string       `json:"open_errors"`
	Pending       []string       `json:"pending"`
	LastEventSeq  int            `json:"last_event_seq"`
}

type Verification struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Summary  string `json:"summary"`
}

type CreateOptions struct {
	Name          *string
	WorkspaceRoot string
	Mode          string
	ModelID       string
	IsGitRepo     bool
}

type EventReadResult struct {
	Events           []Event
	TrailingFragment bool
}
