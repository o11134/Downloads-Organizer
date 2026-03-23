package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenCreatesDefaultConfigFile(t *testing.T) {
	appDir := t.TempDir()
	downloadsDir := filepath.Join(appDir, "Downloads")

	store, runtimeCfg, err := Open(appDir, downloadsDir)
	if err != nil {
		t.Fatalf("open settings: %v", err)
	}

	if got := runtimeCfg.Organizer.DownloadsDir; got != downloadsDir {
		t.Fatalf("unexpected downloads dir: got=%q want=%q", got, downloadsDir)
	}

	if !runtimeCfg.Notifications.Enabled {
		t.Fatalf("expected notifications to be enabled by default")
	}

	if runtimeCfg.Notifications.BatchInterval != 4*time.Second {
		t.Fatalf("unexpected default batch interval: %s", runtimeCfg.Notifications.BatchInterval)
	}

	if _, err := os.Stat(store.Path()); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestOpenMergesExistingConfig(t *testing.T) {
	appDir := t.TempDir()
	downloadsDir := filepath.Join(appDir, "Downloads")

	custom := fileConfig{
		DownloadsDir:                     filepath.Join(appDir, "CustomDownloads"),
		CategoryByExtension:              map[string]string{"PNG": "Photos"},
		IgnoredExtensions:                []string{"TMP", ".partial"},
		StabilityChecks:                  intPtr(2),
		StabilityDelayMS:                 intPtr(500),
		NotificationsEnabled:             boolPtr(false),
		NotificationBatchIntervalSeconds: intPtr(9),
		NotificationBatchMaxFiles:        intPtr(7),
		StartWithWindows:                 boolPtr(true),
	}

	encoded, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		t.Fatalf("marshal custom config: %v", err)
	}

	path := filepath.Join(appDir, configFileName)
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		t.Fatalf("write custom config: %v", err)
	}

	_, runtimeCfg, err := Open(appDir, downloadsDir)
	if err != nil {
		t.Fatalf("open settings: %v", err)
	}

	if got := runtimeCfg.Organizer.DownloadsDir; got != custom.DownloadsDir {
		t.Fatalf("unexpected merged downloads dir: got=%q want=%q", got, custom.DownloadsDir)
	}

	if got := runtimeCfg.Organizer.CategoryByExtension[".png"]; got != "Photos" {
		t.Fatalf("unexpected category merge for .png: got=%q want=%q", got, "Photos")
	}

	if _, exists := runtimeCfg.Organizer.IgnoredExtensions[".tmp"]; !exists {
		t.Fatalf("expected normalized .tmp ignored extension")
	}

	if runtimeCfg.Notifications.Enabled {
		t.Fatalf("expected notifications to be disabled")
	}

	if runtimeCfg.Notifications.BatchInterval != 9*time.Second {
		t.Fatalf("unexpected notification interval: %s", runtimeCfg.Notifications.BatchInterval)
	}

	if runtimeCfg.Notifications.BatchMaxFiles != 7 {
		t.Fatalf("unexpected notification batch max files: %d", runtimeCfg.Notifications.BatchMaxFiles)
	}

	if !runtimeCfg.StartupEnabled {
		t.Fatalf("expected startup to be enabled")
	}
}
