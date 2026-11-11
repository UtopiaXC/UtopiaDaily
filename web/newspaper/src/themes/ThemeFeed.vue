<template>
  <div class="feed-shell">
    <header class="feed-header">
      <div class="feed-header-inner">
        <div class="feed-header-top">
          <div class="feed-title-wrap">
            <h1 class="feed-title">{{ mastheadTitle }}</h1>
            <div class="feed-meta-line">
              <span>{{ formattedCurrentDate }}</span>
              <span v-if="mastheadSubtitle">&middot; {{ mastheadSubtitle }}</span>
              <span v-if="lastUpdatedText">&middot; {{ t('newspaper.updated_at', 'Updated') }} {{ lastUpdatedText }}</span>
            </div>
          </div>

          <div class="feed-actions">
            <div v-if="config.allow_theme_switch" class="feed-theme-switch" role="group" :aria-label="t('newspaper.theme_label', 'Theme')">
              <button
                v-for="mode in ['classic', 'modern', 'feed']"
                :key="mode"
                type="button"
                class="feed-seg-btn"
                :class="{ active: activeTheme === mode }"
                @click="$emit('switch-theme', mode)"
              >
                {{ t('newspaper.theme.' + mode, mode) }}
              </button>
            </div>

            <select
              v-if="languages.length > 0"
              :value="currentLocale"
              class="feed-select"
              aria-label="Language"
              @change="$emit('change-locale', $event.target.value)"
            >
              <option v-for="lang in languages" :key="lang.code" :value="lang.code">
                {{ lang.name }}
              </option>
            </select>

            <a href="/dashboard" class="feed-dash-link">
              {{ t('newspaper.dashboard_link', 'Dashboard') }}
            </a>
          </div>
        </div>

        <nav
          v-if="config.show_source_filter !== false && (sources.length > 1 || selectedSource !== '')"
          class="feed-filter-row"
          aria-label="Sources"
        >
          <button
            type="button"
            class="feed-filter-pill"
            :class="{ active: selectedSource === '' }"
            @click="$emit('select-source', '')"
          >
            {{ t('newspaper.filter_all', 'All Sources') }}
          </button>
          <button
            v-for="src in sources"
            :key="src"
            type="button"
            class="feed-filter-pill"
            :class="{ active: selectedSource === src }"
            @click="$emit('select-source', src)"
          >
            {{ src }}
          </button>
        </nav>
      </div>
    </header>

    <div class="feed-container">
      <div v-if="loading" class="feed-state">
        {{ t('newspaper.loading', 'Loading articles...') }}
      </div>

      <div v-else-if="error" class="feed-state error">
        <p>{{ error }}</p>
        <button type="button" class="feed-retry-btn" @click="$emit('retry')">
          {{ t('newspaper.retry', 'Retry') }}
        </button>
      </div>

      <div v-else-if="newsList.length === 0" class="feed-state">
        {{ t('newspaper.empty', 'No recent articles found') }}
      </div>

      <main v-else class="feed-stream">
        <template v-if="config.sort === 'source'">
          <section v-for="(items, source) in groupedNews" :key="source" class="feed-group">
            <div class="feed-group-label">
              <span>{{ source }}</span>
              <span class="feed-group-count">{{ items.length }}</span>
            </div>

            <article v-for="item in items" :key="item.id" class="feed-item">
              <div class="feed-item-header">
                <span class="feed-item-source">{{ item.source || source }}</span>
                <span class="feed-item-time">{{ formatTime(item.datetime_released) }}</span>
                <span v-if="config.show_author !== false && item.author && item.author !== source" class="feed-item-author">
                  &middot; {{ t('newspaper.author_prefix', 'By') }} {{ item.author }}
                </span>
              </div>

              <h3 class="feed-item-title">{{ plainText(item.title) }}</h3>

              <div v-if="config.show_tags !== false && item.tags && item.tags.length > 0" class="feed-item-tags">
                <span v-for="(tag, idx) in item.tags" :key="idx" class="feed-item-tag">
                  #{{ tag.tag }}
                </span>
              </div>

              <p v-if="showSummary(item)" class="feed-item-summary">{{ plainText(item.summary) }}</p>

              <blockquote v-if="item.quotation && item.quotation.content" class="feed-item-quote">
                <strong>{{ t('newspaper.quote', 'Quote') }} {{ item.quotation.from || source }}:</strong>
                {{ item.quotation.content }}
              </blockquote>

              <div
                v-if="isHtml(item)"
                class="feed-item-body"
                v-html="cleanHtml(item.content)"
              ></div>
              <div
                v-else
                class="feed-item-body"
                v-html="mdHtml(item.content || item.summary || '')"
              ></div>

              <div v-if="!isHtml(item) && attachments(item).length > 0" class="feed-item-attachments">
                <img
                  v-for="(att, idx) in attachments(item)"
                  :key="idx"
                  :src="att.url"
                  loading="lazy"
                  :alt="plainText(item.title)"
                />
              </div>

              <div v-if="safeUrl(item.from_url)" class="feed-item-footer">
                <a
                  :href="safeUrl(item.from_url)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="feed-item-link"
                >
                  {{ t('newspaper.original_link', 'Source Link') }}
                </a>
              </div>
            </article>
          </section>
        </template>

        <template v-else>
          <article v-for="item in newsList" :key="item.id" class="feed-item">
            <div class="feed-item-header">
              <span v-if="item.source" class="feed-item-source">{{ item.source }}</span>
              <span class="feed-item-time">{{ formatTime(item.datetime_released) }}</span>
              <span v-if="config.show_author !== false && item.author && item.author !== item.source" class="feed-item-author">
                &middot; {{ t('newspaper.author_prefix', 'By') }} {{ item.author }}
              </span>
            </div>

            <h3 class="feed-item-title">{{ plainText(item.title) }}</h3>

            <div v-if="config.show_tags !== false && item.tags && item.tags.length > 0" class="feed-item-tags">
              <span v-for="(tag, idx) in item.tags" :key="idx" class="feed-item-tag">
                #{{ tag.tag }}
              </span>
            </div>

            <p v-if="showSummary(item)" class="feed-item-summary">{{ plainText(item.summary) }}</p>

            <blockquote v-if="item.quotation && item.quotation.content" class="feed-item-quote">
              <strong>{{ t('newspaper.quote', 'Quote') }} {{ item.quotation.from || item.source }}:</strong>
              {{ item.quotation.content }}
            </blockquote>

            <div
              v-if="isHtml(item)"
              class="feed-item-body"
              v-html="cleanHtml(item.content)"
            ></div>
            <div
              v-else
              class="feed-item-body"
              v-html="mdHtml(item.content || item.summary || '')"
            ></div>

            <div v-if="!isHtml(item) && attachments(item).length > 0" class="feed-item-attachments">
              <img
                v-for="(att, idx) in attachments(item)"
                :key="idx"
                :src="att.url"
                loading="lazy"
                :alt="plainText(item.title)"
              />
            </div>

            <div v-if="safeUrl(item.from_url)" class="feed-item-footer">
              <a
                :href="safeUrl(item.from_url)"
                target="_blank"
                rel="noopener noreferrer"
                class="feed-item-link"
              >
                {{ t('newspaper.original_link', 'Source Link') }}
              </a>
            </div>
          </article>
        </template>
      </main>

      <footer class="feed-footer">
        {{ t('newspaper.generated_by', 'Published by') }} {{ footerPublisher }} &middot; {{ formattedCurrentDate }}
      </footer>
    </div>
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
  name: 'ThemeFeed',
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

