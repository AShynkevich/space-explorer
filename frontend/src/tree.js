// Lazy file tree: rows are a flat list; expanding a folder fetches its
// children from the backend and inserts them right after the folder's row.
import {el} from './dom.js';
import {depthOf, isExpanded, rowKey} from './row.js';
import {buildRows, createRootRow} from './populate.js';
import {attachKeyboard} from './keyboard.js';

export function createTree(container, fetchChildren) {
    container.classList.add('tree');
    container.setAttribute('role', 'tree');
    container.tabIndex = 0;

    let generation = 0; // bumped on every load/refresh; stale async results are dropped
    let selected = null;
    let rootPath = null;
    let onSelect = () => {};

    const rows = () => Array.from(container.querySelectorAll('.row'));

    function select(row) {
        if (!row) return;
        selected?.classList.remove('selected');
        selected = row;
        row.classList.add('selected');
        row.scrollIntoView({block: 'nearest'});
        onSelect({path: row.dataset.path, kind: row.dataset.kind});
    }

    async function expand(row) {
        if (!row.classList.contains('expandable') || isExpanded(row) || row.classList.contains('loading')) return;
        const gen = generation;
        row.classList.add('loading');
        try {
            const {rows: kids} = await buildRows(fetchChildren, row.dataset.path, depthOf(row) + 1, new Set());
            if (gen !== generation || !row.isConnected) return;
            row.after(...kids);
            row.setAttribute('aria-expanded', 'true');
        } catch (err) {
            row.classList.add('error');
            row.title = String(err);
        } finally {
            row.classList.remove('loading');
        }
    }

    function collapse(row) {
        if (!isExpanded(row)) return;
        const depth = depthOf(row);
        let next = row.nextElementSibling;
        while (next && depthOf(next) > depth) {
            const gone = next;
            next = next.nextElementSibling;
            if (gone === selected) select(row);
            gone.remove();
        }
        row.setAttribute('aria-expanded', 'false');
    }

    function toggle(row) {
        if (isExpanded(row)) collapse(row);
        else expand(row);
    }

    // Rebuilds the whole tree in one DOM swap, re-opening the folders in `open`.
    async function render(open, keepKey) {
        const gen = ++generation;
        const {listing, rows: kids} = await buildRows(fetchChildren, rootPath, 1, open);
        if (gen !== generation) return;
        const root = createRootRow(rootPath, listing);
        const scroll = container.scrollTop;
        container.replaceChildren(root, ...kids);
        container.scrollTop = scroll;
        selected = null;
        select(rows().find((r) => rowKey(r) === keepKey) || root);
    }

    container.addEventListener('click', (ev) => {
        const row = ev.target.closest('.row');
        if (!row) return;
        select(row);
        if (row.classList.contains('expandable')) toggle(row);
    });

    attachKeyboard(container, {rows, selected: () => selected, select, expand, collapse, toggle});

    return {
        clear() {
            generation++;
            selected = null;
            rootPath = null;
            container.replaceChildren();
            onSelect(null);
        },
        // Shows the scan root as an expanded first row.
        async load(path) {
            this.clear();
            rootPath = path;
            await render(new Set([path]), null);
            container.focus({preventScroll: true});
        },
        // Re-fetches everything that is visible (after a rescan or sort change)
        // keeping expanded folders, selection and scroll position.
        async refresh() {
            if (!rootPath) return;
            const open = new Set(rows().filter(isExpanded).map((r) => r.dataset.path));
            await render(open, selected ? rowKey(selected) : null);
        },
        showMessage(text) {
            this.clear();
            container.append(el('div', 'empty', text));
        },
        selected: () => (selected ? {path: selected.dataset.path, kind: selected.dataset.kind} : null),
        rootPath: () => rootPath,
        onSelect(fn) { onSelect = fn; },
    };
}
