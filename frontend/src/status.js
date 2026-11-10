// Status bar below the tree: scan progress and results.
import {$} from './dom.js';
import {formatSize, formatCount, formatDuration} from './format.js';

const set = (text, cls = '') => {
    const s = $('status');
    s.textContent = text;
    s.className = cls;
};

export const status = {
    idle: () => set('Choose a disk or folder to analyse.'),
    scanning: (p) => set(
        `Scanning... ${formatCount(p.Files)} files, ${formatCount(p.Dirs)} folders, ${formatSize(p.Bytes)}` +
        (p.Errors ? `, ${formatCount(p.Errors)} unreadable` : ''), 'busy'),
    done: (s) => set(
        (s.rescanned ? `Rescanned ${s.rescanned}. ` : '') + `${s.root}: ${formatSize(s.size)} in ${formatCount(s.files)} files and ${formatCount(s.dirs)} folders` +
        ` (${formatDuration(s.durationMs)})` +
        (s.errorCount ? `, ${formatCount(s.errorCount)} unreadable paths` : '')),
    cancelled: () => set('Scan cancelled.'),
    failed: (msg) => set(`Scan failed: ${msg}`, 'error'),
};
