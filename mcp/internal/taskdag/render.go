package taskdag

import (
	"fmt"
	"strings"
)

// RenderMermaid generates a valid Mermaid flowchart TD diagram representing the DAG and task states.
func RenderMermaid(dag *DAG) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")

	for idx, t := range dag.Tasks {
		nodeID := fmt.Sprintf("N%d", idx+1)
		label := t.ID
		if t.Title != "" {
			label = fmt.Sprintf("%s: %s", t.ID, sanitizeMermaidLabel(t.Title))
		}
		_, _ = fmt.Fprintf(&b, "    %s[\"%s [%s]\"]:::%s\n", nodeID, label, t.State, strings.ToLower(string(t.State)))
	}

	idxByID := make(map[string]int, len(dag.Tasks))
	for idx, t := range dag.Tasks {
		idxByID[t.ID] = idx + 1
	}

	for idx, t := range dag.Tasks {
		toNode := fmt.Sprintf("N%d", idx+1)
		for _, dep := range t.DependsOn {
			if fromIdx, ok := idxByID[dep]; ok {
				_, _ = fmt.Fprintf(&b, "    N%d --> %s\n", fromIdx, toNode)
			}
		}
	}

	b.WriteString("    classDef completed fill:#d4edda,stroke:#28a745,color:#155724\n")
	b.WriteString("    classDef running fill:#cce5ff,stroke:#007bff,color:#004085\n")
	b.WriteString("    classDef ready fill:#fff3cd,stroke:#ffc107,color:#856404\n")
	b.WriteString("    classDef waiting fill:#e2e3e5,stroke:#6c757d,color:#383d41\n")
	b.WriteString("    classDef blocked fill:#f8f9fa,stroke:#adb5bd,color:#6c757d\n")
	b.WriteString("    classDef halted fill:#f8d7da,stroke:#dc3545,color:#721c24\n")
	return b.String()
}

func sanitizeMermaidLabel(s string) string {
	r := strings.NewReplacer(`"`, `'`, `\`, `/`, "\n", " ", "\r", " ")
	return strings.TrimSpace(r.Replace(s))
}
