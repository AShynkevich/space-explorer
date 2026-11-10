# Space Explorer - Implementation Plan

## 1. Goal
A simple cross-platform desktop application that:
- lets the user choose a disk or folder to analyse;
- scans the file system hierarchy and calculates the space each folder occupies;
- shows a file tree with sizes in B / KB / MB / GB;
- shows a colored progress bar per row, indicating the share of the root folder's total size.

## 2. Constraints
- Must run on low-performance computers with **8 GB RAM** -> low memory and CPU footprint is a first-class requirement.
- Cross-platform: Windows, macOS, Linux.
- Keep the app simple (no heavy frameworks, no database, no network).

## 3. Stack decision

| Option | Cross-platform | Scan speed | UI richness | RAM / size | Verdict |
|---|---|---|---|---|---|
| **Go + Wails (system webview)** | Yes | Excellent | Excellent (HTML/CSS) | Low (~60-120 MB UI) | **Chosen** |
| Go + web UI in browser | Yes | Excellent | Excellent | Browser overhead | Fallback |
| Electron | Yes | OK | Excellent | Heavy (300 MB+) | Rejected (8 GB RAM) |
| Tauri (Rust) | Yes | Excellent | Excellent | Low | Rust not installed |
| Python + Tk/Qt | Yes | Slow | Mediocre | Medium | Rejected |
| Console TUI (Bubble Tea) | Yes | Excellent | Limited | Lowest | Rejected (weak visuals) |

**Decision: Go (scanner/backend) + Wails (native window) + vanilla HTML/CSS/JS (frontend).**
- Go: fast concurrent directory walking, small runtime, single native binary per OS.
- Wails: uses the OS webview (WebView2 / WKWebView / WebKitGTK) - no bundled Chromium; provides native folder-picker dialog.
- Vanilla JS: no React/Vue runtime, no heavy `node_modules`.
- Memory cap via `GOMEMLIMIT` (~256 MB).
- Fallback: same Go core served on localhost and opened in a browser if Wails is unsuitable.

Toolchain: Go 1.27.3 (`/Users/Aliaksandr_Shynkevich/DEV/GO/go`), Wails CLI v2.17.1, Node 24 (frontend build only). Rust not required.

## 4. Architecture

### 4.1 Scanner (Go)
- Recursive walk with a bounded worker pool (2-4 goroutines) to avoid thrashing slow HDDs.
- Does **not** follow symlinks (avoids loops and double counting).
- Hard links deduplicated only for files with link count > 1 (no huge inode map).
- Unreadable folders are skipped and collected into an error list (never fatal).
- Cancellable; reports progress (files/dirs scanned, bytes so far).
- Size is bottom-up aggregation. Disk usage (allocated blocks) on macOS/Linux, apparent size on Windows; configurable.

### 4.2 Memory-bounded data model
1. Only **directories** are stored as nodes; files are summed into the parent's size.
2. Keep only the top ~20-50 largest files per folder; the rest collapse into one "N other files" entry.
3. **Pruning:** once a folder finishes scanning and its subtree is below a threshold (default ~1 MB, configurable), fold it into the parent's "other" total and free it.
4. Compact nodes: name, size, file count, children. Full paths rebuilt from parent links.
5. Target: tens of thousands of nodes (a few MB) even for a full system disk.

### 4.3 UI
- "Choose folder" button (native dialog) plus a list of mounted volumes/disks.
- Lazy-expanding tree: backend returns only the children of the expanded folder (`GetChildren(path)`).
- Each folder shows top ~200 children plus an "other" row to keep the DOM small.
- Rows: name, size (auto B/KB/MB/GB), % of root, colored bar.
- Bar is a CSS `div` with width in %; color gradient green -> yellow -> red by share.
- Sorted by size, descending.
- Progress events throttled to ~4/second; scan progress indicator and Cancel button.
- No animations, no charting library.

Tree sketch:
```
v /Users/me            120.4 GB  100%  ############
   > Library            58.2 GB   48%  ######......
   > Movies             31.0 GB   26%  ###.........
   > Projects           12.7 GB   11%  #...........
   . 214 other items     9.1 GB    8%  #...........
```

## 5. Milestones
1. **Scaffold** - Go module, Wails project, plain frontend. **(done)**
2. **Scanner core + tests** - sizes, symlinks, permission errors, pruning, memory use on a large synthetic tree. **(done)**
3. **Bindings** - folder/disk picker, start/cancel scan, throttled progress, lazy `Children(path)`. **(done)**
4. **Tree UI** - sizes (B/KB/MB/GB), percentage bars, click-to-expand lazy tree, keyboard navigation. **(done)**
5. **Polish** - sorting, unreadable-folders panel, rescan a subtree, "reveal in Finder/Explorer". **(done)**
6. **Verification and release** - memory check on a large disk (target < ~250 MB total while scanning); builds for macOS, Windows, Linux; README. **(done; Windows/Linux not run on real machines)**

## 6. Risks and notes
- Wails requires its CLI (`go install`) and, on Linux, WebKitGTK. macOS and Windows 10/11 include a webview.
- macOS may require Full Disk Access to scan protected folders; unreadable folders are reported in the UI, not treated as failures.
- Allocated size vs apparent size can differ (sparse files, compression, cloud placeholders); expose as a setting.
- Very old machines with HDDs: keep worker count low; scanning speed is I/O-bound.

## 7. Working agreements
- Source files (Go, JS, CSS, tests included) stay under 200 lines; split by responsibility.
- Packaging: one responsibility per package under `internal/`, dependencies point one way
  (`tree`, `platform` <- `scanner`, `volumes` <- `session` <- Wails glue in the root package).
  Platform-specific code lives in build-tagged files behind a neutral API. Details in `CLAUDE.md`.
- `CLAUDE.md` (project context) is updated after each implemented milestone.
