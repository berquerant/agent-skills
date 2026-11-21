package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadJSONRoot reads and decodes a JSON file at relPath inside rootDir using os.OpenRoot
// to enforce directory containment.
func ReadJSONRoot[T any](rootDir, relPath string) (T, error) {
	var zero T
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return zero, err
	}
	defer func() { _ = root.Close() }()

	raw, err := root.ReadFile(relPath)
	if err != nil {
		return zero, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, fmt.Errorf("decode %s: %w", relPath, err)
	}
	return out, nil
}

// WriteJSONAtomic marshals v as indented JSON and writes it atomically to targetPath.
func WriteJSONAtomic[T any](targetPath string, v T, perm os.FileMode) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return WriteAtomic(targetPath, raw, perm)
}
