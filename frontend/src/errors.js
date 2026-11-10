// Panel listing the paths the scanner could not read.
import {$, el} from './dom.js';
import {formatCount} from './format.js';

const MAC_HINT = 'macOS blocks some folders. Grant Space Explorer "Full Disk Access" in ' +
    'System Settings > Privacy & Security to include them, then rescan.';

export function createErrorsPanel({fetchErrors, onPick}) {
    const panel = $('errors-panel');
    const button = $('errors-btn');
    let total = 0;

    async function show() {
        const items = await fetchErrors();
        $('errors-title').textContent = `Unreadable paths (${formatCount(total)})`;
        $('errors-note').textContent = noteFor(items, total);
        $('errors-list').replaceChildren(...items.map((e) => item(e, onPick)));
        panel.hidden = false;
    }

    function hide() {
        panel.hidden = true;
    }

    button.onclick = () => (panel.hidden ? show() : hide());
    $('errors-close').onclick = hide;

    return {
        // Called after every scan with the total number of unreadable paths.
        update(count) {
            total = count || 0;
            button.hidden = total === 0;
            button.textContent = `${formatCount(total)} unreadable`;
            if (total === 0) hide();
            else if (!panel.hidden) show();
        },
    };
}

export function noteFor(items, total) {
    const parts = [];
    if (items.some((e) => /operation not permitted/i.test(e.err))) parts.push(MAC_HINT);
    if (items.length < total) parts.push(`Showing the first ${formatCount(items.length)}.`);
    return parts.join(' ');
}

function item(e, onPick) {
    const li = el('li', 'err-item');
    li.title = 'Click to show in the file manager';
    li.append(el('span', 'err-path', e.path), el('span', 'err-msg', e.err));
    li.onclick = () => onPick(e.path);
    return li;
}
