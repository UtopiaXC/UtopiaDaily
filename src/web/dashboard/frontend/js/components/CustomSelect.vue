<template>
    <div class="relative inline-block text-left w-full" ref="dropdown">
        <button v-if="variant === 'default'" @click="toggle" type="button"
            class="flex items-center justify-between w-full bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition shadow-sm text-gray-900 dark:text-white">
            <span class="block truncate mr-2">{{ selectedLabel }}</span>
            <span class="material-icons text-gray-400 dark:text-gray-300 text-sm">expand_more</span>
        </button>

        <button v-else-if="variant === 'icon'" @click="toggle" type="button"
            class="p-2 rounded-full text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition">
            <span class="material-icons">language</span>
        </button>

        <div v-if="isOpen"
             :class="[
                'absolute z-50 mt-1 bg-white dark:bg-gray-800 shadow-lg rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm',
                'min-w-full w-max max-w-xs',
                variant === 'icon' ? 'right-0 origin-top-right' : 'left-0 origin-top-left'
             ]">
            <ul tabindex="-1" role="listbox">
                <li v-for="option in normalizedOptions" :key="option.value"
                    @click="select(option)"
                    class="text-gray-900 dark:text-gray-200 cursor-pointer select-none relative py-2 pl-3 pr-9 hover:bg-blue-50 dark:hover:bg-gray-700 transition"
                    role="option">
                    <span :class="['block truncate', isSelected(option) ? 'font-semibold text-blue-600 dark:text-blue-400' : 'font-normal']">
                        {{ option.label }}
                    </span>
                    <span v-if="isSelected(option)" class="text-blue-600 dark:text-blue-400 absolute inset-y-0 right-0 flex items-center pr-4">
                        <span class="material-icons text-sm">check</span>
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
        }
    },
    emits: ['update:modelValue'],
    data() {
        return {
            isOpen: false
        }
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
}
</script>
