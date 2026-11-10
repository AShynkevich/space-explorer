// Single place that talks to the Go backend (generated Wails bindings + events).
import {
    Volumes, HomeDir, PickFolder, StartScan, CancelScan,
    IsScanning, Children, ScanErrors, Rescan, Reveal,
} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

export const api = {
    volumes: Volumes,
    homeDir: HomeDir,
    pickFolder: PickFolder,
    startScan: StartScan,
    cancelScan: CancelScan,
    isScanning: IsScanning,
    children: (path, sort) => Children(path, sort || 'size'),
    scanErrors: ScanErrors,
    rescan: Rescan,
    reveal: Reveal,
};

// Event names mirror internal/session/events.go.
export const events = {
    progress: (fn) => EventsOn('scan:progress', fn),
    done: (fn) => EventsOn('scan:done', fn),
    cancelled: (fn) => EventsOn('scan:cancelled', fn),
    failed: (fn) => EventsOn('scan:failed', fn),
};
