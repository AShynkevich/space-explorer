// Keyboard navigation for the tree: arrows move, left/right collapse/expand.
import {depthOf, isExpanded} from './row.js';

export function attachKeyboard(container, t) {
    container.addEventListener('keydown', (ev) => {
        const row = t.selected();
        if (!row) return;
        const all = t.rows();
        const i = all.indexOf(row);
        switch (ev.key) {
            case 'ArrowDown': t.select(all[i + 1]); break;
            case 'ArrowUp': t.select(all[i - 1]); break;
            case 'Home': t.select(all[0]); break;
            case 'End': t.select(all[all.length - 1]); break;
            case 'ArrowRight':
                if (row.classList.contains('expandable') && !isExpanded(row)) t.expand(row);
                else if (isExpanded(row)) t.select(all[i + 1]);
                break;
            case 'ArrowLeft':
                if (isExpanded(row)) t.collapse(row);
                else t.select(parentOf(all, i));
                break;
            case 'Enter':
            case ' ':
                if (row.classList.contains('expandable')) t.toggle(row);
                break;
            default: return;
        }
        ev.preventDefault();
    });
}

// Nearest previous row with a smaller depth.
function parentOf(all, i) {
    const depth = depthOf(all[i]);
    for (let j = i - 1; j >= 0; j--) {
        if (depthOf(all[j]) < depth) return all[j];
    }
    return null;
}
