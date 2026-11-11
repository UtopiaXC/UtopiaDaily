<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('schedule.title') }}</h3>
            <button
                @click="fetchJobs"
                class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow"
            >
                <span class="material-icons text-sm mr-1" :class="{'animate-spin': loading}">refresh</span>
                {{ t('common.refresh') }}
            </button>
        </div>

        <div class="p-6 flex-1 overflow-y-auto">
            <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
                <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                    <thead class="bg-gray-50 dark:bg-gray-700/60 text-xs uppercase text-gray-500 dark:text-gray-300">
                        <tr>
                            <th class="px-4 py-3 text-left font-medium">{{ t('schedule.module') }}</th>
                            <th class="px-4 py-3 text-left font-medium">{{ t('schedule.task') }}</th>
                            <th class="px-4 py-3 text-left font-medium">
                                <div class="flex items-center gap-1">
                                    <span>{{ t('schedule.cron') }}</span>
                                    <div class="relative group flex items-center normal-case font-normal">
                                        <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                        <div class="absolute left-1/2 -translate-x-1/2 top-full mt-1.5 w-52 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                            {{ t('schedule.cron_hint') }}
                                        </div>
                                    </div>
                                </div>
                            </th>
                            <th class="px-4 py-3 text-left font-medium">{{ t('common.enable') }}</th>
                            <th class="px-4 py-3 text-left font-medium">{{ t('schedule.last_run') }}</th>
                            <th class="px-4 py-3 text-left font-medium">{{ t('schedule.last_status') }}</th>
                            <th v-if="canEdit" class="px-4 py-3 text-right font-medium">{{ t('common.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
                        <tr v-if="loading && jobs.length === 0">
                            <td :colspan="canEdit ? 7 : 6" class="px-4 py-10 text-center text-gray-400">
                                {{ t('common.loading') }}
                            </td>
                        </tr>
                        <tr v-else-if="jobs.length === 0">
                            <td :colspan="canEdit ? 7 : 6" class="px-4 py-10 text-center text-gray-400">
                                {{ t('common.no_data') }}
                            </td>
                        </tr>
                        <tr v-for="job in jobs" :key="job.id" class="hover:bg-gray-50/70 dark:hover:bg-gray-700/30">
                            <td class="px-4 py-3.5 align-top">
                                <div class="font-medium text-gray-900 dark:text-white">{{ t(job.module_name, job.module_id) }}</div>
                                <span
                                    v-if="!job.module_enabled"
                                    class="inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
                                >
                                    {{ t('schedule.module_disabled') }}
                                </span>
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <div class="font-medium text-gray-800 dark:text-gray-200">{{ t(job.name_key, job.task_key) }}</div>
                                <div v-if="hasTranslation(job.description_key)" class="text-xs text-gray-400 mt-0.5">
                                    {{ t(job.description_key) }}
                                </div>
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <input
                                    v-model="job.cron_expr"
                                    type="text"
                                    :disabled="!canEdit"
                                    class="w-36 font-mono text-xs border border-gray-300 dark:border-gray-600 rounded px-2.5 py-1.5 bg-white dark:bg-gray-700 text-gray-900 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-900"
                                >
                            </td>
                            <td class="px-4 py-3.5 align-top">
                                <label class="inline-flex items-center cursor-pointer pt-1">
                                    <input
                                        type="checkbox"
                                        v-model="job.enabled"
                                        :disabled="!canEdit"
                                        class="rounded border-gray-300 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                    >
                                </label>
                            </td>
                            <td class="px-4 py-3.5 align-top whitespace-nowrap font-mono text-xs text-gray-500 dark:text-gray-400">
                                {{ job.last_run_at ? formatDate(job.last_run_at) : t('schedule.never') }}
                            </td>
                            <td class="px-4 py-3.5 align-top whitespace-nowrap">
                                <div class="flex items-center gap-1.5">
                                    <span :class="statusBadgeClass(job.last_status)" class="px-2 py-0.5 text-[11px] font-semibold rounded-full border">
                                        {{ t('schedule.status.' + (job.last_status || 'idle'), job.last_status || 'idle') }}
                                    </span>
                                    <button
                                        v-if="job.last_error"
                                        @click="openError(job)"
                                        class="text-red-500 hover:text-red-700 flex items-center"
                                        :title="t('events.details')"
                                    >
                                        <span class="material-icons text-sm">error_outline</span>
                                    </button>
                                </div>
                            </td>
                            <td v-if="canEdit" class="px-4 py-3.5 align-top whitespace-nowrap text-right space-x-2">
                                <button
                                    @click="saveJob(job)"
                                    :disabled="savingId === job.id"
                                    class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium disabled:opacity-50"
                                >
                                    {{ savingId === job.id ? t('common.saving') : t('common.save') }}
                                </button>
                                <button
                                    @click="runJob(job)"
                                    :disabled="runningId === job.id"
                                    class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded-md text-xs font-medium disabled:opacity-50"
                                >
                                    {{ runningId === job.id ? t('common.running') : t('schedule.run_now') }}
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
            savingId: 0,
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
        async saveJob(job) {
            this.savingId = job.id;
            try {
                await http.put(`/api/dashboard/schedule/${job.id}`, {
                    cron_expr: job.cron_expr.trim(),
                    enabled: Boolean(job.enabled)
                });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchJobs();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingId = 0;
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
                    return 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800';
                case 'failed':
                    return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800';
                default:
                    return 'bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600';
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
