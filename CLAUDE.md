# Space Explorer - project context

Cross-platform desktop disk-usage analyser (like WinDirStat / ncdu): pick a disk
or folder, scan it, browse a lazy file tree with sizes (B/KB/MB/GB) and a colored
percentage bar per row (share of the root folder).

Full plan and rationale: `doc/PLAN.md`. **Keep this file updated after every
milestone** (status table, structure, decisions, commands).

## Hard constraints
- Must run well on slow machines with **8 GB RAM**: low memory, few goroutines, no heavy deps.
- Cross-platform: macOS, Windows, Linux. Platform code goes in build-tagged files.
- **Source files must stay under 200 lines** (tests included, Go and JS/CSS). See "Code organisation rules".
- No frontend framework: vanilla HTML/CSS/JS only (Vite is just the bundler).
- Prefer editing existing files; no new docs unless asked.
- License: GPL v3 (`LICENSE`, unmodified text from gnu.org). Dependencies must stay GPL-compatible.

## Code organisation rules
**Source size**
- Hard limit: 200 lines per source file, tests included. Aim for ~150; split at ~170, never "fix it later".
- Split by responsibility (what the code is for), not by arbitrary line ranges. A file name should say what it does.
- Functions: keep them short and single-purpose; extract helpers instead of nesting deeply.
- Verify before finishing any task (must show nothing over 199 lines):
  `find . \( -name '*.go' -o -name '*.js' -o -name '*.css' \) -not -path './frontend/node_modules/*' -not -path './frontend/dist/*' -not -path './frontend/wailsjs/*' -not -path './build/*' | xargs wc -l | awk '$1>199 && $2!="total"'`

**Packaging strategy**
- All non-UI logic lives in `internal/<package>`, one responsibility per package (e.g. `tree`, `platform`, `scanner`, `volumes`, `session`).
- Dependencies point one way, no cycles: `tree`, `platform` (leaves) <- `scanner`, `volumes` <- `session` <- root `package main` (Wails glue).
- The root package (`main.go`, `app.go`) is glue only: thin bound methods that delegate to `internal/*`. No business logic there.
- A new feature area gets a new package; do not grow an existing package with unrelated files.
- Platform-specific code goes into build-tagged files (`*_darwin.go`, `*_linux.go`, `*_windows.go`, or `//go:build`) inside the package that needs it, behind a platform-neutral API. Every OS must compile: check with `GOOS=windows go vet ./internal/...` and `GOOS=linux go vet ./internal/...`.
- Packages expose a small API; keep helpers unexported. Models (`tree`) must not import engines (`scanner`).
- Each package has its own `_test.go` files next to the code; tests use temp dirs, never real user data.
- Frontend mirrors this: one ES module per concern under `frontend/src/` (e.g. `api.js`, `format.js`, `tree.js`, `row.js`), `main.js` only wires them together; `wailsjs/` is generated.
- `cmd/<tool>` holds dev tools; they only use the public API of `internal/*`.

## Stack
Go 1.27.3 + Wails v2.17.1 (system webview, no bundled Chromium) + vanilla JS frontend.

Toolchain paths (not on PATH by default):
```
export PATH=/Users/Aliaksandr_Shynkevich/DEV/GO/go/bin:/Users/Aliaksandr_Shynkevich/DEV/GO/workspace/bin:$PATH
```

## OpenCode LSP
`opencode.json` (project root) enables `gopls` (Go) and `typescript-language-server` (JS, uses `frontend/node_modules/typescript` 5.9.3, a devDependency; the global TS 7 has no `tsserver`, so the server refuses to start without the local one). Absolute tool paths are used because opencode's PATH may lack the Go/nvm dirs. Config is read at startup: restart opencode after changing it.

## Commands
```
go test -race ./...              # unit tests
go vet ./... && gofmt -l .       # must print nothing
(cd frontend && npm test)        # JS unit tests (node:test + jsdom)
wails build                      # -> build/bin/space-explorer.app (needs frontend/dist/gitkeep to exist)
wails dev                        # live-reload dev mode
go build -o /tmp/scantool ./cmd/scantool && /tmp/scantool -depth 2 <folder>   # CLI scanner, prints tree + memory
```
Check line limit: `find . -name '*.go' -not -path './frontend/*' | xargs wc -l | sort -n | tail`

## Milestones
| # | Milestone | Status |
|---|---|---|
| 1 | Scaffold (Go module, Wails, vanilla frontend) | done |
| 2 | Scanner core + tests | done |
| 3 | Bindings: folder/disk picker, start/cancel scan, throttled progress, lazy `Children` | done |
| 4 | Tree UI with sizes and percentage bars | done (not yet visually verified in the window) |
| 5 | Polish: sorting, unreadable-folders panel, rescan subtree, reveal in Finder/Explorer | done (UI not yet visually verified) |
| 6 | Verification (memory < ~250 MB), cross-platform builds, README | done (Windows/Linux builds not run on real machines) |

## Structure
Code lives in small sub-packages under `internal/`, one responsibility each. Dependencies point one way:
`tree`, `platform` (leaves) <- `scanner`, `volumes`, `reveal` <- `session` <- root `package main`.
```
main.go, app.go            Wails entry (GC memory limit 256 MB) + bound App: thin delegates to internal/*
frontend/index.html        static layout: header (volumes + actions), column header, #tree, #status
frontend/src/              vanilla ES modules, one concern each (wiring only in main.js):
  api.js                   only file touching wailsjs (bound methods + events)
  main.js                  wires api -> tree/volumes/toolbar/status
  populate.js              buildRows (fetch + rows, re-opens given folders), createRootRow
  sort.js, errors.js       clickable Name/Size headers; unreadable-paths panel (+ macOS Full Disk Access hint)
  tree.js                  lazy tree: flat row list, expand inserts children after the row, collapse removes deeper rows,
                           stale-result guard (generation counter), loading/error row states
  keyboard.js              arrows / Home / End / Enter / Space navigation
  row.js                   builds one row (arrow, icon, name, size, percent, colored bar); depth via data-depth
  format.js                formatSize (B..PB, binary), formatPercent, formatCount, formatDuration, barColor (green->red)
  volumes.js, toolbar.js, status.js, dom.js
  style.css + tree.css     theme/layout and tree/bar styles
  *.test.js, testkit.js    format, tree, refresh, errors tests (jsdom); testkit.js = shared fake backend
frontend/wailsjs/          generated by `wails build/dev` - do not edit
internal/tree/             result model, no dependencies
  node.go                  Node, Other, Path(), Count(), SortBySize
  find.go                  Find(root, absPath) -> *Node
  replace.go               Replace(old, fresh): swap a rescanned subtree in, fix ancestor sizes/order
internal/platform/         OS-specific code, build-tagged
  usage.go                 Usage struct (size, dev, ino, nlink, HasID)
  usage_unix.go/usage_other.go  UsageOf: allocated size (blocks*512) + identity; Windows falls back to apparent size
  skip_*.go                SkipPaths never descended (/proc, /sys, /dev, /System/Volumes, /Volumes)
internal/scanner/          scan engine
  types.go                 Scanner, dir (scan-time state), Progress, Result, ScanError, constants
  scanner.go               New, Progress, Run, worker loop
  listdir.go               scanDir (lists one folder) + queueDir (registers sub-folders)
  finish.go                complete (bottom-up aggregation/pruning), hard-link set, error recording
  stack.go                 LIFO task stack shared by workers
  options.go               Options + DefaultOptions
internal/volumes/          disks to pick from: Volume{Name,Path,Total,Free}, List()
  volumes.go               List (system volume first); list_darwin/linux/windows.go, space_unix/other.go; mounts.go parses /proc/mounts
internal/reveal/           Show(path): open in Finder (`open -R`) / Explorer (`/select,`) / xdg-open; command_*.go per OS
internal/session/          scan lifecycle for the UI, no Wails dependency (events via injected emit func)
  manager.go               Manager: Start (cancels previous), Cancel, Running, Errors; scan() + finish() shared with rescan
  rescan.go                Rescan(path): scan one folder, graft into the tree (tree.Replace), merge errors
  events.go                event names, Summary, throttled progress pump (250 ms)
  view.go                  Children(path, sort) -> Listing{Entries}: top 200 rows (size or name order) + synthetic "N other items" row, Percent of root
cmd/scantool/              dev CLI to validate against `du` and measure memory
build/                     Wails platform assets (icons, plist, NSIS); build/bin is output
```
Each package has `_test.go` files next to the code (`go test -race ./...`).

## Bindings (milestone 3)
Bound methods on `App` (JS: `frontend/wailsjs/go/main/App`, wrapped by `src/api.js`):
`Volumes()`, `HomeDir()`, `PickFolder()`, `StartScan(root)`, `CancelScan()`, `IsScanning()`, `Children(path, sort)`, `ScanErrors()`, `Rescan(path)`, `Reveal(path)`.
Events (Go -> JS, names in `session/events.go`): `scan:progress` (Progress, ~4/s), `scan:done` (Summary), `scan:cancelled`, `scan:failed` (string).
- Flow: `StartScan` -> progress events -> `scan:done` -> UI calls `Children(summary.root)`, then `Children(entry.path)` on expand.
- `Listing.entries[].kind` is `dir` | `file` | `other`; `percent` is the share of the scan root (0..100); `hasChildren` tells whether to render an expander.
- Only one scan at a time; `Start` cancels the previous scan and drops its result (memory).
- JSON field names are camelCase (`session`/`volumes` structs); `scan:progress` payload is the raw `scanner.Progress` (PascalCase: Files, Dirs, Bytes, Errors, Current).

## Tree UI (milestone 4)
- Click anywhere on a folder row (or Enter/Space/Right arrow) expands it: `Children(path)` is fetched and rows are inserted right below, indented by `depth`. Click again / Left arrow collapses (nested expansions are dropped, re-expanding refetches).
- Only `kind=dir` rows with `hasChildren` are expandable; files and the "N other items" row are leaves.
- Bar width = `percent` of the scan root; color goes green (small) -> red (>= 50%) via `barColor`.
- The scan root is shown as the first, pre-expanded row (100%). Selection survives collapse (moves to the collapsed row).
- Flat DOM, no virtualisation: each folder shows <= 200 rows (+ "other"), so DOM stays small.
- Dev tip: `wails dev` for live reload; headless screenshots are not possible in the agent sandbox, so UI changes need a manual look.

## Polish (milestone 5)
- **Sort:** click the Name / Size column header; `Children(path, "size"|"name")`. The "N other items" row always stays last. Changing sort calls `tree.refresh()`.
- **Rescan:** select a folder row, press Rescan. `Session.Rescan` scans that folder alone and grafts it in (`tree.Replace`), ancestors' sizes/percentages update; old data stays visible and is kept if the rescan is cancelled or fails. Rescanning the root = a new full scan. Event `scan:done` carries `rescanned` (path) so the UI calls `tree.refresh()` (keeps expanded folders, selection, scroll) instead of `load()`.
- Rescan caveat: the fresh scan has its own hard-link set, so a file hard-linked from outside the rescanned folder is counted again inside it. Other (not rescanned) folders keep their old numbers until rescanned.
- **Reveal:** selected row -> Finder/Explorer/xdg-open (Linux can only open the parent folder).
- **Unreadable panel:** footer button "N unreadable" (only when N > 0) opens a list (first 500, from `ScanErrors()`); clicking an item reveals it. Shows a Full Disk Access hint when errors are "operation not permitted".
- Concurrency: `Manager.mu` is held by `Children`/`Errors` while reading and by `graft` while mutating the tree.

## Scanner design (milestone 2)
- Worker pool (2-4 goroutines, `Options.Workers`) pulling folders from a LIFO stack (keeps in-flight set small).
- Each folder: `ReadDir` in batches, files summed locally, sub-folders pushed as tasks, results merged once under a mutex.
- A folder completes when `pending` (1 + unfinished sub-folders) reaches 0; `complete()` then adds it to the parent
  (retained if `Size >= MinDirSize`, otherwise folded into parent's `Other` and freed) and walks up.
- Only folders and the top `MaxFilesPerDir` (30) files >= `MinFileSize` (1 MB) are retained; everything else is in `Node.Other{Items,Size}`.
- **Invariant:** folder `Size == sum(Children.Size) + Other.Size`; children sorted by size desc.
- Symlinks never followed or counted; hard links (nlink>1) counted once via a (dev,ino) set.
- Size = allocated blocks (`du`-like) by default; `Options.Apparent` uses file length. The folder entry's own size is included.
- Unreadable paths: recorded (first 500 details, all counted), `Node.Unreadable` propagates up; never fatal.
- Cancellation via `context`; `Progress()` returns atomic counters, cheap to poll (UI should throttle ~4/s).
- Node paths are rebuilt through `Parent` links (`Node.Path()`), not stored.
- Scan-time state (mutex, pending counter, folder-entry size) lives in the unexported `scanner.dir`, not in `tree.Node`, so the retained tree stays small and the model package has no scanner concepts.

Measured (M4 Pro, SSD): home 200 GB / 1.4M files -> 16.6 s, 22.6k nodes, 20 MB RSS; `/` 272 GB / 2.3M files -> 23.5 s, 43.7k nodes, 30 MB RSS. Totals match `du`.

## Verification (milestone 6)
- Memory (target < ~250 MB): scanner peak RSS 20 MB (home, 1.4M files) / 29 MB (`/`, 2.3M files). Whole app idle on macOS = Go process ~95 MB + WebKit helpers ~70 MB = ~160 MB (measured by pid, excluding other apps' WebKit processes). Not yet measured: app RSS *during* a GUI scan.
- Cross-compile: `go build ./internal/... ./cmd/...` OK for windows/{amd64,arm64}, linux/{amd64,arm64}, darwin/amd64; root package builds for windows and vets for linux. `wails build -platform windows/amd64 -o space-explorer.exe` works from macOS (output in `build/bin`, git-ignored). Linux GUI build needs GTK3/WebKitGTK on a Linux host.
- `vite build` / `npm run build` deletes `frontend/dist/gitkeep`; `touch frontend/dist/gitkeep` before `wails build`.
- README.md is user-facing (usage, build per platform, measurements).

## Known notes
- `wails build` warns about macOS 11 vs 13 linker versions: harmless.
- macOS needs Full Disk Access for protected folders (~140 errors in a home scan without it); they show as unreadable, not failures.
- Windows code paths (apparent size, no hard-link dedup) are untested - no Windows machine yet.
