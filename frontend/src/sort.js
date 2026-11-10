// Clickable column headers: choose how folder contents are ordered.
export function initSort(onChange) {
    let mode = 'size';
    const heads = document.querySelectorAll('.columns [data-sort]');
    const paint = () => heads.forEach((h) => {
        const active = h.dataset.sort === mode;
        h.classList.toggle('active', active);
        h.setAttribute('aria-pressed', String(active));
    });
    heads.forEach((h) => h.addEventListener('click', () => {
        if (h.dataset.sort === mode) return;
        mode = h.dataset.sort;
        paint();
        onChange(mode);
    }));
    paint();
    return {mode: () => mode};
}
