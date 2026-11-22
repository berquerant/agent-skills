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
	"sync"
	"time"
)

//go:embed python_ast.py
var pythonASTScript string

const pythonParseTimeout = 5 * time.Second

var resolvePythonBin = sync.OnceValue(func() string {
	if _, err := exec.LookPath("mise"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if out, err := exec.CommandContext(ctx, "mise", "which", "python3").Output(); err == nil {
			if p := strings.TrimSpace(string(out)); p != "" {
				return p
			}
		}
	}
	return "python3"
})

type pythonASTResponse struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error"`
	Metrics Metrics  `json:"metrics"`
	Hazards []Hazard `json:"hazards"`
}

func analyzePythonAST(code string) (analysisResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pythonParseTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, resolvePythonBin(), "-I", "-c", pythonASTScript)
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
			return analysisResult{}, fmt.Errorf("python3 ast runner failed: %w: %s", err, errDetail)
		}
		return analysisResult{}, fmt.Errorf("python3 ast runner failed: %w", err)
	}

	var resp pythonASTResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return analysisResult{}, fmt.Errorf("decode python ast output: %w", err)
	}
	if !resp.OK {
		return analysisResult{}, fmt.Errorf("%s", resp.Error)
	}

	return analysisResult{
		metrics: resp.Metrics,
		hazards: resp.Hazards,
	}, nil
}

func isPythonLanguage(lang string) bool {
	return lang == "python"
}
