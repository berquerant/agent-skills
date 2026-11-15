package cache

import (
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
)

// Config holds the configuration for the cache feature.
type Config struct {
	Dir                         string
	MaxTTL                      int
	MaxCacheSize                int64
	MaxTotalCompressedCacheSize int64
	Clock                       feature.Clock
}

func (c Config) nowUnix() int64 {
	if c.Clock != nil {
		return c.Clock.Now().Unix()
	}
	return time.Now().Unix()
}

// CacheEntry represents an individual cached execution record.
type CacheEntry struct {
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Args             []string `json:"args"`
	Dir              string   `json:"dir"`
	CreatedTimestamp int64    `json:"created_timestamp"`
	TTL              int      `json:"ttl"`
	TTLTimestamp     int64    `json:"ttl_timestamp"`
	StartTimestamp   *int64   `json:"start_timestamp"`
	EndTimestamp     *int64   `json:"end_timestamp"`
	Description      *string  `json:"description"`
	Stdin            *string  `json:"stdin"`
	Stdout           string   `json:"stdout"`
	Stderr           string   `json:"stderr"`
	Success          bool     `json:"success"`
	StatusCode       *int     `json:"status_code"`
	StatusString     *string  `json:"status_string"`
}

// Normalize ensures non-nil Args slice before serialization.
func (e *CacheEntry) Normalize() {
	if e.Args == nil {
		e.Args = []string{}
	}
}

// VersionInfo represents metadata for an existing cache entry version.
type VersionInfo struct {
	Version          int    `json:"version"`
	Key              string `json:"key"`
	CreatedTimestamp int64  `json:"created_timestamp"`
	TTLTimestamp     int64  `json:"ttl_timestamp"`
	Expired          bool   `json:"expired"`
}

// StorageResult is returned by Storage.Set.
type StorageResult struct {
	Success bool
	Key     *string
	Version *int
	Message *string
}

// GetResult is returned by Storage.Get.
type GetResult struct {
	Hit     bool
	Key     string
	Version *int
	Data    *CacheEntry
	Reason  string
}

// DeleteResult is returned by Storage.Delete.
type DeleteResult struct {
	Hit          bool
	Deleted      bool
	Key          string
	Version      any
	DeletedCount int
}

// VersionResult is returned by Storage.Version.
type VersionResult struct {
	Hit      bool
	Key      string
	Versions []VersionInfo
	Reason   string
}

// PurgeResult is returned by Storage.Purge.
type PurgeResult struct {
	Success      bool
	DeletedCount int
	FreedBytes   int64
	All          bool
	Message      *string
}

// GetOutput is the typed output struct for cache_get tool.
type GetOutput struct {
	Hit     bool        `json:"hit"`
	Key     string      `json:"key,omitempty"`
	Version *int        `json:"version,omitempty"`
	Data    *CacheEntry `json:"data,omitempty"`
	Reason  string      `json:"reason,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SetOutput is the typed output struct for cache_set tool.
type SetOutput struct {
	Success bool    `json:"success"`
	Key     *string `json:"key,omitempty"`
	Version *int    `json:"version,omitempty"`
	Message *string `json:"message,omitempty"`
}

// ExistOutput is the typed output struct for cache_exist tool.
type ExistOutput struct {
	Hit     bool   `json:"hit"`
	Key     string `json:"key,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// DeleteOutput is the typed output struct for cache_delete tool.
type DeleteOutput struct {
	Hit          bool   `json:"hit"`
	Deleted      bool   `json:"deleted"`
	Key          string `json:"key,omitempty"`
	Version      any    `json:"version,omitempty"`
	DeletedCount int    `json:"deleted_count"`
	Reason       string `json:"reason,omitempty"`
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
}

// VersionOutput is the typed output struct for cache_version tool.
type VersionOutput struct {
	Hit      bool          `json:"hit"`
	Key      string        `json:"key,omitempty"`
	Versions []VersionInfo `json:"versions"`
	Reason   string        `json:"reason,omitempty"`
	Error    string        `json:"error,omitempty"`
	Message  string        `json:"message,omitempty"`
}

// PurgeOutput is the typed output struct for cache_purge tool.
type PurgeOutput struct {
	Success      bool    `json:"success"`
	DeletedCount int     `json:"deleted_count"`
	FreedBytes   int64   `json:"freed_bytes"`
	All          bool    `json:"all"`
	Message      *string `json:"message,omitempty"`
}
