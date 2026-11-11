<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
            <div class="bg-white dark:bg-slate-900 rounded-lg p-5 w-96 max-w-full border border-slate-200 dark:border-slate-800">
                <h3 class="text-base font-semibold mb-2 text-slate-900 dark:text-white">{{ title }}</h3>
                <p class="text-sm text-slate-600 dark:text-slate-300 mb-5 leading-relaxed">{{ message }}</p>
                <div class="flex justify-end gap-2">
                    <button @click="$emit('cancel')" class="ui-btn-secondary">
                        {{ t('common.cancel') }}
                    </button>
                    <button @click="$emit('confirm')" class="ui-btn-danger">
                        {{ t('common.confirm') }}
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
    emits: ['confirm', 'cancel']
};
</script>
