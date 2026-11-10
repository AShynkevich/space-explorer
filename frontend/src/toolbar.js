// Toolbar: folder picker, cancel, and the actions on the selected row.
import {$} from './dom.js';

let scanning = false;
let selection = null;

export function initToolbar({onPick, onCancel, onRescan, onReveal}) {
    $('pick').onclick = onPick;
    $('cancel').onclick = onCancel;
    $('rescan').onclick = onRescan;
    $('reveal').onclick = onReveal;
    update();
}

export function setScanning(on) {
    scanning = on;
    update();
}

// selection: {path, kind} of the selected tree row, or null.
export function setSelection(sel) {
    selection = sel;
    update();
}

function update() {
    $('cancel').hidden = !scanning;
    $('pick').disabled = scanning;
    $('rescan').disabled = scanning || selection?.kind !== 'dir';
    $('reveal').disabled = !selection;
    document.querySelectorAll('.volume').forEach((b) => { b.disabled = scanning; });
}
