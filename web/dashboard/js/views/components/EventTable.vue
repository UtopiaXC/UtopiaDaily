<template>
    <div class="ui-surface flex flex-col">
        <div class="ui-surface-header flex flex-wrap justify-between items-center gap-3">
            <div class="flex items-center gap-2">
                <span class="material-icons text-slate-500 dark:text-slate-400 text-[20px]">notifications</span>
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('events.title') }}</h3>
            </div>
            <button
                @click="fetchEvents(1)"
                class="ui-btn-secondary"
                :title="t('common.refresh')"
            >
                <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                <span>{{ t('common.refresh') }}</span>
            </button>
        </div>

        <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
                <thead>
                    <tr class="bg-slate-50 dark:bg-slate-900/80 text-slate-500 dark:text-slate-400 text-[11px] uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
                        <th class="p-3 font-semibold cursor-pointer hover:text-slate-800 dark:hover:text-slate-200 select-none w-32" @click="toggleSort('level')">
                            <div class="flex items-center gap-1">
                                <span>{{ t('events.level') }}</span>
                                <span class="material-icons text-[14px]">{{ getSortIcon('level') }}</span>
                            </div>
                        </th>
                        <th class="p-3 font-semibold cursor-pointer hover:text-slate-800 dark:hover:text-slate-200 select-none w-32" @click="toggleSort('category')">
                            <div class="flex items-center gap-1">
                                <span>{{ t('events.category') }}</span>
                                <span class="material-icons text-[14px]">{{ getSortIcon('category') }}</span>
                            </div>
                        </th>
                        <th class="p-3 font-semibold min-w-[260px]">{{ t('events.message') }}</th>
                        <th class="p-3 font-semibold cursor-pointer hover:text-slate-800 dark:hover:text-slate-200 select-none w-48" @click="toggleSort('created_at')">
                            <div class="flex items-center gap-1">
                                <span>{{ t('events.time') }}</span>
                                <span class="material-icons text-[14px]">{{ getSortIcon('created_at') }}</span>
                            </div>
                        </th>
                        <th class="p-3 font-semibold text-right w-24">{{ t('common.actions') }}</th>
                    </tr>

                    <tr class="bg-slate-50/50 dark:bg-slate-950/40 border-b border-slate-200 dark:border-slate-800">
                        <th class="p-2">
                            <select
                                v-model="filters.level"
                                @change="fetchEvents(1)"
                                class="ui-input py-1.5 px-2 text-xs font-normal"
                            >
                                <option value="">{{ t('events.all_levels') }}</option>
                                <option value="NORMAL">{{ t('event.level.NORMAL') }}</option>
                                <option value="WARNING">{{ t('event.level.WARNING') }}</option>
                                <option value="CRITICAL">{{ t('event.level.CRITICAL') }}</option>
                                <option value="FATAL">{{ t('event.level.FATAL') }}</option>
                            </select>
                        </th>
                        <th class="p-2">
                            <select
                                v-model="filters.category"
                                @change="fetchEvents(1)"
                                class="ui-input py-1.5 px-2 text-xs font-normal"
                            >
                                <option value="">{{ t('events.all_categories') }}</option>
                                <option v-for="cat in categories" :key="cat" :value="cat">
                                    {{ t('event.category.' + cat, cat) }}
                                </option>
                            </select>
                        </th>
                        <th class="p-2">
                            <div class="relative">
                                <span class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none">
                                    <span class="material-icons text-slate-400 text-[15px]">search</span>
                                </span>
                                <input
                                    v-model="filters.search"
                                    @keyup.enter="fetchEvents(1)"
                                    type="text"
                                    class="ui-input py-1.5 pl-8 pr-2 text-xs font-normal"
                                    :placeholder="t('events.search_placeholder')"
                                >
                            </div>
                        </th>
                        <th class="p-2">
                            <input
                                type="date"
                                v-model="filters.date"
                                @change="fetchEvents(1)"
                                class="ui-input py-1.5 px-2 text-xs font-normal"
                            >
                        </th>
                        <th class="p-2 text-right">
                            <button
                                @click="resetFilters"
                                class="ui-btn-secondary py-1 px-2.5 text-xs font-normal"
                            >
                                {{ t('common.reset') }}
                            </button>
                        </th>
                    </tr>
                </thead>

                <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                    <tr v-if="loading">
                        <td colspan="5" class="p-8 text-center">
                            <div class="inline-block animate-spin rounded-full h-5 w-5 border-2 border-slate-200 dark:border-slate-700 border-t-slate-900 dark:border-t-blue-500"></div>
                            <p class="mt-2 text-slate-400 text-xs">{{ t('common.loading') }}</p>
                        </td>
                    </tr>
                    <tr v-else-if="events.length === 0">
                        <td colspan="5" class="p-10 text-center text-slate-400">
                            <span class="material-icons text-3xl mb-1 opacity-40 block">inbox</span>
                            <p class="text-xs">{{ t('events.no_events') }}</p>
                        </td>
                    </tr>
                    <tr v-for="ev in events" :key="ev.id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors text-sm">
                        <td class="p-3 whitespace-nowrap align-top">
                            <span :class="getLevelClass(ev.level)" class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border">
                                {{ t('event.level.' + ev.level, ev.level) }}
                            </span>
                        </td>
                        <td class="p-3 whitespace-nowrap align-top">
                            <span class="text-slate-600 dark:text-slate-300 text-xs font-medium">
                                {{ t('event.category.' + ev.category, ev.category) }}
                            </span>
                        </td>
                        <td class="p-3 align-top">
                            <div class="font-medium text-slate-900 dark:text-slate-100 leading-snug text-[13px]">
                                {{ formatEventMessage(ev) }}
                            </div>
                            <div class="text-[11px] text-slate-400 mt-1 font-mono flex items-center gap-2">
                                <span>{{ ev.event_key }}</span>
                                <span v-if="ev.source_type || ev.source_id">
                                    {{ [ev.source_type, ev.source_id].filter(Boolean).join(':') }}
                                </span>
                            </div>
                        </td>
                        <td class="p-3 text-slate-500 dark:text-slate-400 whitespace-nowrap font-mono text-xs align-top">
                            {{ formatDate(ev.created_at) }}
                        </td>
                        <td class="p-3 text-right align-top">
                            <button
                                @click="openDetails(ev)"
                                class="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 text-xs font-medium hover:underline"
                            >
                                {{ t('events.details') }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="px-4 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs rounded-b-lg">
            <span class="text-slate-500 dark:text-slate-400">
                {{ pageInfoText }}
            </span>

            <div class="flex items-center gap-2">
                <button
                    @click="fetchEvents(page - 1)"
                    :disabled="page <= 1"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                >
                    <span class="material-icons text-[15px]">chevron_left</span>
                    <span>{{ t('common.prev') }}</span>
                </button>
                <button
                    @click="fetchEvents(page + 1)"
                    :disabled="page >= totalPages"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                >
                    <span>{{ t('common.next') }}</span>
                    <span class="material-icons text-[15px]">chevron_right</span>
                </button>
            </div>
        </div>

        <InfoModal
            :show="detailModal.show"
            :title="detailModal.title"
            :message="detailModal.message"
            :t="t"
            @close="detailModal.show = false"
        />
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import InfoModal from '../../components/InfoModal.vue';

export default {
    components: { InfoModal },
    props: ['t'],
    inject: ['showToast'],
    data() {
        return {
            events: [],
            categories: ['SYSTEM', 'USER', 'MODULE', 'TASK', 'SECURITY', 'PUSH'],
            total: 0,
            page: 1,
            pageSize: 15,
            loading: false,
            sortBy: 'created_at',
            order: 'desc',
            filters: {
                level: '',
                category: '',
                search: '',
                date: ''
            },
            detailModal: {
                show: false,
                title: '',
                message: ''
            }
        };
    },
    computed: {
        totalPages() {
            return Math.max(1, Math.ceil(this.total / this.pageSize));
        },
        pageInfoText() {
            return this.t('common.page_info')
                .replace('{page}', String(this.page))
                .replace('{total}', String(this.totalPages));
        }
    },
    mounted() {
        this.fetchEvents(1);
    },
    methods: {
        async fetchEvents(targetPage = 1) {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/events', {
                    params: {
                        page: targetPage,
                        page_size: this.pageSize,
                        level: this.filters.level,
                        category: this.filters.category,
                        search: this.filters.search.trim(),
                        date: this.filters.date,
                        sort_by: this.sortBy,
                        order: this.order
                    }
                });
                const data = res.data || {};
                this.events = data.items || [];
                this.total = data.total || 0;
                this.page = data.page || targetPage;
                if (Array.isArray(data.categories) && data.categories.length > 0) {
                    const merged = new Set(['SYSTEM', 'USER', 'MODULE', 'TASK', 'SECURITY', 'PUSH', ...data.categories]);
                    this.categories = Array.from(merged);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        toggleSort(col) {
            if (this.sortBy === col) {
                this.order = this.order === 'desc' ? 'asc' : 'desc';
            } else {
                this.sortBy = col;
                this.order = 'desc';
            }
            this.fetchEvents(1);
        },
        getSortIcon(col) {
            if (this.sortBy !== col) return 'unfold_more';
            return this.order === 'asc' ? 'arrow_upward' : 'arrow_downward';
        },
        resetFilters() {
            this.filters = { level: '', category: '', search: '', date: '' };
            this.sortBy = 'created_at';
            this.order = 'desc';
            this.fetchEvents(1);
        },
        parseJSONSafe(raw) {
            if (!raw) return null;
            if (typeof raw === 'object') return raw;
            try {
                return JSON.parse(raw);
            } catch {
                return raw;
            }
        },
        formatEventMessage(ev) {
            let text = this.t(ev.event_key, ev.event_key);
            const params = this.parseJSONSafe(ev.params_json);
            if (params && typeof params === 'object') {
                Object.keys(params).forEach(k => {
                    const val = params[k];
                    if (val !== undefined && val !== null) {
                        text = text.split(`{${k}}`).join(String(val));
                    }
                });
            }
            return text;
        },
        openDetails(ev) {
            const details = this.parseJSONSafe(ev.details_json);
            const params = this.parseJSONSafe(ev.params_json);
            const payload = {};
            if (params) payload.params = params;
            if (details) payload.details = details;
            if (ev.source_type || ev.source_id) {
                payload.source = [ev.source_type, ev.source_id].filter(Boolean).join(':');
            }
            this.detailModal = {
                show: true,
                title: this.t('events.details'),
                message: Object.keys(payload).length > 0 ? JSON.stringify(payload, null, 2) : this.t('events.no_details')
            };
        },
        getLevelClass(level) {
            switch (level) {
                case 'WARNING':
                    return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900';
                case 'CRITICAL':
                case 'FATAL':
                    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
                default:
                    return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
            }
        },
        formatDate(iso) {
            if (!iso) return '-';
            const d = new Date(iso);
            if (Number.isNaN(d.getTime())) return iso;
            const pad = n => String(n).padStart(2, '0');
            return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
        }
    }
};
</script>
