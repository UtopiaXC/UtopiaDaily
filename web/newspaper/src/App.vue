<template>
  <div class="paper-container">
    <header class="paper-header">
      <div class="top-actions">
        <select
          v-if="languages.length > 0"
          :value="currentLocale"
          class="locale-select"
          aria-label="Language"
          @change="changeLocale($event.target.value)"
        >
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.name }}
          </option>
        </select>

        <a href="/dashboard" class="dashboard-btn">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z" />
          </svg>
          <span>{{ t('newspaper.dashboard_link', 'Dashboard') }}</span>
        </a>
      </div>

      <h1 class="masthead-title">{{ mastheadTitle }}</h1>

      <div class="date-bar">
        <span>{{ formattedCurrentDate }}</span>
        <span v-if="mastheadSubtitle">&middot; {{ mastheadSubtitle }}</span>
        <span v-if="lastUpdatedText">&middot; {{ t('newspaper.updated_at', 'Updated') }} {{ lastUpdatedText }}</span>
      </div>
    </header>

    <nav v-if="sources.length > 1" class="source-filter-bar" aria-label="Sources">
      <button
        type="button"
        class="filter-pill"
        :class="{ active: selectedSource === '' }"
        @click="selectSource('')"
      >
        {{ t('newspaper.filter_all', 'All Sources') }}
      </button>
      <button
        v-for="src in sources"
        :key="src"
        type="button"
        class="filter-pill"
        :class="{ active: selectedSource === src }"
        @click="selectSource(src)"
      >
        {{ src }}
      </button>
    </nav>

    <div v-if="loading" class="state-box">
      {{ t('newspaper.loading', 'Loading articles...') }}
    </div>

    <div v-else-if="error" class="state-box error">
      {{ error }}
    </div>

    <div v-else-if="newsList.length === 0" class="state-box">
      {{ t('newspaper.empty', 'No recent articles found') }}
    </div>

    <main v-else>
      <section v-for="(items, source) in groupedNews" :key="source" class="source-section">
        <h2 class="channel-title">
          <span>{{ t('newspaper.source_prefix', 'Source') }}: {{ source }}</span>
          <span class="channel-count">{{ items.length }}</span>
        </h2>

        <article v-for="item in items" :key="item.id" class="article-entry">
          <h3 class="news-title">{{ item.title }}</h3>

          <div v-if="item.tags && item.tags.length > 0" class="tags-container">
            <span v-for="(tag, idx) in item.tags" :key="idx" class="tag-badge">
              #{{ tag.tag }}
            </span>
          </div>

          <p class="article-meta">
            <span>{{ formatDateTime(item.datetime_released) }}</span>
            <span v-if="item.author && item.author !== source"> &middot; {{ item.author }}</span>
          </p>

          <blockquote v-if="item.quotation && item.quotation.content" class="article-quote">
            <p>
              <strong>{{ t('newspaper.quote', 'Quote') }} {{ item.quotation.from || source }}:</strong>
              {{ item.quotation.content }}
            </p>
          </blockquote>

          <div
            v-if="isHtmlArticle(item)"
            class="article-body"
            v-html="sanitizeRichHtml(item.content)"
          ></div>
          <div
            v-else
            class="article-body"
            v-html="renderMarkdown(item.content || item.summary || '')"
          ></div>

          <div v-if="!isHtmlArticle(item) && imageAttachments(item).length > 0" class="article-attachments">
            <img
              v-for="(att, idx) in imageAttachments(item)"
              :key="idx"
              :src="att.url"
              loading="lazy"
              :alt="item.title"
            />
          </div>

          <p v-if="safeUrl(item.from_url)" class="original-link-wrap">
            <a
              :href="safeUrl(item.from_url)"
              target="_blank"
              rel="noopener noreferrer"
              class="original-link"
            >
              <svg class="icon-svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M19 19H5V5h7V3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
              </svg>
              <span>{{ t('newspaper.original_link', 'Source Link') }}</span>
            </a>
          </p>

          <hr class="article-divider" />
        </article>
      </section>
    </main>

    <footer class="paper-footer">
      {{ t('newspaper.generated_by', 'Published by') }} {{ footerPublisher }} &middot; {{ formattedCurrentDate }}
    </footer>
  </div>
</template>

<script>
import { Marked } from 'marked'

function escapeHtml(raw) {
  return String(raw || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function sanitizeHttpUrl(raw) {
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

export default {
  name: 'NewspaperApp',
  data() {
    return {
      loading: true,
      error: '',
      siteInfo: {
        site_name: 'Utopia Daily',
        description: '',
        version: '2.0.0',
        default_locale: 'en_US',
        newspaper_title: 'Utopia Daily',
        newspaper_subtitle: 'Daily News Digest'
      },
      newsConfig: {
        title: '',
        subtitle: ''
      },
      languages: [],
      currentLocale: 'en_US',
      messages: {},
      sources: [],
      selectedSource: '',
      newsList: [],
      lastFetchedAt: null
    }
  },
  computed: {
    mastheadTitle() {
      return (
        this.newsConfig.title ||
        this.siteInfo.newspaper_title ||
        this.siteInfo.site_name ||
        'Utopia Daily'
      )
    },
    mastheadSubtitle() {
      return (
        this.newsConfig.subtitle ||
        this.siteInfo.newspaper_subtitle ||
        this.siteInfo.description ||
        ''
      )
    },
    footerPublisher() {
      return this.siteInfo.site_name || this.mastheadTitle || 'Utopia Daily'
    },
    browserLocaleTag() {
      return (this.currentLocale || 'en_US').replace('_', '-')
    },
    formattedCurrentDate() {
      try {
        return new Date().toLocaleDateString(this.browserLocaleTag, {
          year: 'numeric',
          month: 'long',
          day: 'numeric'
        })
      } catch {
        return new Date().toISOString().slice(0, 10)
      }
    },
    lastUpdatedText() {
      if (!this.lastFetchedAt) return ''
      try {
        return this.lastFetchedAt.toLocaleTimeString(this.browserLocaleTag, {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        })
      } catch {
        return ''
      }
    },
    groupedNews() {
      const groups = {}
      for (const item of this.newsList) {
        const key = item.source || item.module_id || 'Unknown'
        if (!groups[key]) {
          groups[key] = []
        }
        groups[key].push(item)
      }
      return groups
    }
  },
  watch: {
    mastheadTitle(newVal) {
      if (newVal) {
        document.title = newVal
      }
    }
  },
  async created() {
    await Promise.all([this.fetchSiteInfo(), this.fetchLocales()])
    this.resolveInitialLocale()
    await Promise.all([this.loadMessages(this.currentLocale), this.fetchNews()])
  },
  methods: {
    t(key, fallback) {
      return this.messages[key] || fallback || key
    },
    async fetchSiteInfo() {
      try {
        const res = await fetch('/api/common/site')
        if (res.ok) {
          this.siteInfo = { ...this.siteInfo, ...(await res.json()) }
          document.title = this.mastheadTitle
        }
      } catch {
        // retain defaults
      }
    },
    async fetchLocales() {
      try {
        const res = await fetch('/api/common/locales')
        if (res.ok) {
          const data = await res.json()
          this.languages = Array.isArray(data.languages) ? data.languages : []
          if (data.default_locale && !this.siteInfo.default_locale) {
            this.siteInfo.default_locale = data.default_locale
          }
          return
        }
      } catch {
        // fallback below
      }
      this.languages = [
        { code: 'en_US', name: 'English' },
        { code: 'zh_CN', name: '中文' }
      ]
    },
    resolveInitialLocale() {
      const available = new Set(this.languages.map((l) => l.code))
      const stored = localStorage.getItem('locale')
      if (stored && stored !== 'auto' && (available.size === 0 || available.has(stored))) {
        this.currentLocale = stored
        return
      }
      const nav = (navigator.language || '').toLowerCase()
      const guessed = nav.includes('zh') ? 'zh_CN' : (this.siteInfo.default_locale || 'en_US')
      if (available.size === 0 || available.has(guessed)) {
        this.currentLocale = guessed
        return
      }
      this.currentLocale = this.languages[0]?.code || 'en_US'
    },
    async loadMessages(locale) {
      const target = locale || 'en_US'
      document.documentElement.lang = target.replace('_', '-')
      try {
        const res = await fetch(`/api/common/i18n/${encodeURIComponent(target)}`)
        if (res.ok) {
          this.messages = await res.json()
        }
      } catch {
        // keep current messages
      }
    },
    async changeLocale(locale) {
      this.currentLocale = locale
      localStorage.setItem('locale', locale)
      await this.loadMessages(locale)
    },
    async selectSource(source) {
      if (this.selectedSource === source && !this.error) return
      this.selectedSource = source
      await this.fetchNews()
    },
    async fetchNews() {
      this.loading = true
      this.error = ''
      try {
        const query = this.selectedSource
          ? `?source=${encodeURIComponent(this.selectedSource)}`
          : ''
        const res = await fetch(`/api/newspaper/latest${query}`)
        if (!res.ok) {
          let errCode = `HTTP ${res.status}`
          try {
            const payload = await res.json()
            if (payload && payload.code) {
              errCode = this.t(`error.${payload.code}`, payload.code)
            }
          } catch {
            // keep status text
          }
          throw new Error(errCode)
        }
        const data = await res.json()
        if (data.config) {
          this.newsConfig = data.config
        }
        if (Array.isArray(data.sources)) {
          this.sources = data.sources
        }
        this.newsList = Array.isArray(data.news) ? data.news : []
        this.lastFetchedAt = new Date()
      } catch (err) {
        this.error = err?.message || 'Failed to load articles'
      } finally {
        this.loading = false
      }
    },
    isHtmlArticle(item) {
      return item?.render_type === 'html' || item?.content_type === 'html'
    },
    safeUrl(url) {
      return sanitizeHttpUrl(url)
    },
    imageAttachments(item) {
      if (!Array.isArray(item?.attachments)) return []
      const rawContent = String(item?.content || '')
      return item.attachments
        .filter((att) => att && (att.type === 'image' || !att.type))
        .map((att) => ({ ...att, url: sanitizeHttpUrl(att.url) }))
        .filter((att) => Boolean(att.url) && !rawContent.includes(att.url))
    },
    renderMarkdown(text) {
      if (!text) return ''
      return markdownParser.parse(String(text))
    },
    sanitizeRichHtml(raw) {
      if (!raw) return ''
      return String(raw)
        .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
        .replace(/\son[a-z]+\s*=\s*(['"]).*?\1/gi, '')
        .replace(/javascript:/gi, '')
    },
    formatDateTime(iso) {
      if (!iso) return ''
      const dt = new Date(iso)
      if (Number.isNaN(dt.getTime())) return ''
      try {
        return dt
          .toLocaleString(this.browserLocaleTag, {
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
  }
}
</script>
