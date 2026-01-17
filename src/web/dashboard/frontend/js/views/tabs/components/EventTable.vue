<template>
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <h3 class="text-lg font-bold text-gray-800 dark:text-white flex items-center gap-2">
                <span class="material-icons text-blue-500">notifications</span>
                {{ t('dashboard.events.title', 'System Events') }}
            </h3>

            <div class="flex flex-wrap gap-2">
                <select v-model="filters.level" @change="fetchEvents(1)" class="px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-gray-200">
                    <option value="">{{ t('common.all_levels', 'All Levels') }}</option>
                    <option value="NORMAL">NORMAL</option>
                    <option value="WARNING">WARNING</option>
                    <option value="CRITICAL">CRITICAL</option>
                    <option value="FATAL">FATAL</option>
                </select>

                <select v-model="filters.category" @change="fetchEvents(1)" class="px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-gray-200">
                    <option value="">{{ t('common.all_categories', 'All Categories') }}</option>
                    <option value="SYSTEM">SYSTEM</option>
                    <option value="USER">USER</option>
                    <option value="MODULE">MODULE</option>
                    <option value="TASK">TASK</option>
                    <option value="SECURITY">SECURITY</option>
                </select>

                <button @click="fetchEvents(1)" class="p-2 text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 transition-colors">
                    <span class="material-icons">refresh</span>
                </button>
            </div>
        </div>

        <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
                <thead>
                    <tr class="bg-gray-50 dark:bg-gray-700/50 text-gray-500 dark:text-gray-400 text-xs uppercase tracking-wider">
                        <th class="p-4 font-medium">{{ t('common.level', 'Level') }}</th>
                        <th class="p-4 font-medium">{{ t('common.category', 'Category') }}</th>
                        <th class="p-4 font-medium">{{ t('common.summary', 'Summary') }}</th>
                        <th class="p-4 font-medium">{{ t('common.time', 'Time') }}</th>
                        <th class="p-4 font-medium text-right">{{ t('common.actions', 'Actions') }}</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                    <tr v-if="loading" class="animate-pulse">
                        <td colspan="5" class="p-4 text-center text-gray-400">{{ t('common.loading', 'Loading...') }}</td>
                    </tr>
                    <tr v-else-if="events.length === 0">
                        <td colspan="5" class="p-8 text-center text-gray-400">
                            <span class="material-icons text-4xl mb-2 opacity-20">inbox</span>
                            <p>{{ t('common.no_data', 'No events found') }}</p>
                        </td>
                    </tr>
                    <tr v-for="event in events" :key="event.id" class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors group">
                        <td class="p-4">
                            <span :class="getLevelClass(event.level)" class="px-2 py-1 rounded text-xs font-medium border">
                                {{ event.level }}
                            </span>
                        </td>
                        <td class="p-4 text-sm text-gray-600 dark:text-gray-300">{{ event.category }}</td>
                        <td class="p-4">
                            <div class="text-sm font-medium text-gray-800 dark:text-white">{{ event.summary }}</div>
                            <div class="text-xs text-gray-400 mt-0.5">{{ event.event_type }}</div>
                        </td>
                        <td class="p-4 text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">
                            {{ formatDate(event.created_at) }}
                        </td>
                        <td class="p-4 text-right">
                            <button @click="showDetails(event)" class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 text-sm font-medium">
                                {{ t('common.details', 'Details') }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Pagination -->
        <div class="p-4 border-t border-gray-100 dark:border-gray-700 flex items-center justify-between" v-if="total > 0">
            <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('common.showing', 'Showing') }} {{ (page - 1) * pageSize + 1 }} - {{ Math.min(page * pageSize, total) }} {{ t('common.of', 'of') }} {{ total }}
            </div>
            <div class="flex gap-2">
                <button
                    @click="fetchEvents(page - 1)"
                    :disabled="page === 1"
                    class="px-3 py-1 border border-gray-200 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm dark:text-gray-300"
                >
                    {{ t('common.prev', 'Prev') }}
                </button>
                <button
                    @click="fetchEvents(page + 1)"
                    :disabled="page * pageSize >= total"
                    class="px-3 py-1 border border-gray-200 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm dark:text-gray-300"
                >
                    {{ t('common.next', 'Next') }}
                </button>
            </div>
        </div>

        <!-- Details Modal -->
        <div v-if="selectedEvent" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm" @click.self="selectedEvent = null">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col animate-scale-in">
                <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex justify-between items-start">
                    <div>
                        <h3 class="text-xl font-bold text-gray-900 dark:text-white mb-1">{{ t('event.details', 'Event Details') }}</h3>
                        <p class="text-sm text-gray-500 dark:text-gray-400">{{ selectedEvent.id }}</p>
                    </div>
                    <button @click="selectedEvent = null" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="p-6 overflow-y-auto">
                    <div class="grid grid-cols-2 gap-4 mb-6">
                        <div>
                            <label class="text-xs font-semibold text-gray-400 uppercase">{{ t('common.level', 'Level') }}</label>
                            <div class="mt-1">
                                <span :class="getLevelClass(selectedEvent.level)" class="px-2 py-1 rounded text-xs font-medium border">
                                    {{ selectedEvent.level }}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="text-xs font-semibold text-gray-400 uppercase">{{ t('common.category', 'Category') }}</label>
                            <p class="mt-1 text-gray-800 dark:text-gray-200">{{ selectedEvent.category }}</p>
                        </div>
                        <div class="col-span-2">
                            <label class="text-xs font-semibold text-gray-400 uppercase">{{ t('common.summary', 'Summary') }}</label>
                            <p class="mt-1 text-gray-800 dark:text-gray-200 font-medium">{{ selectedEvent.summary }}</p>
                        </div>
                        <div class="col-span-2" v-if="selectedEvent.details">
                            <label class="text-xs font-semibold text-gray-400 uppercase">{{ t('common.details', 'Details') }}</label>
                            <pre class="mt-2 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg text-xs font-mono overflow-x-auto text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700">{{ JSON.stringify(selectedEvent.details, null, 2) }}</pre>
                        </div>
                    </div>
                </div>
                <div class="p-6 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 flex justify-end">
                    <button @click="selectedEvent = null" class="px-4 py-2 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 font-medium text-sm">
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
    data() {
        return {
            events: [],
            loading: false,
            page: 1,
            pageSize: 10,
            total: 0,
            filters: {
                level: '',
                category: ''
            },
            selectedEvent: null
        }
    },
    mounted() {
        this.fetchEvents();
    },
    methods: {
        async fetchEvents(page = 1) {
            this.loading = true;
            this.page = page;
            try {
                const params = {
                    page: this.page,
                    page_size: this.pageSize,
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
            } finally {
                this.loading = false;
            }
        },
        getLevelClass(level) {
            switch (level) {
                case 'NORMAL': return 'bg-blue-50 text-blue-700 border-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-900/30';
                case 'WARNING': return 'bg-yellow-50 text-yellow-700 border-yellow-100 dark:bg-yellow-900/20 dark:text-yellow-400 dark:border-yellow-900/30';
                case 'CRITICAL': return 'bg-orange-50 text-orange-700 border-orange-100 dark:bg-orange-900/20 dark:text-orange-400 dark:border-orange-900/30';
                case 'FATAL': return 'bg-red-50 text-red-700 border-red-100 dark:bg-red-900/20 dark:text-red-400 dark:border-red-900/30';
                default: return 'bg-gray-50 text-gray-700 border-gray-100 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700';
            }
        },
        formatDate(isoString) {
            if (!isoString) return '-';
            return new Date(isoString).toLocaleString();
        },
        showDetails(event) {
            this.selectedEvent = event;
        }
    }
}
</script>
