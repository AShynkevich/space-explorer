// Disk list: one clickable card per volume with a used-space bar.
import {el} from './dom.js';
import {formatSize, barColor} from './format.js';

export function renderVolumes(container, volumes, onPick) {
    container.replaceChildren(...volumes.map((v) => card(v, onPick)));
}

function card(v, onPick) {
    const btn = el('button', 'volume');
    btn.title = v.path;
    btn.append(el('span', 'v-name', v.name));
    if (v.total > 0) {
        const used = v.total - v.free;
        const pct = (used / v.total) * 100;
        const bar = el('span', 'v-bar');
        const fill = el('span', 'fill');
        fill.style.width = `${pct}%`;
        fill.style.background = barColor(pct);
        bar.append(fill);
        btn.append(bar, el('span', 'v-info', `${formatSize(used)} used of ${formatSize(v.total)}`));
    } else {
        btn.append(el('span', 'v-info', v.path));
    }
    btn.onclick = () => onPick(v.path);
    return btn;
}
