<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
            <div class="bg-white dark:bg-slate-900 rounded-lg p-5 w-[440px] max-w-full border border-slate-200 dark:border-slate-800">
                <h3 class="text-base font-semibold mb-3 text-slate-900 dark:text-white">{{ title }}</h3>
                <div class="text-xs font-mono text-slate-700 dark:text-slate-300 mb-5 break-all whitespace-pre-wrap bg-slate-50 dark:bg-slate-950 p-3 rounded-md border border-slate-200 dark:border-slate-800 max-h-72 overflow-y-auto">{{ message }}</div>
                <div class="flex justify-end">
                    <button @click="$emit('close')" class="ui-btn-primary">
                        {{ t('common.close') }}
                    </button>
                </div>
            </div>
        </div>
    </transition>
</template>

<script>
export default {
    props: {
        show: Boolean,
        title: String,
        message: String,
        t: Function
    },
    emits: ['close']
};
</script>
