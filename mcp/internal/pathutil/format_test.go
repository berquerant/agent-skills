package pathutil_test

import (
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/stretchr/testify/assert"
)

func TestFormatDualPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unix path",
			input:    "/path/to/doc.md",
			expected: "[doc.md](file:///path/to/doc.md) (/path/to/doc.md)",
		},
		{
			name:     "root file",
			input:    "/file.txt",
			expected: "[file.txt](file:///file.txt) (/file.txt)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pathutil.FormatDualPath(tt.input))
		})
	}
}
