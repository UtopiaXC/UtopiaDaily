<template>
    <div class="flex flex-col gap-6 h-full overflow-y-auto">
        <div class="flex flex-col md:flex-row gap-4 md:gap-6">
            <div
                class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col"
                :class="selectedModule ? 'hidden md:flex md:w-72 lg:w-80 flex-shrink-0' : 'w-full flex-1 md:w-72 lg:w-80 md:flex-none'"
            >
                <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
                    <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('push.modules') }}</h3>
                    <button
                        v-if="canEdit"
                        @click="reloadModules"
                        :disabled="reloading"
                        class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg text-xs font-medium transition-colors flex items-center gap-1"
                    >
                        <span class="material-icons text-sm" :class="{'animate-spin': reloading}">refresh</span>
                        {{ t('scraper.reload_modules') }}
                    </button>
                    <button
                        v-else
                        @click="fetchModules"
                        class="p-1.5 text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
                        :title="t('common.refresh')"
                    >
                        <span class="material-icons text-sm" :class="{'animate-spin': loadingList}">refresh</span>
                    </button>
                </div>

                <div class="flex-1 overflow-y-auto p-2 space-y-2">
                    <div v-if="loadingList" class="p-6 text-center text-gray-400 text-sm">
                        {{ t('common.loading') }}
                    </div>
                    <div v-else-if="modules.length === 0" class="p-6 text-center text-gray-400 text-sm">
                        {{ t('common.no_data') }}
                    </div>
                    <div
                        v-for="mod in modules"
                        :key="mod.module_id"
                        @click="selectModule(mod)"
                        class="p-3 rounded-lg cursor-pointer transition-colors flex flex-col gap-2 border"
                        :class="selectedModule && selectedModule.module_id === mod.module_id
                            ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800'
                            : 'hover:bg-gray-50 dark:hover:bg-gray-700/50 border-transparent'"
                    >
                        <div class="flex justify-between items-start gap-2">
                            <div class="font-medium text-gray-900 dark:text-white truncate">{{ t(mod.name, mod.module_id) }}</div>
                            <div class="text-xs text-gray-400 font-mono flex-shrink-0">{{ mod.version }}</div>
                        </div>

                        <div class="flex items-center gap-1.5 flex-wrap">
                            <span
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium uppercase tracking-wider"
                                :class="mod.source === 'external'
                                    ? 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-900/20 dark:text-purple-300 dark:border-purple-800'
                                    : 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-900/20 dark:text-blue-300 dark:border-blue-800'"
                            >
                                {{ t('scraper.source.' + mod.source, mod.source) }}
                            </span>
                            <span
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                                :class="mod.is_enable
                                    ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800'
                                    : 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800'"
                            >
                                {{ mod.is_enable ? t('scraper.status.enabled') : t('scraper.status.disabled') }}
                            </span>
                            <span
                                v-if="mod.source === 'external'"
                                class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                                :class="mod.env_ready
                                    ? 'bg-teal-50 text-teal-700 border-teal-200 dark:bg-teal-900/20 dark:text-teal-300 dark:border-teal-800'
                                    : 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:border-amber-800'"
                            >
                                {{ mod.env_ready ? t('scraper.env.ready') : t('scraper.env.pending') }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden"
                :class="selectedModule ? 'w-full flex-1' : 'hidden md:flex md:flex-1'"
            >
                <div v-if="selectedModule" class="flex flex-col">
                    <div class="p-4 md:p-6 border-b border-gray-200 dark:border-gray-700">
                        <div class="flex items-start gap-3 mb-3">
                            <button
                                @click="selectedModule = null"
                                class="md:hidden p-1 -ml-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500"
                            >
                                <span class="material-icons">arrow_back</span>
                            </button>
                            <div class="flex-1 min-w-0">
                                <h2 class="text-xl font-bold text-gray-900 dark:text-white break-words">
                                    {{ t(selectedModule.name, selectedModule.module_id) }}
                                </h2>
                                <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">
                                    {{ t(selectedModule.description, '') }}
                                </p>
                            </div>
                        </div>

                        <div class="flex flex-wrap gap-x-6 gap-y-2 text-xs text-gray-500 dark:text-gray-400 mb-4">
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
                                <span>{{ selectedModule.version }}</span>
                            </span>
                            <span class="flex items-center gap-1">
                                <span class="material-icons text-[14px]">extension</span>
                                <span>{{ t('scraper.source.' + selectedModule.source, selectedModule.source) }}</span>
                            </span>
                        </div>

                        <div v-if="canEdit" class="flex flex-wrap gap-3 justify-end">
                            <div v-if="selectedModule.source === 'external'" class="flex items-center gap-1.5">
                                <button
                                    @click="prepareEnv"
                                    :disabled="preparingEnv"
                                    class="px-3.5 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-100 rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors disabled:opacity-50"
                                >
                                    <span class="material-icons text-sm" :class="{'animate-spin': preparingEnv}">build</span>
                                    {{ preparingEnv ? t('scraper.env.rebuilding') : t('scraper.prepare_env') }}
                                </button>
                                <div class="relative group flex items-center">
                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute right-0 bottom-full mb-2 w-60 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('scraper.prepare_env_hint') }}
                                    </div>
                                </div>
                            </div>

                            <button
                                @click="requestToggleEnable"
                                :disabled="enabling"
                                class="px-4 py-2 rounded-lg text-white transition-colors text-sm font-medium flex items-center gap-1.5 disabled:opacity-50"
                                :class="selectedModule.is_enable ? 'bg-red-500 hover:bg-red-600' : 'bg-emerald-600 hover:bg-emerald-700'"
                            >
                                <span class="material-icons text-sm" :class="{'animate-spin': enabling}">
                                    {{ enabling ? 'sync' : (selectedModule.is_enable ? 'power_settings_new' : 'play_arrow') }}
                                </span>
                                {{ enabling ? t('common.enabling') : (selectedModule.is_enable ? t('common.disable') : t('common.enable')) }}
                            </button>
                        </div>
                    </div>

                    <div class="p-4 md:p-6 space-y-6">
                        <div v-if="loadingDetail" class="flex justify-center py-8">
                            <span class="material-icons animate-spin text-3xl text-blue-500">sync</span>
                        </div>

                        <template v-else-if="moduleDetail">
                            <div class="space-y-4 bg-gray-50/70 dark:bg-gray-900/30 p-5 rounded-xl border border-gray-200 dark:border-gray-700">
                                <h4 class="text-sm font-bold text-gray-800 dark:text-white flex items-center gap-2">
                                    <span class="material-icons text-blue-500 text-lg">tune</span>
                                    {{ t('scraper.config.settings') }}
                                </h4>

                                <div v-if="!moduleDetail.config || moduleDetail.config.length === 0" class="text-sm text-gray-400 py-4 text-center">
                                    {{ t('push.no_config') }}
                                </div>

                                <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div v-for="field in sortedConfigFields" :key="field.key" class="flex flex-col gap-1.5">
                                        <div class="flex items-center gap-1.5">
                                            <label class="text-sm font-semibold text-gray-700 dark:text-gray-200">
                                                {{ t(field.description, field.key) }}
                                            </label>
                                            <div v-if="hasTranslation(field.hint)" class="relative group flex items-center">
                                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                                    {{ t(field.hint) }}
                                                </div>
                                            </div>
                                        </div>

                                        <div v-if="field.type === 'array'" class="space-y-2">
                                            <div class="flex flex-wrap gap-2">
                                                <span
                                                    v-for="(item, idx) in (formValues[field.key] || [])"
                                                    :key="idx"
                                                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 border border-blue-200 dark:border-blue-800"
                                                >
                                                    <span>{{ item }}</span>
                                                    <button
                                                        v-if="canEdit && !isFieldDisabled(field)"
                                                        type="button"
                                                        @click="removeArrayItem(field.key, idx)"
                                                        class="hover:text-red-500 flex items-center"
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
                                                    class="flex-1 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                                                >
                                                <button
                                                    type="button"
                                                    @click="addArrayItem(field.key)"
                                                    class="px-3 py-1.5 bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-gray-800 dark:text-white rounded-md text-xs font-medium flex items-center gap-1"
                                                >
                                                    <span class="material-icons text-sm">add</span>
                                                    {{ t('common.add') }}
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
                                            class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:opacity-60"
                                        >

                                        <input
                                            v-else
                                            v-model="formValues[field.key]"
                                            :type="field.secret || field.type === 'password' ? 'password' : 'text'"
                                            :disabled="!canEdit || isFieldDisabled(field)"
                                            class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:opacity-60"
                                        >
                                    </div>
                                </div>

                                <div v-if="canEdit" class="flex flex-wrap justify-end gap-3 pt-2">
                                    <button
                                        @click="testModule"
                                        :disabled="testing"
                                        class="px-4 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-800 dark:text-white rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors disabled:opacity-50"
                                    >
                                        <span class="material-icons text-sm" :class="{'animate-spin': testing}">
                                            {{ testing ? 'sync' : 'wifi_tethering' }}
                                        </span>
                                        {{ testing ? t('common.testing') : t('scraper.config.test') }}
                                    </button>
                                    <button
                                        @click="saveConfig"
                                        :disabled="saving"
                                        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors shadow-sm disabled:opacity-50"
                                    >
                                        <span class="material-icons text-sm" :class="{'animate-spin': saving}">
                                            {{ saving ? 'sync' : 'save' }}
                                        </span>
                                        {{ saving ? t('common.saving') : t('common.save') }}
                                    </button>
                                </div>
                            </div>

                            <div class="space-y-3">
                                <h4 class="text-sm font-bold text-gray-800 dark:text-white flex items-center gap-2">
                                    <span class="material-icons text-blue-500 text-lg">playlist_play</span>
                                    {{ t('scraper.config.tasks') }}
                                </h4>
                                <div
                                    v-for="task in (moduleDetail.tasks || [])"
                                    :key="task.task_key"
                                    class="p-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
                                >
                                    <div>
                                        <div class="flex items-center gap-2 flex-wrap">
                                            <span class="font-semibold text-sm text-gray-900 dark:text-white">
                                                {{ t(task.name, task.task_key) }}
                                            </span>
                                            <span class="text-xs font-mono px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
                                                {{ task.cron_expr }}
                                            </span>
                                        </div>
                                        <p v-if="hasTranslation(task.description)" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                            {{ t(task.description) }}
                                        </p>
                                    </div>
                                    <button
                                        v-if="canEdit"
                                        @click="runTask(task.task_key)"
                                        :disabled="runningTask === task.task_key"
                                        class="px-3.5 py-2 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/40 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors disabled:opacity-50 flex-shrink-0"
                                    >
                                        <span class="material-icons text-sm" :class="{'animate-spin': runningTask === task.task_key}">
                                            {{ runningTask === task.task_key ? 'sync' : 'send' }}
                                        </span>
                                        {{ runningTask === task.task_key ? t('common.running') : t('scraper.run_task') }}
                                    </button>
                                </div>
                            </div>
                        </template>
                    </div>
                </div>

                <div v-else class="flex-1 flex flex-col items-center justify-center text-gray-400 dark:text-gray-500 p-8 text-center">
                    <span class="material-icons text-5xl mb-3 opacity-30">send</span>
                    <p class="text-sm">{{ t('scraper.select_module') }}</p>
                </div>
            </div>
        </div>

        <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <div class="flex justify-between items-center mb-4">
                <h3 class="text-base font-bold text-gray-800 dark:text-white flex items-center gap-2">
                    <span class="material-icons text-blue-500 text-lg">history</span>
                    {{ t('push.logs') }}
                </h3>
                <button
                    @click="fetchLogs"
                    class="p-1.5 text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
                    :title="t('common.refresh')"
                >
                    <span class="material-icons text-sm" :class="{'animate-spin': loadingLogs}">refresh</span>
                </button>
            </div>

            <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
                <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
                    <thead class="bg-gray-50 dark:bg-gray-700/60 text-xs uppercase text-gray-500 dark:text-gray-300">
                        <tr>
                            <th class="px-4 py-2.5 text-left font-medium">{{ t('push.time') }}</th>
                            <th class="px-4 py-2.5 text-left font-medium">{{ t('push.module') }}</th>
                            <th class="px-4 py-2.5 text-left font-medium">{{ t('push.recipient') }}</th>
                            <th class="px-4 py-2.5 text-left font-medium">{{ t('push.items') }}</th>
                            <th class="px-4 py-2.5 text-left font-medium">{{ t('common.status') }}</th>
                            <th class="px-4 py-2.5 text-right font-medium">{{ t('events.details') }}</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                        <tr v-if="logs.length === 0">
                            <td colspan="6" class="px-4 py-6 text-center text-gray-400 text-xs">
                                {{ t('push.logs_empty') }}
                            </td>
                        </tr>
                        <tr v-for="log in logs" :key="log.id">
                            <td class="px-4 py-2.5 font-mono text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ formatDate(log.created_at) }}</td>
                            <td class="px-4 py-2.5 text-gray-800 dark:text-gray-200 font-mono text-xs">{{ log.push_module_id }}</td>
                            <td class="px-4 py-2.5 text-gray-700 dark:text-gray-300 text-xs">{{ log.recipient }}</td>
                            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400 text-xs">{{ log.item_count }}</td>
                            <td class="px-4 py-2.5">
                                <span
                                    class="px-2 py-0.5 text-[11px] font-semibold rounded-full border"
                                    :class="log.status === 'success'
                                        ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800'
                                        : 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800'"
                                >
                                    {{ t('push.status.' + log.status, log.status) }}
                                </span>
                            </td>
                            <td class="px-4 py-2.5 text-right">
                                <button
                                    v-if="log.details"
                                    @click="openLogDetail(log)"
                                    class="text-blue-600 dark:text-blue-400 hover:underline text-xs"
                                >
                                    {{ t('events.details') }}
                                </button>
                                <span v-else class="text-xs text-gray-400">-</span>
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
    props: ['t', 'currentUser'],
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
                if (this.selectedModule) {
                    const updated = this.modules.find(m => m.module_id === this.selectedModule.module_id);
                    if (updated) this.selectedModule = updated;
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
            await this.fetchModuleDetail(mod.module_id);
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
