<template>
    <div class="space-y-4">
        <!-- Group Tabs -->
        <div class="flex border-b border-gray-200 dark:border-gray-700 overflow-x-auto items-center">
            <div
                v-for="(group, index) in groups"
                :key="index"
                @click="activeGroupIndex = index"
                class="group relative px-4 py-2 text-sm font-medium border-b-2 transition-colors whitespace-nowrap flex items-center gap-2 cursor-pointer select-none"
                :class="activeGroupIndex === index ? 'border-blue-500 text-blue-600 dark:text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
            >
                {{ group.name || t('config.ai.tags.group_name') }}
                <span class="bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs px-1.5 rounded-full">{{ getGroupTagCount(group.name) }}</span>

                <!-- Delete Group Button on Tab (Always visible) -->
                <button @click.stop="promptDeleteGroup(index)" class="ml-1 text-gray-400 hover:text-red-500 p-0.5 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors" :title="t('config.ai.tags.delete_group')">
                    <span class="material-icons text-[14px]">delete</span>
                </button>
            </div>
            <button @click="addGroup" class="px-3 py-2 text-gray-400 hover:text-blue-500 transition-colors" :title="t('config.ai.tags.add_group')">
                <span class="material-icons text-sm">add</span>
            </button>
        </div>

        <!-- Group Settings & Content -->
        <div v-if="activeGroup" class="space-y-6 pt-2">
            <!-- Group Config -->
            <div class="bg-gray-50 dark:bg-gray-800/50 p-4 rounded-lg border border-gray-200 dark:border-gray-700 flex flex-wrap gap-4 items-end">
                <div class="flex-1 min-w-[200px]">
                    <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.group_name') }}</label>
                    <input ref="groupNameInput" v-model="activeGroup.name" @change="updateGroupName(activeGroupIndex, $event.target.value)" type="text" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                </div>
                <div class="w-32">
                    <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.max_results') }}</label>
                    <input v-model.number="activeGroup.max_results" type="number" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                </div>
            </div>

            <!-- Tags List -->
            <div class="space-y-4">
                <div v-if="activeGroupTags.length === 0" class="text-center py-8 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg">
                    <p class="text-gray-500 dark:text-gray-400 mb-4">{{ t('common.no_data') }}</p>
                    <button @click="addTag" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition flex items-center mx-auto shadow-sm">
                        <span class="material-icons mr-2">add</span> {{ t('config.ai.tags.add_tag') }}
                    </button>
                </div>

                <div v-else class="space-y-4">
                    <div v-for="(tag, index) in activeGroupTags" :key="tag._id" class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-800">
                        <!-- Tag Header Row -->
                        <div class="flex items-end gap-3 mb-3">
                            <div class="flex-1">
                                <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.tag_name') }}</label>
                                <input :ref="el => tagNameInputs[index] = el" v-model="tag.name" type="text" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                            </div>
                            <div class="w-24">
                                <label class="text-xs text-gray-500 dark:text-gray-400 uppercase font-bold mb-1 block">{{ t('config.ai.tags.priority') }}</label>
                                <input v-model.number="tag.priority" type="number" class="w-full shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                            </div>

                            <!-- Tag Actions -->
                            <div class="flex gap-1 pb-0.5">
                                <button @click="addTag" class="text-blue-600 hover:text-blue-800 p-2 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20 transition" title="Add Tag">
                                    <span class="material-icons text-lg">add</span>
                                </button>
                                <button @click="removeTag(tag)" class="text-red-500 hover:text-red-700 p-2 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition" title="Remove Tag">
                                    <span class="material-icons text-lg">remove</span>
                                </button>
                            </div>
                        </div>

                        <!-- Tag Keys Section -->
                        <div class="flex items-start">
                            <!-- Expand/Collapse Indicator + Label -->
                            <div class="flex flex-col items-center w-24 pt-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded p-1 mr-2" @click="toggleExpand(tag._id)">
                                <div class="flex items-center gap-1 text-gray-500 dark:text-gray-400 select-none">
                                    <span class="material-icons text-sm transition-transform duration-200" :class="{'rotate-180': isExpanded(tag._id)}">expand_more</span>
                                    <span class="text-xs uppercase font-bold">{{ t('config.ai.tags.keywords') }}</span>
                                </div>
                            </div>

                            <div class="flex-1 pl-4 border-l-2 border-gray-200 dark:border-gray-600 space-y-2" v-show="isExpanded(tag._id)">
                                <div v-for="(key, kIndex) in tag.keywords" :key="kIndex" class="flex items-center gap-2">
                                    <input v-model="tag.keywords[kIndex]" class="flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm bg-white dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500">
                                    <button @click="removeKey(tag, kIndex)" class="text-gray-400 hover:text-red-500 p-1 rounded transition">
                                        <span class="material-icons text-lg">remove</span>
                                    </button>
                                </div>

                                <button @click="addKey(tag)" class="flex items-center text-blue-600 hover:text-blue-800 text-sm font-medium mt-2 px-1 py-1 rounded hover:bg-blue-50 dark:hover:bg-blue-900/10 transition">
                                    <span class="material-icons text-sm mr-1">add</span> {{ t('config.ai.tags.add_key') }}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <button v-if="activeGroupTags.length > 0" @click="addTag" class="w-full py-2 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg text-gray-500 hover:text-blue-500 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/10 transition flex justify-center items-center font-medium">
                    <span class="material-icons mr-2">add_circle_outline</span> {{ t('config.ai.tags.add_tag') }}
                </button>
            </div>
        </div>

        <div v-else class="text-center py-12 text-gray-500 dark:text-gray-400">
            <p class="mb-4">{{ t('config.ai.tags.no_groups') }}</p>
            <button @click="addGroup" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition flex items-center gap-2 mx-auto">
                <span class="material-icons text-sm">add</span> {{ t('config.ai.tags.add_group') }}
            </button>
        </div>

        <ConfirmModal
            :show="showDeleteModal"
            :title="t('config.ai.tags.delete_group')"
            :message="deleteMessage"
            :t="t"
            @confirm="confirmDeleteGroup"
            @cancel="showDeleteModal = false"
        />
    </div>
</template>

<script>
import ConfirmModal from '../../components/ConfirmModal.vue';

export default {
    components: { ConfirmModal },
    props: {
        modelValue: {
            type: String,
            default: "{}"
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
    emits: ['update:modelValue', 'save'],
    data() {
        return {
            groups: [],
            tags: [],
            activeGroupIndex: 0,
            expandedIndices: [], // Stores tag._id
            tagNameInputs: [], // Refs for tag name inputs
            showDeleteModal: false,
            pendingDeleteGroupIndex: null
        }
    },
    computed: {
        activeGroup() {
            return this.groups[this.activeGroupIndex];
        },
        activeGroupTags() {
            if (!this.activeGroup) return [];
            return this.tags.filter(t => t.group === this.activeGroup.name);
        },
        deleteMessage() {
            if (this.pendingDeleteGroupIndex === null) return "";
            const group = this.groups[this.pendingDeleteGroupIndex];
            return this.t('config.ai.tags.confirm_delete_group').replace('{name}', group.name);
        }
    },
    watch: {
        modelValue: {
            immediate: true,
            handler(val) {
                try {
                    const parsed = JSON.parse(val || "{}");
                    // Deep compare to avoid loop
                    const currentData = { groups: this.groups, tags: this.tags };
                    // Remove internal _id for comparison
                    const cleanTags = this.tags.map(({_id, ...rest}) => rest);
                    const cleanCurrent = { groups: this.groups, tags: cleanTags };

                    if (JSON.stringify(parsed) !== JSON.stringify(cleanCurrent)) {
                        if (Array.isArray(parsed)) {
                            // Old format
                            this.groups = [{ name: "Default", max_results: 5 }];
                            this.tags = parsed.map((t, i) => ({ ...t, group: "Default", _id: Date.now() + i }));
                        } else {
                            this.groups = parsed.groups || [];
                            this.tags = (parsed.tags || []).map((t, i) => ({ ...t, _id: Date.now() + i }));
                        }
                        // Expand all by default
                        this.expandedIndices = this.tags.map(t => t._id);

                        // If no groups, ensure activeGroupIndex is valid (e.g. -1 or handle in template)
                        if (this.groups.length === 0) {
                            this.activeGroupIndex = -1;
                        } else if (this.activeGroupIndex >= this.groups.length) {
                            this.activeGroupIndex = 0;
                        }
                    }
                } catch (e) {
                    console.error("Failed to parse tags JSON", e);
                    this.groups = [];
                    this.tags = [];
                }
            }
        },
        groups: {
            deep: true,
            handler() { this.emitUpdate(); }
        },
        tags: {
            deep: true,
            handler() { this.emitUpdate(); }
        },
        activeGroupIndex() {
            // Reset refs array when switching groups
            this.tagNameInputs = [];
        }
    },
    methods: {
        emitUpdate() {
            const cleanTags = this.tags.map(({_id, ...rest}) => rest);
            const data = {
                groups: this.groups,
                tags: cleanTags
            };
            this.$emit('update:modelValue', JSON.stringify(data));
        },
        getGroupTagCount(groupName) {
            return this.tags.filter(t => t.group === groupName).length;
        },
        addGroup() {
            const name = this.t('config.ai.tags.new_group_prefix') + " " + (this.groups.length + 1);
            this.groups.push({ name: name, max_results: 3 });
            this.activeGroupIndex = this.groups.length - 1;

            this.$nextTick(() => {
                if (this.$refs.groupNameInput) {
                    this.$refs.groupNameInput.focus();
                }
            });
        },
        promptDeleteGroup(index) {
            this.pendingDeleteGroupIndex = index;
            this.showDeleteModal = true;
        },
        confirmDeleteGroup() {
            const index = this.pendingDeleteGroupIndex;
            const group = this.groups[index];

            // Remove tags in this group
            this.tags = this.tags.filter(t => t.group !== group.name);
            this.groups.splice(index, 1);

            if (this.groups.length === 0) {
                this.activeGroupIndex = -1;
            } else if (this.activeGroupIndex >= this.groups.length) {
                this.activeGroupIndex = Math.max(0, this.groups.length - 1);
            }

            this.showDeleteModal = false;
            this.pendingDeleteGroupIndex = null;
        },
        updateGroupName(index, newName) {
            const oldName = this.groups[index].name;
            // Update tags
            this.tags.forEach(t => {
                if (t.group === oldName) {
                    t.group = newName;
                }
            });
            // Name is already updated via v-model, but we need to ensure consistency if user typed
        },
        addTag() {
            if (!this.activeGroup) return;
            const newTag = {
                group: this.activeGroup.name,
                name: "New Tag",
                keywords: [],
                priority: 1,
                _id: Date.now()
            };
            this.tags.push(newTag);
            this.expandedIndices.push(newTag._id);

            this.$nextTick(() => {
                // Focus the last input in the list
                const inputs = this.tagNameInputs;
                if (inputs && inputs.length > 0) {
                    const lastInput = inputs[inputs.length - 1];
                    if (lastInput) lastInput.focus();
                }
            });
        },
        removeTag(tag) {
            const idx = this.tags.findIndex(t => t._id === tag._id);
            if (idx !== -1) {
                this.tags.splice(idx, 1);
            }
        },
        addKey(tag) {
            tag.keywords.push("keyword");
        },
        removeKey(tag, index) {
            tag.keywords.splice(index, 1);
        },
        loadDefault() {
            try {
                const templateStr = this.t('config.ai.tags.template');
                if (templateStr && templateStr !== 'config.ai.tags.template') {
                    const parsed = JSON.parse(templateStr);
                    this.groups = parsed.groups || [];
                    this.tags = (parsed.tags || []).map((t, i) => ({ ...t, _id: Date.now() + i }));
                    this.activeGroupIndex = 0;
                    this.expandedIndices = this.tags.map(t => t._id);
                }
            } catch (e) {
                console.error("Failed to load template", e);
            }
        },
        toggleExpand(id) {
            const idx = this.expandedIndices.indexOf(id);
            if (idx === -1) {
                this.expandedIndices.push(id);
            } else {
                this.expandedIndices.splice(idx, 1);
            }
        },
        isExpanded(id) {
            return this.expandedIndices.includes(id);
        }
    }
}
</script>
