// Pure formatting helpers (no DOM, unit-tested with `npm test`).

const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

// 1536 -> "1.5 KB". Binary units (1024), like the scanner's `du`-style sizes.
export function formatSize(bytes) {
    if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
    if (bytes < 1024) return `${Math.round(bytes)} B`;
    let v = bytes;
    let i = 0;
    while (v >= 1024 && i < UNITS.length - 1) {
        v /= 1024;
        i++;
    }
    if (v.toFixed(1) === '1024.0' && i < UNITS.length - 1) {
        v = 1;
        i++;
    }
    return `${v.toFixed(1)} ${UNITS[i]}`;
}

export function formatPercent(p) {
    if (!(p > 0)) return '0%';
    if (p < 0.1) return '<0.1%';
    const s = p.toFixed(1);
    return s === '100.0' ? '100%' : `${s}%`;
}

export function formatCount(n) {
    return Number(n || 0).toLocaleString('en-US');
}

export function formatDuration(ms) {
    if (ms < 1000) return `${ms} ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)} s`;
    const m = Math.floor(ms / 60000);
    const s = Math.round((ms % 60000) / 1000);
    return `${m} min ${s} s`;
}

// Bar color by share of the root: green (small) -> yellow -> red (>= 50%).
export function barColor(percent) {
    const p = Math.min(Math.max(percent || 0, 0), 100);
    const hue = Math.round(130 * (1 - Math.min(p / 50, 1)));
    return `hsl(${hue} 70% 48%)`;
}
