<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-[900px] max-w-full transform transition-all scale-100 flex flex-col max-h-[90vh] relative">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ t('config.ai.model.download_title') }}</h3>

                <div class="flex-1 overflow-y-auto mb-4 border border-gray-200 dark:border-gray-700 rounded-lg no-scrollbar">
                    <table class="w-full text-sm text-left table-fixed">
                        <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400 sticky top-0 z-10">
                            <tr>
                                <th scope="col" class="px-4 py-3 w-10"></th>
                                <th scope="col" class="px-4 py-3 w-48">{{ t('config.ai.model.name') }}</th>
                                <th scope="col" class="px-4 py-3 w-24">{{ t('config.ai.model.size_about') }}</th>
                                <th scope="col" class="px-4 py-3">{{ t('config.ai.model.desc') }}</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                            <template v-for="group in presetModels" :key="group.group">
                                <tr class="bg-gray-50 dark:bg-gray-700/50">
                                    <td colspan="4" class="px-4 py-2 text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                                        {{ t(group.group) }}
                                    </td>
                                </tr>
                                <tr v-for="model in group.models" :key="model.id"
                                    class="bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
                                    :class="{'opacity-50 cursor-not-allowed': isInstalled(model.hf_repo), 'cursor-pointer': !isInstalled(model.hf_repo)}"
                                    @click="!isInstalled(model.hf_repo) && (selectedModel = model.hf_repo)">
                                    <td class="px-4 py-3">
                                        <input type="radio" :value="model.hf_repo" v-model="selectedModel"
                                               :disabled="isInstalled(model.hf_repo)"
                                               class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600 disabled:cursor-not-allowed">
                                    </td>
                                    <td class="px-4 py-3 font-medium text-gray-900 dark:text-white truncate"
                                        @mouseenter="showTooltip($event, model.name)"
                                        @mousemove="moveTooltip($event)"
                                        @mouseleave="hideTooltip">
                                        {{ model.name }}
                                        <span v-if="isInstalled(model.hf_repo)" class="ml-2 text-xs text-green-600 dark:text-green-400">(Installed)</span>
                                    </td>
                                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400 truncate">
                                        {{ model.size_about }}
                                    </td>
                                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400 truncate"
                                        @mouseenter="showTooltip($event, t(model.desc))"
                                        @mousemove="moveTooltip($event)"
                                        @mouseleave="hideTooltip">
                                        {{ t(model.desc) }}
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>
                </div>

                <div class="flex justify-end gap-2">
                    <button type="button" @click="cancel" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">{{ t('common.cancel') }}</button>
                    <button type="button" @click="confirm" class="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition flex items-center gap-2" :disabled="!selectedModel">
                        {{ t('config.ai.model.download') }}
                    </button>
                </div>
            </div>

            <!-- Global Tooltip for this Modal -->
            <div v-if="tooltip.visible"
                 class="fixed z-[60] p-3 bg-gray-900 text-white text-xs rounded shadow-xl w-64 pointer-events-none transform -translate-x-1/2 -translate-y-full mt-[-10px]"
                 :style="tooltip.style">
                <div class="font-medium mb-1 break-words">{{ tooltip.content }}</div>
                <div v-if="tooltip.meta" class="text-gray-400 break-words">{{ tooltip.meta }}</div>
                <!-- Arrow -->
                <div class="absolute left-1/2 bottom-0 transform -translate-x-1/2 translate-y-full w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[6px] border-t-gray-900"></div>
            </div>
        </div>
    </transition>
</template>

<script>
import http from '../../utils/http';

export default {
    props: {
        show: Boolean,
        t: Function,
        installedModels: {
            type: Array,
            default: () => []
        }
    },
    emits: ['confirm', 'cancel'],
    data() {
        return {
            selectedModel: null,
            tooltip: {
                visible: false,
                content: '',
                meta: '',
                style: { top: '0px', left: '0px' }
            },
            presetModels: []
        }
    },
    watch: {
        show(newVal) {
            if (newVal) {
                this.fetchDefaultModels();
            }
        }
    },
    methods: {
        async fetchDefaultModels() {
            try {
                const response = await http.get('/api/dashboard/system-config/default-models');
                this.presetModels = response.data;
            } catch (error) {
                console.error("Failed to fetch default models", error);
            }
        },
        isInstalled(repoId) {
            const name = repoId.split('/').pop();
            return this.installedModels.includes(name);
        },
        confirm() {
            this.$emit('confirm', this.selectedModel);
        },
        cancel() {
            this.$emit('cancel');
        },
        showTooltip(event, content, meta) {
            const el = event.currentTarget;
            if (el.scrollWidth > el.clientWidth) {
                this.updateTooltipPosition(event, content, meta);
            }
        },
        moveTooltip(event) {
            if (this.tooltip.visible) {
                this.updateTooltipPosition(event, this.tooltip.content, this.tooltip.meta);
            }
        },
        updateTooltipPosition(event, content, meta) {
            const x = event.clientX;
            const y = event.clientY;

            this.tooltip = {
                visible: true,
                content: content,
                meta: meta,
                style: {
                    top: y + 'px',
                    left: x + 'px'
                }
            };
        },
        hideTooltip() {
            this.tooltip.visible = false;
        }
    }
}
</script>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
    display: none;
}
.no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
}
</style>