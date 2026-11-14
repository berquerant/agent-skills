package snippet

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed python_ast.py
var pythonASTScript string

const pythonParseTimeout = 5 * time.Second

type pythonASTResponse struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error"`
	Metrics Metrics  `json:"metrics"`
	Hazards []Hazard `json:"hazards"`
}

func analyzePythonAST(code string) (Metrics, []Hazard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pythonParseTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-I", "-c", pythonASTScript)
	cmd.Stdin = strings.NewReader(code)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"PYTHONIOENCODING=utf-8",
		"LC_ALL=C.UTF-8",
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errDetail := strings.TrimSpace(stderr.String())
		if errDetail != "" {
			return Metrics{}, nil, fmt.Errorf("python3 ast runner failed: %w: %s", err, errDetail)
		}
		return Metrics{}, nil, fmt.Errorf("python3 ast runner failed: %w", err)
	}

	var resp pythonASTResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return Metrics{}, nil, fmt.Errorf("decode python ast output: %w", err)
	}
	if !resp.OK {
		return Metrics{}, nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Metrics, resp.Hazards, nil
}

func isPythonLanguage(lang string) bool {
	switch lang {
	case "python", "python3", "py":
		return true
	default:
		return false
	}
}

func isSupportedLanguage(lang string) bool {
	if isPythonLanguage(lang) {
		return true
	}
	switch lang {
	case "bash", "sh", "posix", "zsh", "mksh":
		return true
	default:
		return false
	}
}
