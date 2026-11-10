// Wiring only: connects the backend (api.js) to the UI modules.
import './style.css';
import {api, events} from './api.js';
import {$} from './dom.js';
import {createTree} from './tree.js';
import {renderVolumes} from './volumes.js';
import {initToolbar, setScanning, setSelection} from './toolbar.js';
import {initSort} from './sort.js';
import {createErrorsPanel} from './errors.js';
import {status} from './status.js';

const sort = initSort(() => tree.refresh().catch(status.failed));
const tree = createTree($('tree'), (path) => api.children(path, sort.mode()));
const errorsPanel = createErrorsPanel({fetchErrors: api.scanErrors, onPick: api.reveal});
let keepTreeOnStop = false; // a cancelled/failed sub-folder rescan leaves the old tree usable

function begin(keepTree) {
    keepTreeOnStop = keepTree;
    setScanning(true);
    if (!keepTree) {
        tree.showMessage('Scanning...');
        errorsPanel.update(0);
    }
}

function end(showStatus) {
    setScanning(false);
    showStatus();
}

function startScan(root) {
    if (!root) return;
    begin(false);
    api.startScan(root).catch((err) => end(() => status.failed(err)));
}

function rescanSelected() {
    const sel = tree.selected();
    if (sel?.kind !== 'dir') return;
    begin(sel.path !== tree.rootPath());
    api.rescan(sel.path).catch((err) => end(() => status.failed(err)));
}

events.progress((p) => status.scanning(p));
events.done(async (summary) => {
    end(() => status.done(summary));
    errorsPanel.update(summary.errorCount);
    try {
        if (summary.rescanned) await tree.refresh();
        else await tree.load(summary.root);
    } catch (err) {
        status.failed(err);
    }
});
events.cancelled(() => {
    end(status.cancelled);
    if (!keepTreeOnStop) tree.showMessage('Scan cancelled.');
});
events.failed((msg) => {
    end(() => status.failed(msg));
    if (!keepTreeOnStop) tree.showMessage('Scan failed.');
});

tree.onSelect(setSelection);
initToolbar({
    onPick: async () => startScan(await api.pickFolder()),
    onCancel: () => api.cancelScan(),
    onRescan: rescanSelected,
    onReveal: () => tree.selected() && api.reveal(tree.selected().path).catch(status.failed),
});

api.volumes().then((vs) => renderVolumes($('volumes'), vs, startScan));
tree.showMessage('Pick a disk above or choose a folder to see where the space went.');
status.idle();
