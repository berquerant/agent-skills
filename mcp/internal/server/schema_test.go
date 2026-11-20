// Package server_test contains integration and schema compliance tests for the MCP server.
package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/server"
	"github.com/berquerant/agent-skills/mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findUntypedProperties(path string, val any) []string {
	var issues []string
	switch v := val.(type) {
	case bool:
		issues = append(issues, fmt.Sprintf("%s is boolean schema (%v)", path, v))
	case map[string]any:
		if len(v) == 0 {
			issues = append(issues, fmt.Sprintf("%s is empty schema {}", path))
		}
		if props, ok := v["properties"].(map[string]any); ok {
			for pName, pVal := range props {
				subPath := path + "." + pName
				if b, isBool := pVal.(bool); isBool {
					issues = append(issues, fmt.Sprintf("%s is boolean schema (%v)", subPath, b))
				} else if pMap, isMap := pVal.(map[string]any); isMap {
					_, hasType := pMap["type"]
					_, hasAnyOf := pMap["anyOf"]
					_, hasOneOf := pMap["oneOf"]
					_, hasAllOf := pMap["allOf"]
					_, hasRef := pMap["$ref"]
					if !hasType && !hasAnyOf && !hasOneOf && !hasAllOf && !hasRef {
						issues = append(issues, fmt.Sprintf("%s has no type/anyOf/oneOf/allOf/$ref: %v", subPath, pMap))
					}
					issues = append(issues, findUntypedProperties(subPath, pVal)...)
				}
			}
		}
		if items, hasItems := v["items"]; hasItems {
			issues = append(issues, findUntypedProperties(path+"[]", items)...)
		}
	}
	return issues
}

// TestToolSchemas verifies that every registered MCP tool provides unambiguous input
// and output JSON schemas without untyped, empty, or boolean wildcard (any) properties.
func TestToolSchemas(t *testing.T) {
	ctx := t.Context()
	tmp := t.TempDir()

	args := []string{
		"--state-dir=" + tmp + "/state",
		"--cache-dir=" + tmp + "/cache",
		"--checkpoint-root=" + tmp + "/checkpoints",
	}
	cfg, err := feature.ParseConfig(args, tmp)
	require.NoError(t, err)

	var stderrBuf bytes.Buffer
	inst, err := server.New(cfg, nil, nil, &stderrBuf)
	require.NoError(t, err)
	defer func() { _ = inst.Close() }()

	session := testutil.ConnectInMemory(t, inst.MCPServer)

	listRes, err := session.ListTools(ctx, nil)
	require.NoError(t, err)

	var allIssues []string
	for _, tool := range listRes.Tools {
		if tool.InputSchema != nil {
			var inMap any
			inJSON, _ := json.Marshal(tool.InputSchema)
			_ = json.Unmarshal(inJSON, &inMap)
			issues := findUntypedProperties(tool.Name+"(IN)", inMap)
			allIssues = append(allIssues, issues...)
		}
		if tool.OutputSchema != nil {
			var outMap any
			outJSON, _ := json.Marshal(tool.OutputSchema)
			_ = json.Unmarshal(outJSON, &outMap)
			issues := findUntypedProperties(tool.Name+"(OUT)", outMap)
			allIssues = append(allIssues, issues...)
		}
	}

	for _, issue := range allIssues {
		t.Logf("SCHEMA ISSUE: %s", issue)
	}
	assert.Empty(t, allIssues, "all registered MCP tools must have strictly-typed schemas without ambiguous/any properties")
}
