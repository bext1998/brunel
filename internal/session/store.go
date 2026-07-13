package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const abortedRetention = 72 * time.Hour

type Store struct {
	root string
	now  func() time.Time
}

type Session struct {
	store *Store
	dir   string
	meta  Meta

	mu             sync.Mutex
	lastSeq        int
	trailingEvents bool
	closed         bool
}

func NewStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		var base string
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			base = local
		} else if cache, err := os.UserCacheDir(); err == nil {
			base = cache
		} else {
			return nil, codeError(ErrInvalidArgument.Code, "cannot determine local application data directory", err)
		}
		root = filepath.Join(base, "Brunel", "sessions")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, codeError(ErrInvalidArgument.Code, "invalid session root", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, codeError("E_SESSION_STORAGE", "cannot create session root", err)
	}
	store := &Store{root: root, now: time.Now}
	// Expired unnamed aborted sessions are cleaned during the next explicit
	// store startup; there is intentionally no background cleanup goroutine.
	if _, err := store.CleanupExpired(store.now()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Root() string { return s.root }

func (s *Store) Create(options CreateOptions) (*Session, error) {
	if options.Mode == "" {
		options.Mode = ModeWorkspace
	}
	if options.Mode != ModeWorkspace && options.Mode != ModeReadonly && options.Mode != ModeBenchmark {
		return nil, codeError(ErrInvalidArgument.Code, "unsupported session mode", nil)
	}
	if strings.TrimSpace(options.WorkspaceRoot) == "" {
		return nil, codeError(ErrInvalidArgument.Code, "workspace root is required", nil)
	}
	if strings.IndexByte(options.WorkspaceRoot, 0) >= 0 {
		return nil, codeError(ErrInvalidArgument.Code, "workspace root contains NUL", nil)
	}
	workspaceRoot := options.WorkspaceRoot
	if workspaceRoot != "" {
		var err error
		workspaceRoot, err = filepath.Abs(workspaceRoot)
		if err != nil {
			return nil, codeError(ErrInvalidArgument.Code, "invalid workspace root", err)
		}
	}
	if options.Name != nil && strings.IndexByte(*options.Name, 0) >= 0 {
		return nil, codeError(ErrInvalidArgument.Code, "session name contains NUL", nil)
	}

	for attempt := 0; attempt < 5; attempt++ {
		id, err := newULID(s.now())
		if err != nil {
			return nil, codeError("E_SESSION_ID", "cannot generate session id", err)
		}
		dir := filepath.Join(s.root, id)
		if err := os.Mkdir(dir, 0o700); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return nil, codeError("E_SESSION_STORAGE", "cannot create session directory", err)
		}
		meta := Meta{ID: id, Name: cloneStringPtr(options.Name), WorkspaceRoot: workspaceRoot, Mode: options.Mode, ModelID: options.ModelID, CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC(), ExitStatus: ExitStatusRunning, IsGitRepo: options.IsGitRepo}
		if err := writeJSONAtomic(filepath.Join(dir, "meta.json"), meta); err != nil {
			_ = os.RemoveAll(dir)
			return nil, err
		}
		if err := writeJSONAtomic(filepath.Join(dir, "summary.json"), emptySummary()); err != nil {
			_ = os.RemoveAll(dir)
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), nil, 0o600); err != nil {
			_ = os.RemoveAll(dir)
			return nil, codeError("E_SESSION_STORAGE", "cannot create events log", err)
		}
		if !options.IsGitRepo {
			if err := os.Mkdir(filepath.Join(dir, "snapshot"), 0o700); err != nil {
				_ = os.RemoveAll(dir)
				return nil, codeError("E_SESSION_STORAGE", "cannot create snapshot directory", err)
			}
		}
		return &Session{store: s, dir: dir, meta: meta}, nil
	}
	return nil, codeError("E_SESSION_ID", "could not allocate unique session id", nil)
}

func (s *Store) Resume(ref string) (*Session, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, codeError(ErrInvalidArgument.Code, "session reference is empty", nil)
	}
	if isULID(ref) {
		meta, err := readMeta(filepath.Join(s.root, ref, "meta.json"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, codeError(ErrSessionNotFound.Code, "session id not found", err)
			}
			return nil, err
		}
		if meta.ID != ref {
			return nil, codeError(ErrSessionCorrupt.Code, "metadata id does not match session directory", nil)
		}
		return s.openSession(meta)
	}

	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, codeError("E_SESSION_STORAGE", "cannot enumerate sessions", err)
	}
	var matches []Meta
	for _, entry := range entries {
		if !entry.IsDir() || !isULID(entry.Name()) {
			continue
		}
		meta, err := readMeta(filepath.Join(s.root, entry.Name(), "meta.json"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if meta.ID != entry.Name() {
			return nil, codeError(ErrSessionCorrupt.Code, "metadata id does not match session directory", nil)
		}
		if meta.Name != nil && *meta.Name == ref {
			matches = append(matches, meta)
		}
	}
	if len(matches) == 0 {
		return nil, codeError(ErrSessionNotFound.Code, "session name not found", nil)
	}
	if len(matches) > 1 {
		sort.Slice(matches, func(i, j int) bool { return matches[i].CreatedAt.Before(matches[j].CreatedAt) })
		details := make([]string, 0, len(matches))
		for _, match := range matches {
			details = append(details, fmt.Sprintf("%s@%s", match.ID, match.CreatedAt.UTC().Format(time.RFC3339)))
		}
		return nil, codeError(ErrSessionAmbiguous.Code, "session name matches multiple sessions: "+strings.Join(details, ", "), nil)
	}
	return s.openSession(matches[0])
}

func (s *Store) openSession(meta Meta) (*Session, error) {
	dir := filepath.Join(s.root, meta.ID)
	if _, err := os.Stat(filepath.Join(dir, "events.jsonl")); err != nil {
		return nil, codeError(ErrSessionCorrupt.Code, "events log is missing", err)
	}
	var summary Summary
	if err := readJSON(filepath.Join(dir, "summary.json"), &summary); err != nil {
		return nil, codeError(ErrSessionCorrupt.Code, "summary is missing or unreadable", err)
	}
	result, err := readEvents(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		return nil, err
	}
	return &Session{store: s, dir: dir, meta: meta, lastSeq: len(result.Events), trailingEvents: result.TrailingFragment}, nil
}

func (s *Store) CleanupExpired(now time.Time) (int, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return 0, codeError("E_SESSION_STORAGE", "cannot enumerate sessions", err)
	}
	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() || !isULID(entry.Name()) {
			continue
		}
		meta, err := readMeta(filepath.Join(s.root, entry.Name(), "meta.json"))
		if err != nil {
			continue
		}
		if meta.Name == nil && meta.ExitStatus == ExitStatusAborted && now.Sub(meta.UpdatedAt) >= abortedRetention {
			if err := os.RemoveAll(filepath.Join(s.root, entry.Name())); err != nil {
				return removed, codeError("E_SESSION_STORAGE", "cannot remove expired session", err)
			}
			removed++
		}
	}
	return removed, nil
}

func (s *Session) Metadata() Meta { return cloneMeta(s.meta) }

func (s *Session) Dir() string { return s.dir }

func (s *Session) SetName(name *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	if name != nil && strings.IndexByte(*name, 0) >= 0 {
		return codeError(ErrInvalidArgument.Code, "session name contains NUL", nil)
	}
	s.meta.Name = cloneStringPtr(name)
	s.meta.UpdatedAt = s.store.now().UTC()
	return s.persistMetaLocked()
}

func (s *Session) SaveMeta(meta Meta) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	if meta.ID != s.meta.ID || meta.WorkspaceRoot != s.meta.WorkspaceRoot || meta.CreatedAt != s.meta.CreatedAt {
		return ErrImmutableMetadata
	}
	meta.UpdatedAt = s.store.now().UTC()
	s.meta = cloneMeta(meta)
	return s.persistMetaLocked()
}

func (s *Session) SaveSummary(summary Summary) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	if summary.LastEventSeq < s.lastSeq {
		summary.LastEventSeq = s.lastSeq
	}
	return writeJSONAtomic(filepath.Join(s.dir, "summary.json"), maskSummary(summary))
}

func (s *Session) LoadSummary() (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var summary Summary
	if err := readJSON(filepath.Join(s.dir, "summary.json"), &summary); err != nil {
		return Summary{}, codeError(ErrSessionCorrupt.Code, "summary is unreadable", err)
	}
	return normalizeSummary(summary), nil
}

func (s *Session) AppendEvent(kind EventKind, payload any, tokens int, source string) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return Event{}, ErrSessionClosed
	}
	if s.trailingEvents {
		return Event{}, ErrEventLogTail
	}
	if kind == "" {
		return Event{}, codeError(ErrInvalidArgument.Code, "event kind is empty", nil)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Event{}, codeError(ErrInvalidArgument.Code, "event payload is not JSON serializable", err)
	}
	encoded = []byte(MaskSecrets(string(encoded)))
	event := Event{Seq: s.lastSeq + 1, Timestamp: s.store.now().UTC(), Kind: kind, Payload: encoded, Tokens: tokens, Source: source}
	line, err := json.Marshal(event)
	if err != nil {
		return Event{}, codeError(ErrInvalidArgument.Code, "event is not JSON serializable", err)
	}
	line = append(line, '\n')
	file, err := os.OpenFile(filepath.Join(s.dir, "events.jsonl"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return Event{}, codeError("E_SESSION_STORAGE", "cannot open events log", err)
	}
	_, writeErr := file.Write(line)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return Event{}, codeError("E_SESSION_STORAGE", "cannot append event", writeErr)
	}
	if closeErr != nil {
		return Event{}, codeError("E_SESSION_STORAGE", "cannot close events log", closeErr)
	}
	s.lastSeq++
	s.meta.UpdatedAt = s.store.now().UTC()
	if err := s.persistMetaLocked(); err != nil {
		return Event{}, err
	}
	return event, nil
}

func (s *Session) ReadEvents() (EventReadResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := readEvents(filepath.Join(s.dir, "events.jsonl"))
	if err != nil {
		return EventReadResult{}, err
	}
	s.lastSeq = len(result.Events)
	s.trailingEvents = result.TrailingFragment
	return result, nil
}

func (s *Session) Close(status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	if status != ExitStatusClean && status != ExitStatusAborted {
		return codeError(ErrInvalidArgument.Code, "invalid exit status", nil)
	}
	s.meta.ExitStatus = status
	s.meta.UpdatedAt = s.store.now().UTC()
	if err := s.persistMetaLocked(); err != nil {
		return err
	}
	if status == ExitStatusClean && s.meta.Name == nil {
		if err := os.RemoveAll(s.dir); err != nil {
			return codeError("E_SESSION_STORAGE", "cannot remove unnamed clean session", err)
		}
	}
	s.closed = true
	return nil
}

func (s *Session) persistMetaLocked() error {
	return writeJSONAtomic(filepath.Join(s.dir, "meta.json"), s.meta)
}

func readMeta(path string) (Meta, error) {
	var meta Meta
	if err := readJSON(path, &meta); err != nil {
		return Meta{}, codeError(ErrSessionCorrupt.Code, "metadata is unreadable", err)
	}
	if !isULID(meta.ID) || meta.ExitStatus == "" {
		return Meta{}, codeError(ErrSessionCorrupt.Code, "metadata is invalid", nil)
	}
	return meta, nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return codeError("E_SESSION_STORAGE", "cannot encode JSON", err)
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".brunel-session-*")
	if err != nil {
		return codeError("E_SESSION_STORAGE", "cannot create temporary JSON file", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return codeError("E_SESSION_STORAGE", "cannot write temporary JSON file", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return codeError("E_SESSION_STORAGE", "cannot sync temporary JSON file", err)
	}
	if err := tmp.Close(); err != nil {
		return codeError("E_SESSION_STORAGE", "cannot close temporary JSON file", err)
	}
	if err := replaceExistingFile(tmpName, path); err != nil {
		return codeError("E_SESSION_STORAGE", "cannot replace JSON file", err)
	}
	return nil
}

func readEvents(path string) (EventReadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EventReadResult{}, codeError(ErrSessionCorrupt.Code, "events log is unreadable", err)
	}
	var result EventReadResult
	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			result.TrailingFragment = len(data) > 0
			break
		}
		line := bytes.TrimSpace(data[:idx])
		data = data[idx+1:]
		if len(line) == 0 {
			return EventReadResult{}, codeError(ErrSessionCorrupt.Code, "events log contains an empty line", nil)
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			return EventReadResult{}, codeError(ErrSessionCorrupt.Code, "events log contains malformed JSON", err)
		}
		if event.Seq != len(result.Events)+1 {
			return EventReadResult{}, codeError(ErrSessionCorrupt.Code, "events log sequence is not contiguous", nil)
		}
		result.Events = append(result.Events, event)
	}
	return result, nil
}

func emptySummary() Summary {
	return Summary{Decisions: []string{}, ModifiedFiles: []string{}, Verifications: []Verification{}, OpenErrors: []string{}, Pending: []string{}}
}

func normalizeSummary(summary Summary) Summary {
	if summary.Decisions == nil {
		summary.Decisions = []string{}
	}
	if summary.ModifiedFiles == nil {
		summary.ModifiedFiles = []string{}
	}
	if summary.Verifications == nil {
		summary.Verifications = []Verification{}
	}
	if summary.OpenErrors == nil {
		summary.OpenErrors = []string{}
	}
	if summary.Pending == nil {
		summary.Pending = []string{}
	}
	return summary
}

func maskSummary(summary Summary) Summary {
	summary.Goal = MaskSecrets(summary.Goal)
	summary.LatestDiff = MaskSecrets(summary.LatestDiff)
	for i := range summary.Decisions {
		summary.Decisions[i] = MaskSecrets(summary.Decisions[i])
	}
	for i := range summary.ModifiedFiles {
		summary.ModifiedFiles[i] = MaskSecrets(summary.ModifiedFiles[i])
	}
	for i := range summary.OpenErrors {
		summary.OpenErrors[i] = MaskSecrets(summary.OpenErrors[i])
	}
	for i := range summary.Pending {
		summary.Pending[i] = MaskSecrets(summary.Pending[i])
	}
	for i := range summary.Verifications {
		summary.Verifications[i].Command = MaskSecrets(summary.Verifications[i].Command)
		summary.Verifications[i].Summary = MaskSecrets(summary.Verifications[i].Summary)
	}
	return normalizeSummary(summary)
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneMeta(meta Meta) Meta {
	meta.Name = cloneStringPtr(meta.Name)
	return meta
}
