# Downloads Organizer (Go)

Background tray app that keeps your `Downloads` folder clean by moving files into category folders based on extension.

## What it does

- Watches your `Downloads` folder continuously.
- Runs an initial scan on startup.
- Moves files into folders inside `Downloads`:
  - `Images`
  - `Documents`
  - `Programs`
  - `Archives`
- Ignores temporary/incomplete downloads (`.crdownload`, `.part`, `.tmp`, `.download`, `.opdownload`).
- Handles name conflicts by appending a number, for example `report (1).pdf`.
- Shows grouped Windows notifications when files are moved.

## Extension mapping

- `Images`: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.bmp`, `.tif`, `.tiff`, `.svg`, `.heic`
- `Documents`: `.pdf`, `.doc`, `.docx`, `.txt`, `.rtf`, `.xls`, `.xlsx`, `.ppt`, `.pptx`, `.csv`
- `Programs`: `.exe`, `.msi`, `.bat`, `.cmd`, `.ps1`, `.appx`
- `Archives`: `.zip`, `.rar`, `.7z`, `.tar`, `.gz`, `.bz2`

Unknown extensions are left in `Downloads`.

## Project layout

- `cmd/downloads-organizer`: app entrypoint
- `internal/organizer`: extension rules + move logic
- `internal/watcher`: filesystem watcher
- `internal/app`: service lifecycle (start/stop/scan)

## Running

1. Install Go 1.22+.
2. From project root:

```bash
go run ./cmd/downloads-organizer
```

On Windows, this starts as a tray app.
On non-Windows systems, it runs in console mode.

### Windows tray controls

- Pause/resume organizing
- Organize now (manual scan)
- Open Downloads folder
- Open log file
- Enable/disable notifications
- Enable/disable startup with Windows

## Building

```bash
go build ./cmd/downloads-organizer
```

For a Windows executable:

```bash
go build -o DownloadsOrganizer.exe ./cmd/downloads-organizer
```

## Logging

Logs are written to:

- Windows: `%AppData%\\DownloadsOrganizer\\organizer.log`
- Other platforms: config directory equivalent from `os.UserConfigDir()`

## Tests

```bash
go test ./...
```
