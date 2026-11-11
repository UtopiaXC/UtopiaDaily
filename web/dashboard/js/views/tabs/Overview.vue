<template>
    <div class="space-y-6 animate-fade-in">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-5 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">{{ t('dashboard.api_status') }}</p>
                        <div class="flex items-baseline gap-2 mt-1">
                            <p class="text-xl font-bold text-gray-800 dark:text-white">{{ t('common.online') }}</p>
                            <span v-if="stats.version" class="text-xs font-mono text-gray-400 dark:text-gray-500">v{{ stats.version }}</span>
                        </div>
                    </div>
                    <div class="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-full text-blue-600 dark:text-blue-400">
                        <span class="material-icons text-2xl">api</span>
                    </div>
                </div>
            </div>

            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-5 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">{{ t('dashboard.db_status') }}</p>
                        <p class="text-xl font-bold text-gray-800 dark:text-white mt-1">{{ t('common.connected') }}</p>
                    </div>
                    <div class="p-3 bg-green-50 dark:bg-green-900/20 rounded-full text-green-600 dark:text-green-400">
                        <span class="material-icons text-2xl">storage</span>
                    </div>
                </div>
            </div>

            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-5 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">{{ t('dashboard.active_modules') }}</p>
                        <p class="text-2xl font-bold text-gray-800 dark:text-white mt-1">{{ stats.active_modules }}</p>
                    </div>
                    <div class="p-3 bg-purple-50 dark:bg-purple-900/20 rounded-full text-purple-600 dark:text-purple-400">
                        <span class="material-icons text-2xl">extension</span>
                    </div>
                </div>
            </div>

            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-5 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">{{ t('dashboard.scraped_articles') }}</p>
                        <p class="text-2xl font-bold text-gray-800 dark:text-white mt-1">{{ stats.scraped_articles }}</p>
                    </div>
                    <div class="p-3 bg-amber-50 dark:bg-amber-900/20 rounded-full text-amber-600 dark:text-amber-400">
                        <span class="material-icons text-2xl">article</span>
                    </div>
                </div>
            </div>

            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-5 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">{{ t('dashboard.classification_slot') }}</p>
                        <p class="text-base font-bold text-gray-800 dark:text-white mt-1.5">{{ t('dashboard.slot_reserved') }}</p>
                    </div>
                    <div class="p-3 bg-gray-100 dark:bg-gray-700 rounded-full text-gray-500 dark:text-gray-300">
                        <span class="material-icons text-2xl">hub</span>
                    </div>
                </div>
            </div>
        </div>

        <EventTable :t="t" />
    </div>
</template>

<script>
import http from '../../utils/http.js';
import EventTable from '../components/EventTable.vue';

export default {
    components: { EventTable },
    props: ['t'],
    data() {
        return {
            stats: {
                api_status: 'online',
                db_status: 'connected',
                version: '',
                active_modules: 0,
                scraped_articles: 0,
                classification_slot: {
                    provider: 'slot',
                    slot_ready: true
                }
            }
        };
    },
    mounted() {
        this.fetchOverview();
    },
    methods: {
        async fetchOverview() {
            try {
                const res = await http.get('/api/dashboard/overview');
                if (res.data) {
                    this.stats = res.data;
                }
            } catch {
                // keep default counters
            }
        }
    }
};
</script>
