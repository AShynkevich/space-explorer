import test from 'node:test';
import assert from 'node:assert/strict';
import {entry, data, setup} from './testkit.js';

test('refresh keeps expanded folders, selection and re-fetches', async () => {
    const t = setup();
    await t.tree.load('/r');
    t.click('/r/sub');
    await t.tick();
    t.click('/r/sub/deep');
    await t.tick();
    t.click('/r/sub/deep/c.bin'); // select a nested file
    const before = t.calls.length;

    await t.tree.refresh();
    assert.deepEqual(t.names(), ['/r', 'sub', 'deep', 'c.bin', 'b.bin', 'a.bin']);
    assert.ok(t.calls.length > before, 'refresh must hit the backend again');
    assert.ok(t.row('/r/sub/deep/c.bin').classList.contains('selected'));
    assert.equal(t.row('/r/sub').getAttribute('aria-expanded'), 'true');
    assert.equal(t.row('/r/sub/deep').getAttribute('aria-expanded'), 'true');
});

test('refresh shows changed data and falls back when a folder disappeared', async () => {
    let current = structuredClone(data);
    const t = setup(async (p) => {
        if (!current[p]) throw new Error('gone');
        return {path: p, size: 300, total: 300, entries: current[p]};
    });
    await t.tree.load('/r');
    t.click('/r/sub');
    await t.tick();
    t.click('/r/sub/deep');
    await t.tick();

    current['/r/sub'] = [entry('new.bin', '/r/sub/new.bin', 'file')]; // deep is gone
    delete current['/r/sub/deep'];
    await t.tree.refresh();
    assert.deepEqual(t.names(), ['/r', 'sub', 'new.bin', 'a.bin']);
    assert.ok(t.row('/r/sub/deep') === null);
    assert.ok(t.row('/r/sub/deep/c.bin') === null);
    // selection fell back to the root row
    assert.ok(t.row('/r').classList.contains('selected'));
});

test('refresh before any load is a no-op', async () => {
    const t = setup();
    await t.tree.refresh();
    assert.deepEqual(t.names(), []);
    assert.equal(t.calls.length, 0);
});

test('onSelect reports the selected row and null after clear', async () => {
    const t = setup();
    const seen = [];
    t.tree.onSelect((s) => seen.push(s));
    await t.tree.load('/r');
    t.click('/r/a.bin');
    t.key('ArrowUp');
    assert.deepEqual(seen.at(-3), {path: '/r', kind: 'dir'});
    assert.deepEqual(seen.at(-2), {path: '/r/a.bin', kind: 'file'});
    assert.deepEqual(seen.at(-1), {path: '/r/sub', kind: 'dir'});
    assert.deepEqual(t.tree.selected(), {path: '/r/sub', kind: 'dir'});
    t.tree.clear();
    assert.equal(seen.at(-1), null);
    assert.equal(t.tree.selected(), null);
});

test('a refresh that finishes after a new load is dropped', async () => {
    const t = setup();
    await t.tree.load('/r');
    const pending = t.tree.refresh();
    t.tree.clear(); // new scan started meanwhile
    await pending;
    assert.deepEqual(t.names(), []);
});

test('the synthetic "other" row keeps selection across refresh', async () => {
    const extra = structuredClone(data);
    extra['/r'].push(entry('3 other items', '/r', 'other', {size: 5}));
    const t = setup(async (p) => ({path: p, size: 300, total: 300, entries: extra[p]}));
    await t.tree.load('/r');
    const other = t.container.querySelector('.row.kind-other');
    other.dispatchEvent(new (other.ownerDocument.defaultView.MouseEvent)('click', {bubbles: true}));
    await t.tree.refresh();
    assert.ok(t.container.querySelector('.row.kind-other').classList.contains('selected'));
});
