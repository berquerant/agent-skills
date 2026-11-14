package prtriage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

var makefileTargetRe = regexp.MustCompile(`(?m)^([a-zA-Z0-9][a-zA-Z0-9_-]*):(?:\s|$)`)
var justfileRecipeRe = regexp.MustCompile(`(?m)^([a-zA-Z0-9][a-zA-Z0-9_-]*):(?:\s|$)`)
var miseTaskSectionRe = regexp.MustCompile(`(?m)^\[tasks\.([a-zA-Z0-9:_-]+)\]`)

// DetectTasks scans repoDir for standard task runner definitions and CI workflows.
func DetectTasks(repoDir string) (DetectTasksOutput, error) {
	if err := guard.CheckPathAllowed(repoDir); err != nil {
		return DetectTasksOutput{}, err
	}
	absDir, err := filepath.Abs(repoDir)
	if err != nil {
		return DetectTasksOutput{}, fmt.Errorf("resolve repo dir: %w", err)
	}

	root, err := os.OpenRoot(absDir)
	if err != nil {
		return DetectTasksOutput{}, fmt.Errorf("open repo dir: %w", err)
	}
	defer func() { _ = root.Close() }()

	var tasks []TaskCandidate
	var runners []string

	scanTaskRunners(root, &runners, &tasks)
	scanWorkflowConfigs(root, &runners, &tasks)

	hasTests := slices.ContainsFunc(tasks, func(t TaskCandidate) bool { return t.Category == "test" })
	hasLinters := slices.ContainsFunc(tasks, func(t TaskCandidate) bool { return t.Category == "lint" })

	return DetectTasksOutput{
		OK:         true,
		RepoDir:    absDir,
		Runners:    runners,
		Tasks:      tasks,
		HasTests:   hasTests,
		HasLinters: hasLinters,
	}, nil
}

func scanTaskRunners(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	inspectPackageJSON(root, runners, tasks)
	inspectMiseToml(root, runners, tasks)
	inspectMakefile(root, runners, tasks)
	inspectJustfile(root, runners, tasks)
	inspectGoMod(root, runners, tasks)
	inspectCargoToml(root, runners, tasks)
}

func inspectPackageJSON(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	raw, err := root.ReadFile("package.json")
	if err != nil {
		return
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil || len(pkg.Scripts) == 0 {
		return
	}
	*runners = append(*runners, "npm/package.json")
	scripts := make([]string, 0, len(pkg.Scripts))
	for script := range pkg.Scripts {
		scripts = append(scripts, script)
	}
	slices.Sort(scripts)
	for _, script := range scripts {
		cat := categorizeTaskName(script)
		if cat != "" {
			*tasks = append(*tasks, TaskCandidate{
				Category: cat,
				Source:   "package.json",
				Command:  "npm run " + script,
			})
		}
	}
}

func inspectMiseToml(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	raw, err := root.ReadFile(".mise.toml")
	if err != nil {
		raw, err = root.ReadFile("mise.toml")
		if err != nil {
			return
		}
	}
	*runners = append(*runners, "mise")
	matches := miseTaskSectionRe.FindAllStringSubmatch(string(raw), -1)
	for _, m := range matches {
		taskName := m[1]
		cat := categorizeTaskName(taskName)
		if cat != "" {
			*tasks = append(*tasks, TaskCandidate{
				Category: cat,
				Source:   "mise.toml",
				Command:  "mise run " + taskName,
			})
		}
	}
}

func inspectMakefile(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	raw, err := root.ReadFile("Makefile")
	if err != nil {
		return
	}
	*runners = append(*runners, "make")
	matches := makefileTargetRe.FindAllStringSubmatch(string(raw), -1)
	seen := make(map[string]bool)
	for _, m := range matches {
		target := m[1]
		if seen[target] {
			continue
		}
		seen[target] = true
		cat := categorizeTaskName(target)
		if cat != "" {
			*tasks = append(*tasks, TaskCandidate{
				Category: cat,
				Source:   "Makefile",
				Command:  "make " + target,
			})
		}
	}
}

func inspectJustfile(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	raw, err := root.ReadFile("justfile")
	if err != nil {
		return
	}
	*runners = append(*runners, "just")
	matches := justfileRecipeRe.FindAllStringSubmatch(string(raw), -1)
	seen := make(map[string]bool)
	for _, m := range matches {
		recipe := m[1]
		if seen[recipe] {
			continue
		}
		seen[recipe] = true
		cat := categorizeTaskName(recipe)
		if cat != "" {
			*tasks = append(*tasks, TaskCandidate{
				Category: cat,
				Source:   "justfile",
				Command:  "just " + recipe,
			})
		}
	}
}

func inspectGoMod(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	if _, err := root.Stat("go.mod"); err == nil {
		*runners = append(*runners, "go")
		*tasks = append(*tasks,
			TaskCandidate{Category: "test", Source: "go.mod", Command: "go test ./..."},
			TaskCandidate{Category: "lint", Source: "go.mod", Command: "go vet ./..."},
		)
	}
}

func inspectCargoToml(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	if _, err := root.Stat("Cargo.toml"); err == nil {
		*runners = append(*runners, "cargo")
		*tasks = append(*tasks,
			TaskCandidate{Category: "test", Source: "Cargo.toml", Command: "cargo test"},
			TaskCandidate{Category: "lint", Source: "Cargo.toml", Command: "cargo clippy"},
		)
	}
}

func scanWorkflowConfigs(root *os.Root, runners *[]string, tasks *[]TaskCandidate) {
	if _, err := root.Stat(".github/workflows"); err == nil {
		*runners = append(*runners, "github-actions")
	}
	if _, err := root.Stat(".gitlab-ci.yml"); err == nil {
		*runners = append(*runners, "gitlab-ci")
	}
}

func categorizeTaskName(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "test"):
		return "test"
	case strings.Contains(n, "lint") || strings.Contains(n, "clippy") || strings.Contains(n, "vet"):
		return "lint"
	case strings.Contains(n, "check") || strings.Contains(n, "typecheck"):
		return "check"
	case strings.Contains(n, "build") || strings.Contains(n, "compile"):
		return "build"
	default:
		return ""
	}
}
