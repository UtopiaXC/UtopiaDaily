<template>
    <div class="space-y-5">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
            <div class="ui-card p-4">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.api_status') }}</p>
                        <div class="flex items-baseline gap-2 mt-1.5">
                            <p class="text-lg font-semibold text-slate-900 dark:text-white">{{ t('common.online') }}</p>
                            <span v-if="stats.version" class="text-xs font-mono text-slate-400">v{{ stats.version }}</span>
                        </div>
                    </div>
                    <div class="p-2 rounded-md border border-blue-200 dark:border-blue-900/60 bg-blue-50 dark:bg-blue-950/40 text-blue-600 dark:text-blue-400">
                        <span class="material-icons text-[20px]">api</span>
                    </div>
                </div>
            </div>

            <div class="ui-card p-4">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.db_status') }}</p>
                        <p class="text-lg font-semibold text-slate-900 dark:text-white mt-1.5">{{ t('common.connected') }}</p>
                    </div>
                    <div class="p-2 rounded-md border border-emerald-200 dark:border-emerald-900/60 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400">
                        <span class="material-icons text-[20px]">storage</span>
                    </div>
                </div>
            </div>

            <div class="ui-card p-4">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.active_modules') }}</p>
                        <p class="text-2xl font-semibold text-slate-900 dark:text-white mt-1 font-mono">{{ stats.active_modules }}</p>
                    </div>
                    <div class="p-2 rounded-md border border-indigo-200 dark:border-indigo-900/60 bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400">
                        <span class="material-icons text-[20px]">extension</span>
                    </div>
                </div>
            </div>

            <div class="ui-card p-4">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.scraped_articles') }}</p>
                        <p class="text-2xl font-semibold text-slate-900 dark:text-white mt-1 font-mono">{{ stats.scraped_articles }}</p>
                    </div>
                    <div class="p-2 rounded-md border border-amber-200 dark:border-amber-900/60 bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400">
                        <span class="material-icons text-[20px]">article</span>
                    </div>
                </div>
            </div>

            <button
                type="button"
                @click="$emit('navigate-tab', 'classification')"
                class="ui-card p-4 text-left hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
            >
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.classification_slot') }}</p>
                        <p class="text-base font-semibold text-slate-900 dark:text-white mt-1.5 truncate">
                            {{ providerLabel }}
                        </p>
                        <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-1 flex items-center gap-2 font-mono">
                            <span>{{ workerStateLabel }}</span>
                            <span>{{ stats.classification_slot?.pending_tasks || 0 }}</span>
                        </p>
                    </div>
                    <div class="p-2 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                        <span class="material-icons text-[20px]">auto_awesome</span>
                    </div>
                </div>
            </button>
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
    emits: ['navigate-tab'],
    data() {
        return {
            stats: {
                api_status: 'online',
                db_status: 'connected',
                version: '',
                active_modules: 0,
                scraped_articles: 0,
                classification_slot: {
                    provider: 'none',
                    slot_ready: true,
                    tag_group_count: 0,
                    tag_count: 0,
                    worker_state: 'idle',
                    pending_tasks: 0
                }
            }
        };
    },
    computed: {
        providerLabel() {
            const p = this.stats.classification_slot?.provider || 'none';
            return this.t('config.classification.provider.' + p, p);
        },
        workerStateLabel() {
            const s = this.stats.classification_slot?.worker_state || 'idle';
            return this.t('classification.queue_state.' + s, s);
        }
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
