// Shared jsdom setup for tree tests (not a test file itself).
import {JSDOM} from 'jsdom';

export const dom = new JSDOM('<body></body>');
globalThis.document = dom.window.document;
dom.window.HTMLElement.prototype.scrollIntoView = () => {};
const {createTree} = await import('./tree.js');

export const entry = (name, path, kind, extra = {}) =>
    ({name, path, kind, size: 100, percent: 10, items: 1, hasChildren: false, unreadable: false, ...extra});

// Fake backend: /r -> [sub (expandable), a.bin]; /r/sub -> [deep (expandable), b.bin]; /r/sub/deep -> [c.bin]
export const data = {
    '/r': [entry('sub', '/r/sub', 'dir', {hasChildren: true}), entry('a.bin', '/r/a.bin', 'file')],
    '/r/sub': [entry('deep', '/r/sub/deep', 'dir', {hasChildren: true}), entry('b.bin', '/r/sub/b.bin', 'file')],
    '/r/sub/deep': [entry('c.bin', '/r/sub/deep/c.bin', 'file')],
};

// Every test gets its own container so listeners from earlier trees don't leak in.
export function freshContainer() {
    const c = document.createElement('main');
    document.body.replaceChildren(c);
    return c;
}

export function setup(fetch) {
    const calls = [];
    const container = freshContainer();
    const tree = createTree(container, fetch || (async (p) => {
        calls.push(p);
        return {path: p, size: 300, total: 300, entries: data[p]};
    }));
    const names = () => [...container.querySelectorAll('.row .label')].map((n) => n.textContent);
    const row = (path) => container.querySelector(`.row[data-path="${path}"]`);
    const click = (path) => row(path).dispatchEvent(new dom.window.MouseEvent('click', {bubbles: true}));
    const key = (k) => container.dispatchEvent(new dom.window.KeyboardEvent('keydown', {key: k, bubbles: true}));
    const tick = () => new Promise((r) => setTimeout(r, 0));
    return {tree, container, calls, names, row, click, key, tick};
}

