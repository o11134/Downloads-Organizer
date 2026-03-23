package organizer

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanAndOrganizeMovesFilesToCategoryFolder(t *testing.T) {
	downloads := t.TempDir()
	cfg := DefaultConfig(downloads)
	cfg.StabilityChecks = 2
	cfg.StabilityDelay = 1 * time.Millisecond

	org := New(cfg, log.New(os.Stderr, "", 0))

	input := filepath.Join(downloads, "report.pdf")
	if err := os.WriteFile(input, []byte("pdf"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := org.ScanAndOrganize(); err != nil {
		t.Fatalf("scan and organize: %v", err)
	}

	expected := filepath.Join(downloads, "Documents", "report.pdf")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected moved file at %q: %v", expected, err)
	}

	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Fatalf("expected source file to be moved, got err=%v", err)
	}
}

func TestScanAndOrganizeAddsNumericSuffixOnConflict(t *testing.T) {
	downloads := t.TempDir()
	cfg := DefaultConfig(downloads)
	cfg.StabilityChecks = 2
	cfg.StabilityDelay = 1 * time.Millisecond

	targetDir := filepath.Join(downloads, "Documents")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}

	existing := filepath.Join(targetDir, "report.pdf")
	if err := os.WriteFile(existing, []byte("old"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	incoming := filepath.Join(downloads, "report.pdf")
	if err := os.WriteFile(incoming, []byte("new"), 0o644); err != nil {
		t.Fatalf("write incoming: %v", err)
	}

	org := New(cfg, log.New(os.Stderr, "", 0))
	if err := org.ScanAndOrganize(); err != nil {
		t.Fatalf("scan and organize: %v", err)
	}

	newPath := filepath.Join(targetDir, "report (1).pdf")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected conflict file at %q: %v", newPath, err)
	}
}

func TestScanAndOrganizeLeavesUnknownExtensions(t *testing.T) {
	downloads := t.TempDir()
	cfg := DefaultConfig(downloads)
	cfg.StabilityChecks = 2
	cfg.StabilityDelay = 1 * time.Millisecond

	input := filepath.Join(downloads, "notes.xyz")
	if err := os.WriteFile(input, []byte("data"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	org := New(cfg, log.New(os.Stderr, "", 0))
	if err := org.ScanAndOrganize(); err != nil {
		t.Fatalf("scan and organize: %v", err)
	}

	if _, err := os.Stat(input); err != nil {
		t.Fatalf("expected unknown extension file to stay in place: %v", err)
	}
}
