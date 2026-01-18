<template>
    <div class="space-y-4">
        <!-- Empty State -->
        <div v-if="tags.length === 0" class="text-center py-8 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800/30">
            <p class="text-gray-500 dark:text-gray-400 mb-4">{{ t('common.no_data') }}</p>
            <button @click="addTag" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition flex items-center mx-auto shadow-sm">
                <span class="material-icons mr-2">add</span> {{ t('config.ai.tags.add_tag') }}
            </button>
        </div>

        <!-- Tags List -->
        <div v-else class="space-y-6">
            <div v-for="(tag, index) in tags" :key="index" class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-gray-50 dark:bg-gray-800/50">
                <!-- Tag Header Row -->
                <div class="flex items-end gap-3 mb-3">
                    <div class="flex-1">
                        <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.tag_name') }}</label>
                        <input v-model="tag.tag_name" type="text" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                    </div>
                    <div class="w-24">
                        <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.priority') }}</label>
                        <input v-model.number="tag.tag_priority" type="number" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                    </div>

                    <!-- Tag Actions -->
                    <div class="flex gap-1 pb-0.5">
                        <button @click="addTag" class="text-blue-600 hover:text-blue-800 p-2 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20 transition" title="Add Tag">
                            <span class="material-icons text-lg">add</span>
                        </button>
                        <button @click="removeTag(index)" class="text-red-500 hover:text-red-700 p-2 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition" title="Remove Tag">
                            <span class="material-icons text-lg">remove</span>
                        </button>
                    </div>
                </div>

                <!-- Tag Keys Section -->
                <div>
                    <!-- Expand/Collapse Header -->
                    <div class="flex items-center cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 rounded py-1 px-2 -ml-2 select-none" @click="toggleExpand(index)">
                        <span class="material-icons text-gray-400 text-sm transition-transform duration-200 mr-2" :class="{'rotate-180': isExpanded(index)}">expand_more</span>
                        <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold cursor-pointer">{{ t('config.ai.tags.keywords') }}</label>
                    </div>

                    <div class="pl-6 border-l-2 border-gray-200 dark:border-gray-600 ml-2 mt-2 space-y-2" v-show="isExpanded(index)">
                        <div v-for="(key, kIndex) in tag.tag_keys" :key="kIndex" class="flex items-center gap-2">
                            <input v-model="tag.tag_keys[kIndex]" class="flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                            <button @click="removeKey(index, kIndex)" class="text-gray-400 hover:text-red-500 p-1 rounded transition">
                                <span class="material-icons text-lg">remove</span>
                            </button>
                        </div>

                        <button @click="addKey(index)" class="flex items-center text-blue-600 hover:text-blue-800 text-sm font-medium mt-2 px-1 py-1 rounded hover:bg-blue-50 dark:hover:bg-blue-900/10 transition">
                            <span class="material-icons text-sm mr-1">add</span> {{ t('config.ai.tags.add_key') }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    props: {
        modelValue: {
            type: String,
            default: "[]"
        },
        t: {
            type: Function,
            default: (k) => k
        },
        currentLocale: {
            type: String,
            default: "en_US"
        }
    },
    data() {
        return {
            tags: [],
            expandedIndices: [] // Track expanded state
        }
    },
    watch: {
        modelValue: {
            immediate: true,
            handler(val) {
                try {
                    const parsed = JSON.parse(val || "[]");
                    if (JSON.stringify(parsed) !== JSON.stringify(this.tags)) {
                        this.tags = parsed;
                        // Default expand all on load
                        this.expandedIndices = this.tags.map((_, i) => i);
                    }
                } catch (e) {
                    console.error("Failed to parse tags JSON", e);
                    this.tags = [];
                }
            }
        },
        tags: {
            deep: true,
            handler(newTags) {
                const str = JSON.stringify(newTags);
                if (str !== this.modelValue) {
                    this.$emit('update:modelValue', str);
                }
            }
        }
    },
    methods: {
        addTag() {
            this.tags.push({
                tag_name: "New Tag",
                tag_keys: [],
                tag_priority: 1
            });
            this.expandedIndices.push(this.tags.length - 1);
        },
        removeTag(index) {
            this.tags.splice(index, 1);
            // Adjust expanded indices
            this.expandedIndices = this.expandedIndices.filter(i => i !== index).map(i => i > index ? i - 1 : i);
        },
        addKey(tagIndex) {
            this.tags[tagIndex].tag_keys.push("keyword");
        },
        removeKey(tagIndex, keyIndex) {
            this.tags[tagIndex].tag_keys.splice(keyIndex, 1);
        },
        loadDefault() {
            try {
                const templateStr = this.t('config.ai.tags.template');
                if (templateStr && templateStr !== 'config.ai.tags.template') {
                    this.tags = JSON.parse(templateStr);
                    this.expandedIndices = this.tags.map((_, i) => i);
                } else {
                    this.tags = [];
                }
            } catch (e) {
                console.error("Failed to load template", e);
            }
        },
        toggleExpand(index) {
            const idx = this.expandedIndices.indexOf(index);
            if (idx === -1) {
                this.expandedIndices.push(index);
            } else {
                this.expandedIndices.splice(idx, 1);
            }
        },
        isExpanded(index) {
            return this.expandedIndices.includes(index);
        }
    }
}
</script>
