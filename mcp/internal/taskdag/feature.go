package taskdag

import (
	"context"
	"path/filepath"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "task-dag"

var toolNames = []string{
	"dag_create",
	"dag_next",
	"dag_transition",
	"dag_status",
	"dag_render",
	"dag_list",
	"dag_delete",
}

// Feature registers the task-dag MCP tools.
type Feature struct {
	stateDir string
	storage  *Storage
	clock    feature.Clock
}

// NewFeature creates a task-dag Feature.
func NewFeature() *Feature {
	return &Feature{}
}

// NewFeatureWithDir creates a task-dag Feature with a custom state directory.
func NewFeatureWithDir(stateDir string) *Feature {
	return &Feature{stateDir: stateDir}
}

// Name returns the canonical feature name ("task-dag").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for task-dag.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register initializes storage and adds the 7 task-dag tools to the MCP server.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	clock := env.Clock
	if clock == nil {
		clock = feature.SystemClock{}
	}
	baseDir := f.stateDir
	if baseDir == "" {
		baseDir = filepath.Join(env.StateDir, featureName)
	}
	st, err := NewStorage(baseDir)
	if err != nil {
		return err
	}
	f.storage = st
	f.clock = clock

	f.registerCreateAndNext(s, env)
	f.registerTransitionAndStatus(s, env)
	f.registerRenderListDelete(s, env)
	return nil
}

func (f *Feature) registerCreateAndNext(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_create",
			Description: "Validate and create a Directed Acyclic Graph (DAG) of tasks with explicit dependencies, resource scopes, and wait profiles.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in CreateInput) (CreateOutput, error) {
			return f.handleCreate(in)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_next",
			Description: "Select the highest-priority READY task using cooperative idle-wait scheduling (unblockers first, local over async, shorter estimate).",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in NextInput) (NextOutput, error) {
			return f.handleNext(in)
		},
	)
}

func (f *Feature) registerTransitionAndStatus(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_transition",
			Description: "Transition a task to a new state (RUNNING, WAITING, COMPLETED, HALTED, READY) with resource collision and anti-polling guards.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in TransitionInput) (TransitionOutput, error) {
			return f.handleTransition(in)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_status",
			Description: "Inspect all task states, counts, critical path, and independent branches for a DAG.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in StatusInput) (StatusOutput, error) {
			return f.handleStatus(in)
		},
	)
}

func (f *Feature) registerRenderListDelete(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_render",
			Description: "Render the task dependency graph and current task states as a Mermaid flowchart TD diagram.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in RenderInput) (RenderOutput, error) {
			dag, err := f.storage.Load(in.DAGID)
			if err != nil {
				return RenderOutput{}, err
			}
			return RenderOutput{OK: true, DAGID: dag.DAGID, Mermaid: RenderMermaid(&dag)}, nil
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_list",
			Description: "List all stored task dependency graphs and their completion progress.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, _ ListInput) (ListOutput, error) {
			dags, err := f.storage.List()
			if err != nil {
				return ListOutput{}, err
			}
			return ListOutput{OK: true, DAGs: dags}, nil
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "dag_delete",
			Description: "Delete a stored task dependency graph and its journal.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in DeleteInput) (DeleteOutput, error) {
			deleted, err := f.storage.Delete(in.DAGID)
			if err != nil {
				return DeleteOutput{}, err
			}
			return DeleteOutput{OK: true, DAGID: in.DAGID, Deleted: deleted}, nil
		},
	)
}

func (f *Feature) handleCreate(in CreateInput) (CreateOutput, error) {
	if err := guard.CheckNoSecrets(in.DAGID, in.Name); err != nil {
		return CreateOutput{}, err
	}
	for _, t := range in.Tasks {
		if err := guard.CheckPathAllowed(t.Resource); err != nil {
			return CreateOutput{}, err
		}
		if err := guard.CheckNoSecrets(t.ID, t.Title, t.Resource, t.Note); err != nil {
			return CreateOutput{}, err
		}
	}
	dagID := in.DAGID
	if dagID == "" {
		id, err := store.GenerateID()
		if err != nil {
			return CreateOutput{}, err
		}
		dagID = "dag-" + id[:12]
	}
	nowStr := f.clock.Now().UTC().Format(time.RFC3339)
	tasks, err := ValidateAndInitTasks(in.Tasks, nowStr)
	if err != nil {
		return CreateOutput{}, err
	}

	dag := DAG{
		DAGID:     dagID,
		Name:      in.Name,
		CreatedAt: nowStr,
		UpdatedAt: nowStr,
		Tasks:     tasks,
	}
	if err := f.storage.Create(dag); err != nil {
		return CreateOutput{}, err
	}

	ready := []string{}
	blocked := []string{}
	for _, t := range tasks {
		if t.State == StateReady {
			ready = append(ready, t.ID)
		} else {
			blocked = append(blocked, t.ID)
		}
	}
	return CreateOutput{
		OK:                 true,
		DAGID:              dagID,
		Name:               in.Name,
		TotalTasks:         len(tasks),
		ReadyTasks:         ready,
		BlockedTasks:       blocked,
		CriticalPath:       ComputeCriticalPath(tasks),
		IndependentStreams: len(ComputeIndependentBranches(tasks)),
	}, nil
}

func (f *Feature) handleNext(in NextInput) (NextOutput, error) {
	var out NextOutput
	_, err := f.storage.Update(in.DAGID, func(dag *DAG) (*JournalEntry, error) {
		var mutated bool
		out, mutated = SelectNext(dag, f.clock.Now())
		if !mutated || out.Task == nil {
			return nil, nil
		}
		return &JournalEntry{
			Timestamp: f.clock.Now().UTC().Format(time.RFC3339),
			TaskID:    out.Task.ID,
			From:      StateWaiting,
			To:        StateReady,
			Note:      "auto-awakened after poll interval elapsed",
		}, nil
	})
	return out, err
}

func (f *Feature) handleTransition(in TransitionInput) (TransitionOutput, error) {
	var out TransitionOutput
	_, err := f.storage.Update(in.DAGID, func(dag *DAG) (*JournalEntry, error) {
		res, err := ApplyTransition(dag, in, f.clock.Now())
		if err != nil {
			return nil, err
		}
		out = TransitionOutput{
			OK:                true,
			DAGID:             dag.DAGID,
			Task:              res.Task,
			PromotedTasks:     res.PromotedTasks,
			HaltedDescendants: res.HaltedDescendants,
			PipelineStatus:    SummarizePipelineStatus(dag.Tasks),
		}
		return &res.Journal, nil
	})
	return out, err
}

func (f *Feature) handleStatus(in StatusInput) (StatusOutput, error) {
	dag, err := f.storage.Load(in.DAGID)
	if err != nil {
		return StatusOutput{}, err
	}
	branches := ComputeIndependentBranches(dag.Tasks)
	if branches == nil {
		branches = [][]string{}
	}
	crit := ComputeCriticalPath(dag.Tasks)
	if crit == nil {
		crit = []string{}
	}
	return StatusOutput{
		OK:                  true,
		DAG:                 dag,
		Counts:              CountStates(dag.Tasks),
		CriticalPath:        crit,
		IndependentBranches: branches,
		PipelineStatus:      SummarizePipelineStatus(dag.Tasks),
	}, nil
}

// Close is a no-op because locks are per-operation.
func (f *Feature) Close() error {
	return nil
}
