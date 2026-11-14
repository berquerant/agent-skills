package cache

import (
	"errors"

	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const missingKeyOrParamsMsg = "Either 'key' or ('type', 'name', 'dir') must be provided"

// Service coordinates cache storage operations and response formatting.
type Service struct {
	config  Config
	storage *Storage
}

// NewService creates a Service with storage.
func NewService(cfg Config) (*Service, error) {
	st, err := NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{
		config:  cfg,
		storage: st,
	}, nil
}

// Close releases storage resources.
func (s *Service) Close() error {
	if s.storage != nil {
		return s.storage.Close()
	}
	return nil
}

// QueryInput holds parameters for get, exist, and delete operations.
type QueryInput struct {
	Key     string   `json:"key,omitempty" jsonschema:"Cache key (optionally including @<version>)"`
	Version any      `json:"version,omitempty" jsonschema:"Target version (integer >= 1 or 'latest')"`
	Type    string   `json:"type,omitempty" jsonschema:"Invocation type ('command', 'mcp', etc.)"`
	Name    string   `json:"name,omitempty" jsonschema:"Command or tool name"`
	Args    []string `json:"args,omitempty" jsonschema:"Ordered list of command or tool arguments"`
	Dir     string   `json:"dir,omitempty" jsonschema:"Working directory of the invocation"`
	Stdin   *string  `json:"stdin,omitempty" jsonschema:"Standard input string if provided"`
}

// VersionInput holds parameters for the version operation.
type VersionInput struct {
	Key   string   `json:"key,omitempty" jsonschema:"Cache key (optionally including @<version>)"`
	Type  string   `json:"type,omitempty" jsonschema:"Invocation type ('command', 'mcp', etc.)"`
	Name  string   `json:"name,omitempty" jsonschema:"Command or tool name"`
	Args  []string `json:"args,omitempty" jsonschema:"Ordered list of command or tool arguments"`
	Dir   string   `json:"dir,omitempty" jsonschema:"Working directory of the invocation"`
	Stdin *string  `json:"stdin,omitempty" jsonschema:"Standard input string if provided"`
}

// SetInput holds parameters for the set operation.
type SetInput struct {
	Type           string   `json:"type" jsonschema:"Invocation type ('command', 'mcp', etc.)"`
	Name           string   `json:"name" jsonschema:"Command or tool name"`
	Dir            string   `json:"dir" jsonschema:"Working directory of the invocation"`
	TTL            int      `json:"ttl" jsonschema:"Requested TTL in seconds (clamped to max_ttl)"`
	Stdout         string   `json:"stdout" jsonschema:"Standard output content"`
	Stderr         string   `json:"stderr" jsonschema:"Standard error content"`
	Success        bool     `json:"success" jsonschema:"Whether the invocation succeeded"`
	StartTimestamp *int64   `json:"start_timestamp,omitempty" jsonschema:"Optional start unix timestamp"`
	EndTimestamp   *int64   `json:"end_timestamp,omitempty" jsonschema:"Optional end unix timestamp"`
	Args           []string `json:"args,omitempty" jsonschema:"Ordered list of command or tool arguments"`
	Description    *string  `json:"description,omitempty" jsonschema:"Optional human-readable description"`
	Stdin          *string  `json:"stdin,omitempty" jsonschema:"Optional standard input"`
	StatusCode     *int     `json:"status_code,omitempty" jsonschema:"Optional process exit status code"`
	StatusString   *string  `json:"status_string,omitempty" jsonschema:"Optional status string"`
	Key            string   `json:"key,omitempty" jsonschema:"Optional explicit cache key"`
}

// PurgeInput holds parameters for the purge operation.
type PurgeInput struct {
	All bool `json:"all,omitempty" jsonschema:"When true, purges all cache files; when false, purges only expired files"`
}

// Get retrieves a cached entry.
func (s *Service) Get(in QueryInput) map[string]any {
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return map[string]any{
			"hit":     false,
			"message": missingKeyOrParamsMsg,
		}
	}

	getRes, err := s.storage.Get(targetKey, in.Version)
	if err != nil {
		return map[string]any{
			"hit":     false,
			"key":     targetKey,
			"version": nil,
			"reason":  classifyStorageErr(err),
			"error":   err.Error(),
		}
	}

	result := map[string]any{
		"hit":     getRes.Hit,
		"key":     getRes.Key,
		"version": getRes.Version,
		"data":    getRes.Data,
	}
	if getRes.Reason != "" {
		result["reason"] = getRes.Reason
	}
	return result
}

// Set constructs a CacheEntry with server-assigned timestamps and stores it.
func (s *Service) Set(in SetInput) map[string]any {
	resolvedKey := in.Key
	if resolvedKey == "" {
		resolvedKey = ComputeCacheKey(in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	}
	if in.TTL <= 0 {
		return map[string]any{
			"success": false,
			"key":     resolvedKey,
			"version": nil,
			"message": "ttl must be > 0",
		}
	}

	now := s.config.nowUnix()
	effectiveTTL := min(in.TTL, s.config.MaxTTL)
	ttlTS := now + int64(effectiveTTL)

	args := in.Args
	if args == nil {
		args = []string{}
	}
	entry := CacheEntry{
		Type:             in.Type,
		Name:             in.Name,
		Args:             args,
		Dir:              in.Dir,
		StartTimestamp:   in.StartTimestamp,
		EndTimestamp:     in.EndTimestamp,
		CreatedTimestamp: now,
		TTL:              effectiveTTL,
		TTLTimestamp:     ttlTS,
		Description:      in.Description,
		Stdin:            in.Stdin,
		Stdout:           in.Stdout,
		Stderr:           in.Stderr,
		Success:          in.Success,
		StatusCode:       in.StatusCode,
		StatusString:     in.StatusString,
	}

	stRes, err := s.storage.Set(resolvedKey, entry)
	if err != nil {
		return map[string]any{
			"success": false,
			"key":     resolvedKey,
			"version": nil,
			"message": err.Error(),
		}
	}
	return map[string]any{
		"success": stRes.Success,
		"key":     stRes.Key,
		"version": stRes.Version,
		"message": stRes.Message,
	}
}

// Exist checks whether a non-expired cache entry exists.
func (s *Service) Exist(in QueryInput) map[string]any {
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return map[string]any{
			"hit":     false,
			"message": missingKeyOrParamsMsg,
		}
	}

	hit, err := s.storage.Exist(targetKey, in.Version)
	if err != nil {
		return map[string]any{
			"hit":    false,
			"key":    targetKey,
			"reason": classifyStorageErr(err),
			"error":  err.Error(),
		}
	}
	return map[string]any{
		"hit": hit,
		"key": targetKey,
	}
}

// Delete removes cache entries for a key (all versions or a specific version).
func (s *Service) Delete(in QueryInput) map[string]any {
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return map[string]any{
			"hit":     false,
			"deleted": false,
			"message": missingKeyOrParamsMsg,
		}
	}

	delRes, err := s.storage.Delete(targetKey, in.Version)
	if err != nil {
		return map[string]any{
			"hit":           false,
			"deleted":       false,
			"key":           targetKey,
			"version":       in.Version,
			"deleted_count": 0,
			"reason":        classifyStorageErr(err),
			"error":         err.Error(),
		}
	}
	return map[string]any{
		"hit":           delRes.Hit,
		"deleted":       delRes.Deleted,
		"key":           delRes.Key,
		"version":       delRes.Version,
		"deleted_count": delRes.DeletedCount,
	}
}

// Version lists all valid, non-expired versions for a key without reading payloads.
func (s *Service) Version(in VersionInput) map[string]any {
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return map[string]any{
			"hit":      false,
			"key":      nil,
			"versions": []VersionInfo{},
			"message":  missingKeyOrParamsMsg,
		}
	}

	verRes, err := s.storage.Version(targetKey)
	if err != nil {
		return map[string]any{
			"hit":      false,
			"key":      targetKey,
			"versions": []VersionInfo{},
			"reason":   classifyStorageErr(err),
			"error":    err.Error(),
		}
	}
	result := map[string]any{
		"hit":      verRes.Hit,
		"key":      verRes.Key,
		"versions": verRes.Versions,
	}
	if verRes.Reason != "" {
		result["reason"] = verRes.Reason
	}
	return result
}

// Purge deletes expired (or all) cache files across all shard directories.
func (s *Service) Purge(in PurgeInput) map[string]any {
	purgeRes, err := s.storage.Purge(in.All)
	if err != nil {
		return map[string]any{
			"success":       false,
			"deleted_count": 0,
			"freed_bytes":   0,
			"all":           in.All,
			"message":       err.Error(),
		}
	}
	return map[string]any{
		"success":       purgeRes.Success,
		"deleted_count": purgeRes.DeletedCount,
		"freed_bytes":   purgeRes.FreedBytes,
		"all":           purgeRes.All,
		"message":       purgeRes.Message,
	}
}

func classifyStorageErr(err error) string {
	if errors.Is(err, store.ErrLockTimeout) {
		return "timeout"
	}
	return "system_error"
}
