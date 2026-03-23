package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"downloads-organizer/internal/organizer"
)

const configFileName = "config.json"

type NotificationSettings struct {
	Enabled       bool
	BatchInterval time.Duration
	BatchMaxFiles int
}

type RuntimeConfig struct {
	Organizer      organizer.Config
	Notifications  NotificationSettings
	StartupEnabled bool
}

type Store struct {
	mu   sync.Mutex
	path string
	data fileConfig
}

type fileConfig struct {
	DownloadsDir                     string            `json:"downloads_dir"`
	CategoryByExtension              map[string]string `json:"category_by_extension"`
	IgnoredExtensions                []string          `json:"ignored_extensions"`
	StabilityChecks                  *int              `json:"stability_checks"`
	StabilityDelayMS                 *int              `json:"stability_delay_ms"`
	NotificationsEnabled             *bool             `json:"notifications_enabled"`
	NotificationBatchIntervalSeconds *int              `json:"notification_batch_interval_seconds"`
	NotificationBatchMaxFiles        *int              `json:"notification_batch_max_files"`
	StartWithWindows                 *bool             `json:"start_with_windows"`
}

func Open(appDir string, defaultDownloadsDir string) (*Store, RuntimeConfig, error) {
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, RuntimeConfig{}, err
	}

	path := filepath.Join(appDir, configFileName)
	defaults := defaultFileConfig(defaultDownloadsDir)
	merged := defaults

	raw, err := os.ReadFile(path)
	if err == nil {
		var loaded fileConfig
		if unmarshalErr := json.Unmarshal(raw, &loaded); unmarshalErr != nil {
			return nil, RuntimeConfig{}, fmt.Errorf("parse config %s: %w", path, unmarshalErr)
		}
		merged = mergeFileConfig(defaults, loaded)
	} else if !os.IsNotExist(err) {
		return nil, RuntimeConfig{}, err
	}

	store := &Store{path: path, data: merged}
	if err := store.saveLocked(); err != nil {
		return nil, RuntimeConfig{}, err
	}

	return store, toRuntimeConfig(merged), nil
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) SetNotificationsEnabled(enabled bool) (RuntimeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.NotificationsEnabled = boolPtr(enabled)
	if err := s.saveLocked(); err != nil {
		return RuntimeConfig{}, err
	}

	return toRuntimeConfig(s.data), nil
}

func (s *Store) SetStartupEnabled(enabled bool) (RuntimeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.StartWithWindows = boolPtr(enabled)
	if err := s.saveLocked(); err != nil {
		return RuntimeConfig{}, err
	}

	return toRuntimeConfig(s.data), nil
}

func (s *Store) saveLocked() error {
	encoded, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(encoded, '\n'), 0o644); err != nil {
		return err
	}

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.Rename(tmp, s.path)
}

func defaultFileConfig(downloadsDir string) fileConfig {
	def := organizer.DefaultConfig(downloadsDir)
	extMap := make(map[string]string, len(def.CategoryByExtension))
	for ext, category := range def.CategoryByExtension {
		extMap[ext] = category
	}

	ignored := make([]string, 0, len(def.IgnoredExtensions))
	for ext := range def.IgnoredExtensions {
		ignored = append(ignored, ext)
	}
	sort.Strings(ignored)

	stabilityChecks := def.StabilityChecks
	stabilityDelayMS := int(def.StabilityDelay / time.Millisecond)
	notificationsEnabled := true
	notificationBatchIntervalSeconds := 4
	notificationBatchMaxFiles := 20
	startWithWindows := false

	return fileConfig{
		DownloadsDir:                     def.DownloadsDir,
		CategoryByExtension:              extMap,
		IgnoredExtensions:                ignored,
		StabilityChecks:                  intPtr(stabilityChecks),
		StabilityDelayMS:                 intPtr(stabilityDelayMS),
		NotificationsEnabled:             boolPtr(notificationsEnabled),
		NotificationBatchIntervalSeconds: intPtr(notificationBatchIntervalSeconds),
		NotificationBatchMaxFiles:        intPtr(notificationBatchMaxFiles),
		StartWithWindows:                 boolPtr(startWithWindows),
	}
}

func mergeFileConfig(defaults fileConfig, loaded fileConfig) fileConfig {
	result := defaults

	if dir := strings.TrimSpace(loaded.DownloadsDir); dir != "" {
		result.DownloadsDir = dir
	}

	if loaded.CategoryByExtension != nil {
		result.CategoryByExtension = normalizeCategoryMap(loaded.CategoryByExtension)
	}

	if loaded.IgnoredExtensions != nil {
		result.IgnoredExtensions = normalizeExtensions(loaded.IgnoredExtensions)
	}

	if loaded.StabilityChecks != nil {
		result.StabilityChecks = intPtr(*loaded.StabilityChecks)
	}

	if loaded.StabilityDelayMS != nil {
		result.StabilityDelayMS = intPtr(*loaded.StabilityDelayMS)
	}

	if loaded.NotificationsEnabled != nil {
		result.NotificationsEnabled = boolPtr(*loaded.NotificationsEnabled)
	}

	if loaded.NotificationBatchIntervalSeconds != nil {
		result.NotificationBatchIntervalSeconds = intPtr(*loaded.NotificationBatchIntervalSeconds)
	}

	if loaded.NotificationBatchMaxFiles != nil {
		result.NotificationBatchMaxFiles = intPtr(*loaded.NotificationBatchMaxFiles)
	}

	if loaded.StartWithWindows != nil {
		result.StartWithWindows = boolPtr(*loaded.StartWithWindows)
	}

	return result
}

func toRuntimeConfig(cfg fileConfig) RuntimeConfig {
	stabilityChecks := derefInt(cfg.StabilityChecks, 6)
	if stabilityChecks < 0 {
		stabilityChecks = 0
	}

	stabilityDelayMS := derefInt(cfg.StabilityDelayMS, 2000)
	if stabilityDelayMS < 0 {
		stabilityDelayMS = 0
	}

	batchIntervalSeconds := derefInt(cfg.NotificationBatchIntervalSeconds, 4)
	if batchIntervalSeconds <= 0 {
		batchIntervalSeconds = 4
	}

	batchMaxFiles := derefInt(cfg.NotificationBatchMaxFiles, 20)
	if batchMaxFiles <= 0 {
		batchMaxFiles = 20
	}

	ignoredMap := make(map[string]struct{}, len(cfg.IgnoredExtensions))
	for _, ext := range normalizeExtensions(cfg.IgnoredExtensions) {
		ignoredMap[ext] = struct{}{}
	}

	return RuntimeConfig{
		Organizer: organizer.Config{
			DownloadsDir:        strings.TrimSpace(cfg.DownloadsDir),
			CategoryByExtension: normalizeCategoryMap(cfg.CategoryByExtension),
			IgnoredExtensions:   ignoredMap,
			StabilityChecks:     stabilityChecks,
			StabilityDelay:      time.Duration(stabilityDelayMS) * time.Millisecond,
		},
		Notifications: NotificationSettings{
			Enabled:       derefBool(cfg.NotificationsEnabled, true),
			BatchInterval: time.Duration(batchIntervalSeconds) * time.Second,
			BatchMaxFiles: batchMaxFiles,
		},
		StartupEnabled: derefBool(cfg.StartWithWindows, false),
	}
}

func normalizeCategoryMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for ext, category := range input {
		normalizedExt := normalizeExtension(ext)
		normalizedCategory := strings.TrimSpace(category)
		if normalizedExt == "" || normalizedCategory == "" {
			continue
		}
		result[normalizedExt] = normalizedCategory
	}
	return result
}

func normalizeExtensions(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, ext := range input {
		normalized := normalizeExtension(ext)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func normalizeExtension(ext string) string {
	trimmed := strings.TrimSpace(strings.ToLower(ext))
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, ".") {
		return "." + trimmed
	}
	return trimmed
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func derefBool(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func derefInt(v *int, fallback int) int {
	if v == nil {
		return fallback
	}
	return *v
}
