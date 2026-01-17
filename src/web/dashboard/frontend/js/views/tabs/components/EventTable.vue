<template>
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 flex flex-col h-full">
        <div class="mb-6 flex justify-between items-center">
            <div class="flex items-center gap-2">
                <span class="material-icons text-blue-500 text-2xl">notifications</span>
                <h3 class="text-xl font-bold text-gray-800 dark:text-white">{{ t('dashboard.events.title', 'System Events') }}</h3>
            </div>
            <button @click="fetchEvents(1)" class="p-2 text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 transition-colors rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700" title="Refresh">
                <span class="material-icons">refresh</span>
            </button>
        </div>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col flex-1 bg-white dark:bg-gray-900">
            <div class="overflow-x-auto flex-1">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 text-xs uppercase tracking-wider border-b border-gray-200 dark:border-gray-700">
                            <th class="p-4 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 transition-colors select-none w-32" @click="toggleSort('level')">
                                <div class="flex items-center gap-1">
                                    {{ t('common.level', 'Level') }}
                                    <span class="material-icons text-[10px]" :class="getSortIconClass('level')">{{ getSortIcon('level') }}</span>
                                </div>
                            </th>
                            <th class="p-4 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 transition-colors select-none w-32" @click="toggleSort('category')">
                                <div class="flex items-center gap-1">
                                    {{ t('common.category', 'Category') }}
                                    <span class="material-icons text-[10px]" :class="getSortIconClass('category')">{{ getSortIcon('category') }}</span>
                                </div>
                            </th>
                            <th class="p-4 font-medium min-w-[300px]">{{ t('common.summary', 'Summary') }}</th>
                            <th class="p-4 font-medium cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 transition-colors select-none w-48" @click="toggleSort('created_at')">
                                <div class="flex items-center gap-1">
                                    {{ t('common.time', 'Time') }}
                                    <span class="material-icons text-[10px]" :class="getSortIconClass('created_at')">{{ getSortIcon('created_at') }}</span>
                                </div>
                            </th>
                            <th class="p-4 font-medium text-right w-24">{{ t('common.actions', 'Actions') }}</th>
                        </tr>

                        <!-- Filter Row -->
                        <tr class="bg-gray-50/50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
                            <th class="p-2">
                                <select v-model="filters.level" @change="fetchEvents(1)" class="w-full px-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white">
                                    <option value="">{{ t('common.all_levels', 'All') }}</option>
                                    <option value="NORMAL">NORMAL</option>
                                    <option value="WARNING">WARNING</option>
                                    <option value="CRITICAL">CRITICAL</option>
                                    <option value="FATAL">FATAL</option>
                                </select>
                            </th>
                            <th class="p-2">
                                <select v-model="filters.category" @change="fetchEvents(1)" class="w-full px-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white">
                                    <option value="">{{ t('common.all_categories', 'All') }}</option>
                                    <option value="SYSTEM">SYSTEM</option>
                                    <option value="USER">USER</option>
                                    <option value="MODULE">MODULE</option>
                                    <option value="TASK">TASK</option>
                                    <option value="SECURITY">SECURITY</option>
                                </select>
                            </th>
                            <th class="p-2">
                                <div class="relative">
                                    <span class="absolute inset-y-0 left-0 pl-2 flex items-center pointer-events-none">
                                        <span class="material-icons text-gray-400 text-[14px]">search</span>
                                    </span>
                                    <input
                                        v-model="searchQuery"
                                        @keyup.enter="fetchEvents(1)"
                                        type="text"
                                        class="w-full pl-7 pr-2 py-1.5 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-xs focus:ring-1 focus:ring-blue-500 dark:text-white"
                                        :placeholder="t('common.search', 'Search summary...')"
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
                            <th class="p-2"></th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                        <tr v-if="loading" class="animate-pulse">
                            <td colspan="5" class="p-8 text-center">
                                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
                                <p class="mt-2 text-gray-400 text-xs">{{ t('common.loading', 'Loading...') }}</p>
                            </td>
                        </tr>
                        <tr v-else-if="events.length === 0">
                            <td colspan="5" class="p-12 text-center text-gray-400">
                                <span class="material-icons text-4xl mb-2 opacity-20">inbox</span>
                                <p class="text-sm">{{ t('common.no_data', 'No events found') }}</p>
                            </td>
                        </tr>
                        <tr v-for="event in events" :key="event.id" class="hover:bg-blue-50/30 dark:hover:bg-blue-900/10 transition-colors group text-sm">
                            <td class="p-4 whitespace-nowrap align-top">
                                <span :class="getLevelClass(event.level)" class="px-2 py-0.5 rounded text-[10px] font-bold border tracking-wide">
                                    {{ event.level }}
                                </span>
                            </td>
                            <td class="p-4 whitespace-nowrap align-top">
                                <span class="text-gray-600 dark:text-gray-300 text-xs font-medium">
                                    {{ event.category }}
                                </span>
                            </td>
                            <td class="p-4 align-top">
                                <div class="font-medium text-gray-800 dark:text-white leading-snug">{{ event.summary }}</div>
                                <div class="text-[10px] text-gray-400 mt-1 font-mono">{{ event.event_type }}</div>
                            </td>
                            <td class="p-4 text-gray-500 dark:text-gray-400 whitespace-nowrap font-mono text-xs align-top">
                                {{ formatDate(event.created_at) }}
                            </td>
                            <td class="p-4 text-right align-top">
                                <button @click="showDetails(event)" class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 text-xs font-medium hover:underline">
                                    {{ t('common.details', 'Details') }}
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <!-- Pagination Footer -->
            <div class="p-3 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs">
                <div class="text-gray-500 dark:text-gray-400">
                    {{ t('common.showing', 'Showing') }} <span class="font-medium text-gray-900 dark:text-white">{{ (page - 1) * pageSize + 1 }}</span> - <span class="font-medium text-gray-900 dark:text-white">{{ Math.min(page * pageSize, total) }}</span> {{ t('common.of', 'of') }} <span class="font-medium text-gray-900 dark:text-white">{{ total }}</span>
                </div>

                <div class="flex items-center gap-2">
                    <button
                        @click="fetchEvents(page - 1)"
                        :disabled="page === 1"
                        class="p-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-white dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-gray-600 dark:text-gray-300 transition-colors"
                    >
                        <span class="material-icons text-sm">chevron_left</span>
                    </button>

                    <div class="flex items-center gap-1">
                        <span class="text-gray-700 dark:text-gray-200 font-medium">
                            {{ page }}
                        </span>
                        <span class="text-gray-400">/</span>
                        <span class="text-gray-500 dark:text-gray-400">{{ Math.ceil(total / pageSize) || 1 }}</span>
                    </div>

                    <button
                        @click="fetchEvents(page + 1)"
                        :disabled="page * pageSize >= total"
                        class="p-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-white dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-gray-600 dark:text-gray-300 transition-colors"
                    >
                        <span class="material-icons text-sm">chevron_right</span>
                    </button>

                    <!-- Go to Page -->
                    <div class="flex items-center gap-2 ml-2 pl-2 border-l border-gray-300 dark:border-gray-600">
                        <div class="flex">
                            <input
                                type="number"
                                min="1"
                                :max="Math.ceil(total / pageSize) || 1"
                                v-model.number="jumpPage"
                                @keyup.enter="handleJump"
                                class="w-10 px-1 py-0.5 text-center bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-l focus:ring-1 focus:ring-blue-500 dark:text-white"
                            >
                            <button
                                @click="handleJump"
                                class="px-2 py-0.5 bg-gray-100 dark:bg-gray-600 border border-l-0 border-gray-300 dark:border-gray-500 rounded-r hover:bg-gray-200 dark:hover:bg-gray-500 text-gray-600 dark:text-gray-300 transition-colors"
                            >
                                Go
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Details Modal -->
        <div v-if="selectedEvent" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm transition-opacity">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col animate-scale-in border border-gray-100 dark:border-gray-700">
                <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex justify-between items-start bg-gray-50/50 dark:bg-gray-800/50">
                    <div>
                        <h3 class="text-xl font-bold text-gray-900 dark:text-white mb-1">{{ t('event.details', 'Event Details') }}</h3>
                        <p class="text-xs font-mono text-gray-400 select-all">{{ selectedEvent.id }}</p>
                    </div>
                    <button @click="selectedEvent = null" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 p-1 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="p-6 overflow-y-auto custom-scrollbar">
                    <div class="grid grid-cols-2 gap-6 mb-6">
                        <div>
                            <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">{{ t('common.level', 'Level') }}</label>
                            <div class="mt-2">
                                <span :class="getLevelClass(selectedEvent.level)" class="px-3 py-1 rounded-full text-xs font-bold border shadow-sm">
                                    {{ selectedEvent.level }}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">{{ t('common.category', 'Category') }}</label>
                            <div class="mt-2">
                                <span class="px-3 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded-full text-xs font-medium border border-gray-200 dark:border-gray-600">
                                    {{ selectedEvent.category }}
                                </span>
                            </div>
                        </div>
                        <div class="col-span-2">
                            <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">{{ t('common.summary', 'Summary') }}</label>
                            <p class="mt-2 text-gray-800 dark:text-gray-200 font-medium text-lg leading-relaxed">{{ selectedEvent.summary }}</p>
                        </div>
                        <div class="col-span-2" v-if="selectedEvent.details">
                            <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">{{ t('common.details', 'Details') }}</label>
                            <div class="mt-2 relative group">
                                <pre class="p-4 bg-gray-900 text-gray-300 rounded-lg text-xs font-mono overflow-x-auto border border-gray-700 shadow-inner">{{ JSON.stringify(selectedEvent.details, null, 2) }}</pre>
                                <button @click="copyToClipboard(JSON.stringify(selectedEvent.details, null, 2))" class="absolute top-2 right-2 p-1.5 bg-gray-800 text-gray-400 rounded hover:text-white opacity-0 group-hover:opacity-100 transition-opacity" title="Copy">
                                    <span class="material-icons text-xs">content_copy</span>
                                </button>
                            </div>
                        </div>
                        <div class="col-span-2 grid grid-cols-2 gap-4 pt-4 border-t border-gray-100 dark:border-gray-700">
                            <div>
                                <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">{{ t('common.time', 'Time') }}</label>
                                <p class="mt-1 text-sm text-gray-600 dark:text-gray-300 font-mono">{{ formatDate(selectedEvent.created_at) }}</p>
                            </div>
                            <div v-if="selectedEvent.source_id">
                                <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">Source ID</label>
                                <p class="mt-1 text-sm text-gray-600 dark:text-gray-300 font-mono">{{ selectedEvent.source_id }}</p>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="p-4 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 flex justify-end">
                    <button @click="selectedEvent = null" class="px-5 py-2 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 font-medium text-sm shadow-sm transition-all hover:shadow">
                        {{ t('common.close', 'Close') }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import http from '../../../utils/http';

export default {
    props: ['t'],
    inject: ['showToast'],
    data() {
        return {
            events: [],
            loading: false,
            page: 1,
            pageSize: 15,
            total: 0,
            filters: {
                level: '',
                category: '',
                date: ''
            },
            searchQuery: '',
            sortBy: 'created_at',
            sortOrder: 'desc',
            selectedEvent: null,
            jumpPage: 1
        }
    },
    mounted() {
        this.fetchEvents();
    },
    methods: {
        async fetchEvents(page = 1) {
            this.loading = true;
            this.page = page;
            this.jumpPage = page;
            try {
                const params = {
                    page: this.page,
                    page_size: this.pageSize,
                    search: this.searchQuery,
                    sort_by: this.sortBy,
                    sort_order: this.sortOrder,
                    ...this.filters
                };
                // Remove empty filters
                Object.keys(params).forEach(key => {
                    if (params[key] === '') delete params[key];
                });

                const res = await http.get('/api/dashboard/events', { params });
                this.events = res.data.items;
                this.total = res.data.total;
            } catch (err) {
                console.error("Failed to fetch events", err);
                this.showToast('Failed to load events', 'error');
            } finally {
                this.loading = false;
            }
        },
        handleJump() {
            const maxPage = Math.ceil(this.total / this.pageSize) || 1;
            let p = parseInt(this.jumpPage);
            if (isNaN(p) || p < 1) p = 1;
            if (p > maxPage) p = maxPage;
            this.fetchEvents(p);
        },
        toggleSort(field) {
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'desc'; // Default to desc for new field
            }
            this.fetchEvents(1);
        },
        getSortIcon(field) {
            if (this.sortBy !== field) return 'unfold_more';
            return this.sortOrder === 'asc' ? 'expand_less' : 'expand_more';
        },
        getSortIconClass(field) {
            if (this.sortBy !== field) return 'text-gray-300 opacity-50';
            return 'text-blue-500 font-bold';
        },
        getLevelClass(level) {
            switch (level) {
                case 'NORMAL': return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800';
                case 'WARNING': return 'bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-800';
                case 'CRITICAL': return 'bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-900/30 dark:text-orange-300 dark:border-orange-800';
                case 'FATAL': return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-800';
                default: return 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700';
            }
        },
        formatDate(timestamp) {
            if (!timestamp) return '-';
            return new Date(timestamp).toLocaleString();
        },
        showDetails(event) {
            this.selectedEvent = event;
        },
        copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(() => {
                this.showToast(this.t('common.copied'), 'success');
            });
        }
    }
}
</script>
