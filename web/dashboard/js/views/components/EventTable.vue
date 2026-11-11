<template>
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 flex flex-col">
        <div class="mb-5 flex flex-wrap justify-between items-center gap-3">
            <div class="flex items-center gap-2">
                <span class="material-icons text-blue-500 text-2xl">notifications</span>
                <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('events.title') }}</h3>
            </div>
            <button
                @click="fetchEvents(1)"
                class="p-2 text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 transition-colors rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
                :title="t('common.refresh')"
            >
                <span class="material-icons" :class="{'animate-spin': loading}">refresh</span>
            </button>
        </div>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col bg-white dark:bg-gray-900">
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 text-xs uppercase tracking-wider border-b border-gray-200 dark:border-gray-700">
                            <th class="p-3.5 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none w-32" @click="toggleSort('level')">
                                <div class="flex items-center gap-1">
                                    {{ t('events.level') }}
                                    <span class="material-icons text-xs">{{ getSortIcon('level') }}</span>
                                </div>
                            </th>
                            <th class="p-3.5 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none w-32" @click="toggleSort('category')">
                                <div class="flex items-center gap-1">
                                    {{ t('events.category') }}
                                    <span class="material-icons text-xs">{{ getSortIcon('category') }}</span>
                                </div>
                            </th>
                            <th class="p-3.5 font-medium min-w-[260px]">{{ t('events.message') }}</th>
                            <th class="p-3.5 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none w-48" @click="toggleSort('created_at')">
                                <div class="flex items-center gap-1">
                                    {{ t('events.time') }}
                                    <span class="material-icons text-xs">{{ getSortIcon('created_at') }}</span>
                                </div>
                            </th>
                            <th class="p-3.5 font-medium text-right w-24">{{ t('common.actions') }}</th>
                        </tr>

                        <tr class="bg-gray-50/60 dark:bg-gray-800/60 border-b border-gray-200 dark:border-gray-700">
                            <th class="p-2">
                                <select
                                    v-model="filters.level"
                                    @change="fetchEvents(1)"
                                    class="w-full px-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white"
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
                                    class="w-full px-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white"
                                >
                                    <option value="">{{ t('events.all_categories') }}</option>
                                    <option v-for="cat in categories" :key="cat" :value="cat">
                                        {{ t('event.category.' + cat, cat) }}
                                    </option>
                                </select>
                            </th>
                            <th class="p-2">
                                <div class="relative">
                                    <span class="absolute inset-y-0 left-0 pl-2 flex items-center pointer-events-none">
                                        <span class="material-icons text-gray-400 text-[14px]">search</span>
                                    </span>
                                    <input
                                        v-model="filters.search"
                                        @keyup.enter="fetchEvents(1)"
                                        type="text"
                                        class="w-full pl-7 pr-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white"
                                        :placeholder="t('events.search_placeholder')"
                                    >
                                </div>
                            </th>
                            <th class="p-2">
                                <input
                                    type="date"
                                    v-model="filters.date"
                                    @change="fetchEvents(1)"
                                    class="w-full px-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white"
                                >
                            </th>
                            <th class="p-2 text-right">
                                <button
                                    @click="resetFilters"
                                    class="px-2 py-1 text-xs text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 rounded hover:bg-gray-200/60 dark:hover:bg-gray-700"
                                >
                                    {{ t('common.reset') }}
                                </button>
                            </th>
                        </tr>
                    </thead>

                    <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                        <tr v-if="loading">
                            <td colspan="5" class="p-8 text-center">
                                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
                                <p class="mt-2 text-gray-400 text-xs">{{ t('common.loading') }}</p>
                            </td>
                        </tr>
                        <tr v-else-if="events.length === 0">
                            <td colspan="5" class="p-12 text-center text-gray-400">
                                <span class="material-icons text-4xl mb-2 opacity-30">inbox</span>
                                <p class="text-sm">{{ t('events.no_events') }}</p>
                            </td>
                        </tr>
                        <tr v-for="ev in events" :key="ev.id" class="hover:bg-blue-50/30 dark:hover:bg-blue-900/10 transition-colors text-sm">
                            <td class="p-3.5 whitespace-nowrap align-top">
                                <span :class="getLevelClass(ev.level)" class="px-2 py-0.5 rounded text-[11px] font-semibold border tracking-wide">
                                    {{ t('event.level.' + ev.level, ev.level) }}
                                </span>
                            </td>
                            <td class="p-3.5 whitespace-nowrap align-top">
                                <span class="text-gray-600 dark:text-gray-300 text-xs font-medium">
                                    {{ t('event.category.' + ev.category, ev.category) }}
                                </span>
                            </td>
                            <td class="p-3.5 align-top">
                                <div class="font-medium text-gray-800 dark:text-white leading-snug">
                                    {{ formatEventMessage(ev) }}
                                </div>
                                <div class="text-[11px] text-gray-400 mt-1 font-mono flex items-center gap-2">
                                    <span>{{ ev.event_key }}</span>
                                    <span v-if="ev.source_type || ev.source_id">
                                        {{ [ev.source_type, ev.source_id].filter(Boolean).join(':') }}
                                    </span>
                                </div>
                            </td>
                            <td class="p-3.5 text-gray-500 dark:text-gray-400 whitespace-nowrap font-mono text-xs align-top">
                                {{ formatDate(ev.created_at) }}
                            </td>
                            <td class="p-3.5 text-right align-top">
                                <button
                                    @click="openDetails(ev)"
                                    class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 text-xs font-medium hover:underline"
                                >
                                    {{ t('events.details') }}
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <div class="p-3 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs">
                <span class="text-gray-500 dark:text-gray-400">
                    {{ pageInfoText }}
                </span>

                <div class="flex items-center gap-2">
                    <button
                        @click="fetchEvents(page - 1)"
                        :disabled="page <= 1"
                        class="px-2.5 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-white dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed text-gray-600 dark:text-gray-300 transition-colors flex items-center gap-1"
                    >
                        <span class="material-icons text-sm">chevron_left</span>
                        {{ t('common.prev') }}
                    </button>
                    <button
                        @click="fetchEvents(page + 1)"
                        :disabled="page >= totalPages"
                        class="px-2.5 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-white dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed text-gray-600 dark:text-gray-300 transition-colors flex items-center gap-1"
                    >
                        {{ t('common.next') }}
                        <span class="material-icons text-sm">chevron_right</span>
                    </button>
                </div>
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
                    return 'bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-900/20 dark:text-yellow-300 dark:border-yellow-800';
                case 'CRITICAL':
                case 'FATAL':
                    return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800';
                default:
                    return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-900/20 dark:text-blue-300 dark:border-blue-800';
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
