package organizer

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DownloadsDir        string
	CategoryByExtension map[string]string
	IgnoredExtensions   map[string]struct{}
	StabilityChecks     int
	StabilityDelay      time.Duration
}

func DefaultDownloadsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Downloads"), nil
}

func DefaultConfig(downloadsDir string) Config {
	return Config{
		DownloadsDir: downloadsDir,
		CategoryByExtension: map[string]string{
			".jpg":  "Images",
			".jpeg": "Images",
			".png":  "Images",
			".gif":  "Images",
			".webp": "Images",
			".bmp":  "Images",
			".tif":  "Images",
			".tiff": "Images",
			".svg":  "Images",
			".heic": "Images",

			".pdf":  "Documents",
			".doc":  "Documents",
			".docx": "Documents",
			".txt":  "Documents",
			".rtf":  "Documents",
			".xls":  "Documents",
			".xlsx": "Documents",
			".ppt":  "Documents",
			".pptx": "Documents",
			".csv":  "Documents",

			".exe":  "Programs",
			".msi":  "Programs",
			".bat":  "Programs",
			".cmd":  "Programs",
			".ps1":  "Programs",
			".appx": "Programs",

			".zip": "Archives",
			".rar": "Archives",
			".7z":  "Archives",
			".tar": "Archives",
			".gz":  "Archives",
			".bz2": "Archives",
		},
		IgnoredExtensions: map[string]struct{}{
			".crdownload": {},
			".part":       {},
			".tmp":        {},
			".download":   {},
			".opdownload": {},
		},
		StabilityChecks: 6,
		StabilityDelay:  2 * time.Second,
	}
}
