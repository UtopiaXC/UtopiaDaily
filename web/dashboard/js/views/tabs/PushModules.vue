<template>
    <div class="flex flex-col gap-5 h-full overflow-y-auto">
        <div class="flex flex-col md:flex-row gap-4 md:gap-5">
            <div
                class="ui-surface flex flex-col"
                :class="selectedModule ? 'hidden md:flex md:w-72 lg:w-80 flex-shrink-0' : 'w-full flex-1 md:w-72 lg:w-80 md:flex-none'"
            >
                <div class="ui-surface-header px-4 py-3 flex justify-between items-center">
                    <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('push.modules') }}</h3>
                    <button
                        v-if="canEdit"
                        @click="reloadModules"
                        :disabled="reloading"
                        class="ui-btn-secondary py-1 px-2.5 text-xs"
                    >
                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': reloading }">refresh</span>
                        <span>{{ t('scraper.reload_modules') }}</span>
                    </button>
                    <button
                        v-else
                        @click="fetchModules"
                        class="p-1.5 text-slate-500 hover:text-blue-600 dark:text-slate-400 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800"
                        :title="t('common.refresh')"
                    >
                        <span class="material-icons text-[16px]" :class="{ 'animate-spin': loadingList }">refresh</span>
                    </button>
                </div>

                <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
                    <div v-if="loadingList" class="p-6 text-center text-slate-400 text-xs">
                        {{ t('common.loading') }}
                    </div>
                    <div v-else-if="modules.length === 0" class="p-6 text-center text-slate-400 text-xs">
                        {{ t('common.no_data') }}
                    </div>
                    <div
                        v-for="mod in modules"
                        :key="mod.module_id"
                        @click="selectModule(mod)"
                        class="p-3 rounded-md cursor-pointer transition-colors flex flex-col gap-2 border"
                        :class="selectedModule && selectedModule.module_id === mod.module_id
                            ? 'bg-blue-50/70 dark:bg-blue-950/30 border-blue-200 dark:border-blue-900'
                            : 'hover:bg-slate-50 dark:hover:bg-slate-800/50 border-transparent'"
                    >
                        <div class="flex justify-between items-start gap-2">
                            <div class="font-medium text-sm text-slate-900 dark:text-white truncate">{{ t(mod.name, mod.module_id) }}</div>
                            <div class="text-[11px] text-slate-400 font-mono flex-shrink-0">{{ mod.version }}</div>
                        </div>

                        <div class="flex items-center gap-1.5 flex-wrap">
                            <span
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium uppercase tracking-wider"
                                :class="mod.source === 'external'
                                    ? 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/40 dark:text-indigo-300 dark:border-indigo-900'
                                    : 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900'"
                            >
                                {{ t('scraper.source.' + mod.source, mod.source) }}
                            </span>
                            <span
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                                :class="mod.is_enable
                                    ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900'
                                    : 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900'"
                            >
                                {{ mod.is_enable ? t('scraper.status.enabled') : t('scraper.status.disabled') }}
                            </span>
                            <span
                                v-if="mod.source === 'external'"
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                                :class="mod.env_ready
                                    ? 'bg-teal-50 text-teal-700 border-teal-200 dark:bg-teal-950/40 dark:text-teal-300 dark:border-teal-900'
                                    : 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900'"
                            >
                                {{ mod.env_ready ? t('scraper.env.ready') : t('scraper.env.pending') }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="ui-surface flex flex-col overflow-hidden"
                :class="selectedModule ? 'w-full flex-1' : 'hidden md:flex md:flex-1'"
            >
                <div v-if="selectedModule" class="flex flex-col">
                    <div class="p-4 md:p-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-900/40">
                        <div class="flex items-start gap-3 mb-3">
                            <button
                                @click="clearSelectedModule"
                                class="md:hidden p-1 -ml-1 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500"
                            >
                                <span class="material-icons">arrow_back</span>
                            </button>
                            <div class="flex-1 min-w-0">
                                <h2 class="text-lg font-semibold text-slate-900 dark:text-white break-words">
                                    {{ t(selectedModule.name, selectedModule.module_id) }}
                                </h2>
                                <p class="text-slate-500 dark:text-slate-400 text-xs mt-1">
                                    {{ t(selectedModule.description, '') }}
                                </p>
                            </div>
                        </div>

                        <div class="flex flex-wrap gap-x-5 gap-y-1.5 text-xs text-slate-500 dark:text-slate-400 mb-4">
                            <span class="flex items-center gap-1">
                                <span class="material-icons text-[14px]">fingerprint</span>
                                <span class="font-mono">{{ selectedModule.module_id }}</span>
                            </span>
                            <span class="flex items-center gap-1">
                                <span class="material-icons text-[14px]">person</span>
                                <span>{{ selectedModule.author }}</span>
                            </span>
                            <span class="flex items-center gap-1">
                                <span class="material-icons text-[14px]">tag</span>
                                <span class="font-mono">{{ selectedModule.version }}</span>
                            </span>
                            <span class="flex items-center gap-1">
                                <span class="material-icons text-[14px]">extension</span>
                                <span>{{ t('scraper.source.' + selectedModule.source, selectedModule.source) }}</span>
                            </span>
                        </div>

                        <div v-if="canEdit" class="flex flex-wrap gap-2.5 justify-end">
                            <div v-if="selectedModule.source === 'external'" class="flex items-center gap-1.5">
                                <button
                                    @click="prepareEnv"
                                    :disabled="preparingEnv"
                                    class="ui-btn-secondary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': preparingEnv }">build</span>
                                    <span>{{ preparingEnv ? t('scraper.env.rebuilding') : t('scraper.prepare_env') }}</span>
                                </button>
                                <div class="relative group flex items-center">
                                    <span class="material-icons text-slate-400 text-[16px] cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute right-0 bottom-full mb-2 w-60 p-2 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('scraper.prepare_env_hint') }}
                                    </div>
                                </div>
                            </div>

                            <button
                                @click="requestToggleEnable"
                                :disabled="enabling"
                                class="text-xs py-1.5 px-3.5"
                                :class="selectedModule.is_enable ? 'ui-btn-danger' : 'ui-btn-emerald'"
                            >
                                <span class="material-icons text-[15px]" :class="{ 'animate-spin': enabling }">
                                    {{ enabling ? 'sync' : (selectedModule.is_enable ? 'power_settings_new' : 'play_arrow') }}
                                </span>
                                <span>{{ enabling ? t('common.enabling') : (selectedModule.is_enable ? t('common.disable') : t('common.enable')) }}</span>
                            </button>
                        </div>
                    </div>

                    <div class="p-4 md:p-5 space-y-5">
                        <div v-if="loadingDetail" class="flex justify-center py-8">
                            <span class="material-icons animate-spin text-2xl text-blue-600">sync</span>
                        </div>

                        <template v-else-if="moduleDetail">
                            <div class="ui-card">
                                <div class="ui-card-header">
                                    <div class="flex items-center gap-2">
                                        <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">tune</span>
                                        <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                            {{ t('scraper.config.settings') }}
                                        </span>
                                    </div>
                                    <div v-if="canEdit" class="flex items-center gap-2">
                                        <button
                                            @click="testModule"
                                            :disabled="testing"
                                            class="ui-btn-secondary text-xs py-1.5 px-3"
                                        >
                                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': testing }">
                                                {{ testing ? 'sync' : 'wifi_tethering' }}
                                            </span>
                                            <span>{{ testing ? t('common.testing') : t('scraper.config.test') }}</span>
                                        </button>
                                        <button
                                            @click="saveConfig"
                                            :disabled="saving"
                                            class="ui-btn-primary text-xs py-1.5 px-3"
                                        >
                                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': saving }">
                                                {{ saving ? 'sync' : 'save' }}
                                            </span>
                                            <span>{{ saving ? t('common.saving') : t('common.save') }}</span>
                                        </button>
                                    </div>
                                </div>

                                <div class="p-4">
                                    <div v-if="!moduleDetail.config || moduleDetail.config.length === 0" class="text-xs text-slate-400 py-4 text-center">
                                        {{ t('push.no_config') }}
                                    </div>

                                    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div v-for="field in sortedConfigFields" :key="field.key" class="flex flex-col gap-1.5">
                                            <div class="flex items-center gap-1.5">
                                                <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">
                                                    {{ t(field.description, field.key) }}
                                                </label>
                                                <div v-if="hasTranslation(field.hint)" class="relative group flex items-center">
                                                    <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                                        {{ t(field.hint) }}
                                                    </div>
                                                </div>
                                            </div>

                                            <div v-if="field.type === 'array'" class="space-y-2">
                                                <div class="flex flex-wrap gap-1.5">
                                                    <span
                                                        v-for="(item, idx) in (formValues[field.key] || [])"
                                                        :key="idx"
                                                        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200 dark:border-blue-900"
                                                    >
                                                        <span>{{ item }}</span>
                                                        <button
                                                            v-if="canEdit && !isFieldDisabled(field)"
                                                            type="button"
                                                            @click="removeArrayItem(field.key, idx)"
                                                            class="hover:text-rose-500 flex items-center"
                                                        >
                                                            <span class="material-icons text-[14px]">close</span>
                                                        </button>
                                                    </span>
                                                </div>
                                                <div v-if="canEdit && !isFieldDisabled(field)" class="flex gap-2">
                                                    <input
                                                        v-model="arrayDrafts[field.key]"
                                                        @keyup.enter.prevent="addArrayItem(field.key)"
                                                        type="text"
                                                        :placeholder="t('scraper.array_placeholder')"
                                                        class="ui-input flex-1 py-1.5"
                                                    >
                                                    <button
                                                        type="button"
                                                        @click="addArrayItem(field.key)"
                                                        class="ui-btn-secondary py-1.5 px-3 text-xs"
                                                    >
                                                        <span class="material-icons text-[15px]">add</span>
                                                        <span>{{ t('common.add') }}</span>
                                                    </button>
                                                </div>
                                            </div>

                                            <CustomSelect
                                                v-else-if="field.type === 'select'"
                                                v-model="formValues[field.key]"
                                                :options="field.options || []"
                                                :disabled="!canEdit || isFieldDisabled(field)"
                                            />

                                            <CustomSelect
                                                v-else-if="field.type === 'bool' || field.type === 'boolean'"
                                                v-model="formValues[field.key]"
                                                :options="[
                                                    { label: t('common.true'), value: true },
                                                    { label: t('common.false'), value: false }
                                                ]"
                                                :disabled="!canEdit || isFieldDisabled(field)"
                                            />

                                            <input
                                                v-else-if="field.type === 'int'"
                                                v-model="formValues[field.key]"
                                                type="number"
                                                :disabled="!canEdit || isFieldDisabled(field)"
                                                class="ui-input"
                                            >

                                            <input
                                                v-else
                                                v-model="formValues[field.key]"
                                                :type="field.secret || field.type === 'password' ? 'password' : 'text'"
                                                :disabled="!canEdit || isFieldDisabled(field)"
                                                class="ui-input"
                                            >
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="ui-card">
                                <div class="ui-card-header">
                                    <div class="flex items-center gap-2">
                                        <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">playlist_play</span>
                                        <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                            {{ t('scraper.config.tasks') }}
                                        </span>
                                    </div>
                                </div>
                                <div class="divide-y divide-slate-200 dark:divide-slate-800">
                                    <div
                                        v-for="task in (moduleDetail.tasks || [])"
                                        :key="task.task_key"
                                        class="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
                                    >
                                        <div>
                                            <div class="flex items-center gap-2 flex-wrap">
                                                <span class="font-medium text-sm text-slate-900 dark:text-white">
                                                    {{ t(task.name, task.task_key) }}
                                                </span>
                                                <span class="text-xs font-mono px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                                                    {{ task.cron_expr }}
                                                </span>
                                            </div>
                                            <p v-if="hasTranslation(task.description)" class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                                                {{ t(task.description) }}
                                            </p>
                                        </div>
                                        <button
                                            v-if="canEdit"
                                            @click="runTask(task.task_key)"
                                            :disabled="runningTask === task.task_key"
                                            class="ui-btn-secondary text-xs py-1.5 px-3 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60 flex-shrink-0"
                                        >
                                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': runningTask === task.task_key }">
                                                {{ runningTask === task.task_key ? 'sync' : 'send' }}
                                            </span>
                                            <span>{{ runningTask === task.task_key ? t('common.running') : t('scraper.run_task') }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </template>
                    </div>
                </div>

                <div v-else class="flex-1 flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 p-8 text-center">
                    <span class="material-icons text-4xl mb-2 opacity-40">send</span>
                    <p class="text-xs">{{ t('scraper.select_module') }}</p>
                </div>
            </div>
        </div>

        <div class="ui-surface">
            <div class="ui-surface-header flex justify-between items-center">
                <div class="flex items-center gap-2">
                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">history</span>
                    <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('push.logs') }}</h3>
                </div>
                <button
                    @click="fetchLogs"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                    :title="t('common.refresh')"
                >
                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': loadingLogs }">refresh</span>
                    <span>{{ t('common.refresh') }}</span>
                </button>
            </div>

            <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                    <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                        <tr>
                            <th class="px-4 py-2.5 text-left font-semibold">{{ t('push.time') }}</th>
                            <th class="px-4 py-2.5 text-left font-semibold">{{ t('push.module') }}</th>
                            <th class="px-4 py-2.5 text-left font-semibold">{{ t('push.recipient') }}</th>
                            <th class="px-4 py-2.5 text-left font-semibold">{{ t('push.items') }}</th>
                            <th class="px-4 py-2.5 text-left font-semibold">{{ t('common.status') }}</th>
                            <th class="px-4 py-2.5 text-right font-semibold">{{ t('events.details') }}</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                        <tr v-if="logs.length === 0">
                            <td colspan="6" class="px-4 py-6 text-center text-slate-400 text-xs">
                                {{ t('push.logs_empty') }}
                            </td>
                        </tr>
                        <tr v-for="log in logs" :key="log.id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                            <td class="px-4 py-2.5 font-mono text-xs text-slate-500 dark:text-slate-400 whitespace-nowrap">{{ formatDate(log.created_at) }}</td>
                            <td class="px-4 py-2.5 text-slate-800 dark:text-slate-200 font-mono text-xs">{{ log.push_module_id }}</td>
                            <td class="px-4 py-2.5 text-slate-700 dark:text-slate-300 text-xs">{{ log.recipient }}</td>
                            <td class="px-4 py-2.5 text-slate-600 dark:text-slate-400 text-xs font-mono">{{ log.item_count }}</td>
                            <td class="px-4 py-2.5">
                                <span
                                    class="px-2 py-0.5 text-[11px] font-medium rounded border"
                                    :class="log.status === 'success'
                                        ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900'
                                        : 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900'"
                                >
                                    {{ t('push.status.' + log.status, log.status) }}
                                </span>
                            </td>
                            <td class="px-4 py-2.5 text-right">
                                <button
                                    v-if="log.details"
                                    @click="openLogDetail(log)"
                                    class="text-blue-600 dark:text-blue-400 hover:underline text-xs font-medium"
                                >
                                    {{ t('events.details') }}
                                </button>
                                <span v-else class="text-xs text-slate-400">-</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <ConfirmModal
            :show="externalConfirm.show"
            :title="t('scraper.external_warning.title')"
            :message="t('scraper.external_warning.content')"
            :t="t"
            @confirm="performEnableModule"
            @cancel="externalConfirm.show = false"
        />

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
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';
import InfoModal from '../../components/InfoModal.vue';

const OVERRIDABLE_SMTP_KEYS = new Set([
    'smtp_host',
    'smtp_port',
    'smtp_encryption',
    'smtp_user',
    'smtp_password',
    'smtp_from',
    'smtp_from_name'
]);

export default {
    components: { CustomSelect, ConfirmModal, InfoModal },
    props: ['t', 'currentUser', 'subtab'],
    emits: ['update:subtab'],
    inject: ['showToast'],
    data() {
        return {
            modules: [],
            selectedModule: null,
            moduleDetail: null,
            formValues: {},
            arrayDrafts: {},
            loadingList: false,
            loadingDetail: false,
            loadingLogs: false,
            reloading: false,
            preparingEnv: false,
            enabling: false,
            saving: false,
            testing: false,
            runningTask: '',
            externalConfirm: { show: false },
            detailModal: { show: false, title: '', message: '' },
            logs: []
        };
    },
    watch: {
        async subtab(next) {
            const targetId = String(next || '').trim();
            if (!targetId) {
                this.selectedModule = null;
                this.moduleDetail = null;
                return;
            }
            if (this.selectedModule && this.selectedModule.module_id === targetId) {
                return;
            }
            const found = (this.modules || []).find(m => m.module_id === targetId);
            if (found) {
                this.selectedModule = found;
                await this.fetchModuleDetail(found.module_id);
            }
        }
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('push_module.edit');
        },
        sortedConfigFields() {
            const cfg = this.moduleDetail?.config || [];
            return cfg.slice().sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
        }
    },
    mounted() {
        this.fetchModules();
        this.fetchLogs();
    },
    methods: {
        hasTranslation(key) {
            if (!key) return false;
            const v = this.t(key, '');
            return Boolean(v && v !== key);
        },
        isFieldDisabled(field) {
            if (this.selectedModule?.module_id === 'smtp_pusher' && OVERRIDABLE_SMTP_KEYS.has(field.key)) {
                return !this.formValues.override_smtp;
            }
            return false;
        },
        async fetchModules() {
            this.loadingList = true;
            try {
                const res = await http.get('/api/dashboard/push/modules');
                this.modules = res.data || [];
                const targetSubtab = String(this.subtab || '').trim();
                if (this.selectedModule) {
                    const updated = this.modules.find(m => m.module_id === this.selectedModule.module_id);
                    if (updated) this.selectedModule = updated;
                } else if (targetSubtab) {
                    const matched = this.modules.find(m => m.module_id === targetSubtab);
                    if (matched) {
                        this.selectedModule = matched;
                        await this.fetchModuleDetail(matched.module_id);
                    } else {
                        this.$emit('update:subtab', { subtab: '', replace: true });
                    }
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loadingList = false;
            }
        },
        async reloadModules() {
            this.reloading = true;
            try {
                await http.post('/api/dashboard/push/modules/reload');
                await this.fetchModules();
                if (this.selectedModule) {
                    await this.fetchModuleDetail(this.selectedModule.module_id);
                }
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.reloading = false;
            }
        },
        async selectModule(mod) {
            this.selectedModule = mod;
            this.$emit('update:subtab', mod.module_id);
            await this.fetchModuleDetail(mod.module_id);
        },
        clearSelectedModule() {
            this.selectedModule = null;
            this.moduleDetail = null;
            this.$emit('update:subtab', '');
        },
        async fetchModuleDetail(id) {
            this.loadingDetail = true;
            try {
                const res = await http.get(`/api/dashboard/push/modules/${encodeURIComponent(id)}`);
                this.moduleDetail = res.data;
                const vals = {};
                const drafts = {};
                (res.data?.config || []).forEach(f => {
                    if (f.type === 'array') {
                        vals[f.key] = Array.isArray(f.value) ? [...f.value] : [];
                        drafts[f.key] = '';
                    } else if (f.type === 'bool' || f.type === 'boolean') {
                        vals[f.key] = Boolean(f.value === true || f.value === 'true');
                    } else {
                        vals[f.key] = f.value ?? '';
                    }
                });
                this.formValues = vals;
                this.arrayDrafts = drafts;
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loadingDetail = false;
            }
        },
        addArrayItem(key) {
            const val = (this.arrayDrafts[key] || '').trim();
            if (!val) return;
            if (!Array.isArray(this.formValues[key])) {
                this.formValues[key] = [];
            }
            if (!this.formValues[key].includes(val)) {
                this.formValues[key].push(val);
            }
            this.arrayDrafts[key] = '';
        },
        removeArrayItem(key, idx) {
            if (Array.isArray(this.formValues[key])) {
                this.formValues[key].splice(idx, 1);
            }
        },
        requestToggleEnable() {
            if (!this.selectedModule) return;
            if (this.selectedModule.is_enable) {
                this.performDisableModule();
            } else if (this.selectedModule.source === 'external') {
                this.externalConfirm.show = true;
            } else {
                this.performEnableModule();
            }
        },
        async performEnableModule() {
            this.externalConfirm.show = false;
            this.enabling = true;
            try {
                await http.post(`/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/enable`);
                await this.fetchModules();
                await this.fetchModuleDetail(this.selectedModule.module_id);
                this.showToast(this.t('scraper.status.enabled'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.enabling = false;
            }
        },
        async performDisableModule() {
            this.enabling = true;
            try {
                await http.post(`/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/disable`);
                await this.fetchModules();
                await this.fetchModuleDetail(this.selectedModule.module_id);
                this.showToast(this.t('scraper.status.disabled'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.enabling = false;
            }
        },
        async prepareEnv() {
            if (!this.selectedModule) return;
            this.preparingEnv = true;
            try {
                const res = await http.post(`/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/prepare-env`);
                await this.fetchModules();
                this.showToast(this.t('scraper.env.ready'), 'success');
                if (res.data?.output && res.data.output.trim()) {
                    this.detailModal = {
                        show: true,
                        title: this.t('scraper.prepare_env'),
                        message: res.data.output.trim()
                    };
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.preparingEnv = false;
            }
        },
        buildConfigPayload() {
            const payload = {};
            (this.moduleDetail?.config || []).forEach(f => {
                const val = this.formValues[f.key];
                if (f.secret && val === '********') return;
                payload[f.key] = f.type === 'int' ? String(val ?? '0') : val;
            });
            return payload;
        },
        async saveConfig() {
            if (!this.selectedModule) return;
            this.saving = true;
            try {
                await http.post(
                    `/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/config`,
                    this.buildConfigPayload()
                );
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchModuleDetail(this.selectedModule.module_id);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.saving = false;
            }
        },
        async testModule() {
            if (!this.selectedModule) return;
            this.testing = true;
            try {
                const res = await http.post(
                    `/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/test`,
                    this.buildConfigPayload()
                );
                const result = res.data || {};
                const baseMsg = this.t(result.message_key, result.success ? this.t('module.test.passed') : this.t('module.test.failed'));
                const msg = result.detail ? `${baseMsg}: ${result.detail}` : baseMsg;
                this.showToast(msg, result.success ? 'success' : 'error');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.testing = false;
            }
        },
        async runTask(taskKey) {
            if (!this.selectedModule) return;
            this.runningTask = taskKey;
            try {
                await http.post(
                    `/api/dashboard/push/modules/${encodeURIComponent(this.selectedModule.module_id)}/tasks/${encodeURIComponent(taskKey)}/run`
                );
                this.showToast(this.t('scraper.task_started'), 'success');
                await this.fetchLogs();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
                await this.fetchLogs();
            } finally {
                this.runningTask = '';
            }
        },
        async fetchLogs() {
            this.loadingLogs = true;
            try {
                const res = await http.get('/api/dashboard/push/logs', { params: { limit: 30 } });
                this.logs = res.data || [];
            } catch {
                // ignore transient fetch failure
            } finally {
                this.loadingLogs = false;
            }
        },
        openLogDetail(log) {
            this.detailModal = {
                show: true,
                title: this.t('events.details'),
                message: log.details
            };
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
