package cache

import (
	"errors"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
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
func (s *Service) Get(in QueryInput) GetOutput {
	if err := validatePathArgs(in.Dir, in.Args); err != nil {
		return GetOutput{
			Hit:     false,
			Reason:  "sensitive_path",
			Error:   err.Error(),
			Message: err.Error(),
		}
	}
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return GetOutput{
			Hit:     false,
			Message: missingKeyOrParamsMsg,
		}
	}

	getRes, err := s.storage.Get(targetKey, in.Version)
	if err != nil {
		return GetOutput{
			Hit:     false,
			Key:     targetKey,
			Version: nil,
			Reason:  classifyStorageErr(err),
			Error:   err.Error(),
		}
	}

	return GetOutput{
		Hit:     getRes.Hit,
		Key:     getRes.Key,
		Version: getRes.Version,
		Data:    getRes.Data,
		Reason:  getRes.Reason,
	}
}

// Set constructs a CacheEntry with server-assigned timestamps and stores it.
func (s *Service) Set(in SetInput) SetOutput {
	resolvedKey := in.Key
	if resolvedKey == "" {
		resolvedKey = ComputeCacheKey(in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	}
	if in.TTL <= 0 {
		msg := "ttl must be > 0"
		return SetOutput{
			Success: false,
			Key:     &resolvedKey,
			Version: nil,
			Message: &msg,
		}
	}
	if err := validateSetSensitivity(in); err != nil {
		msg := err.Error()
		return SetOutput{
			Success: false,
			Key:     &resolvedKey,
			Version: nil,
			Message: &msg,
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
		msg := err.Error()
		return SetOutput{
			Success: false,
			Key:     &resolvedKey,
			Version: nil,
			Message: &msg,
		}
	}
	return SetOutput(stRes)
}

// Exist checks whether a non-expired cache entry exists.
func (s *Service) Exist(in QueryInput) ExistOutput {
	if err := validatePathArgs(in.Dir, in.Args); err != nil {
		return ExistOutput{
			Hit:     false,
			Reason:  "sensitive_path",
			Error:   err.Error(),
			Message: err.Error(),
		}
	}
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return ExistOutput{
			Hit:     false,
			Message: missingKeyOrParamsMsg,
		}
	}

	hit, err := s.storage.Exist(targetKey, in.Version)
	if err != nil {
		return ExistOutput{
			Hit:    false,
			Key:    targetKey,
			Reason: classifyStorageErr(err),
			Error:  err.Error(),
		}
	}
	return ExistOutput{
		Hit: hit,
		Key: targetKey,
	}
}

// Delete removes cache entries for a key (all versions or a specific version).
func (s *Service) Delete(in QueryInput) DeleteOutput {
	if err := validatePathArgs(in.Dir, in.Args); err != nil {
		return DeleteOutput{
			Hit:          false,
			Deleted:      false,
			DeletedCount: 0,
			Reason:       "sensitive_path",
			Error:        err.Error(),
			Message:      err.Error(),
		}
	}
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return DeleteOutput{
			Hit:     false,
			Deleted: false,
			Message: missingKeyOrParamsMsg,
		}
	}

	delRes, err := s.storage.Delete(targetKey, in.Version)
	if err != nil {
		return DeleteOutput{
			Hit:          false,
			Deleted:      false,
			Key:          targetKey,
			Version:      in.Version,
			DeletedCount: 0,
			Reason:       classifyStorageErr(err),
			Error:        err.Error(),
		}
	}
	return DeleteOutput{
		Hit:          delRes.Hit,
		Deleted:      delRes.Deleted,
		Key:          delRes.Key,
		Version:      delRes.Version,
		DeletedCount: delRes.DeletedCount,
	}
}

// Version lists all valid, non-expired versions for a key without reading payloads.
func (s *Service) Version(in VersionInput) VersionOutput {
	if err := validatePathArgs(in.Dir, in.Args); err != nil {
		return VersionOutput{
			Hit:      false,
			Versions: []VersionInfo{},
			Reason:   "sensitive_path",
			Error:    err.Error(),
			Message:  err.Error(),
		}
	}
	targetKey, ok := ResolveKey(in.Key, in.Type, in.Name, in.Dir, in.Args, in.Stdin)
	if !ok {
		return VersionOutput{
			Hit:      false,
			Versions: []VersionInfo{},
			Message:  missingKeyOrParamsMsg,
		}
	}

	verRes, err := s.storage.Version(targetKey)
	if err != nil {
		return VersionOutput{
			Hit:      false,
			Key:      targetKey,
			Versions: []VersionInfo{},
			Reason:   classifyStorageErr(err),
			Error:    err.Error(),
		}
	}
	vers := verRes.Versions
	if vers == nil {
		vers = []VersionInfo{}
	}
	return VersionOutput{
		Hit:      verRes.Hit,
		Key:      verRes.Key,
		Versions: vers,
		Reason:   verRes.Reason,
	}
}

// Purge deletes expired (or all) cache files across all shard directories.
func (s *Service) Purge(in PurgeInput) PurgeOutput {
	purgeRes, err := s.storage.Purge(in.All)
	if err != nil {
		msg := err.Error()
		return PurgeOutput{
			Success:      false,
			DeletedCount: 0,
			FreedBytes:   0,
			All:          in.All,
			Message:      &msg,
		}
	}
	return PurgeOutput(purgeRes)
}

func validatePathArgs(dir string, args []string) error {
	if err := guard.CheckPathAllowed(dir); err != nil {
		return err
	}
	for _, arg := range args {
		if err := guard.CheckPathAllowed(arg); err != nil {
			return err
		}
	}
	return nil
}

func validateSetSensitivity(in SetInput) error {
	if err := validatePathArgs(in.Dir, in.Args); err != nil {
		return err
	}
	texts := []string{in.Name, in.Stdout, in.Stderr}
	texts = append(texts, in.Args...)
	if in.Stdin != nil {
		texts = append(texts, *in.Stdin)
	}
	if in.Description != nil {
		texts = append(texts, *in.Description)
	}
	return guard.CheckNoSecrets(texts...)
}

func classifyStorageErr(err error) string {
	if errors.Is(err, store.ErrLockTimeout) {
		return "timeout"
	}
	return "system_error"
}
