package pathutil

import (
	"fmt"
	"path/filepath"
)

// FormatDualPath formats an absolute file path according to the output-presentation convention:
// [<basename>](file://<abs_path>) (<abs_path>).
func FormatDualPath(absPath string) string {
	slashPath := filepath.ToSlash(absPath)
	base := filepath.Base(slashPath)
	return fmt.Sprintf("[%s](file://%s) (%s)", base, slashPath, slashPath)
}
