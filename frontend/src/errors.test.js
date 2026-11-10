import test from 'node:test';
import assert from 'node:assert/strict';
import {noteFor} from './errors.js';

test('noteFor adds the macOS hint for "operation not permitted"', () => {
    const note = noteFor([{path: '/x', err: 'open /x: operation not permitted'}], 1);
    assert.match(note, /Full Disk Access/);
});

test('noteFor says when the list is truncated', () => {
    const items = Array.from({length: 500}, (_, i) => ({path: `/p${i}`, err: 'permission denied'}));
    const note = noteFor(items, 1432);
    assert.match(note, /first 500/);
    assert.doesNotMatch(note, /Full Disk Access/);
});

test('noteFor is empty when nothing special', () => {
    assert.equal(noteFor([{path: '/x', err: 'permission denied'}], 1), '');
});
