<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header flex justify-between items-center">
            <div>
                <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('user_push.title') }}</h3>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{{ t('user_push.subtitle') }}</p>
            </div>
            <button @click="fetchData" class="ui-btn-secondary">
                <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                <span>{{ t('common.refresh') }}</span>
            </button>
        </div>

        <div v-if="loading" class="p-12 text-center text-slate-500 dark:text-slate-400 flex-1">
            <span class="material-icons animate-spin text-2xl mb-2 text-blue-600">sync</span>
            <p class="text-xs">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-5 flex-1 overflow-y-auto space-y-5">
            <div v-if="enabledPushModules.length === 0" class="p-4 rounded-md border border-amber-200 dark:border-amber-900 bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-300 text-xs flex items-center gap-2">
                <span class="material-icons text-[18px]">warning</span>
                <span>{{ t('user_push.no_modules') }}</span>
            </div>

            <form v-else @submit.prevent="saveSubscription" class="max-w-2xl space-y-5">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">{{ t('user_push.channel') }}</label>
                        <CustomSelect
                            v-model="form.push_module_id"
                            @update:model-value="onChannelChange"
                            :options="pushModuleOptions"
                        />
                    </div>

                    <div class="flex items-end pb-2">
                        <label class="inline-flex items-center gap-2 cursor-pointer">
                            <input
                                type="checkbox"
                                v-model="form.enabled"
                                class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                            >
                            <span class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ t('common.enable') }}</span>
                        </label>
                    </div>
                </div>

                <div>
                    <div class="flex items-center gap-1.5 mb-1.5">
                        <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">{{ t('user_push.target') }}</label>
                        <div class="relative group flex items-center">
                            <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-60 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                {{ t('user_push.target_hint') }}
                            </div>
                        </div>
                    </div>
                    <input
                        v-model="form.target"
                        type="text"
                        :placeholder="currentUser?.email || ''"
                        class="ui-input"
                    >
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                        <div class="flex items-center gap-1.5 mb-1.5">
                            <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">{{ t('user_push.cron') }}</label>
                            <div class="relative group flex items-center">
                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-52 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t('schedule.cron_hint') }}
                                </div>
                            </div>
                        </div>
                        <input v-model="form.cron_expr" type="text" class="ui-input font-mono text-xs">
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">{{ t('user_push.digest_title') }}</label>
                        <input v-model="form.digest_title" type="text" class="ui-input">
                    </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">{{ t('user_push.max_items') }}</label>
                        <input v-model.number="form.max_items" type="number" min="1" max="200" class="ui-input">
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">{{ t('user_push.lookback_hours') }}</label>
                        <input v-model.number="form.lookback_hours" type="number" min="1" max="720" class="ui-input">
                    </div>
                </div>

                <div>
                    <div class="flex items-center gap-1.5 mb-2">
                        <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">{{ t('user_push.sources') }}</label>
                        <div class="relative group flex items-center">
                            <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-60 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                {{ t('user_push.sources_hint') }}
                            </div>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5 p-4 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-950/40">
                        <label
                            v-for="mod in scraperModules"
                            :key="mod.module_id"
                            class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer"
                        >
                            <input
                                type="checkbox"
                                :value="mod.module_id"
                                v-model="form.source_ids"
                                class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500"
                            >
                            <span>{{ t(mod.name, mod.module_id) }}</span>
                        </label>
                    </div>
                </div>

                <div class="flex flex-wrap items-center justify-between gap-3 pt-4 border-t border-slate-200 dark:border-slate-800">
                    <div class="flex gap-2">
                        <button type="button" @click="sendTest" :disabled="testing" class="ui-btn-emerald">
                            <span class="material-icons text-[16px]" :class="{ 'animate-spin': testing }">
                                {{ testing ? 'sync' : 'send' }}
                            </span>
                            <span>{{ testing ? t('common.testing') : t('user_push.test_send') }}</span>
                        </button>
                        <button
                            v-if="form.id > 0"
                            type="button"
                            @click="confirmDelete"
                            class="ui-btn-secondary text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                        >
                            {{ t('common.delete') }}
                        </button>
                    </div>

                    <button type="submit" :disabled="saving" class="ui-btn-primary">
                        <span class="material-icons text-[16px]" :class="{ 'animate-spin': saving }">
                            {{ saving ? 'sync' : 'save' }}
                        </span>
                        <span>{{ saving ? t('common.saving') : t('common.save') }}</span>
                    </button>
                </div>
            </form>
        </div>

        <ConfirmModal
            :show="deleteConfirm"
            :title="t('common.delete')"
            :message="t('user_push.title')"
            :t="t"
            @confirm="deleteSubscription"
            @cancel="deleteConfirm = false"
        />
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';

export default {
    components: { CustomSelect, ConfirmModal },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            loading: false,
            saving: false,
            testing: false,
            deleteConfirm: false,
            subscriptions: [],
            pushModules: [],
            scraperModules: [],
            form: {
                id: 0,
                push_module_id: '',
                enabled: true,
                target: '',
                cron_expr: '0 8 * * *',
                digest_title: 'Daily Digest',
                max_items: 25,
                lookback_hours: 24,
                source_ids: []
            }
        };
    },
    computed: {
        enabledPushModules() {
            return this.pushModules.filter(m => m.is_enable);
        },
        pushModuleOptions() {
            return this.enabledPushModules.map(m => ({
                value: m.module_id,
                label: this.t(m.name, m.module_id)
            }));
        }
    },
    mounted() {
        this.fetchData();
    },
    methods: {
        async fetchData() {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/user-push');
                const data = res.data || {};
                this.subscriptions = data.subscriptions || [];
                this.pushModules = data.push_modules || [];
                this.scraperModules = data.scraper_modules || [];

                const firstMod = this.enabledPushModules[0];
                if (firstMod) {
                    const hasCurrent = this.enabledPushModules.some(m => m.module_id === this.form.push_module_id);
                    const targetModId = hasCurrent ? this.form.push_module_id : firstMod.module_id;
                    this.loadModuleIntoForm(targetModId);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        loadModuleIntoForm(moduleId) {
            const existing = this.subscriptions.find(s => s.push_module_id === moduleId);
            if (existing) {
                this.form = {
                    id: existing.id,
                    push_module_id: existing.push_module_id,
                    enabled: Boolean(existing.enabled),
                    target: existing.target || '',
                    cron_expr: existing.cron_expr || '0 8 * * *',
                    digest_title: existing.digest_title || 'Daily Digest',
                    max_items: existing.max_items || 25,
                    lookback_hours: existing.lookback_hours || 24,
                    source_ids: Array.isArray(existing.source_ids) ? [...existing.source_ids] : []
                };
            } else {
                this.form = {
                    id: 0,
                    push_module_id: moduleId,
                    enabled: true,
                    target: this.currentUser?.email || '',
                    cron_expr: '0 8 * * *',
                    digest_title: 'Daily Digest',
                    max_items: 25,
                    lookback_hours: 24,
                    source_ids: []
                };
            }
        },
        onChannelChange(moduleId) {
            this.loadModuleIntoForm(moduleId);
        },
        async saveSubscription() {
            this.saving = true;
            try {
                const res = await http.post('/api/dashboard/user-push', {
                    push_module_id: this.form.push_module_id,
                    enabled: Boolean(this.form.enabled),
                    target: this.form.target.trim(),
                    cron_expr: this.form.cron_expr.trim(),
                    digest_title: this.form.digest_title.trim(),
                    max_items: Number(this.form.max_items) || 25,
                    lookback_hours: Number(this.form.lookback_hours) || 24,
                    source_ids: this.form.source_ids
                });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchData();
                if (res.data?.push_module_id) {
                    this.loadModuleIntoForm(res.data.push_module_id);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.saving = false;
            }
        },
        async sendTest() {
            this.testing = true;
            try {
                await http.post('/api/dashboard/user-push/test', {
                    push_module_id: this.form.push_module_id,
                    target: this.form.target.trim()
                });
                this.showToast(this.t('user_push.test_sent'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.testing = false;
            }
        },
        confirmDelete() {
            if (this.form.id > 0) {
                this.deleteConfirm = true;
            }
        },
        async deleteSubscription() {
            this.deleteConfirm = false;
            if (!this.form.id) return;
            try {
                await http.delete(`/api/dashboard/user-push/${this.form.id}`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchData();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        }
    }
};
</script>
