<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 rounded-t-xl flex flex-wrap justify-between items-center gap-3">
            <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('frontend_cfg.title') }}</h3>
            <div class="flex items-center gap-2">
                <a
                    href="/"
                    target="_blank"
                    class="text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 text-sm flex items-center transition-colors border border-blue-200 dark:border-blue-800 px-3 py-1.5 rounded-md font-medium"
                >
                    <span class="material-icons text-sm mr-1">open_in_new</span>
                    {{ t('frontend_cfg.open_newspaper') }}
                </a>
                <button
                    @click="fetchConfigs"
                    class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow"
                >
                    <span class="material-icons text-sm mr-1" :class="{'animate-spin': loading}">refresh</span>
                    {{ t('common.refresh') }}
                </button>
            </div>
        </div>

        <div v-if="loading" class="p-12 text-center text-gray-500 dark:text-gray-400 flex-1">
            <span class="material-icons animate-spin text-3xl mb-2 text-blue-500">sync</span>
            <p class="text-sm">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-6 space-y-6 overflow-y-auto flex-1">
            <div v-if="configs.length === 0" class="text-center text-gray-500 dark:text-gray-400 py-8">
                {{ t('common.no_data') }}
            </div>

            <div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm">
                <div class="px-4 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700 font-semibold text-gray-700 dark:text-gray-200 text-xs uppercase tracking-wide rounded-t-lg">
                    {{ t('config.group.newspaper') }}
                </div>

                <div class="p-4 space-y-5">
                    <div v-for="conf in configs" :key="conf.key" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5">
                            <label class="text-sm font-semibold text-gray-800 dark:text-gray-200">
                                {{ t(conf.description, conf.key) }}
                            </label>
                            <div v-if="conf.hint" class="relative group flex items-center">
                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t(conf.hint) }}
                                </div>
                            </div>
                        </div>

                        <div class="flex flex-col sm:flex-row gap-3 sm:items-center">
                            <input
                                v-model="conf.value"
                                :type="conf.type === 'int' ? 'number' : 'text'"
                                :disabled="!conf.is_editable || !canEdit"
                                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-gray-800"
                            >
                            <button
                                v-if="conf.is_editable && canEdit"
                                @click="saveConfig(conf)"
                                :disabled="savingKey === conf.key"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-1.5 transition-colors shadow-sm disabled:opacity-50 flex-shrink-0"
                            >
                                <span class="material-icons text-sm" :class="{'animate-spin': savingKey === conf.key}">
                                    {{ savingKey === conf.key ? 'sync' : 'save' }}
                                </span>
                                {{ savingKey === conf.key ? t('common.saving') : t('common.save') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';

export default {
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            configs: [],
            loading: false,
            savingKey: ''
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('frontend.config.edit');
        }
    },
    mounted() {
        this.fetchConfigs();
    },
    methods: {
        async fetchConfigs() {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/frontend-config');
                this.configs = (res.data || []).sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        async saveConfig(conf) {
            this.savingKey = conf.key;
            try {
                await http.put(`/api/dashboard/frontend-config/${encodeURIComponent(conf.key)}`, {
                    value: String(conf.value ?? '')
                });
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingKey = '';
            }
        }
    }
};
</script>
