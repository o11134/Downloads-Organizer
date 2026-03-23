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
- Loads behavior from `config.json` in the app config directory.

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

## Configuration

On startup, the app creates/updates `config.json` in:

- Windows: `%AppData%\\DownloadsOrganizer\\config.json`

You can open it from tray with `Open Config`.

Example:

```json
{
  "downloads_dir": "C:/Users/you/Downloads",
  "notifications_enabled": true,
  "notification_batch_interval_seconds": 4,
  "notification_batch_max_files": 20,
  "start_with_windows": false,
  "stability_checks": 6,
  "stability_delay_ms": 2000
}
```

You can also customize:

- `category_by_extension` (map of extension -> folder name)
- `ignored_extensions` (extensions skipped by organizer)

### Windows tray controls

- Pause/resume organizing
- Organize now (manual scan)
- Open Downloads folder
- Open log file
- Open config file
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

## Windows installer (Inno Setup)

1. Build the executable in project root:

```bash
go build -o DownloadsOrganizer.exe ./cmd/downloads-organizer
```

2. Open `installer/DownloadsOrganizer.iss` in Inno Setup Compiler.
3. Build the installer.

Output installer:

- `dist/installer/DownloadsOrganizerSetup.exe`

## Logging

Logs are written to:

- Windows: `%AppData%\\DownloadsOrganizer\\organizer.log`
- Other platforms: config directory equivalent from `os.UserConfigDir()`

## Tests

```bash
go test ./...
```
