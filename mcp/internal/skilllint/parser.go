package skilllint

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	inlineCodeRegex = regexp.MustCompile("`[^`]*`")
	mdLinkRegex     = regexp.MustCompile(`\]\(([^)]*)\)`)
)

type frontmatterData struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func readFileInRoot(rootDir, relPath string) ([]byte, error) {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	return root.ReadFile(relPath)
}

func countNewlines(b []byte) int {
	return bytes.Count(b, []byte{'\n'})
}

func splitLines(content string) []string {
	trimmed := strings.TrimSuffix(content, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func isFenceDelimiter(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "```")
}

func isShellFenceStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "```sh" || trimmed == "```bash" || trimmed == "```zsh"
}

func stripInlineCode(line string) string {
	return inlineCodeRegex.ReplaceAllString(line, "")
}

func extractFrontmatterLines(lines []string) []string {
	delimCount := 0
	var fm []string
	for _, line := range lines {
		if strings.TrimRight(line, " \t\r") == "---" {
			delimCount++
			if delimCount == 2 {
				break
			}
			continue
		}
		if delimCount == 1 {
			fm = append(fm, line)
		}
	}
	return fm
}

func parseFrontmatter(lines []string) (frontmatterData, error) {
	fmLines := extractFrontmatterLines(lines)
	if len(fmLines) == 0 {
		return frontmatterData{}, nil
	}
	var data frontmatterData
	if err := yaml.Unmarshal([]byte(strings.Join(fmLines, "\n")), &data); err != nil {
		return frontmatterData{}, err
	}
	data.Name = strings.TrimSpace(data.Name)
	data.Description = strings.TrimSpace(strings.Join(strings.Fields(data.Description), " "))
	return data, nil
}

func extractRelativeLinkTargets(lines []string) []string {
	inCode := false
	var targets []string
	for _, line := range lines {
		if isFenceDelimiter(line) {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		cleaned := stripInlineCode(line)
		matches := mdLinkRegex.FindAllStringSubmatch(cleaned, -1)
		for _, m := range matches {
			target := cleanLinkTarget(m[1])
			if target != "" {
				targets = append(targets, target)
			}
		}
	}
	return targets
}

func cleanLinkTarget(raw string) string {
	target := raw
	if idx := strings.IndexByte(target, '#'); idx >= 0 {
		target = target[:idx]
	}
	if fields := strings.Fields(target); len(fields) > 0 {
		target = fields[0]
	} else {
		target = ""
	}
	if target == "" ||
		strings.HasPrefix(target, "http:") ||
		strings.HasPrefix(target, "https:") ||
		strings.HasPrefix(target, "mailto:") {
		return ""
	}
	return target
}

func resolveRepoRoot(skillDirAbs string) string {
	curr := filepath.Clean(skillDirAbs)
	for {
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	parent2 := filepath.Clean(filepath.Join(skillDirAbs, "..", ".."))
	return parent2
}
