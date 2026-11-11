<template>
  <div v-if="forbidden" class="forbidden-shell">
    <div class="forbidden-card">
      <div class="forbidden-code">403</div>
      <h1 class="forbidden-title">{{ t('newspaper.forbidden_title', 'Newspaper Access Denied') }}</h1>
      <p class="forbidden-desc">{{ forbiddenMessage || t('newspaper.forbidden_desc', 'Access restricted') }}</p>
      <div class="forbidden-actions">
        <a href="/dashboard" class="forbidden-btn primary">
          {{ t('newspaper.go_dashboard', 'Open Dashboard') }}
        </a>
        <button type="button" class="forbidden-btn" @click="signOutAndRedirect">
          {{ t('newspaper.sign_out', 'Sign Out') }}
        </button>
      </div>
    </div>
  </div>

  <div v-else class="newspaper-shell" :class="'theme-' + activeTheme">
    <component
      :is="activeThemeComponent"
      :masthead-title="mastheadTitle"
      :masthead-subtitle="mastheadSubtitle"
      :footer-publisher="footerPublisher"
      :formatted-current-date="formattedCurrentDate"
      :last-updated-text="lastUpdatedText"
      :news-list="newsList"
      :grouped-news="groupedNews"
      :sources="sources"
      :selected-source="selectedSource"
      :config="newsConfig"
      :languages="languages"
      :current-locale="currentLocale"
      :active-theme="activeTheme"
      :loading="loading"
      :error="error"
      :browser-locale-tag="browserLocaleTag"
      :t="t"
      @select-source="selectSource"
      @change-locale="changeLocale"
      @switch-theme="switchTheme"
      @retry="fetchNews"
    />
  </div>
</template>

<script>
import ThemeClassic from './themes/ThemeClassic.vue'
import ThemeModern from './themes/ThemeModern.vue'
import ThemeFeed from './themes/ThemeFeed.vue'

const ACCESS_KEY = 'utd_access_token'
const REFRESH_KEY = 'utd_refresh_token'
const USER_KEY = 'utd_user'
const THEME_STORAGE_KEY = 'utd_newspaper_theme'
const VALID_THEMES = new Set(['classic', 'modern', 'feed'])

function getStoredToken(key) {
  return localStorage.getItem(key) || sessionStorage.getItem(key) || ''
}

function saveRefreshedTokens(data) {
  const store = localStorage.getItem(REFRESH_KEY) ? localStorage : sessionStorage
  if (data.token) store.setItem(ACCESS_KEY, data.token)
  if (data.refresh_token) store.setItem(REFRESH_KEY, data.refresh_token)
  if (data.user) store.setItem(USER_KEY, JSON.stringify(data.user))
}

function clearStoredAuth() {
  ;[localStorage, sessionStorage].forEach((store) => {
    store.removeItem(ACCESS_KEY)
    store.removeItem(REFRESH_KEY)
    store.removeItem(USER_KEY)
  })
}

export default {
  name: 'NewspaperApp',
  components: { ThemeClassic, ThemeModern, ThemeFeed },
  data() {
    return {
      loading: true,
      error: '',
      forbidden: false,
      forbiddenMessage: '',
      siteInfo: {
        site_name: 'Utopia Daily',
        description: '',
        version: __APP_VERSION__,
        default_locale: 'en_US',
        newspaper_title: 'Utopia Daily',
        newspaper_subtitle: 'Daily News Digest',
        newspaper_visibility: 'public'
      },
      newsConfig: {
        title: '',
        subtitle: '',
        lookback_days: 3,
        max_items: 100,
        theme: 'classic',
        sort: 'newest',
        show_tags: true,
        show_summary: false,
        show_author: true,
        show_source_filter: true,
        allow_theme_switch: true
      },
      themeOverride: '',
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
    activeTheme() {
      if (this.newsConfig.allow_theme_switch && VALID_THEMES.has(this.themeOverride)) {
        return this.themeOverride
      }
      if (VALID_THEMES.has(this.newsConfig.theme)) {
        return this.newsConfig.theme
      }
      return 'classic'
    },
    activeThemeComponent() {
      if (this.activeTheme === 'modern') return 'ThemeModern'
      if (this.activeTheme === 'feed') return 'ThemeFeed'
      return 'ThemeClassic'
    },
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
    const savedTheme = (localStorage.getItem(THEME_STORAGE_KEY) || '').toLowerCase()
    if (VALID_THEMES.has(savedTheme)) {
      this.themeOverride = savedTheme
    }
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
    switchTheme(mode) {
      if (!VALID_THEMES.has(mode)) return
      this.themeOverride = mode
      localStorage.setItem(THEME_STORAGE_KEY, mode)
    },
    async selectSource(source) {
      if (this.selectedSource === source && !this.error) return
      this.selectedSource = source
      await this.fetchNews()
    },
    async tryRefreshAccessToken() {
      const refreshToken = getStoredToken(REFRESH_KEY)
      if (!refreshToken) return false
      try {
        const res = await fetch('/api/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken })
        })
        if (!res.ok) return false
        const data = await res.json()
        if (!data || !data.token) return false
        saveRefreshedTokens(data)
        return true
      } catch {
        return false
      }
    },
    async requestLatestOnce(query) {
      const headers = {}
      const token = getStoredToken(ACCESS_KEY)
      if (token) {
        headers.Authorization = `Bearer ${token}`
      }
      return fetch(`/api/newspaper/latest${query}`, { headers })
    },
    async fetchNews() {
      this.loading = true
      this.error = ''
      this.forbidden = false
      this.forbiddenMessage = ''

      try {
        const query = this.selectedSource
          ? `?source=${encodeURIComponent(this.selectedSource)}`
          : ''

        let res = await this.requestLatestOnce(query)
        if (res.status === 401) {
          const refreshed = await this.tryRefreshAccessToken()
          if (refreshed) {
            res = await this.requestLatestOnce(query)
          }
          if (res.status === 401) {
            clearStoredAuth()
            window.location.replace('/dashboard?redirect=/')
            return
          }
        }

        if (res.status === 403) {
          this.forbidden = true
          try {
            const payload = await res.json()
            if (payload && payload.code) {
              this.forbiddenMessage = this.t(`error.${payload.code}`, this.t('newspaper.forbidden_desc', 'Access denied'))
            }
          } catch {
            // keep default forbidden description
          }
          return
        }

        if (!res.ok) {
          let errText = `HTTP ${res.status}`
          try {
            const payload = await res.json()
            if (payload && payload.code) {
              errText = this.t(`error.${payload.code}`, payload.code)
            }
          } catch {
            // keep status text
          }
          throw new Error(errText)
        }

        const data = await res.json()
        if (data.config) {
          this.newsConfig = { ...this.newsConfig, ...data.config }
        }
        if (Array.isArray(data.sources)) {
          this.sources = data.sources
        }
        this.newsList = Array.isArray(data.news) ? data.news : []
        this.lastFetchedAt = new Date()
      } catch (err) {
        this.error = err?.message || this.t('newspaper.load_error', 'Failed to load articles')
      } finally {
        this.loading = false
      }
    },
    async signOutAndRedirect() {
      const refreshToken = getStoredToken(REFRESH_KEY)
      const accessToken = getStoredToken(ACCESS_KEY)
      if (refreshToken) {
        try {
          const headers = { 'Content-Type': 'application/json' }
          if (accessToken) headers.Authorization = `Bearer ${accessToken}`
          await fetch('/api/auth/logout', {
            method: 'POST',
            headers,
            body: JSON.stringify({ refresh_token: refreshToken })
          })
        } catch {
          // ignore logout error
        }
      }
      clearStoredAuth()
      window.location.replace('/dashboard?redirect=/')
    }
  }
}
</script>
