import { Marked } from 'marked'

export function escapeHtml(raw) {
  return String(raw || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

export function sanitizeHttpUrl(raw) {
  const value = String(raw || '').trim()
  if (!value) return ''
  if (value.startsWith('https://') || value.startsWith('http://')) {
    return value
  }
  return ''
}

function createMarkdownParser() {
  const parser = new Marked({
    breaks: true,
    gfm: true
  })

  parser.use({
    renderer: {
      html({ text }) {
        return escapeHtml(text)
      },
      link({ href, title, text, tokens }) {
        const cleanHref = sanitizeHttpUrl(href)
        const label = Array.isArray(tokens) && tokens.length > 0
          ? this.parser.parseInline(tokens)
          : escapeHtml(text || href || '')
        if (!cleanHref) {
          return label
        }
        const titleAttr = title ? ` title="${escapeHtml(title)}"` : ''
        return `<a href="${escapeHtml(cleanHref)}" target="_blank" rel="noopener noreferrer"${titleAttr}>${label}</a>`
      },
      image({ href, title, text }) {
        const cleanSrc = sanitizeHttpUrl(href)
        if (!cleanSrc) {
          return escapeHtml(text || '')
        }
        const altAttr = escapeHtml(text || '')
        const titleAttr = title ? ` title="${escapeHtml(title)}"` : ''
        return `<img src="${escapeHtml(cleanSrc)}" alt="${altAttr}"${titleAttr} loading="lazy" />`
      }
    }
  })

  return parser
}

const markdownParser = createMarkdownParser()

export function isHtmlArticle(item) {
  return item?.render_type === 'html' || item?.content_type === 'html'
}

export function imageAttachments(item) {
  if (!Array.isArray(item?.attachments)) return []
  const rawContent = String(item?.content || '')
  return item.attachments
    .filter((att) => att && (att.type === 'image' || !att.type))
    .map((att) => ({ ...att, url: sanitizeHttpUrl(att.url) }))
    .filter((att) => Boolean(att.url) && !rawContent.includes(att.url))
}

export function renderMarkdown(text) {
  if (!text) return ''
  return markdownParser.parse(String(text))
}

export function sanitizeRichHtml(raw) {
  if (!raw) return ''
  return String(raw)
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/\son[a-z]+\s*=\s*(['"]).*?\1/gi, '')
    .replace(/javascript:/gi, '')
}

function scanBracketed(text, startIdx) {
  if (text[startIdx] !== '[') return -1
  let depth = 0
  for (let i = startIdx; i < text.length; i++) {
    const ch = text[i]
    if (ch === '\\' && i + 1 < text.length) {
      i++
      continue
    }
    if (ch === '[') depth++
    else if (ch === ']') {
      depth--
      if (depth === 0) return i
    }
  }
  return -1
}

function scanParenthesized(text, startIdx) {
  if (text[startIdx] !== '(') return -1
  let depth = 0
  for (let i = startIdx; i < text.length; i++) {
    const ch = text[i]
    if (ch === '\\' && i + 1 < text.length) {
      i++
      continue
    }
    if (ch === '(') depth++
    else if (ch === ')') {
      depth--
      if (depth === 0) return i
    }
  }
  return -1
}

function parseLinkTarget(rawInsideParens) {
  let s = String(rawInsideParens || '').trim()
  if (!s) return ''
  if (s.startsWith('<')) {
    const endAngle = s.indexOf('>')
    if (endAngle > 1) {
      return s.slice(1, endAngle).trim()
    }
  }
  const spaceIdx = s.search(/\s/)
  if (spaceIdx > 0) {
    s = s.slice(0, spaceIdx).trim()
  }
  return s
}

export function stripMarkdownLinks(raw) {
  const text = String(raw || '')
  if (!text) return ''
  let out = ''
  for (let i = 0; i < text.length; ) {
    if (text[i] === '!' && i + 1 < text.length && text[i + 1] === '[') {
      const labelEnd = scanBracketed(text, i + 1)
      if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
        const urlEnd = scanParenthesized(text, labelEnd + 1)
        if (urlEnd !== -1) {
          const alt = text.slice(i + 2, labelEnd).trim()
          if (alt) out += alt
          i = urlEnd + 1
          continue
        }
      }
    }
    if (text[i] === '[') {
      const labelEnd = scanBracketed(text, i)
      if (labelEnd !== -1 && labelEnd + 1 < text.length && text[labelEnd + 1] === '(') {
        const urlEnd = scanParenthesized(text, labelEnd + 1)
        if (urlEnd !== -1) {
          const inner = stripMarkdownLinks(text.slice(i + 1, labelEnd))
          const url = parseLinkTarget(text.slice(labelEnd + 2, urlEnd))
          out += inner || url
          i = urlEnd + 1
          continue
        }
      }
    }
    out += text[i]
    i++
  }
  return out
}

export function stripMarkdownPlain(raw) {
  const normalized = stripMarkdownLinks(String(raw || '').replace(/\r\n?/g, '\n'))
  const lines = normalized.split('\n').map((line) => {
    const s = line
      .replace(/^\s*#{1,6}\s+/, '')
      .replace(/^\s*(?:>\s*)+/, '')
      .replace(/^\s*[-*+]\s+/, '')
      .replace(/\\([\\`*_{}[\]()#+\-.!~|>])/g, '$1')
      .replace(/`([^`]*)`/g, '$1')
      .replace(/\*\*([^*\n]+)\*\*/g, '$1')
      .replace(/__([^_\n]+)__/g, '$1')
      .replace(/~~([^~\n]+)~~/g, '$1')
      .replace(/\|\|([^|\n]+)\|\|/g, '$1')
      .replace(/(^|[\s([{\u2018\u201c\u300c\u300e\uff08])\*([^*\n]+?)\*(?=$|[\s)\]}\u2019\u201d\u300d\u300f\uff09.,!?;:\u3002\uff0c\uff1b\uff1a])/g, '$1$2')
      .replace(/(^|[\s([{\u2018\u201c\u300c\u300e\uff08])_([^_\n]+?)_(?=$|[\s)\]}\u2019\u201d\u300d\u300f\uff09.,!?;:\u3002\uff0c\uff1b\uff1a])/g, '$1$2')
      .trim()
    return s.replace(/\s{2,}/g, ' ')
  })
  return lines.filter(Boolean).join(' ').replace(/\s{2,}/g, ' ').trim()
}

export function shouldShowSummary(item, config) {
  if (!config?.show_summary) return false
  const sum = stripMarkdownPlain(item?.summary || '')
  const body = stripMarkdownPlain(item?.content || '')
  return Boolean(sum && body && sum !== body)
}

export function formatDateTime(iso, browserLocaleTag) {
  if (!iso) return ''
  const dt = new Date(iso)
  if (Number.isNaN(dt.getTime())) return ''
  try {
    return dt
      .toLocaleString(browserLocaleTag || 'en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      })
      .replace(/\//g, '-')
  } catch {
    return iso
  }
}

