<template>
    <header class="h-16 bg-white dark:bg-gray-800 shadow-sm flex items-center justify-between px-4 md:px-8 z-10 border-b border-gray-200 dark:border-gray-700 transition-colors duration-200">
        <div class="flex items-center">
            <!-- Mobile Menu Button -->
            <button
                @click="$emit('toggle-sidebar')"
                class="md:hidden p-2 -ml-2 mr-2 rounded-md text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white focus:outline-none"
            >
                <span class="material-icons">menu</span>
            </button>
            <h2 class="text-xl font-bold text-gray-800 dark:text-white truncate">{{ currentTabLabel }}</h2>
        </div>

        <div class="flex items-center space-x-2 md:space-x-4">
            <button @click="$emit('toggle-dark-mode')" class="p-2 rounded-full text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 transition focus:outline-none">
                <span class="material-icons text-xl">{{ isDarkMode ? 'light_mode' : 'dark_mode' }}</span>
            </button>

            <div class="w-24 md:w-32">
                <CustomSelect
                    :model-value="currentLocale"
                    @update:model-value="$emit('update:locale', $event)"
                    :options="uiLocaleOptions"
                />
            </div>
            <div class="hidden md:block h-8 w-px bg-gray-200 dark:bg-gray-700 mx-2"></div>
            <div class="hidden md:flex items-center text-sm text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-3 py-1 rounded-full border border-green-100 dark:border-green-900/30">
                <span class="h-2 w-2 rounded-full bg-green-500 mr-2"></span>
                Online
            </div>
        </div>
    </header>
</template>

<script>
import CustomSelect from '../components/CustomSelect.vue';

export default {
    components: { CustomSelect },
    props: {
        currentTabLabel: String,
        isDarkMode: Boolean,
        currentLocale: String,
        uiLocaleOptions: Array
    },
    emits: ['toggle-dark-mode', 'update:locale', 'toggle-sidebar']
}
</script>
