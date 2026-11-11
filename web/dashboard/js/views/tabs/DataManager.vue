<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header flex flex-wrap justify-between items-center gap-3">
            <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('data.title') }}</h3>
            <button @click="fetchArticles(page)" class="ui-btn-secondary">
                <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                <span>{{ t('common.refresh') }}</span>
            </button>
        </div>

        <!-- Filter Bar -->
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-950/30 space-y-3">
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-12 gap-3 items-center">
                <div class="lg:col-span-4 relative">
                    <span class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none">
                        <span class="material-icons text-slate-400 text-[16px]">search</span>
                    </span>
                    <input
                        v-model="filters.search"
                        @keyup.enter="applyFilters"
                        type="text"
                        :placeholder="t('data.search_placeholder')"
                        class="ui-input pl-8 py-1.5 text-xs"
                    >
                </div>

                <div class="lg:col-span-3">
                    <CustomSelect
                        v-model="filters.module_id"
                        :options="moduleSelectOptions"
                    />
                </div>

                <div class="lg:col-span-3">
                    <CustomSelect
                        v-model="filters.tag_id"
                        :disabled="filters.untagged"
                        :options="tagFilterOptions"
                    />
                </div>

                <div class="lg:col-span-2 flex items-center justify-end gap-2">
                    <button @click="applyFilters" class="ui-btn-primary text-xs py-1.5 px-3 flex-1 sm:flex-none">
                        {{ t('common.search') }}
                    </button>
                    <button @click="resetFilters" class="ui-btn-secondary text-xs py-1.5 px-2.5">
                        {{ t('common.reset') }}
                    </button>
                </div>
            </div>

            <div class="flex flex-wrap items-center gap-4 text-xs">
                <label class="inline-flex items-center gap-2 cursor-pointer select-none">
                    <input
                        type="checkbox"
                        v-model="filters.untagged"
                        @change="applyFilters"
                        class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                    >
                    <span class="font-medium text-slate-700 dark:text-slate-300">{{ t('data.untagged_only') }}</span>
                </label>

                <div class="flex items-center gap-2">
                    <span class="text-slate-500 dark:text-slate-400">{{ t('data.date_from') }}</span>
                    <input
                        type="date"
                        v-model="filters.date_from"
                        @change="applyFilters"
                        class="ui-input py-1 px-2 text-xs w-auto"
                    >
                </div>

                <div class="flex items-center gap-2">
                    <span class="text-slate-500 dark:text-slate-400">{{ t('data.date_to') }}</span>
                    <input
                        type="date"
                        v-model="filters.date_to"
                        @change="applyFilters"
                        class="ui-input py-1 px-2 text-xs w-auto"
                    >
                </div>
            </div>
        </div>

        <!-- Batch Action Bar -->
        <div
            v-if="canEdit && selectedIds.length > 0"
            class="px-5 py-2.5 border-b border-blue-200 dark:border-blue-900 bg-blue-50/70 dark:bg-blue-950/30 flex flex-wrap items-center justify-between gap-3"
        >
            <span class="text-xs font-semibold text-blue-800 dark:text-blue-200">
                {{ t('data.selected_count').replace('{count}', String(selectedIds.length)) }}
            </span>
            <div class="flex items-center gap-2">
                <button
                    @click="reclassifySelected"
                    :disabled="batchBusy"
                    class="ui-btn-primary text-xs py-1 px-3"
                >
                    <span class="material-icons text-[14px]">auto_awesome</span>
                    <span>{{ t('data.batch_reclassify') }}</span>
                </button>
                <button
                    @click="batchDeleteConfirm = true"
                    :disabled="batchBusy"
                    class="ui-btn-danger text-xs py-1 px-3"
                >
                    <span class="material-icons text-[14px]">delete_outline</span>
                    <span>{{ t('data.batch_delete') }}</span>
                </button>
                <button
                    @click="selectedIds = []"
                    class="ui-btn-secondary text-xs py-1 px-2.5"
                >
                    {{ t('data.clear_selection') }}
                </button>
            </div>
        </div>

        <!-- Data Table -->
        <div class="flex-1 overflow-auto">
            <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400 sticky top-0 z-10">
                    <tr>
                        <th v-if="canEdit" class="px-4 py-2.5 w-10 text-left">
                            <input
                                type="checkbox"
                                :checked="allPageSelected"
                                @change="toggleSelectAll"
                                class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                            >
                        </th>
                        <th class="px-4 py-2.5 text-left font-semibold">{{ t('data.col_title') }}</th>
                        <th class="px-4 py-2.5 text-left font-semibold w-44">{{ t('data.col_source') }}</th>
                        <th class="px-4 py-2.5 text-left font-semibold w-60">{{ t('data.col_tags') }}</th>
                        <th class="px-4 py-2.5 text-left font-semibold w-40">{{ t('data.col_released') }}</th>
                        <th class="px-4 py-2.5 text-right font-semibold w-36">{{ t('common.actions') }}</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                    <tr v-if="loading">
                        <td :colspan="canEdit ? 6 : 5" class="p-10 text-center text-slate-400 text-xs">
                            <span class="material-icons animate-spin text-2xl text-blue-600 mb-1 block">sync</span>
                            <span>{{ t('common.loading') }}</span>
                        </td>
                    </tr>
                    <tr v-else-if="articles.length === 0">
                        <td :colspan="canEdit ? 6 : 5" class="p-10 text-center text-slate-400 text-xs">
                            <span class="material-icons text-3xl mb-1 opacity-40 block">inbox</span>
                            <span>{{ t('data.empty') }}</span>
                        </td>
                    </tr>
                    <tr
                        v-for="item in articles"
                        :key="item.id"
                        class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors"
                    >
                        <td v-if="canEdit" class="px-4 py-3 align-top">
                            <input
                                type="checkbox"
                                :value="item.id"
                                v-model="selectedIds"
                                class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                            >
                        </td>
                        <td class="px-4 py-3 align-top max-w-md">
                            <button
                                type="button"
                                @click="openDetail(item.id)"
                                class="text-left font-medium text-slate-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors line-clamp-1"
                            >
                                {{ plainText(item.title) || `#${item.id}` }}
                            </button>
                            <p v-if="plainText(item.summary)" class="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">
                                {{ plainText(item.summary) }}
                            </p>
                        </td>
                        <td class="px-4 py-3 align-top">
                            <div class="text-xs font-medium text-slate-800 dark:text-slate-200 truncate">{{ item.source || '-' }}</div>
                            <div class="text-[11px] font-mono text-slate-400 mt-0.5 truncate">{{ item.module_id }}</div>
                        </td>
                        <td class="px-4 py-3 align-top">
                            <div v-if="item.tags && item.tags.length > 0" class="flex flex-wrap gap-1">
                                <span
                                    v-for="tag in item.tags"
                                    :key="tag.tag_id"
                                    class="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900"
                                >
                                    #{{ tag.tag }}
                                </span>
                            </div>
                            <span
                                v-else
                                class="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-slate-100 text-slate-500 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700"
                            >
                                {{ t('data.untagged_badge') }}
                            </span>
                        </td>
                        <td class="px-4 py-3 align-top text-xs font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap">
                            {{ formatTime(item.datetime_released) }}
                        </td>
                        <td class="px-4 py-3 align-top text-right whitespace-nowrap">
                            <div class="inline-flex items-center gap-1.5">
                                <button
                                    @click="openDetail(item.id)"
                                    class="ui-btn-secondary text-xs py-1 px-2 text-blue-600 dark:text-blue-400"
                                >
                                    {{ t('events.details') }}
                                </button>
                                <button
                                    v-if="canEdit"
                                    @click="requestDeleteSingle(item)"
                                    class="ui-btn-secondary text-xs py-1 px-2 text-rose-600 dark:text-rose-400"
                                >
                                    {{ t('common.delete') }}
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Pagination Footer -->
        <div class="px-4 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs">
            <span class="text-slate-500 dark:text-slate-400">{{ pageInfoText }}</span>
            <div class="flex items-center gap-2">
                <button
                    @click="fetchArticles(page - 1)"
                    :disabled="page <= 1"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                >
                    <span class="material-icons text-[15px]">chevron_left</span>
                    <span>{{ t('common.prev') }}</span>
                </button>
                <button
                    @click="fetchArticles(page + 1)"
                    :disabled="page >= totalPages"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                >
                    <span>{{ t('common.next') }}</span>
                    <span class="material-icons text-[15px]">chevron_right</span>
                </button>
            </div>
        </div>

        <!-- Article Detail & Tag Management Modal -->
        <transition name="modal">
            <div v-if="detailModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 w-full max-w-3xl max-h-[90vh] flex flex-col overflow-hidden">
                    <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-3">
                        <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('data.detail_title') }}</h3>
                        <button @click="closeDetail" class="p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded">
                            <span class="material-icons text-[18px]">close</span>
                        </button>
                    </div>

                    <div v-if="detailModal.loading" class="p-12 text-center text-slate-400 text-xs flex-1">
                        <span class="material-icons animate-spin text-2xl text-blue-600 mb-2 block">sync</span>
                        <span>{{ t('common.loading') }}</span>
                    </div>

                    <div v-else-if="detailModal.item" class="p-5 overflow-y-auto space-y-5 flex-1">
                        <div>
                            <h2 class="text-lg font-semibold text-slate-900 dark:text-white leading-snug">
                                {{ plainText(detailModal.item.title) }}
                            </h2>
                            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500 dark:text-slate-400 mt-2">
                                <span>{{ detailModal.item.source }}</span>
                                <span class="font-mono">{{ detailModal.item.module_id }}</span>
                                <span v-if="detailModal.item.author">{{ detailModal.item.author }}</span>
                                <span class="font-mono">{{ formatTime(detailModal.item.datetime_released) }}</span>
                                <a
                                    v-if="safeUrl(detailModal.item.from_url)"
                                    :href="safeUrl(detailModal.item.from_url)"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:underline"
                                >
                                    <span>{{ t('data.open_original') }}</span>
                                    <span class="material-icons text-[13px]">open_in_new</span>
                                </a>
                            </div>
                        </div>

                        <!-- Current Tags & Tag Management -->
                        <div class="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 space-y-3">
                            <div class="flex flex-wrap items-center justify-between gap-2">
                                <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                    {{ t('data.col_tags') }}
                                </span>
                                <button
                                    v-if="canEdit"
                                    @click="reclassifyDetail"
                                    :disabled="detailModal.busy"
                                    class="ui-btn-secondary text-xs py-1 px-2.5 text-blue-600 dark:text-blue-400"
                                >
                                    <span class="material-icons text-[14px]">auto_awesome</span>
                                    <span>{{ t('data.reclassify_single') }}</span>
                                </button>
                            </div>

                            <div class="flex flex-wrap gap-1.5">
                                <span
                                    v-if="!detailModal.item.tags || detailModal.item.tags.length === 0"
                                    class="text-xs text-slate-400"
                                >
                                    {{ t('data.untagged_badge') }}
                                </span>
                                <span
                                    v-for="tag in (detailModal.item.tags || [])"
                                    :key="tag.tag_id"
                                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-medium bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-200"
                                >
                                    <span v-if="tag.group_name" class="text-[11px] text-slate-400">{{ tag.group_name }} /</span>
                                    <span>{{ tag.tag }}</span>
                                    <span class="text-[10px] font-mono px-1 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-500">
                                        {{ formatTagProvider(tag) }}
                                    </span>
                                    <button
                                        v-if="canEdit"
                                        type="button"
                                        @click="removeDetailTag(tag.tag_id)"
                                        :disabled="detailModal.busy"
                                        :title="t('data.remove_tag')"
                                        class="text-slate-400 hover:text-rose-500 flex items-center ml-0.5"
                                    >
                                        <span class="material-icons text-[14px]">close</span>
                                    </button>
                                </span>
                            </div>

                            <div v-if="canEdit" class="flex flex-col sm:flex-row gap-2 pt-2 border-t border-slate-200 dark:border-slate-800">
                                <div class="flex-1">
                                    <CustomSelect
                                        v-model="detailModal.selectedTagId"
                                        :options="addTagOptions"
                                    />
                                </div>
                                <button
                                    @click="addDetailTag"
                                    :disabled="detailModal.busy || !detailModal.selectedTagId"
                                    class="ui-btn-primary text-xs py-1.5 px-3 flex-shrink-0"
                                >
                                    <span class="material-icons text-[15px]">add</span>
                                    <span>{{ t('data.add_tag') }}</span>
                                </button>
                            </div>
                        </div>

                        <blockquote
                            v-if="detailModal.item.quotation && detailModal.item.quotation.content"
                            class="p-3 rounded-md border-l-4 border-slate-400 bg-slate-50 dark:bg-slate-950/60 text-xs text-slate-600 dark:text-slate-300"
                        >
                            <div v-if="detailModal.item.quotation.from" class="font-semibold mb-1">{{ detailModal.item.quotation.from }}</div>
                            <div>{{ detailModal.item.quotation.content }}</div>
                        </blockquote>

                        <div
                            v-if="isHtmlArticle(detailModal.item)"
                            class="text-sm text-slate-700 dark:text-slate-200 leading-relaxed space-y-3 overflow-x-auto"
                            v-html="sanitizeRichHtml(detailModal.item.content)"
                        ></div>
                        <div
                            v-else-if="isMarkdownArticle(detailModal.item)"
                            class="text-sm text-slate-700 dark:text-slate-200 leading-relaxed space-y-2.5 break-words"
                            v-html="renderMd(detailModal.item.content || detailModal.item.summary || '')"
                        ></div>
                        <div
                            v-else
                            class="text-sm text-slate-700 dark:text-slate-200 leading-relaxed whitespace-pre-wrap break-words"
                        >{{ detailModal.item.content || detailModal.item.summary || '' }}</div>

                        <div v-if="detailAttachments.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                            <img
                                v-for="(att, idx) in detailAttachments"
                                :key="idx"
                                :src="att.url"
                                loading="lazy"
                                :alt="plainText(detailModal.item.title)"
                                class="rounded border border-slate-200 dark:border-slate-800 max-h-64 object-contain bg-slate-50 dark:bg-slate-950"
                            >
                        </div>
                    </div>

                    <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 flex justify-end">
                        <button @click="closeDetail" class="ui-btn-secondary text-xs">{{ t('common.close') }}</button>
                    </div>
                </div>
            </div>
        </transition>

        <ConfirmModal
            :show="singleDeleteConfirm.show"
            :title="t('common.delete')"
            :message="t('data.delete_confirm').replace('{title}', singleDeleteConfirm.title)"
            :t="t"
            @confirm="performDeleteSingle"
            @cancel="singleDeleteConfirm.show = false"
        />

        <ConfirmModal
            :show="batchDeleteConfirm"
            :title="t('data.batch_delete')"
            :message="t('data.batch_delete_confirm').replace('{count}', String(selectedIds.length))"
            :t="t"
            @confirm="performDeleteBatch"
            @cancel="batchDeleteConfirm = false"
        />
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import {
    stripMarkdownPlain,
    renderMarkdown,
    sanitizeRichHtml,
    sanitizeHttpUrl
} from '../../utils/markdown.js';
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';

export default {
    components: { CustomSelect, ConfirmModal },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            articles: [],
            modules: [],
            allTags: [],
            total: 0,
            page: 1,
            pageSize: 20,
            loading: false,
            batchBusy: false,

            filters: {
                search: '',
                module_id: '',
                tag_id: 0,
                untagged: false,
                date_from: '',
                date_to: ''
            },

            selectedIds: [],
            singleDeleteConfirm: {
                show: false,
                id: 0,
                title: ''
            },
            batchDeleteConfirm: false,

            detailModal: {
                show: false,
                loading: false,
                busy: false,
                item: null,
                selectedTagId: 0
            }
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('data.edit');
        },
        totalPages() {
            return Math.max(1, Math.ceil((this.total || 0) / this.pageSize));
        },
        pageInfoText() {
            return this.t('common.page_info')
                .replace('{page}', String(this.page))
                .replace('{total}', String(this.totalPages));
        },
        allPageSelected() {
            return this.articles.length > 0 && this.articles.every(a => this.selectedIds.includes(a.id));
        },
        moduleSelectOptions() {
            const opts = [{ value: '', label: this.t('data.all_modules') }];
            (this.modules || []).forEach(m => {
                opts.push({ value: m, label: m });
            });
            return opts;
        },
        tagFilterOptions() {
            const opts = [{ value: 0, label: this.t('data.all_tags') }];
            (this.allTags || []).forEach(tag => {
                const grp = tag.group?.name ? `${tag.group.name} / ` : '';
                opts.push({ value: tag.id, label: `${grp}${tag.name}` });
            });
            return opts;
        },
        addTagOptions() {
            const assigned = new Set((this.detailModal.item?.tags || []).map(x => x.tag_id));
            const opts = [{ value: 0, label: this.t('data.choose_tag') }];
            (this.allTags || []).forEach(tag => {
                if (!assigned.has(tag.id)) {
                    const grp = tag.group?.name ? `${tag.group.name} / ` : '';
                    opts.push({ value: tag.id, label: `${grp}${tag.name}` });
                }
            });
            return opts;
        },
        detailAttachments() {
            const atts = this.detailModal.item?.attachments || [];
            return atts
                .filter(a => a && (a.type === 'image' || !a.type))
                .map(a => ({ ...a, url: this.safeUrl(a.url) }))
                .filter(a => Boolean(a.url));
        }
    },
    watch: {
        'filters.module_id'() {
            if (this.initializedFilters) this.applyFilters();
        },
        'filters.tag_id'() {
            if (this.initializedFilters) this.applyFilters();
        }
    },
    async mounted() {
        await Promise.all([this.fetchTags(), this.fetchArticles(1)]);
        this.initializedFilters = true;
    },
    methods: {
        plainText: stripMarkdownPlain,
        renderMd: renderMarkdown,
        safeUrl: sanitizeHttpUrl,
        sanitizeRichHtml,
        isHtmlArticle(item) {
            return item?.render_type === 'html' || item?.content_type === 'html';
        },
        isMarkdownArticle(item) {
            return item?.render_type === 'markdown' || item?.content_type === 'markdown';
        },
        formatTime(iso) {
            if (!iso) return '-';
            const d = new Date(iso);
            if (Number.isNaN(d.getTime())) return iso;
            const pad = n => String(n).padStart(2, '0');
            return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
        },
        formatTagProvider(tag) {
            if (!tag.provider || tag.provider === 'manual') {
                return this.t('data.provider_manual');
            }
            const pct = Math.round((Number(tag.confidence) || 0) * 100);
            return pct > 0 ? `${tag.provider} ${pct}%` : tag.provider;
        },
        async fetchTags() {
            try {
                const res = await http.get('/api/dashboard/data/tags');
                this.allTags = Array.isArray(res.data) ? res.data : [];
            } catch {
                // ignore optional tags error
            }
        },
        async fetchArticles(targetPage = 1, clearSelection = true) {
            if (targetPage < 1) targetPage = 1;
            this.loading = true;
            if (clearSelection) {
                this.selectedIds = [];
            }
            try {
                const res = await http.get('/api/dashboard/data/articles', {
                    params: {
                        page: targetPage,
                        page_size: this.pageSize,
                        search: this.filters.search.trim(),
                        module_id: this.filters.module_id,
                        tag_id: !this.filters.untagged && this.filters.tag_id > 0 ? this.filters.tag_id : '',
                        untagged: this.filters.untagged ? 'true' : '',
                        date_from: this.filters.date_from,
                        date_to: this.filters.date_to
                    }
                });
                const data = res.data || {};
                this.articles = Array.isArray(data.items) ? data.items : [];
                this.total = Number(data.total) || 0;
                this.page = Number(data.page) || targetPage;
                if (Array.isArray(data.modules)) {
                    this.modules = data.modules;
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        applyFilters() {
            this.fetchArticles(1);
        },
        resetFilters() {
            this.initializedFilters = false;
            this.filters = {
                search: '',
                module_id: '',
                tag_id: 0,
                untagged: false,
                date_from: '',
                date_to: ''
            };
            this.$nextTick(() => {
                this.initializedFilters = true;
            });
            this.fetchArticles(1);
        },
        toggleSelectAll(e) {
            if (e.target.checked) {
                this.selectedIds = this.articles.map(a => a.id);
            } else {
                this.selectedIds = [];
            }
        },
        requestDeleteSingle(item) {
            this.singleDeleteConfirm = {
                show: true,
                id: item.id,
                title: this.plainText(item.title) || `#${item.id}`
            };
        },
        async performDeleteSingle() {
            const id = this.singleDeleteConfirm.id;
            this.singleDeleteConfirm.show = false;
            if (!id) return;
            try {
                await http.delete(`/api/dashboard/data/articles/${id}`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchArticles(this.page);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async performDeleteBatch() {
            this.batchDeleteConfirm = false;
            if (this.selectedIds.length === 0) return;
            this.batchBusy = true;
            try {
                await http.post('/api/dashboard/data/articles/delete', { ids: this.selectedIds });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchArticles(this.page);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.batchBusy = false;
            }
        },
        async reclassifySelected() {
            if (this.selectedIds.length === 0) return;
            this.batchBusy = true;
            try {
                const res = await http.post('/api/dashboard/data/articles/reclassify', { ids: this.selectedIds });
                const added = res.data?.added ?? 0;
                this.showToast(this.t('classification.queue_enqueued').replace('{count}', String(added)), 'success');
                this.selectedIds = [];
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.batchBusy = false;
            }
        },
        async openDetail(id) {
            this.detailModal = {
                show: true,
                loading: true,
                busy: false,
                item: null,
                selectedTagId: 0
            };
            await this.reloadDetail(id);
        },
        async reloadDetail(id) {
            try {
                const res = await http.get(`/api/dashboard/data/articles/${id}`);
                this.detailModal.item = res.data || null;
                this.detailModal.selectedTagId = 0;
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
                this.detailModal.show = false;
            } finally {
                this.detailModal.loading = false;
            }
        },
        closeDetail() {
            this.detailModal.show = false;
            this.detailModal.item = null;
        },
        async addDetailTag() {
            const item = this.detailModal.item;
            const tagId = Number(this.detailModal.selectedTagId);
            if (!item || !tagId) return;
            this.detailModal.busy = true;
            try {
                await http.post(`/api/dashboard/data/articles/${item.id}/tags`, { tag_id: tagId });
                this.showToast(this.t('common.saved_success'), 'success');
                await Promise.all([
                    this.reloadDetail(item.id),
                    this.fetchArticles(this.page, false)
                ]);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.detailModal.busy = false;
            }
        },
        async removeDetailTag(tagId) {
            const item = this.detailModal.item;
            if (!item || !tagId) return;
            this.detailModal.busy = true;
            try {
                await http.delete(`/api/dashboard/data/articles/${item.id}/tags/${tagId}`);
                this.showToast(this.t('common.saved_success'), 'success');
                await Promise.all([
                    this.reloadDetail(item.id),
                    this.fetchArticles(this.page, false)
                ]);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.detailModal.busy = false;
            }
        },
        async reclassifyDetail() {
            const item = this.detailModal.item;
            if (!item) return;
            this.detailModal.busy = true;
            try {
                const res = await http.post('/api/dashboard/data/articles/reclassify', { ids: [item.id] });
                const added = res.data?.added ?? 0;
                this.showToast(this.t('classification.queue_enqueued').replace('{count}', String(added)), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.detailModal.busy = false;
            }
        }
    }
};
</script>

