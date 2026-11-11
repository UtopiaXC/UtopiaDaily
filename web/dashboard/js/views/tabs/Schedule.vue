<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header flex flex-wrap justify-between items-center gap-3">
            <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('schedule.title') }}</h3>
            <div class="flex items-center gap-2">
                <button
                    v-if="canEdit && jobs.length > 0"
                    @click="saveAll"
                    :disabled="savingAll"
                    class="ui-btn-primary text-xs py-1.5 px-3"
                >
                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingAll }">
                        {{ savingAll ? 'sync' : 'save' }}
                    </span>
                    <span>{{ savingAll ? t('common.saving') : t('schedule.save_all') }}</span>
                </button>
                <button @click="fetchJobs" class="ui-btn-secondary">
                    <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                    <span>{{ t('common.refresh') }}</span>
                </button>
            </div>
        </div>

        <div class="p-5 flex-1 overflow-y-auto">
            <div class="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-lg">
                <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                    <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                        <tr>
                            <th class="px-4 py-3 text-left font-semibold">{{ t('schedule.module') }}</th>
                            <th class="px-4 py-3 text-left font-semibold">{{ t('schedule.task') }}</th>
                            <th class="px-4 py-3 text-left font-semibold">
                                <div class="flex items-center gap-1">
                                    <span>{{ t('schedule.cron') }}</span>
                                    <div class="relative group flex items-center normal-case font-normal">
                                        <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                        <div class="absolute left-1/2 -translate-x-1/2 top-full mt-1.5 w-52 p-2 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                            {{ t('schedule.cron_hint') }}
                                        </div>
                                    </div>
                                </div>
                            </th>
                            <th class="px-4 py-3 text-left font-semibold">{{ t('common.enable') }}</th>
                            <th class="px-4 py-3 text-left font-semibold">{{ t('schedule.last_run') }}</th>
                            <th class="px-4 py-3 text-left font-semibold">{{ t('schedule.last_status') }}</th>
                            <th v-if="canEdit" class="px-4 py-3 text-right font-semibold">{{ t('common.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                        <tr v-if="loading && jobs.length === 0">
                            <td :colspan="canEdit ? 7 : 6" class="px-4 py-10 text-center text-slate-400 text-xs">
                                {{ t('common.loading') }}
                            </td>
                        </tr>
                        <tr v-else-if="jobs.length === 0">
                            <td :colspan="canEdit ? 7 : 6" class="px-4 py-10 text-center text-slate-400 text-xs">
                                {{ t('common.no_data') }}
                            </td>
                        </tr>
                        <tr v-for="job in jobs" :key="job.id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                            <td class="px-4 py-3.5 align-top">
                                <div class="font-medium text-slate-900 dark:text-white">{{ t(job.module_name, job.module_id) }}</div>
                                <span
                                    v-if="!job.module_enabled"
                                    class="inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700"
                                >
                                    {{ t('schedule.module_disabled') }}
                                </span>
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <div class="font-medium text-slate-800 dark:text-slate-200">{{ t(job.name_key, job.task_key) }}</div>
                                <div v-if="hasTranslation(job.description_key)" class="text-xs text-slate-400 mt-0.5">
                                    {{ t(job.description_key) }}
                                </div>
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <input
                                    v-model="job.cron_expr"
                                    type="text"
                                    :disabled="!canEdit"
                                    class="ui-input w-36 font-mono text-xs py-1.5 px-2.5"
                                >
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <label class="inline-flex items-center cursor-pointer pt-1">
                                    <input
                                        type="checkbox"
                                        v-model="job.enabled"
                                        :disabled="!canEdit"
                                        class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                    >
                                </label>
                            </td>
                            <td class="px-4 py-3.5 align-top whitespace-nowrap font-mono text-xs text-slate-500 dark:text-slate-400">
                                {{ job.last_run_at ? formatDate(job.last_run_at) : t('schedule.never') }}
                            </td>
                            <td class="px-4 py-3.5 align-top whitespace-nowrap">
                                <div class="flex items-center gap-1.5">
                                    <span :class="statusBadgeClass(job.last_status)" class="px-2 py-0.5 text-[11px] font-medium rounded border">
                                        {{ t('schedule.status.' + (job.last_status || 'idle'), job.last_status || 'idle') }}
                                    </span>
                                    <button
                                        v-if="job.last_error"
                                        @click="openError(job)"
                                        class="text-rose-500 hover:text-rose-700 flex items-center"
                                        :title="t('events.details')"
                                    >
                                        <span class="material-icons text-[16px]">error_outline</span>
                                    </button>
                                </div>
                            </td>
                            <td v-if="canEdit" class="px-4 py-3.5 align-top whitespace-nowrap text-right">
                                <button
                                    @click="runJob(job)"
                                    :disabled="runningId === job.id"
                                    class="ui-btn-secondary text-xs py-1 px-2.5"
                                >
                                    <span class="material-icons text-[14px]" :class="{ 'animate-spin': runningId === job.id }">
                                        {{ runningId === job.id ? 'sync' : 'play_arrow' }}
                                    </span>
                                    <span>{{ runningId === job.id ? t('common.running') : t('schedule.run_now') }}</span>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <InfoModal
            :show="errorModal.show"
            :title="t('events.details')"
            :message="errorModal.message"
            :t="t"
            @close="errorModal.show = false"
        />
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import InfoModal from '../../components/InfoModal.vue';

export default {
    components: { InfoModal },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            jobs: [],
            loading: false,
            savingAll: false,
            runningId: 0,
            errorModal: { show: false, message: '' }
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('schedule.edit');
        }
    },
    mounted() {
        this.fetchJobs();
    },
    methods: {
        hasTranslation(key) {
            if (!key) return false;
            const v = this.t(key, '');
            return Boolean(v && v !== key);
        },
        async fetchJobs() {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/schedule');
                this.jobs = res.data || [];
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        async saveAll() {
            this.savingAll = true;
            try {
                const payload = this.jobs.map(job => ({
                    id: job.id,
                    cron_expr: String(job.cron_expr || '').trim(),
                    enabled: Boolean(job.enabled)
                }));
                await http.put('/api/dashboard/schedule', { jobs: payload });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchJobs();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingAll = false;
            }
        },
        async runJob(job) {
            this.runningId = job.id;
            try {
                await http.post(`/api/dashboard/schedule/${job.id}/run`);
                this.showToast(this.t('scraper.task_started'), 'success');
                await this.fetchJobs();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
                await this.fetchJobs();
            } finally {
                this.runningId = 0;
            }
        },
        openError(job) {
            this.errorModal = { show: true, message: job.last_error };
        },
        statusBadgeClass(status) {
            switch (status) {
                case 'success':
                    return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900';
                case 'failed':
                    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
                default:
                    return 'bg-slate-50 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
            }
        },
        formatDate(iso) {
            if (!iso) return '-';
            const d = new Date(iso);
            if (Number.isNaN(d.getTime())) return iso;
            const pad = n => String(n).padStart(2, '0');
            return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
        }
    }
};
</script>
