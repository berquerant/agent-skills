package taskdag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const lockTimeout = 10 * time.Second

// Storage manages persistent DAG snapshots and append-only journals under baseDir.
type Storage struct {
	baseDir string
	mu      sync.Mutex
	locks   map[string]*store.FileLock
}

// NewStorage creates a task-dag Storage rooted at baseDir.
func NewStorage(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, fmt.Errorf("create task-dag dir: %w", err)
	}
	return &Storage{
		baseDir: baseDir,
		locks:   make(map[string]*store.FileLock),
	}, nil
}

func (s *Storage) lockFor(dagID string) *store.FileLock {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.locks[dagID]; ok {
		return l
	}
	lockPath := filepath.Join(s.baseDir, dagID+".lock")
	l := store.NewFileLock(lockPath, lockTimeout)
	s.locks[dagID] = l
	return l
}

func validateDAGID(dagID string) error {
	if !safeIDRegex.MatchString(dagID) {
		return fmt.Errorf("invalid dag_id %q", dagID)
	}
	return nil
}

// Create stores a newly validated DAG and writes its initial snapshot, rejecting duplicate dag_id values.
func (s *Storage) Create(dag DAG) error {
	if err := validateDAGID(dag.DAGID); err != nil {
		return err
	}
	lock := s.lockFor(dag.DAGID)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()

	if _, err := s.readSnapshotUnlocked(dag.DAGID); err == nil {
		return fmt.Errorf("dag %q already exists", dag.DAGID)
	}
	return s.writeSnapshotUnlocked(dag)
}

// Load reads a DAG snapshot from disk under a shared file lock.
func (s *Storage) Load(dagID string) (DAG, error) {
	if err := validateDAGID(dagID); err != nil {
		return DAG{}, err
	}
	lock := s.lockFor(dagID)
	if err := lock.RLock(); err != nil {
		return DAG{}, err
	}
	defer func() { _ = lock.Unlock() }()

	return s.readSnapshotUnlocked(dagID)
}

// Update atomically loads, mutates, saves the DAG snapshot, and optionally appends a journal entry under an exclusive lock.
func (s *Storage) Update(dagID string, fn func(dag *DAG) (*JournalEntry, error)) (DAG, error) {
	if err := validateDAGID(dagID); err != nil {
		return DAG{}, err
	}
	lock := s.lockFor(dagID)
	if err := lock.Lock(); err != nil {
		return DAG{}, err
	}
	defer func() { _ = lock.Unlock() }()

	dag, err := s.readSnapshotUnlocked(dagID)
	if err != nil {
		return DAG{}, err
	}
	entry, err := fn(&dag)
	if err != nil {
		return DAG{}, err
	}
	if err := s.writeSnapshotUnlocked(dag); err != nil {
		return DAG{}, err
	}
	if entry != nil {
		_ = s.appendJournalUnlocked(dagID, *entry)
	}
	return dag, nil
}

// List returns summaries of all stored DAGs sorted by DAGID.
func (s *Storage) List() ([]DAGSummary, error) {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []DAGSummary{}, nil
		}
		return nil, err
	}
	defer func() { _ = root.Close() }()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	summaries := []DAGSummary{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), ".tmp_") {
			continue
		}
		raw, err := root.ReadFile(e.Name())
		if err != nil {
			continue
		}
		var d DAG
		if err := json.Unmarshal(raw, &d); err != nil {
			continue
		}
		counts := CountStates(d.Tasks)
		summaries = append(summaries, DAGSummary{
			DAGID:          d.DAGID,
			Name:           d.Name,
			TotalTasks:     len(d.Tasks),
			CompletedTasks: counts[StateCompleted],
			PipelineStatus: SummarizePipelineStatus(d.Tasks),
			UpdatedAt:      d.UpdatedAt,
		})
	}
	slices.SortFunc(summaries, func(a, b DAGSummary) int {
		return strings.Compare(a.DAGID, b.DAGID)
	})
	return summaries, nil
}

// Delete removes a DAG snapshot and its journal file under an exclusive lock.
func (s *Storage) Delete(dagID string) (bool, error) {
	if err := validateDAGID(dagID); err != nil {
		return false, err
	}
	lock := s.lockFor(dagID)
	if err := lock.Lock(); err != nil {
		return false, err
	}
	defer func() { _ = lock.Unlock() }()

	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return false, err
	}
	defer func() { _ = root.Close() }()

	err = root.Remove(dagID + ".json")
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_ = root.Remove(dagID + ".journal.jsonl")
	return true, nil
}

func (s *Storage) readSnapshotUnlocked(dagID string) (DAG, error) {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return DAG{}, err
	}
	defer func() { _ = root.Close() }()

	raw, err := root.ReadFile(dagID + ".json")
	if err != nil {
		if os.IsNotExist(err) {
			return DAG{}, fmt.Errorf("dag %q not found", dagID)
		}
		return DAG{}, err
	}
	var dag DAG
	if err := json.Unmarshal(raw, &dag); err != nil {
		return DAG{}, fmt.Errorf("decode dag %q: %w", dagID, err)
	}
	return dag, nil
}

func (s *Storage) writeSnapshotUnlocked(dag DAG) error {
	raw, err := json.MarshalIndent(dag, "", "  ")
	if err != nil {
		return fmt.Errorf("encode dag %q: %w", dag.DAGID, err)
	}
	target := filepath.Join(s.baseDir, dag.DAGID+".json")
	return store.WriteAtomic(target, raw, 0o600)
}

func (s *Storage) appendJournalUnlocked(dagID string, entry JournalEntry) error {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	f, err := root.OpenFile(dagID+".journal.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(line)
	return err
}
