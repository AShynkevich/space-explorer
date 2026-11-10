// Fetches a folder's listing and builds its rows, recursing into the folders
// listed in `open` (used to restore expanded state after a refresh).
import {createRow, isExpandable} from './row.js';

export async function buildRows(fetchChildren, path, depth, open) {
    const listing = await fetchChildren(path);
    const rows = [];
    for (const entry of listing.entries) {
        const row = createRow(entry, depth);
        rows.push(row);
        if (!isExpandable(entry) || !open.has(entry.path)) continue;
        try {
            const nested = await buildRows(fetchChildren, entry.path, depth + 1, open);
            row.setAttribute('aria-expanded', 'true');
            rows.push(...nested.rows);
        } catch {
            // folder vanished or fails to load: leave it collapsed
        }
    }
    return {listing, rows};
}

// The first row of the tree: the scan root, shown expanded at 100%.
export function createRootRow(path, listing) {
    const row = createRow({
        name: path, path, kind: 'dir', size: listing.size,
        percent: 100, items: 0, hasChildren: listing.entries.length > 0,
    }, 0);
    row.classList.add('root');
    row.setAttribute('aria-expanded', 'true');
    return row;
}
