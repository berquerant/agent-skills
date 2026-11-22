package taskdag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const lockTimeout = 10 * time.Second

// Storage manages persistent DAG snapshots and append-only journals under baseDir.
type Storage struct {
	baseDir string
	locks   *store.LockPool
}

// NewStorage creates a task-dag Storage rooted at baseDir.
func NewStorage(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, fmt.Errorf("create task-dag dir: %w", err)
	}
	return &Storage{
		baseDir: baseDir,
		locks:   store.NewLockPool(baseDir, lockTimeout),
	}, nil
}

func validateDAGID(dagID string) error {
	if !safeIDRegex.MatchString(dagID) {
		return fmt.Errorf("invalid dag_id %q", dagID)
	}
	return nil
}

func snapshotFile(dagID string) string { return dagID + ".json" }
func journalFile(dagID string) string  { return dagID + ".journal.jsonl" }

func (s *Storage) withDAGLock(dagID string, fn func() error) error {
	if err := validateDAGID(dagID); err != nil {
		return err
	}
	return s.locks.WithLock(dagID, fn)
}

func (s *Storage) withDAGRLock(dagID string, fn func() error) error {
	if err := validateDAGID(dagID); err != nil {
		return err
	}
	return s.locks.WithRLock(dagID, fn)
}

// Create stores a newly validated DAG and writes its initial snapshot, rejecting duplicate dag_id values.
func (s *Storage) Create(dag DAG) error {
	return s.withDAGLock(dag.DAGID, func() error {
		if _, err := s.readSnapshotUnlocked(dag.DAGID); err == nil {
			return fmt.Errorf("dag %q already exists", dag.DAGID)
		}
		return s.writeSnapshotUnlocked(dag)
	})
}

// Load reads a DAG snapshot from disk under a shared file lock.
func (s *Storage) Load(dagID string) (DAG, error) {
	var dag DAG
	err := s.withDAGRLock(dagID, func() error {
		var readErr error
		dag, readErr = s.readSnapshotUnlocked(dagID)
		return readErr
	})
	return dag, err
}

// Update atomically loads, mutates, saves the DAG snapshot, and optionally appends a journal entry under an exclusive lock.
func (s *Storage) Update(dagID string, fn func(dag *DAG) (*JournalEntry, error)) (DAG, error) {
	var dag DAG
	err := s.withDAGLock(dagID, func() error {
		var readErr error
		dag, readErr = s.readSnapshotUnlocked(dagID)
		if readErr != nil {
			return readErr
		}
		entry, fnErr := fn(&dag)
		if fnErr != nil {
			return fnErr
		}
		if writeErr := s.writeSnapshotUnlocked(dag); writeErr != nil {
			return writeErr
		}
		if entry != nil {
			if jErr := s.appendJournalUnlocked(dagID, *entry); jErr != nil {
				return jErr
			}
		}
		return nil
	})
	if err != nil {
		return DAG{}, err
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
	var deleted bool
	err := s.withDAGLock(dagID, func() error {
		root, err := os.OpenRoot(s.baseDir)
		if err != nil {
			return err
		}
		defer func() { _ = root.Close() }()

		err = root.Remove(snapshotFile(dagID))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		_ = root.Remove(journalFile(dagID))
		deleted = true
		return nil
	})
	return deleted, err
}

func (s *Storage) readSnapshotUnlocked(dagID string) (DAG, error) {
	dag, err := store.ReadJSONRoot[DAG](s.baseDir, snapshotFile(dagID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DAG{}, fmt.Errorf("dag %q not found", dagID)
		}
		return DAG{}, fmt.Errorf("load dag %q: %w", dagID, err)
	}
	return dag, nil
}

func (s *Storage) writeSnapshotUnlocked(dag DAG) error {
	target := filepath.Join(s.baseDir, snapshotFile(dag.DAGID))
	if err := store.WriteJSONAtomic(target, dag, 0o600); err != nil {
		return fmt.Errorf("write dag %q: %w", dag.DAGID, err)
	}
	return nil
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
	f, err := root.OpenFile(journalFile(dagID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(line)
	return err
}
