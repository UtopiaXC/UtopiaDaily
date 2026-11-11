<template>
    <div class="relative inline-block text-left w-full" ref="dropdown">
        <button
            v-if="variant === 'default'"
            @click="!disabled && toggle()"
            type="button"
            :disabled="disabled"
            class="flex items-center justify-between w-full bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition text-slate-900 dark:text-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
        >
            <span class="block truncate mr-2">{{ selectedLabel }}</span>
            <span class="material-icons text-slate-400 text-[18px]">expand_more</span>
        </button>

        <button
            v-else-if="variant === 'icon'"
            @click="!disabled && toggle()"
            type="button"
            :disabled="disabled"
            class="p-1.5 rounded-md border border-slate-200 dark:border-slate-800 text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 focus:outline-none transition disabled:opacity-50 disabled:cursor-not-allowed"
        >
            <span class="material-icons text-[18px] leading-none block">language</span>
        </button>

        <div
            v-if="isOpen"
            :class="[
                'absolute z-[100] mt-1 bg-white dark:bg-slate-900 rounded-md py-1 text-sm overflow-auto focus:outline-none max-h-60 border border-slate-200 dark:border-slate-700',
                'min-w-full w-max max-w-xs',
                variant === 'icon' ? 'right-0 origin-top-right' : 'left-0 origin-top-left'
            ]"
        >
            <ul tabindex="-1" role="listbox">
                <li
                    v-for="option in normalizedOptions"
                    :key="option.value"
                    @click="select(option)"
                    class="text-slate-800 dark:text-slate-200 cursor-pointer select-none relative py-1.5 pl-3 pr-9 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                    role="option"
                >
                    <span :class="['block truncate', isSelected(option) ? 'font-semibold text-blue-600 dark:text-blue-400' : 'font-normal']">
                        {{ option.label }}
                    </span>
                    <span v-if="isSelected(option)" class="text-blue-600 dark:text-blue-400 absolute inset-y-0 right-0 flex items-center pr-3">
                        <span class="material-icons text-[16px]">check</span>
                    </span>
                </li>
            </ul>
        </div>
    </div>
</template>

<script>
export default {
    props: {
        modelValue: [String, Number, Boolean],
        options: {
            type: Array,
            default: () => []
        },
        variant: {
            type: String,
            default: 'default'
        },
        disabled: {
            type: Boolean,
            default: false
        }
    },
    emits: ['update:modelValue'],
    data() {
        return {
            isOpen: false
        };
    },
    computed: {
        normalizedOptions() {
            return this.options.map(opt => {
                if (typeof opt === 'object' && opt !== null && 'value' in opt) {
                    return opt;
                }
                return { label: String(opt), value: opt };
            });
        },
        selectedLabel() {
            const option = this.normalizedOptions.find(o => o.value === this.modelValue);
            return option ? option.label : this.modelValue;
        }
    },
    methods: {
        toggle() {
            if (this.disabled) return;
            this.isOpen = !this.isOpen;
        },
        select(option) {
            this.$emit('update:modelValue', option.value);
            this.isOpen = false;
        },
        isSelected(option) {
            return option.value === this.modelValue;
        },
        handleClickOutside(event) {
            if (this.$refs.dropdown && !this.$refs.dropdown.contains(event.target)) {
                this.isOpen = false;
            }
        }
    },
    mounted() {
        document.addEventListener('click', this.handleClickOutside);
    },
    unmounted() {
        document.removeEventListener('click', this.handleClickOutside);
    }
};
</script>
