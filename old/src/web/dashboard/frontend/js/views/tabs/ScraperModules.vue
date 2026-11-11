<template>
    <div class="flex flex-col md:flex-row h-full gap-4 md:gap-6 relative">
        <!-- Module List -->
        <div
            class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col transition-all duration-300"
            :class="[
                selectedModule ? 'hidden md:flex md:w-72 lg:w-80 flex-shrink-0' : 'w-full flex-1 md:w-72 lg:w-80 md:flex-none'
            ]"
        >
            <!-- List Header -->
            <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
                <h3 class="text-lg font-medium text-gray-900 dark:text-white">{{ t('scraper.modules') }}</h3>
                <button
                    @click="reloadModules(false)"
                    class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg text-xs font-medium transition-colors flex items-center gap-1"
                    :title="t('scraper.reload_modules')"
                >
                    <span class="material-icons text-sm" :class="{'animate-spin': reloading}">refresh</span>
                    {{ t('scraper.reload_modules') }}
                </button>
            </div>

            <!-- List Content -->
            <div class="flex-1 overflow-y-auto p-2 space-y-2">
                <div
                    v-for="mod in modules"
                    :key="mod.module_id"
                    @click="selectModule(mod)"
                    class="p-3 rounded-lg cursor-pointer transition-colors duration-200 flex flex-col gap-2 border"
                    :class="selectedModule && selectedModule.module_id === mod.module_id ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800' : 'hover:bg-gray-50 dark:hover:bg-gray-700/50 border-transparent bg-transparent'"
                >
                    <!-- Name & Version -->
                    <div class="flex justify-between items-start">
                        <div class="font-medium text-gray-900 dark:text-white truncate pr-2" :title="t(mod.name)">{{ t(mod.name) }}</div>
                        <div class="text-xs text-gray-400 dark:text-gray-500 font-mono flex-shrink-0">{{ mod.version }}</div>
                    </div>

                    <!-- Badges Row -->
                    <div class="flex items-center justify-between mt-1">
                        <div class="flex items-center gap-2">
                            <!-- Source Badge -->
                            <span class="text-[10px] px-1.5 py-0.5 rounded border font-medium uppercase tracking-wider"
                                :class="getSourceBadgeClass(mod.source)">
                                {{ t('scraper.source.' + mod.source) }}
                            </span>

                            <!-- Status Badge -->
                            <span class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                                :class="mod.is_enable ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800' : 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800'">
                                {{ mod.is_enable ? t('scraper.status.enabled') : t('scraper.status.disabled') }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Module Detail -->
        <div
            class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden transition-all duration-300"
            :class="[
                selectedModule ? 'w-full flex-1' : 'hidden md:flex md:flex-1'
            ]"
        >
            <div v-if="selectedModule" class="flex flex-col h-full">
                <!-- Detail Header -->
                <div class="p-4 md:p-6 border-b border-gray-200 dark:border-gray-700">
                    <!-- Mobile Back Button & Title -->
                    <div class="flex items-start gap-3 mb-4">
                        <button
                            @click="selectedModule = null"
                            class="md:hidden p-1 -ml-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500"
                        >
                            <span class="material-icons">arrow_back</span>
                        </button>
                        <div class="flex-1 min-w-0">
                            <h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white leading-tight break-words">{{ t(selectedModule.name) }}</h2>
                            <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">{{ t(selectedModule.description) }}</p>
                        </div>
                    </div>

                    <!-- Meta Info -->
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
                            <span class="material-icons text-[14px]">folder</span>
                            <span>{{ t('scraper.source.' + selectedModule.source) }}</span>
                        </span>
                    </div>

                    <!-- Actions -->
                    <div class="flex flex-wrap gap-3 justify-end">
                        <button
                            @click="toggleEnable"
                            :disabled="enabling"
                            class="flex-1 md:flex-none px-4 py-2 rounded-lg text-white transition-colors text-sm font-medium flex items-center justify-center gap-2"
                            :class="selectedModule.is_enable ? 'bg-red-500 hover:bg-red-600' : 'bg-green-500 hover:bg-green-600'"
                        >
                            <span class="material-icons text-sm" :class="{'animate-spin': enabling}">
                                {{ enabling ? 'sync' : (selectedModule.is_enable ? 'power_settings_new' : 'play_arrow') }}
                            </span>
                            {{ enabling ? t('common.enabling') : (selectedModule.is_enable ? t('common.disable') : t('common.enable')) }}
                        </button>
                    </div>
                </div>

                <!-- Detail Content -->
                <div class="flex-1 overflow-y-auto p-4 md:p-6 flex flex-col">
                    <div v-if="!moduleDetail" class="flex justify-center py-12">
                        <span class="material-icons animate-spin text-gray-400 text-3xl">refresh</span>
                    </div>

                    <div v-else class="flex-1 flex flex-col">
                        <!-- Config Tabs -->
                        <div class="flex border-b border-gray-200 dark:border-gray-700 mb-6 justify-between items-center">
                            <div class="flex">
                                <button
                                    @click="activeConfigTab = 'general'"
                                    class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
                                    :class="activeConfigTab === 'general' ? 'border-blue-500 text-blue-600 dark:text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
                                >
                                    {{ t('scraper.config.general') }}
                                </button>
                                <button
                                    @click="activeConfigTab = 'custom'"
                                    class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
                                    :class="activeConfigTab === 'custom' ? 'border-blue-500 text-blue-600 dark:text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
                                >
                                    {{ t('scraper.config.custom') }}
                                </button>
                            </div>

                            <!-- Save Button (Top Right) -->
                            <div class="flex gap-2" v-if="(activeConfigTab === 'general' && getConfigsBySource('default').length > 0) || (activeConfigTab === 'custom' && getConfigsBySource('custom').length > 0)">
                                <button
                                    v-if="activeConfigTab === 'custom'"
                                    @click="testConfig"
                                    :disabled="testingConfig"
                                    class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg text-xs font-medium flex items-center gap-1 transition-colors"
                                >
                                    <span class="material-icons text-sm" :class="{'animate-spin': testingConfig}">science</span>
                                    {{ testingConfig ? t('common.testing') : t('scraper.config.test') }}
                                </button>
                                <button
                                    @click="saveConfig"
                                    :disabled="saving"
                                    class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium flex items-center gap-1 transition-colors"
                                >
                                    <span class="material-icons text-sm" :class="{'animate-spin': saving}">save</span>
                                    {{ saving ? t('common.saving') : t('common.save') }}
                                </button>
                            </div>
                        </div>

                        <!-- General Config (Default) -->
                        <div v-if="activeConfigTab === 'general'" class="space-y-6">
                            <div v-if="getConfigsBySource('default').length === 0" class="text-center py-8 text-gray-400">
                                {{ t('scraper.no_config') }}
                            </div>

                            <div v-else class="grid grid-cols-1 gap-6">
                                <div v-for="cfg in getConfigsBySource('default')" :key="cfg.key" class="space-y-2">

                                    <!-- Layout for JSON Type: Label + Hint + Buttons in Header -->
                                    <template v-if="cfg.type === 'json'">
                                        <div class="flex justify-between items-center mb-1">
                                            <div class="flex items-center gap-2">
                                                <label class="text-sm font-bold text-gray-800 dark:text-gray-200" :class="{'opacity-50': isConfigDisabled(cfg.key)}">{{ t(cfg.description, cfg.description) }}</label>
                                                <div v-if="cfg.hint" class="relative group">
                                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                                        {{ t(cfg.hint) }}
                                                    </div>
                                                </div>
                                            </div>
                                            <div class="flex gap-2">
                                                <button @click="loadDefault(cfg.key)"
                                                        :disabled="isConfigDisabled(cfg.key)"
                                                        class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium flex items-center gap-1 transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed">
                                                    <span class="material-icons text-xs">restore</span> {{ t('config.ai.tags.load_default') }}
                                                </button>
                                            </div>
                                        </div>
                                        <div class="w-full" :class="{'opacity-50 pointer-events-none': isConfigDisabled(cfg.key)}">
                                            <TagConfigEditor :ref="'editor_' + cfg.key" v-model="configForm[cfg.key]" :t="t" :current-locale="currentLocale" />
                                        </div>
                                    </template>

                                    <!-- Layout for Model Select Type -->
                                    <template v-else-if="cfg.type === 'model_select'">
                                        <div class="flex items-center gap-2 mb-1">
                                            <label class="text-sm font-bold text-gray-800 dark:text-gray-200" :class="{'opacity-50': isConfigDisabled(cfg.key)}">{{ t(cfg.description, cfg.description) }}</label>
                                            <div v-if="cfg.hint" class="relative group">
                                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                                    {{ t(cfg.hint) }}
                                                </div>
                                            </div>
                                        </div>
                                        <div class="flex gap-3 items-center">
                                            <div class="flex-1 flex gap-2">
                                                <CustomSelect
                                                    v-model="configForm[cfg.key]"
                                                    :options="cfg.options"
                                                    class="flex-1"
                                                    :disabled="isConfigDisabled(cfg.key)"
                                                />

                                                <button @click="openDownloadModal(cfg.options)" :disabled="isConfigDisabled(cfg.key)" class="px-3 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg transition-colors shadow-sm text-sm font-medium whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed">
                                                    {{ t('config.ai.model.download') }}
                                                </button>
                                                <button @click="refreshModels" :disabled="isConfigDisabled(cfg.key)" class="px-3 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg transition-colors shadow-sm text-sm font-medium whitespace-nowrap flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed">
                                                    <span v-if="refreshingModels" class="material-icons text-sm animate-spin">sync</span>
                                                    {{ t('config.ai.model.refresh') }}
                                                </button>
                                            </div>
                                        </div>
                                    </template>

                                    <template v-else>
                                        <div class="flex items-center gap-2 mb-1">
                                            <label class="text-sm font-bold text-gray-800 dark:text-gray-200" :class="{'opacity-50': isConfigDisabled(cfg.key)}">{{ t(cfg.description, cfg.description) }}</label>
                                            <div v-if="cfg.hint" class="relative group">
                                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                                    {{ t(cfg.hint) }}
                                                </div>
                                            </div>
                                        </div>
                                        <div :class="{'opacity-50 pointer-events-none': isConfigDisabled(cfg.key)}">
                                            <component :is="getInputComponent(cfg.type)" :cfg="cfg" v-model="configForm[cfg.key]" :t="t" />
                                        </div>
                                    </template>
                                </div>
                            </div>
                        </div>

                        <!-- Custom Config -->
                        <div v-if="activeConfigTab === 'custom'" class="space-y-6">
                            <div v-if="getConfigsBySource('custom').length === 0" class="text-center py-8 text-gray-400">
                                {{ t('scraper.no_config') }}
                            </div>

                            <div v-else class="grid grid-cols-1 gap-6">
                                <div v-for="cfg in getConfigsBySource('custom')" :key="cfg.key" class="space-y-2">
                                    <div class="flex items-center gap-2 mb-1">
                                        <label class="text-sm font-bold text-gray-800 dark:text-gray-200">{{ t(cfg.description, cfg.description) }}</label>
                                        <div v-if="cfg.hint" class="relative group">
                                            <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2 bg-gray-800 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                                {{ t(cfg.hint) }}
                                            </div>
                                        </div>
                                    </div>
                                    <component :is="getInputComponent(cfg.type)" :cfg="cfg" v-model="configForm[cfg.key]" :t="t" />
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Empty State (Desktop) -->
            <div v-else class="hidden md:flex flex-1 items-center justify-center text-gray-400 flex-col h-full">
                <span class="material-icons text-6xl mb-4 text-gray-300 dark:text-gray-600">extension</span>
                <p>{{ t('scraper.select_module') }}</p>
            </div>
        </div>

        <!-- Warning Modal -->
        <div v-if="showWarningModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 p-4">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6 border border-gray-200 dark:border-gray-700">
                <div class="flex items-center mb-4 text-orange-500">
                    <span class="material-icons text-3xl mr-2">warning</span>
                    <h3 class="text-xl font-bold text-gray-900 dark:text-white">{{ t('scraper.external_warning.title') }}</h3>
                </div>
                <p class="text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
                    {{ t('scraper.external_warning.content') }}
                </p>
                <div class="flex justify-end space-x-3">
                    <button
                        @click="showWarningModal = false"
                        class="px-4 py-2 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                    >
                        {{ t('common.cancel') }}
                    </button>
                    <button
                        @click="confirmEnable"
                        class="px-4 py-2 rounded-lg bg-orange-500 hover:bg-orange-600 text-white transition-colors font-medium"
                    >
                        {{ t('common.continue') }}
                    </button>
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
import http from '../../utils/http';
import { h } from 'vue';
import CustomSelect from '../../components/CustomSelect.vue';
import TagConfigEditor from '../components/TagConfigEditor.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';
import ModelDownloadModal from '../components/ModelDownloadModal.vue';

// Inline components for inputs
const TextInput = {
    props: ['modelValue', 'cfg', 't'],
    emits: ['update:modelValue'],
    template: `
        <input
            :type="cfg.type === 'password' ? 'password' : 'text'"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
            class="w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:text-white"
            :placeholder="t(cfg.hint)"
        >
    `
};

const NumberInput = {
    props: ['modelValue', 'cfg', 't'],
    emits: ['update:modelValue'],
    template: `
        <input
            type="number"
            :step="cfg.type === 'float' || cfg.type === 'double' ? 'any' : '1'"
            :value="modelValue"
            @input="$emit('update:modelValue', Number($event.target.value))"
            class="w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:text-white"
            :placeholder="t(cfg.hint)"
        >
    `
};

const SwitchInput = {
    props: ['modelValue', 'cfg', 't'],
    emits: ['update:modelValue'],
    template: `
        <div class="flex items-center">
            <button
                @click="$emit('update:modelValue', !modelValue)"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:ring-offset-gray-800"
                :class="modelValue ? 'bg-blue-600' : 'bg-gray-200 dark:bg-gray-700'"
            >
                <span
                    class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform"
                    :class="modelValue ? 'translate-x-6' : 'translate-x-1'"
                />
            </button>
            <span class="ml-3 text-sm text-gray-500 dark:text-gray-400">{{ modelValue ? t('common.active') : t('common.inactive') }}</span>
        </div>
    `
};

const SelectInput = {
    props: ['modelValue', 'cfg', 't'],
    emits: ['update:modelValue'],
    template: `
        <select
            :value="modelValue"
            @change="$emit('update:modelValue', $event.target.value)"
            class="w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:text-white"
        >
            <option v-for="opt in (Array.isArray(cfg.options) ? cfg.options : Object.keys(cfg.options))" :key="opt" :value="opt">
                {{ Array.isArray(cfg.options) ? opt : cfg.options[opt] }}
            </option>
        </select>
    `
};

const ArrayInput = {
    props: ['modelValue', 'cfg', 't'],
    emits: ['update:modelValue'],
    data() {
        return {
            items: Array.isArray(this.modelValue) ? [...this.modelValue] : ['']
        }
    },
    watch: {
        modelValue(newVal) {
            if (Array.isArray(newVal)) {
                // Only update if length changed or content significantly different to avoid cursor jump issues
                // For simplicity, we just update.
                this.items = [...newVal];
                if (this.items.length === 0) this.items.push('');
            }
        }
    },
    methods: {
        updateItem(index, value) {
            this.items[index] = value;
            this.emitUpdate();
        },
        addItem(index) {
            this.items.splice(index + 1, 0, '');
            this.emitUpdate();
        },
        removeItem(index) {
            this.items.splice(index, 1);
            if (this.items.length === 0) {
                this.items.push('');
            }
            this.emitUpdate();
        },
        emitUpdate() {
            this.$emit('update:modelValue', [...this.items]);
        }
    },
    template: `
        <div class="space-y-2">
            <div v-for="(item, index) in items" :key="index" class="flex gap-2">
                <input
                    type="text"
                    :value="item"
                    @input="updateItem(index, $event.target.value)"
                    class="flex-1 px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:text-white"
                    :placeholder="t(cfg.hint)"
                >
                <button @click="addItem(index)" class="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors" title="Add">
                    <span class="material-icons">add</span>
                </button>
                <button @click="removeItem(index)" class="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors" title="Remove">
                    <span class="material-icons">remove</span>
                </button>
            </div>
        </div>
    `
};

export default {
    components: { CustomSelect, TagConfigEditor, ConfirmModal, ModelDownloadModal },
    props: ['t', 'currentLocale'],
    inject: ['showToast'],
    data() {
        return {
            modules: [],
            selectedModule: null,
            moduleDetail: null,
            reloading: false,
            showWarningModal: false,
            pendingEnableModule: null,

            activeConfigTab: 'general',
            enabling: false,
            testingConfig: false,
            saving: false,
            configForm: {},

            showLoadDefaultModal: false,
            pendingLoadDefaultKey: null,
            showDownloadModal: false,
            refreshingModels: false,
            downloadingModel: false,
            currentInstalledModels: []
        }
    },
    async created() {
        await this.fetchModules(); // Changed from reloadModules(true)
    },
    methods: {
        getSourceBadgeClass(source) {
            if (source === 'default') {
                return 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-900/20 dark:text-purple-300 dark:border-purple-800';
            } else if (source === 'external') {
                return 'bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-900/20 dark:text-orange-300 dark:border-orange-800';
            }
            return 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700';
        },
        async fetchModules() {
            try {
                const res = await http.get('/api/dashboard/scraper/modules/');
                this.modules = res.data;
            } catch (err) {
                this.showToast('Failed to load modules', 'error');
            }
        },
        async reloadModules(silent = false) {
            this.reloading = true;
            try {
                await http.post('/api/dashboard/scraper/modules/reload');
                await this.fetchModules();
                if (this.selectedModule) {
                    const exists = this.modules.find(m => m.module_id === this.selectedModule.module_id);
                    if (exists) {
                        await this.selectModule(exists);
                    } else {
                        this.selectedModule = null;
                        this.moduleDetail = null;
                    }
                }
                if (!silent) {
                    this.showToast(this.t('common.success'), 'success');
                }
            } catch (err) {
                this.showToast(this.t('common.failed_op') + err.message, 'error');
            } finally {
                this.reloading = false;
            }
        },
        async selectModule(mod) {
            this.selectedModule = mod;
            this.moduleDetail = null;
            this.activeConfigTab = 'general';
            this.configForm = {};

            try {
                const res = await http.get(`/api/dashboard/scraper/modules/${mod.module_id}`);
                this.moduleDetail = res.data;

                // Initialize form data
                if (this.moduleDetail.config) {
                    Object.keys(this.moduleDetail.config).forEach(key => {
                        const cfg = this.moduleDetail.config[key];
                        // Handle array type: ensure it's an array
                        if (cfg.type === 'array') {
                            this.configForm[key] = Array.isArray(cfg.value) ? cfg.value : [];
                            if (this.configForm[key].length === 0) this.configForm[key].push('');
                        } else {
                            this.configForm[key] = cfg.value;
                        }
                    });
                }

                const idx = this.modules.findIndex(m => m.module_id === mod.module_id);
                if (idx !== -1) {
                    this.modules[idx] = { ...this.modules[idx], ...res.data };
                }
            } catch (err) {
                this.showToast('Failed to load module details', 'error');
            }
        },
        getConfigsBySource(source) {
            if (!this.moduleDetail || !this.moduleDetail.config) return [];
            return Object.keys(this.moduleDetail.config)
                .map(key => ({ key, ...this.moduleDetail.config[key] }))
                .filter(cfg => cfg.source === source)
                .sort((a, b) => {
                    // Force override_system_config to top
                    if (a.key === 'override_system_config') return -1;
                    if (b.key === 'override_system_config') return 1;
                    return 0;
                });
        },
        getInputComponent(type) {
            switch (type) {
                case 'number':
                case 'int':
                case 'float':
                case 'double':
                    return NumberInput;
                case 'switch':
                case 'boolean':
                    return SwitchInput;
                case 'select': return SelectInput;
                case 'array': return ArrayInput;
                default: return TextInput;
            }
        },
        isConfigDisabled(key) {
            // If override_system_config is false, disable other AI configs
            if (key === 'override_system_config') return false;

            // Check if we are in default/AI config section
            const aiKeys = ['ai_model', 'ai_confidence', 'ai_tags'];
            if (aiKeys.includes(key)) {
                return !this.configForm['override_system_config'];
            }
            return false;
        },
        async toggleEnable() {
            if (!this.selectedModule || this.enabling) return;

            if (this.selectedModule.is_enable) {
                await this.executeToggle(this.selectedModule);
                return;
            }

            if (this.selectedModule.source === 'external') {
                this.pendingEnableModule = this.selectedModule;
                this.showWarningModal = true;
            } else {
                await this.executeToggle(this.selectedModule);
            }
        },
        async confirmEnable() {
            this.showWarningModal = false;
            if (this.pendingEnableModule) {
                await this.executeToggle(this.pendingEnableModule);
                this.pendingEnableModule = null;
            }
        },
        async executeToggle(mod) {
            const isEnabling = !mod.is_enable;
            this.enabling = true;

            const action = isEnabling ? 'enable' : 'disable';
            try {
                await http.post(`/api/dashboard/scraper/modules/${mod.module_id}/${action}`);

                const successKey = isEnabling ? 'scraper.toast.enable_success' : 'scraper.toast.disable_success';
                this.showToast(this.t(successKey), 'success');

                mod.is_enable = isEnabling;

                const idx = this.modules.findIndex(m => m.module_id === mod.module_id);
                if (idx !== -1) {
                    this.modules[idx].is_enable = mod.is_enable;
                }

                await this.selectModule(mod);
            } catch (err) {
                const errorMsg = err.response && err.response.data && err.response.data.detail
                    ? err.response.data.detail
                    : err.message;

                const errorKey = isEnabling ? 'scraper.toast.enable_failed' : 'scraper.toast.disable_failed';
                this.showToast(this.t(errorKey) + ': ' + errorMsg, 'error');
            } finally {
                this.enabling = false;
            }
        },
        async testConfig() {
            this.testingConfig = true;

            // Prepare payload
            const payload = {};
            const currentConfigs = this.getConfigsBySource(this.activeConfigTab === 'general' ? 'default' : 'custom');
            currentConfigs.forEach(cfg => {
                let val = this.configForm[cfg.key];
                if (cfg.type === 'array' && Array.isArray(val)) {
                    val = val.filter(item => item !== '');
                }
                payload[cfg.key] = val;
            });

            try {
                const res = await http.post(`/api/dashboard/scraper/modules/${this.selectedModule.module_id}/test_config`, payload);
                if (res.data.success) {
                    this.showToast(this.t('scraper.test_passed'), 'success');
                } else {
                    this.showToast(this.t('scraper.test_failed') + ': ' + res.data.message, 'error');
                }
            } catch (err) {
                const errorMsg = err.response && err.response.data && err.response.data.detail
                    ? err.response.data.detail
                    : err.message;
                this.showToast(this.t('common.error') + ': ' + errorMsg, 'error');
            } finally {
                this.testingConfig = false;
            }
        },
        async saveConfig() {
            if (!this.selectedModule) return;

            // Validation
            const currentConfigs = this.getConfigsBySource(this.activeConfigTab === 'general' ? 'default' : 'custom');
            for (const cfg of currentConfigs) {
                const val = this.configForm[cfg.key];

                // Type Validation
                if (cfg.type === 'number' || cfg.type === 'int' || cfg.type === 'float' || cfg.type === 'double') {
                    if (val !== '' && val !== null && isNaN(Number(val))) {
                        this.showToast(`${this.t(cfg.description)} must be a number`, 'error');
                        return;
                    }
                }

                // Regex Validation
                if (cfg.regular) {
                    try {
                        const regex = new RegExp(cfg.regular);
                        if (!regex.test(String(val))) {
                            this.showToast(`${this.t(cfg.description)} format is invalid`, 'error');
                            return;
                        }
                    } catch (e) {
                        console.warn("Invalid regex in config:", cfg.regular);
                    }
                }
            }

            this.saving = true;

            try {
                // Prepare payload
                const payload = {};
                currentConfigs.forEach(cfg => {
                    let val = this.configForm[cfg.key];

                    if (cfg.type === 'array' && Array.isArray(val)) {
                        val = val.filter(item => item !== '');
                    }

                    payload[cfg.key] = val;
                });

                await http.post(`/api/dashboard/scraper/modules/${this.selectedModule.module_id}/config`, payload);
                this.showToast(this.t('common.saved_success'), 'success');

                // Don't reload the whole module to avoid flickering
                // Just update local data if needed, but since we just saved what we have, it's fine.

            } catch (err) {
                const errorMsg = err.response && err.response.data && err.response.data.detail
                    ? err.response.data.detail
                    : err.message;
                this.showToast(this.t('common.saved_failed') + errorMsg, 'error');
            } finally {
                this.saving = false;
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
                // Just re-fetch module details, the backend will re-scan models
                await this.selectModule(this.selectedModule);
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
                await this.selectModule(this.selectedModule); // Refresh list
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
