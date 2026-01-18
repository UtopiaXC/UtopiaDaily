<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <!-- Header & Tabs -->
        <div class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <div class="px-6 py-4 flex justify-between items-center">
                <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('menu.system_config') }}</h3>
                <button @click="fetchSystemConfigs" class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow">
                    <span class="material-icons text-sm mr-1">refresh</span> {{ t('common.refresh') }}
                </button>
            </div>

            <!-- Tabs -->
            <div class="flex px-6 gap-6 overflow-x-auto">
                <button
                    v-for="tab in sortedTabs"
                    :key="tab"
                    @click="currentTab = tab"
                    :class="['pb-3 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        currentTab === tab
                        ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                        : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300']"
                >
                    {{ t('config.tab.' + tab, tab.toUpperCase()) }}
                </button>
            </div>
        </div>

        <div v-if="loading" class="p-12 text-center text-gray-500 dark:text-gray-400 flex-1">
            <span class="material-icons animate-spin text-3xl mb-2 text-blue-500">sync</span>
            <p>{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-6 space-y-8 overflow-y-auto flex-1">
            <div v-if="!currentTabConfigs || Object.keys(currentTabConfigs).length === 0" class="text-center text-gray-500 dark:text-gray-400 py-8">
                {{ t('common.no_data') }}
            </div>

            <!-- Groups as Cards -->
            <div v-for="(configs, group) in currentTabConfigs" :key="group" class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm">
                <div class="px-4 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700 font-semibold text-gray-700 dark:text-gray-200 text-sm uppercase tracking-wide rounded-t-lg">
                    {{ t('config.group.' + group, group) }}
                </div>

                <div class="p-4 space-y-6">
                    <div v-for="conf in configs" :key="conf.key" class="flex flex-col gap-1">

                        <!-- Layout for JSON Type: Label + Hint + Buttons in Header -->
                        <template v-if="conf.type === 'json'">
                            <div class="flex justify-between items-center mb-1">
                                <div class="flex items-center gap-2">
                                    <label class="text-sm font-bold text-gray-800 dark:text-gray-200">{{ t(conf.description, conf.description) }}</label>
                                    <div v-if="conf.hint" class="relative group">
                                        <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                        <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                            {{ t(conf.hint) }}
                                        </div>
                                    </div>
                                </div>
                                <div class="flex gap-2">
                                    <button v-if="conf.is_editable" @click="loadDefault(conf.key)"
                                            class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium flex items-center gap-1 transition-colors shadow-sm">
                                        <span class="material-icons text-xs">restore</span> {{ t('config.ai.tags.load_default') }}
                                    </button>
                                    <button v-if="conf.is_editable" @click="saveConfig(conf)"
                                            class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium flex items-center gap-1 transition-colors shadow-sm"
                                            :disabled="savingConfig === conf.key">
                                        <span class="material-icons text-xs" :class="{'animate-spin': savingConfig === conf.key}">save</span>
                                        {{ savingConfig === conf.key ? t('common.saving') : t('common.save') }}
                                    </button>
                                </div>
                            </div>
                            <div class="w-full">
                                <TagConfigEditor :ref="'editor_' + conf.key" v-model="conf.value" :t="t" :current-locale="currentLocale" />
                            </div>
                        </template>

                        <!-- Layout for Model Select Type -->
                        <template v-else-if="conf.type === 'model_select'">
                            <div class="flex items-center gap-2 mb-1">
                                <label class="text-sm font-bold text-gray-800 dark:text-gray-200">{{ t(conf.description, conf.description) }}</label>
                                <div v-if="conf.hint" class="relative group">
                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                        {{ t(conf.hint) }}
                                    </div>
                                </div>
                            </div>
                            <div class="flex gap-3 items-center">
                                <div class="flex-1 flex gap-2">
                                    <CustomSelect
                                        v-if="conf.is_editable"
                                        v-model="conf.value"
                                        :options="conf.options"
                                        class="flex-1"
                                    />
                                    <input v-else type="text" :value="conf.value" disabled class="flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed">

                                    <button @click="openDownloadModal(conf.options)" class="px-3 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg transition-colors shadow-sm text-sm font-medium whitespace-nowrap">
                                        {{ t('config.ai.model.download') }}
                                    </button>
                                    <button @click="refreshModels" class="px-3 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg transition-colors shadow-sm text-sm font-medium whitespace-nowrap flex items-center gap-1">
                                        <span v-if="refreshingModels" class="material-icons text-sm animate-spin">sync</span>
                                        {{ t('config.ai.model.refresh') }}
                                    </button>
                                </div>

                                <button v-if="conf.is_editable" @click="saveConfig(conf)"
                                        class="px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-medium flex items-center gap-1 transition-colors shadow-sm flex-shrink-0 h-[38px]"
                                        :disabled="savingConfig === conf.key">
                                    <span class="material-icons text-sm" :class="{'animate-spin': savingConfig === conf.key}">save</span>
                                    {{ savingConfig === conf.key ? t('common.saving') : t('common.save') }}
                                </button>
                            </div>
                        </template>

                        <!-- Layout for Other Types: Label + Hint on top, Input + Save Button in Row -->
                        <template v-else>
                            <div class="flex items-center gap-2 mb-1">
                                <label class="text-sm font-bold text-gray-800 dark:text-gray-200">{{ t(conf.description, conf.description) }}</label>
                                <div v-if="conf.hint" class="relative group">
                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                        {{ t(conf.hint) }}
                                    </div>
                                </div>
                            </div>
                            <div class="flex gap-3 items-center">
                                <div class="flex-1">
                                    <!-- String -->
                                    <input v-if="conf.type === 'string'" v-model="conf.value" type="text"
                                           :disabled="!conf.is_editable"
                                           :class="['w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500 transition bg-white dark:bg-gray-700 dark:text-white', !conf.is_editable ? 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed' : '']">

                                    <!-- Int -->
                                    <input v-else-if="conf.type === 'int'" v-model="conf.value" type="number"
                                           :disabled="!conf.is_editable"
                                           :class="['w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500 transition bg-white dark:bg-gray-700 dark:text-white', !conf.is_editable ? 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed' : '']">

                                    <!-- Boolean -->
                                    <div v-else-if="conf.type === 'boolean'">
                                        <CustomSelect
                                            v-if="conf.is_editable"
                                            v-model="conf.value"
                                            :options="[{label: 'True', value: 'true'}, {label: 'False', value: 'false'}]"
                                        />
                                        <input v-else type="text" :value="conf.value" disabled class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed">
                                    </div>

                                    <!-- Select -->
                                    <div v-else-if="conf.type === 'select'">
                                        <CustomSelect
                                            v-if="conf.is_editable"
                                            v-model="conf.value"
                                            :options="conf.options"
                                        />
                                        <input v-else type="text" :value="conf.value" disabled class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed">
                                    </div>
                                </div>

                                <button v-if="conf.is_editable" @click="saveConfig(conf)"
                                        class="px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-medium flex items-center gap-1 transition-colors shadow-sm flex-shrink-0 h-[38px]"
                                        :disabled="savingConfig === conf.key">
                                    <span class="material-icons text-sm" :class="{'animate-spin': savingConfig === conf.key}">save</span>
                                    {{ savingConfig === conf.key ? t('common.saving') : t('common.save') }}
                                </button>
                            </div>
                        </template>

                    </div>
                </div>
            </div>
        </div>

        <!-- Global Loading Overlay -->
        <div v-if="downloadingModel" class="fixed inset-0 z-[60] flex flex-col items-center justify-center bg-black bg-opacity-70 backdrop-blur-sm">
            <span class="material-icons animate-spin text-5xl text-white mb-4">autorenew</span>
            <p class="text-white text-lg font-medium">{{ t('common.downloading') }}</p>
            <p class="text-gray-300 text-sm mt-2">{{ t('common.downloading_hint') }}</p>
        </div>

        <ConfirmModal
            :show="showLoadDefaultModal"
            :title="t('config.ai.tags.load_default')"
            :message="t('config.ai.tags.confirm_load_default')"
            :t="t"
            @confirm="confirmLoadDefault"
            @cancel="showLoadDefaultModal = false"
        />

        <ModelDownloadModal
            :show="showDownloadModal"
            :t="t"
            :installed-models="currentInstalledModels"
            @confirm="downloadModel"
            @cancel="showDownloadModal = false"
        />
    </div>
</template>

<script>
import CustomSelect from '../../components/CustomSelect.vue';
import TagConfigEditor from '../components/TagConfigEditor.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';
import ModelDownloadModal from '../components/ModelDownloadModal.vue';
import http from '../../utils/http.js';

export default {
    components: { CustomSelect, TagConfigEditor, ConfirmModal, ModelDownloadModal },
    props: ['t', 'allLocales', 'currentLocale'],
    inject: ['showToast'],
    data() {
        return {
            loading: false,
            configs: [],
            savingConfig: null,
            currentTab: 'system',
            showLoadDefaultModal: false,
            pendingLoadDefaultKey: null,
            showDownloadModal: false,
            refreshingModels: false,
            downloadingModel: false,
            currentInstalledModels: []
        }
    },
    computed: {
        groupedConfigs() {
            const groups = {};
            this.configs.forEach(conf => {
                const tab = conf.tab || 'system';
                const group = conf.group || 'basic';

                if (!groups[tab]) groups[tab] = {};
                if (!groups[tab][group]) groups[tab][group] = [];

                groups[tab][group].push(conf);
            });
            return groups;
        },
        sortedTabs() {
            // Ensure 'system' is first, others sorted alphabetically or by custom order
            const tabs = Object.keys(this.groupedConfigs);
            return tabs.sort((a, b) => {
                if (a === 'system') return -1;
                if (b === 'system') return 1;
                return a.localeCompare(b);
            });
        },
        currentTabConfigs() {
            return this.groupedConfigs[this.currentTab] || {};
        }
    },
    mounted() {
        this.fetchSystemConfigs();
    },
    methods: {
        async fetchSystemConfigs() {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/system-config/');
                this.configs = res.data.map(conf => {
                    if (conf.key === 'default_locale') {
                        conf.options = this.allLocales;
                    }
                    return conf;
                });

                // Set initial tab if current is invalid
                if (!this.groupedConfigs[this.currentTab] && this.sortedTabs.length > 0) {
                    this.currentTab = this.sortedTabs[0];
                }
            } catch (err) {
                console.error("Failed to fetch configs", err);
            } finally {
                this.loading = false;
            }
        },
        async saveConfig(conf) {
            this.savingConfig = conf.key;
            try {
                await http.put(`/api/dashboard/system-config/${conf.key}`, {
                    value: conf.value
                });
                this.showToast(this.t('common.saved_success', 'Configuration saved'), 'success');
                if (conf.key === 'server_name') {
                    window.dispatchEvent(new CustomEvent('server-name-updated', { detail: conf.value }));
                }
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.saved_failed', 'Failed to save: ') + msg, 'error');
            } finally {
                this.savingConfig = null;
            }
        },
        loadDefault(key) {
            this.pendingLoadDefaultKey = key;
            this.showLoadDefaultModal = true;
        },
        confirmLoadDefault() {
            const key = this.pendingLoadDefaultKey;
            const editor = this.$refs['editor_' + key];
            if (editor && editor[0]) {
                editor[0].loadDefault();
            }
            this.showLoadDefaultModal = false;
            this.pendingLoadDefaultKey = null;
        },
        async refreshModels() {
            this.refreshingModels = true;
            try {
                // Just re-fetch configs, the backend will re-scan models
                await this.fetchSystemConfigs();
                this.showToast(this.t('common.success'), 'success');
            } catch (err) {
                this.showToast(this.t('common.failed_op') + err.message, 'error');
            } finally {
                this.refreshingModels = false;
            }
        },
        openDownloadModal(installedModels) {
            this.currentInstalledModels = installedModels || [];
            this.showDownloadModal = true;
        },
        async downloadModel(repoId) {
            this.showDownloadModal = false;
            this.downloadingModel = true; // Show full screen loader

            try {
                const res = await http.post('/api/dashboard/system-config/download-model', { repo_id: repoId });
                this.showToast(res.data.message, 'success');
                await this.fetchSystemConfigs(); // Refresh list
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            } finally {
                this.downloadingModel = false;
            }
        }
    }
}
</script>
