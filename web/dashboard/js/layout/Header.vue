<template>
    <header class="h-14 bg-white dark:bg-slate-950 flex items-center justify-between px-4 md:px-6 z-10 border-b border-slate-200 dark:border-slate-800 transition-colors duration-150">
        <div class="flex items-center min-w-0 gap-2">
            <button
                @click="$emit('toggle-sidebar')"
                class="md:hidden p-1.5 -ml-1.5 rounded-md text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 focus:outline-none"
            >
                <span class="material-icons text-[20px]">menu</span>
            </button>
            <h2 class="text-base font-semibold text-slate-900 dark:text-white truncate tracking-tight">{{ currentTabLabel }}</h2>
        </div>

        <div class="flex items-center gap-2">
            <a
                href="/"
                target="_blank"
                class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-900 transition-colors"
                :title="t('frontend_cfg.open_newspaper')"
            >
                <span class="material-icons text-[16px] text-slate-500 dark:text-slate-400">newspaper</span>
                <span class="hidden sm:inline">{{ t('common.newspaper') }}</span>
            </a>

            <button
                @click="$emit('toggle-dark-mode')"
                class="p-1.5 rounded-md text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white border border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-900 transition-colors focus:outline-none"
                :title="isDarkMode ? 'Light' : 'Dark'"
            >
                <span class="material-icons text-[18px] leading-none block">{{ isDarkMode ? 'light_mode' : 'dark_mode' }}</span>
            </button>

            <div class="w-28 md:w-32">
                <CustomSelect
                    :model-value="currentLocale"
                    @update:model-value="$emit('update:locale', $event)"
                    :options="uiLocaleOptions"
                />
            </div>

            <div class="hidden md:flex items-center text-xs font-medium text-emerald-700 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/50 px-2.5 py-1 rounded-md border border-emerald-200 dark:border-emerald-900">
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 mr-1.5"></span>
                {{ t('common.online') }}
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
        uiLocaleOptions: Array,
        t: Function
    },
    emits: ['toggle-dark-mode', 'update:locale', 'toggle-sidebar']
};
</script>
