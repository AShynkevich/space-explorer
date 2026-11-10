import test from 'node:test';
import assert from 'node:assert/strict';
import {dom, data, setup, freshContainer} from './testkit.js';
import {createTree} from './tree.js';

test('load shows root expanded with its children', async () => {
    const t = setup();
    await t.tree.load('/r');
    assert.deepEqual(t.names(), ['/r', 'sub', 'a.bin']);
    assert.equal(t.row('/r').getAttribute('aria-expanded'), 'true');
    assert.ok(t.row('/r').classList.contains('selected'));
});

test('click expands lazily, inserts children right below, then collapses', async () => {
    const t = setup();
    await t.tree.load('/r');
    t.click('/r/sub');
    await t.tick();
    assert.deepEqual(t.names(), ['/r', 'sub', 'deep', 'b.bin', 'a.bin']);
    assert.equal(t.row('/r/sub/deep').dataset.depth, '2');
    t.click('/r/sub/deep');
    await t.tick();
    assert.deepEqual(t.names(), ['/r', 'sub', 'deep', 'c.bin', 'b.bin', 'a.bin']);

    t.click('/r/sub'); // collapse removes the whole subtree, including nested expansions
    assert.deepEqual(t.names(), ['/r', 'sub', 'a.bin']);
    assert.equal(t.row('/r/sub').getAttribute('aria-expanded'), 'false');

    const before = t.calls.length;
    t.click('/r/sub'); // expand again fetches fresh data
    await t.tick();
    assert.equal(t.calls.length, before + 1);
    assert.deepEqual(t.names(), ['/r', 'sub', 'deep', 'b.bin', 'a.bin']);
});

test('files are not expandable', async () => {
    const t = setup();
    await t.tree.load('/r');
    const before = t.calls.length;
    t.click('/r/a.bin');
    await t.tick();
    assert.equal(t.calls.length, before);
    assert.ok(t.row('/r/a.bin').classList.contains('selected'));
});

test('double click while loading fetches once', async () => {
    const t = setup();
    await t.tree.load('/r');
    const before = t.calls.length;
    t.click('/r/sub');
    t.click('/r/sub');
    await t.tick();
    assert.equal(t.calls.length, before + 1);
    assert.deepEqual(t.names(), ['/r', 'sub', 'deep', 'b.bin', 'a.bin']);
});

test('keyboard: arrows move, right expands, left collapses / goes to parent', async () => {
    const t = setup();
    await t.tree.load('/r');
    t.key('ArrowDown'); // -> sub
    assert.ok(t.row('/r/sub').classList.contains('selected'));
    t.key('ArrowRight');
    await t.tick();
    assert.equal(t.row('/r/sub').getAttribute('aria-expanded'), 'true');
    t.key('ArrowRight'); // already expanded -> first child
    assert.ok(t.row('/r/sub/deep').classList.contains('selected'));
    t.key('ArrowLeft'); // collapsed leaf-ish row -> parent
    assert.ok(t.row('/r/sub').classList.contains('selected'));
    t.key('ArrowLeft'); // collapse sub
    assert.deepEqual(t.names(), ['/r', 'sub', 'a.bin']);
    t.key('End');
    assert.ok(t.row('/r/a.bin').classList.contains('selected'));
    t.key('Home');
    assert.ok(t.row('/r').classList.contains('selected'));
});

test('collapsing a row that contains the selection moves selection to the row', async () => {
    const t = setup();
    await t.tree.load('/r');
    t.click('/r/sub');
    await t.tick();
    t.click('/r/sub/b.bin');
    t.click('/r/sub');
    assert.ok(t.row('/r/sub').classList.contains('selected'));
});

test('a stale expand result is dropped after the tree is reloaded', async () => {
    const t = setup();
    await t.tree.load('/r');
    t.click('/r/sub');
    t.tree.clear(); // new scan started before the fetch resolved
    await t.tick();
    assert.deepEqual(t.names(), []);
});

test('fetch errors are shown on the row, not thrown', async () => {
    const container = freshContainer();
    let fail = false;
    const tree = createTree(container, async (p) => {
        if (fail) throw new Error('boom');
        return {path: p, size: 1, total: 1, entries: data['/r']};
    });
    await tree.load('/r');
    fail = true;
    container.querySelector('.row[data-path="/r/sub"]')
        .dispatchEvent(new dom.window.MouseEvent('click', {bubbles: true}));
    await new Promise((r) => setTimeout(r, 0));
    const row = container.querySelector('.row[data-path="/r/sub"]');
    assert.ok(row.classList.contains('error'));
    assert.ok(!row.classList.contains('loading'));
});
