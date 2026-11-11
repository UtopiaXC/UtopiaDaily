<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header flex flex-wrap justify-between items-center gap-3">
            <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('frontend_cfg.title') }}</h3>
            <div class="flex items-center gap-2">
                <a
                    href="/"
                    target="_blank"
                    class="ui-btn-secondary text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60"
                >
                    <span class="material-icons text-[16px]">open_in_new</span>
                    <span>{{ t('frontend_cfg.open_newspaper') }}</span>
                </a>
                <button @click="fetchAll" class="ui-btn-secondary">
                    <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                    <span>{{ t('common.refresh') }}</span>
                </button>
            </div>
        </div>

        <div v-if="loading" class="p-12 text-center text-slate-500 dark:text-slate-400 flex-1">
            <span class="material-icons animate-spin text-2xl mb-2 text-blue-600">sync</span>
            <p class="text-xs">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-5 space-y-5 overflow-y-auto flex-1">
            <div class="ui-card">
                <div class="ui-card-header">
                    <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                        {{ t('frontend_cfg.general_title') }}
                    </span>
                    <button
                        v-if="canEdit"
                        @click="saveGeneral"
                        :disabled="savingGeneral"
                        class="ui-btn-primary text-xs py-1.5 px-3"
                    >
                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingGeneral }">
                            {{ savingGeneral ? 'sync' : 'save' }}
                        </span>
                        <span>{{ savingGeneral ? t('common.saving') : t('frontend_cfg.save_general') }}</span>
                    </button>
                </div>

                <div v-if="configs.length === 0" class="p-8 text-center text-slate-400 text-sm">
                    {{ t('common.no_data') }}
                </div>

                <div v-else class="p-4 grid grid-cols-1 md:grid-cols-2 gap-5">
                    <div v-for="conf in configs" :key="conf.key" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5">
                            <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">
                                {{ t(conf.description, conf.key) }}
                            </label>
                            <div v-if="conf.hint" class="relative group flex items-center">
                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t(conf.hint) }}
                                </div>
                            </div>
                        </div>

                        <CustomSelect
                            v-if="conf.type === 'select' && conf.is_editable && canEdit"
                            v-model="conf.value"
                            :options="formatOptions(conf.options)"
                        />
                        <CustomSelect
                            v-else-if="conf.type === 'boolean' && conf.is_editable && canEdit"
                            v-model="conf.value"
                            :options="[
                                { label: t('common.true'), value: 'true' },
                                { label: t('common.false'), value: 'false' }
                            ]"
                        />
                        <input
                            v-else-if="conf.type === 'int'"
                            v-model="conf.value"
                            type="number"
                            :disabled="!conf.is_editable || !canEdit"
                            class="ui-input"
                        >
                        <input
                            v-else
                            v-model="conf.value"
                            type="text"
                            :disabled="!conf.is_editable || !canEdit"
                            class="ui-input"
                        >
                    </div>
                </div>
            </div>

            <div class="ui-card">
                <div class="ui-card-header">
                    <div>
                        <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider block">
                            {{ t('frontend_cfg.modules_title') }}
                        </span>
                    </div>
                    <button
                        v-if="canEdit"
                        @click="saveModules"
                        :disabled="savingModules"
                        class="ui-btn-primary text-xs py-1.5 px-3"
                    >
                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingModules }">
                            {{ savingModules ? 'sync' : 'save' }}
                        </span>
                        <span>{{ savingModules ? t('common.saving') : t('frontend_cfg.save_modules') }}</span>
                    </button>
                </div>

                <div class="px-4 py-2.5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-950/30 text-xs text-slate-500 dark:text-slate-400">
                    {{ t('frontend_cfg.modules_desc') }}
                </div>

                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                        <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                            <tr>
                                <th class="px-4 py-2.5 text-left font-semibold">{{ t('frontend_cfg.col_module') }}</th>
                                <th class="px-4 py-2.5 text-left font-semibold w-28">{{ t('frontend_cfg.col_visible') }}</th>
                                <th class="px-4 py-2.5 text-left font-semibold w-44">{{ t('frontend_cfg.col_lookback') }}</th>
                                <th class="px-4 py-2.5 text-left font-semibold w-44">{{ t('frontend_cfg.col_max_items') }}</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                            <tr v-if="moduleConfigs.length === 0">
                                <td colspan="4" class="px-4 py-8 text-center text-xs text-slate-400">{{ t('common.no_data') }}</td>
                            </tr>
                            <tr v-for="mod in moduleConfigs" :key="mod.module_id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                                <td class="px-4 py-3">
                                    <div class="flex items-center gap-2 flex-wrap">
                                        <span class="font-medium text-slate-900 dark:text-white">{{ t(mod.module_name, mod.module_id) }}</span>
                                        <span
                                            v-if="!mod.module_enabled"
                                            class="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700"
                                        >
                                            {{ t('frontend_cfg.disabled_badge') }}
                                        </span>
                                    </div>
                                    <div class="text-[11px] font-mono text-slate-400 mt-0.5">{{ mod.module_id }}</div>
                                </td>
                                <td class="px-4 py-3">
                                    <label class="inline-flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="checkbox"
                                            v-model="mod.visible"
                                            :disabled="!canEdit"
                                            class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                        >
                                    </label>
                                </td>
                                <td class="px-4 py-3">
                                    <input
                                        v-model="mod.lookback_days_override"
                                        type="number"
                                        min="0"
                                        :disabled="!canEdit"
                                        :placeholder="t('frontend_cfg.use_global')"
                                        class="ui-input py-1.5 text-xs font-mono"
                                    >
                                </td>
                                <td class="px-4 py-3">
                                    <input
                                        v-model="mod.max_items_override"
                                        type="number"
                                        min="0"
                                        :disabled="!canEdit"
                                        :placeholder="t('frontend_cfg.use_global')"
                                        class="ui-input py-1.5 text-xs font-mono"
                                    >
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';

export default {
    components: { CustomSelect },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            configs: [],
            moduleConfigs: [],
            loading: false,
            savingGeneral: false,
            savingModules: false
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('frontend.config.edit');
        }
    },
    mounted() {
        this.fetchAll();
    },
    methods: {
        formatOptions(options) {
            return (options || []).map(opt => {
                if (typeof opt === 'object' && opt !== null) {
                    return {
                        value: opt.value,
                        label: opt.label_key ? this.t(opt.label_key, opt.label || opt.value) : (opt.label || opt.value)
                    };
                }
                return { value: opt, label: String(opt) };
            });
        },
        async fetchAll() {
            this.loading = true;
            try {
                const [cfgRes, modRes] = await Promise.all([
                    http.get('/api/dashboard/frontend-config'),
                    http.get('/api/dashboard/frontend-config/modules')
                ]);
                this.configs = (cfgRes.data || [])
                    .map(c => ({ ...c, value: c.value ?? '' }))
                    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
                this.moduleConfigs = (modRes.data || []).map(m => ({
                    ...m,
                    lookback_days_override: m.lookback_days_override > 0 ? m.lookback_days_override : '',
                    max_items_override: m.max_items_override > 0 ? m.max_items_override : ''
                }));
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        async saveGeneral() {
            const payload = {};
            this.configs.forEach(conf => {
                if (!conf.is_editable) return;
                payload[conf.key] = String(conf.value ?? '');
            });
            this.savingGeneral = true;
            try {
                await http.put('/api/dashboard/frontend-config', { configs: payload });
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingGeneral = false;
            }
        },
        async saveModules() {
            const items = this.moduleConfigs.map((m, idx) => ({
                module_id: m.module_id,
                visible: Boolean(m.visible),
                lookback_days_override: Math.max(0, parseInt(m.lookback_days_override, 10) || 0),
                max_items_override: Math.max(0, parseInt(m.max_items_override, 10) || 0),
                sort_order: typeof m.sort_order === 'number' ? m.sort_order : idx
            }));
            this.savingModules = true;
            try {
                await http.put('/api/dashboard/frontend-config/modules', { modules: items });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchAll();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingModules = false;
            }
        }
    }
};
</script>
