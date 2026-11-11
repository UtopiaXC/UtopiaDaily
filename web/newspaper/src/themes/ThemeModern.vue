<template>
  <div class="modern-shell">
    <header class="modern-topbar">
      <div class="modern-topbar-inner">
        <div class="modern-brand">
          <h1 class="modern-title">{{ mastheadTitle }}</h1>
          <div class="modern-subtitle-row">
            <span>{{ formattedCurrentDate }}</span>
            <span v-if="mastheadSubtitle">&middot; {{ mastheadSubtitle }}</span>
            <span v-if="lastUpdatedText">&middot; {{ t('newspaper.updated_at', 'Updated') }} {{ lastUpdatedText }}</span>
          </div>
        </div>

        <div class="modern-controls">
          <div v-if="config.allow_theme_switch" class="modern-theme-switch" role="group" :aria-label="t('newspaper.theme_label', 'Theme')">
            <button
              v-for="mode in ['classic', 'modern', 'feed']"
              :key="mode"
              type="button"
              class="modern-seg-btn"
              :class="{ active: activeTheme === mode }"
              @click="$emit('switch-theme', mode)"
            >
              {{ t('newspaper.theme.' + mode, mode) }}
            </button>
          </div>

          <select
            v-if="languages.length > 0"
            :value="currentLocale"
            class="modern-select"
            aria-label="Language"
            @change="$emit('change-locale', $event.target.value)"
          >
            <option v-for="lang in languages" :key="lang.code" :value="lang.code">
              {{ lang.name }}
            </option>
          </select>

          <a href="/dashboard" class="modern-dash-btn">
            <svg class="icon-svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z" />
            </svg>
            <span>{{ t('newspaper.dashboard_link', 'Dashboard') }}</span>
          </a>
        </div>
      </div>

      <nav
        v-if="config.show_source_filter !== false && (sources.length > 1 || selectedSource !== '')"
        class="modern-filter-wrap"
        aria-label="Sources"
      >
        <button
          type="button"
          class="modern-filter-chip"
          :class="{ active: selectedSource === '' }"
          @click="$emit('select-source', '')"
        >
          {{ t('newspaper.filter_all', 'All Sources') }}
        </button>
        <button
          v-for="src in sources"
          :key="src"
          type="button"
          class="modern-filter-chip"
          :class="{ active: selectedSource === src }"
          @click="$emit('select-source', src)"
        >
          {{ src }}
        </button>
      </nav>
    </header>

    <div class="modern-content">
      <div v-if="loading" class="modern-state">
        {{ t('newspaper.loading', 'Loading articles...') }}
      </div>

      <div v-else-if="error" class="modern-state error">
        <p>{{ error }}</p>
        <button type="button" class="modern-retry-btn" @click="$emit('retry')">
          {{ t('newspaper.retry', 'Retry') }}
        </button>
      </div>

      <div v-else-if="newsList.length === 0" class="modern-state">
        {{ t('newspaper.empty', 'No recent articles found') }}
      </div>

      <main v-else>
        <template v-if="config.sort === 'source'">
          <section v-for="(items, source) in groupedNews" :key="source" class="modern-group">
            <div class="modern-group-header">
              <h2>{{ source }}</h2>
              <span class="modern-count">{{ items.length }}</span>
            </div>

            <div class="modern-grid">
              <article v-for="item in items" :key="item.id" class="modern-card">
                <div class="modern-card-meta">
                  <span class="modern-source-tag">{{ item.source || source }}</span>
                  <span class="modern-time">{{ formatTime(item.datetime_released) }}</span>
                  <span v-if="config.show_author !== false && item.author && item.author !== source" class="modern-author">
                    &middot; {{ t('newspaper.author_prefix', 'By') }} {{ item.author }}
                  </span>
                </div>

                <h3 class="modern-card-title">{{ plainText(item.title) }}</h3>

                <div v-if="config.show_tags !== false && item.tags && item.tags.length > 0" class="modern-tags">
                  <span v-for="(tag, idx) in item.tags" :key="idx" class="modern-tag">
                    #{{ tag.tag }}
                  </span>
                </div>

                <p v-if="showSummary(item)" class="modern-summary">{{ plainText(item.summary) }}</p>

                <blockquote v-if="item.quotation && item.quotation.content" class="modern-quote">
                  <strong>{{ t('newspaper.quote', 'Quote') }} {{ item.quotation.from || source }}:</strong>
                  {{ item.quotation.content }}
                </blockquote>

                <div
                  v-if="isHtml(item)"
                  class="modern-body"
                  v-html="cleanHtml(item.content)"
                ></div>
                <div
                  v-else
                  class="modern-body"
                  v-html="mdHtml(item.content || item.summary || '')"
                ></div>

                <div v-if="!isHtml(item) && attachments(item).length > 0" class="modern-attachments">
                  <img
                    v-for="(att, idx) in attachments(item)"
                    :key="idx"
                    :src="att.url"
                    loading="lazy"
                    :alt="plainText(item.title)"
                  />
                </div>

                <div v-if="safeUrl(item.from_url)" class="modern-card-footer">
                  <a
                    :href="safeUrl(item.from_url)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="modern-link"
                  >
                    <span>{{ t('newspaper.original_link', 'Source Link') }}</span>
                    <svg class="icon-svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                      <path d="M19 19H5V5h7V3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                    </svg>
                  </a>
                </div>
              </article>
            </div>
          </section>
        </template>

        <template v-else>
          <div class="modern-grid">
            <article v-for="item in newsList" :key="item.id" class="modern-card">
              <div class="modern-card-meta">
                <span v-if="item.source" class="modern-source-tag">{{ item.source }}</span>
                <span class="modern-time">{{ formatTime(item.datetime_released) }}</span>
                <span v-if="config.show_author !== false && item.author && item.author !== item.source" class="modern-author">
                  &middot; {{ t('newspaper.author_prefix', 'By') }} {{ item.author }}
                </span>
              </div>

              <h3 class="modern-card-title">{{ plainText(item.title) }}</h3>

              <div v-if="config.show_tags !== false && item.tags && item.tags.length > 0" class="modern-tags">
                <span v-for="(tag, idx) in item.tags" :key="idx" class="modern-tag">
                  #{{ tag.tag }}
                </span>
              </div>

              <p v-if="showSummary(item)" class="modern-summary">{{ plainText(item.summary) }}</p>

              <blockquote v-if="item.quotation && item.quotation.content" class="modern-quote">
                <strong>{{ t('newspaper.quote', 'Quote') }} {{ item.quotation.from || item.source }}:</strong>
                {{ item.quotation.content }}
              </blockquote>

              <div
                v-if="isHtml(item)"
                class="modern-body"
                v-html="cleanHtml(item.content)"
              ></div>
              <div
                v-else
                class="modern-body"
                v-html="mdHtml(item.content || item.summary || '')"
              ></div>

              <div v-if="!isHtml(item) && attachments(item).length > 0" class="modern-attachments">
                <img
                  v-for="(att, idx) in attachments(item)"
                  :key="idx"
                  :src="att.url"
                  loading="lazy"
                  :alt="plainText(item.title)"
                />
              </div>

              <div v-if="safeUrl(item.from_url)" class="modern-card-footer">
                <a
                  :href="safeUrl(item.from_url)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="modern-link"
                >
                  <span>{{ t('newspaper.original_link', 'Source Link') }}</span>
                  <svg class="icon-svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M19 19H5V5h7V3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                  </svg>
                </a>
              </div>
            </article>
          </div>
        </template>
      </main>
    </div>

    <footer class="modern-footer">
      {{ t('newspaper.generated_by', 'Published by') }} {{ footerPublisher }} &middot; {{ formattedCurrentDate }}
    </footer>
  </div>
</template>

<script>
import {
  sanitizeHttpUrl,
  renderMarkdown,
  stripMarkdownPlain,
  sanitizeRichHtml,
  isHtmlArticle,
  imageAttachments,
  shouldShowSummary,
  formatDateTime
} from '../utils/render.js'

export default {
  name: 'ThemeModern',
  props: {
    mastheadTitle: String,
    mastheadSubtitle: String,
    footerPublisher: String,
    formattedCurrentDate: String,
    lastUpdatedText: String,
    newsList: Array,
    groupedNews: Object,
    sources: Array,
    selectedSource: String,
    config: Object,
    languages: Array,
    currentLocale: String,
    activeTheme: String,
    loading: Boolean,
    error: String,
    browserLocaleTag: String,
    t: Function
  },
  emits: ['select-source', 'change-locale', 'switch-theme', 'retry'],
  methods: {
    isHtml: isHtmlArticle,
    attachments: imageAttachments,
    cleanHtml: sanitizeRichHtml,
    mdHtml: renderMarkdown,
    plainText: stripMarkdownPlain,
    safeUrl: sanitizeHttpUrl,
    showSummary(item) {
      return shouldShowSummary(item, this.config)
    },
    formatTime(iso) {
      return formatDateTime(iso, this.browserLocaleTag)
    }
  }
}
</script>

