// Tiny DOM helper: el('span', 'size', '1.5 MB').
export function el(tag, className, text) {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined) node.textContent = text;
    return node;
}

export const $ = (id) => document.getElementById(id);
