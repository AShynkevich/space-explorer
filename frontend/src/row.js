// One tree row: [arrow] [icon] name | size | percent | bar.
import {el} from './dom.js';
import {formatSize, formatPercent, formatCount, barColor} from './format.js';

const INDENT_PX = 18;

export function isExpandable(entry) {
    return entry.kind === 'dir' && entry.hasChildren;
}

export function createRow(entry, depth) {
    const row = el('div', `row kind-${entry.kind}`);
    row.setAttribute('role', 'treeitem');
    row.dataset.path = entry.path;
    row.dataset.kind = entry.kind;
    row.dataset.depth = String(depth);
    row.title = tooltip(entry);

    const name = el('div', 'cell name');
    name.style.paddingLeft = `${depth * INDENT_PX}px`;
    const arrow = el('span', 'arrow');
    if (isExpandable(entry)) {
        row.classList.add('expandable');
        row.setAttribute('aria-expanded', 'false');
    }
    name.append(arrow, el('span', 'icon'), el('span', 'label', entry.name));
    if (entry.unreadable) name.append(el('span', 'warn', 'partly unreadable'));

    row.append(
        name,
        el('div', 'cell size', formatSize(entry.size)),
        el('div', 'cell percent', formatPercent(entry.percent)),
        createBar(entry.percent),
    );
    return row;
}

function createBar(percent) {
    const cell = el('div', 'cell bar');
    const fill = el('div', 'fill');
    // Tiny shares still get a 1px sliver so they remain visible.
    fill.style.width = `${Math.max(percent || 0, percent > 0 ? 0.4 : 0)}%`;
    fill.style.background = barColor(percent);
    cell.append(fill);
    return cell;
}

function tooltip(entry) {
    const parts = [entry.path, formatSize(entry.size)];
    if (entry.kind === 'dir') parts.push(`${formatCount(entry.items)} files`);
    if (entry.kind === 'other') parts.push(`${formatCount(entry.items)} small items`);
    return parts.join('\n');
}

// Identifies a row across re-renders (the synthetic "other" row shares its folder's path).
export const rowKey = (row) => `${row.dataset.kind === 'other' ? 'other:' : ''}${row.dataset.path}`;
export const depthOf = (row) => Number(row.dataset.depth);
export const isExpanded = (row) => row.getAttribute('aria-expanded') === 'true';
