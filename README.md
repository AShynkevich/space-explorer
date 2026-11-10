# Space Explorer

A small desktop disk-usage analyser (in the spirit of WinDirStat / ncdu) for macOS,
Windows and Linux. Pick a disk or folder, scan it, then browse a lazy tree of folders
and files with sizes (B / KB / MB / GB) and a colored bar showing each row's share of
the scanned root.

Built for slow machines: Go scanner with 2-4 workers, only the biggest items are kept
in memory, and the UI uses the system webview (Wails) instead of a bundled browser.

## Using it

- Click a disk card (shows used space) or press **Choose folder...**; the scan starts immediately and **Cancel** stops it.
- Click a folder (or press Enter / Space / Right) to expand it, Left to collapse.
  Arrows, Home and End move the selection.
- Click the **Name** or **Size** column header to change the sort order.
- Select a folder and press **Rescan** to refresh just that folder; parent sizes update.
- **Reveal** shows the selected item in Finder / Explorer / your file manager.
- The footer shows "N unreadable" when some paths could not be read; click it for the list.
  On macOS, grant the app **Full Disk Access** (System Settings > Privacy & Security) to scan
  protected folders.

Sizes are allocated disk space (like `du`), symlinks are never followed and hard links are
counted once. Each folder lists its 200 largest entries plus an "N other items" row; files
under 1 MB are folded into that row.

## Building

Requirements: Go, Node.js, and the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation)
(`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```
wails dev        # live-reload development
wails build      # release build into build/bin
```

| Platform | Notes |
|---|---|
| macOS | Uses the built-in WebKit. `wails build` produces `build/bin/space-explorer.app`. |
| Windows | Needs WebView2 (included in Windows 10/11). Cross-compiles from macOS with `wails build -platform windows/amd64`; add `-nsis` on Windows for an installer. Not yet run on a real Windows machine. |
| Linux | Needs GTK3 and WebKitGTK dev packages (see the Wails docs). Build on Linux; cannot be cross-compiled from macOS. Not yet run on a real Linux machine. |

## Development

```
go test -race ./...          # Go unit tests
go vet ./... && gofmt -l .   # must print nothing
(cd frontend && npm test)    # JS unit tests
go build -o /tmp/scantool ./cmd/scantool && /tmp/scantool -depth 2 <folder>   # CLI scanner
```

Layout and conventions (source files stay under 200 lines, one package per responsibility)
are described in `CLAUDE.md`; the original plan is in `doc/PLAN.md`.

## Measured resource use (Apple M4 Pro, SSD)

| Scan | Time | Retained nodes | Scanner peak RSS |
|---|---|---|---|
| Home folder, 200 GB, 1.4M files | 15 s | 23k | 20 MB |
| `/`, 274 GB, 2.3M files | 23 s | 44k | 29 MB |

Whole app idle on macOS (Go process + WebKit helpers): about 160 MB. Totals match `du`.

## License

Space Explorer is free software, released under the GNU General Public License v3.0.
See [LICENSE](LICENSE) for the full text.
