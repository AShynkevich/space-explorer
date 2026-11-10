import test from 'node:test';
import assert from 'node:assert/strict';
import {formatSize, formatPercent, formatCount, formatDuration, barColor} from './format.js';

test('formatSize picks B/KB/MB/GB/TB', () => {
    assert.equal(formatSize(0), '0 B');
    assert.equal(formatSize(-5), '0 B');
    assert.equal(formatSize(NaN), '0 B');
    assert.equal(formatSize(512), '512 B');
    assert.equal(formatSize(1023), '1023 B');
    assert.equal(formatSize(1024), '1.0 KB');
    assert.equal(formatSize(1536), '1.5 KB');
    assert.equal(formatSize(5 * 1024 ** 2), '5.0 MB');
    assert.equal(formatSize(1.5 * 1024 ** 3), '1.5 GB');
    assert.equal(formatSize(2 * 1024 ** 4), '2.0 TB');
});

test('formatSize never prints 1024.0 of a unit', () => {
    assert.equal(formatSize(1024 ** 2 - 1), '1.0 MB');
    assert.equal(formatSize(1024 ** 3 - 10), '1.0 GB');
});

test('formatPercent', () => {
    assert.equal(formatPercent(0), '0%');
    assert.equal(formatPercent(undefined), '0%');
    assert.equal(formatPercent(0.04), '<0.1%');
    assert.equal(formatPercent(12.34), '12.3%');
    assert.equal(formatPercent(99.97), '100%');
    assert.equal(formatPercent(100), '100%');
});

test('formatCount groups thousands', () => {
    assert.equal(formatCount(1234567), '1,234,567');
    assert.equal(formatCount(undefined), '0');
});

test('formatDuration', () => {
    assert.equal(formatDuration(250), '250 ms');
    assert.equal(formatDuration(16640), '16.6 s');
    assert.equal(formatDuration(125000), '2 min 5 s');
});

test('barColor goes green -> red and clamps', () => {
    assert.equal(barColor(0), 'hsl(130 70% 48%)');
    assert.equal(barColor(25), 'hsl(65 70% 48%)');
    assert.equal(barColor(50), 'hsl(0 70% 48%)');
    assert.equal(barColor(100), 'hsl(0 70% 48%)');
    assert.equal(barColor(-3), barColor(0));
    assert.equal(barColor(undefined), barColor(0));
});
