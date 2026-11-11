export function escapeHtml(raw) {
    return String(raw || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

export function sanitizeHttpUrl(raw) {
    const val = String(raw || '').trim();
    if (!val) return '';
    if (val.startsWith('https://') || val.startsWith('http://')) {
        return val;
    }
    return '';
}

export function sanitizeRichHtml(raw) {
    if (!raw) return '';
    return String(raw)
        .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
        .replace(/\son[a-z]+\s*=\s*(['"]).*?\1/gi, '')
        .replace(/javascript:/gi, '');
}

function scanBracketed(text, startIdx) {
    if (text[startIdx] !== '[') return -1;
    let depth = 0;
    for (let i = startIdx; i < text.length; i++) {
        const ch = text[i];
        if (ch === '\\' && i + 1 < text.length) {
            i++;
            continue;
        }
        if (ch === '[') depth++;
        else if (ch === ']') {
            depth--;
            if (depth === 0) return i;
        }
    }
    return -1;
}

function scanParenthesized(text, startIdx) {
    if (text[startIdx] !== '(') return -1;
    let depth = 0;
    for (let i = startIdx; i < text.length; i++) {
        const ch = text[i];
        if (ch === '\\' && i + 1 < text.length) {
            i++;
            continue;
        }
        if (ch === '(') depth++;
        else if (ch === ')') {
            depth--;
            if (depth === 0) return i;
        }
    }
    return -1;
}

function parseLinkTarget(rawInsideParens) {
    let s = String(rawInsideParens || '').trim();
    if (!s) return '';
    if (s.startsWith('<')) {
        const endAngle = s.indexOf('>');
        if (endAngle > 1) {
            return s.slice(1, endAngle).trim();
        }
    }
    const spaceIdx = s.search(/\s/);
    if (spaceIdx > 0) {
        s = s.slice(0, spaceIdx).trim();
    }
    return s;
}

export function stripMarkdownLinks(raw) {
    const text = String(raw || '');
    if (!text) return '';
    let out = '';
    for (let i = 0; i < text.length; ) {
        if (text[i] === '!' && i + 1 < text.length && text[i + 1] === '[') {
            const labelEnd = scanBracketed(text, i + 1);
            if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
                const urlEnd = scanParenthesized(text, labelEnd + 1);
                if (urlEnd !== -1) {
                    const alt = text.slice(i + 2, labelEnd).trim();
                    if (alt) out += alt;
                    i = urlEnd + 1;
                    continue;
                }
            }
        }
        if (text[i] === '[') {
            const labelEnd = scanBracketed(text, i);
            if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
                const urlEnd = scanParenthesized(text, labelEnd + 1);
                if (urlEnd !== -1) {
                    const inner = stripMarkdownLinks(text.slice(i + 1, labelEnd));
                    const url = parseLinkTarget(text.slice(labelEnd + 2, urlEnd));
                    out += inner || url;
                    i = urlEnd + 1;
                    continue;
                }
            }
        }
        out += text[i];
        i++;
    }
    return out;
}

export function stripMarkdownPlain(raw) {
    const normalized = stripMarkdownLinks(String(raw || '').replace(/\r\n?/g, '\n'));
    const lines = normalized.split('\n').map(line => {
        let s = line
            .replace(/^\s*#{1,6}\s+/, '')
            .replace(/^\s*(?:>\s*)+/, '')
            .replace(/^\s*[-*+]\s+/, '')
            .replace(/\\([\\`*_{}[\]()#+\-.!~|>])/g, '$1')
            .replace(/`([^`]*)`/g, '$1')
            .replace(/\*\*([^*\n]+)\*\*/g, '$1')
            .replace(/__([^_\n]+)__/g, '$1')
            .replace(/~~([^~\n]+)~~/g, '$1')
            .replace(/\|\|([^\|\n]+)\|\|/g, '$1')
            .replace(/(^|[\s([{\u2018\u201c\u300c\u300e\uff08])\*([^*\n]+?)\*(?=$|[\s)\]}\u2019\u201d\u300d\u300f\uff09.,!?;:\u3002\uff0c\uff1b\uff1a])/g, '$1$2')
            .replace(/(^|[\s([{\u2018\u201c\u300c\u300e\uff08])_([^_\n]+?)_(?=$|[\s)\]}\u2019\u201d\u300d\u300f\uff09.,!?;:\u3002\uff0c\uff1b\uff1a])/g, '$1$2')
            .trim();
        return s.replace(/\s{2,}/g, ' ');
    });
    return lines.filter(Boolean).join(' ').replace(/\s{2,}/g, ' ').trim();
}

function findInlineClose(text, startIdx, marker, wordBounded = false) {
    for (let i = startIdx; i <= text.length - marker.length; i++) {
        if (text[i] === '\\' && i + 1 < text.length) {
            i++;
            continue;
        }
        if (text.slice(i, i + marker.length) === marker) {
            if (i === startIdx) continue;
            if (text[i - 1] === ' ' && marker.length === 1) continue;
            if (wordBounded) {
                const nextCh = text[i + marker.length] || '';
                if (nextCh && /[A-Za-z0-9_]/.test(nextCh)) continue;
            }
            return i;
        }
    }
    return -1;
}

function renderInline(raw) {
    const text = String(raw || '');
    let out = '';
    for (let i = 0; i < text.length; ) {
        const ch = text[i];
        if (ch === '\\' && i + 1 < text.length && /[\\`*_{}[\]()#+\-.!~|>]/.test(text[i + 1])) {
            out += escapeHtml(text[i + 1]);
            i += 2;
            continue;
        }
        if (ch === '`') {
            const end = text.indexOf('`', i + 1);
            if (end !== -1) {
                out += `<code class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 font-mono text-xs">${escapeHtml(text.slice(i + 1, end))}</code>`;
                i = end + 1;
                continue;
            }
        }
        if (ch === '!' && i + 1 < text.length && text[i + 1] === '[') {
            const labelEnd = scanBracketed(text, i + 1);
            if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
                const urlEnd = scanParenthesized(text, labelEnd + 1);
                if (urlEnd !== -1) {
                    const alt = text.slice(i + 2, labelEnd).trim();
                    const href = sanitizeHttpUrl(parseLinkTarget(text.slice(labelEnd + 2, urlEnd)));
                    if (href) {
                        out += `<img src="${escapeHtml(href)}" alt="${escapeHtml(alt)}" loading="lazy" class="my-2 max-w-full rounded border border-slate-200 dark:border-slate-700" />`;
                    } else if (alt) {
                        out += escapeHtml(alt);
                    }
                    i = urlEnd + 1;
                    continue;
                }
            }
        }
        if (ch === '[') {
            const labelEnd = scanBracketed(text, i);
            if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
                const urlEnd = scanParenthesized(text, labelEnd + 1);
                if (urlEnd !== -1) {
                    const labelRaw = text.slice(i + 1, labelEnd);
                    const href = sanitizeHttpUrl(parseLinkTarget(text.slice(labelEnd + 2, urlEnd)));
                    const labelHtml = renderInline(labelRaw || href);
                    if (href) {
                        out += `<a href="${escapeHtml(href)}" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline break-all">${labelHtml}</a>`;
                    } else {
                        out += labelHtml;
                    }
                    i = urlEnd + 1;
                    continue;
                }
            }
        }
        if (ch === '<' && (text.startsWith('<https://', i) || text.startsWith('<http://', i))) {
            const close = text.indexOf('>', i + 1);
            if (close !== -1) {
                const href = sanitizeHttpUrl(text.slice(i + 1, close));
                if (href) {
                    out += `<a href="${escapeHtml(href)}" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline break-all">${escapeHtml(href)}</a>`;
                    i = close + 1;
                    continue;
                }
            }
        }
        if (text.startsWith('**', i)) {
            const end = findInlineClose(text, i + 2, '**');
            if (end !== -1) {
                out += `<strong class="font-semibold text-slate-900 dark:text-white">${renderInline(text.slice(i + 2, end))}</strong>`;
                i = end + 2;
                continue;
            }
        }
        if (text.startsWith('__', i)) {
            const end = findInlineClose(text, i + 2, '__');
            if (end !== -1) {
                out += `<strong class="font-semibold text-slate-900 dark:text-white">${renderInline(text.slice(i + 2, end))}</strong>`;
                i = end + 2;
                continue;
            }
        }
        if (text.startsWith('~~', i)) {
            const end = findInlineClose(text, i + 2, '~~');
            if (end !== -1) {
                out += `<del>${renderInline(text.slice(i + 2, end))}</del>`;
                i = end + 2;
                continue;
            }
        }
        if (text.startsWith('||', i)) {
            const end = findInlineClose(text, i + 2, '||');
            if (end !== -1) {
                out += `<span class="px-1 rounded bg-slate-200 dark:bg-slate-800">${renderInline(text.slice(i + 2, end))}</span>`;
                i = end + 2;
                continue;
            }
        }
        if (ch === '*') {
            const end = findInlineClose(text, i + 1, '*');
            if (end !== -1 && text[i + 1] !== ' ') {
                out += `<em>${renderInline(text.slice(i + 1, end))}</em>`;
                i = end + 1;
                continue;
            }
        }
        if (ch === '_') {
            const prevCh = i > 0 ? text[i - 1] : '';
            if (!prevCh || !/[A-Za-z0-9_]/.test(prevCh)) {
                const end = findInlineClose(text, i + 1, '_', true);
                if (end !== -1 && text[i + 1] !== ' ') {
                    out += `<em>${renderInline(text.slice(i + 1, end))}</em>`;
                    i = end + 1;
                    continue;
                }
            }
        }
        if (text.startsWith('https://', i) || text.startsWith('http://', i)) {
            const prevCh = i > 0 ? text[i - 1] : '';
            if (!prevCh || /[\s([{\u2018\u201c\u300c\u300e\uff08>]/.test(prevCh)) {
                const rest = text.slice(i);
                const match = rest.match(/^https?:\/\/[^\s<>\u3000]+/);
                if (match) {
                    let rawUrl = match[0];
                    let suffix = '';
                    while (rawUrl.length > 8 && /[.,;:!?)\]}'"\u3002\uff0c\uff1b\uff1a\uff01\uff1f\uff09\u300d\u300f]$/.test(rawUrl)) {
                        if (rawUrl.endsWith(')') && (rawUrl.match(/\(/g) || []).length >= (rawUrl.match(/\)/g) || []).length) {
                            break;
                        }
                        suffix = rawUrl.slice(-1) + suffix;
                        rawUrl = rawUrl.slice(0, -1);
                    }
                    const href = sanitizeHttpUrl(rawUrl);
                    if (href) {
                        out += `<a href="${escapeHtml(href)}" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline break-all">${escapeHtml(href)}</a>${escapeHtml(suffix)}`;
                        i += match[0].length;
                        continue;
                    }
                }
            }
        }
        out += escapeHtml(ch);
        i++;
    }
    return out;
}

export function renderMarkdown(raw) {
    if (!raw) return '';
    const lines = String(raw).replace(/\r\n?/g, '\n').split('\n');
    const blocks = [];

    for (let i = 0; i < lines.length; ) {
        const line = lines[i];
        const trimmed = line.trim();
        if (!trimmed) {
            i++;
            continue;
        }

        if (trimmed.startsWith('```')) {
            const codeLines = [];
            i++;
            while (i < lines.length && !lines[i].trim().startsWith('```')) {
                codeLines.push(lines[i]);
                i++;
            }
            if (i < lines.length) i++;
            blocks.push(
                `<pre class="p-3 rounded-md bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 overflow-x-auto text-xs font-mono leading-relaxed"><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`
            );
            continue;
        }

        if (/^([-*_]){3,}$/.test(trimmed)) {
            blocks.push('<hr class="border-slate-200 dark:border-slate-800 my-3" />');
            i++;
            continue;
        }

        const headingMatch = line.match(/^\s{0,3}(#{1,6})\s+(.+)$/);
        if (headingMatch) {
            const level = Math.min(6, Math.max(1, headingMatch[1].length));
            const sizeClass = level <= 2 ? 'text-base font-semibold' : 'text-sm font-semibold';
            blocks.push(`<h${level} class="${sizeClass} text-slate-900 dark:text-white mt-2 mb-1">${renderInline(headingMatch[2].trim())}</h${level}>`);
            i++;
            continue;
        }

        if (/^\s*>\s?/.test(line)) {
            const quoteLines = [];
            while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
                quoteLines.push(lines[i].replace(/^\s*>\s?/, ''));
                i++;
            }
            blocks.push(
                `<blockquote class="pl-3 py-1 border-l-2 border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 space-y-1">${renderMarkdown(quoteLines.join('\n'))}</blockquote>`
            );
            continue;
        }

        if (/^\s*[-*+]\s+/.test(line)) {
            const items = [];
            while (i < lines.length && /^\s*[-*+]\s+/.test(lines[i])) {
                items.push(`<li>${renderInline(lines[i].replace(/^\s*[-*+]\s+/, ''))}</li>`);
                i++;
            }
            blocks.push(`<ul class="list-disc pl-5 space-y-1">${items.join('')}</ul>`);
            continue;
        }

        if (/^\s*\d+\.\s+/.test(line)) {
            const items = [];
            while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
                items.push(`<li>${renderInline(lines[i].replace(/^\s*\d+\.\s+/, ''))}</li>`);
                i++;
            }
            blocks.push(`<ol class="list-decimal pl-5 space-y-1">${items.join('')}</ol>`);
            continue;
        }

        const paraLines = [];
        while (
            i < lines.length &&
            lines[i].trim() &&
            !lines[i].trim().startsWith('```') &&
            !/^([-*_]){3,}$/.test(lines[i].trim()) &&
            !/^\s{0,3}#{1,6}\s+/.test(lines[i]) &&
            !/^\s*>\s?/.test(lines[i]) &&
            !/^\s*[-*+]\s+/.test(lines[i]) &&
            !/^\s*\d+\.\s+/.test(lines[i])
        ) {
            paraLines.push(renderInline(lines[i]));
            i++;
        }
        blocks.push(`<p class="leading-relaxed">${paraLines.join('<br />')}</p>`);
    }

    return blocks.join('');
}

