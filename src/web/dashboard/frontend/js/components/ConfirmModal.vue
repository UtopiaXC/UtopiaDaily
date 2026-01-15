<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-96 max-w-full transform transition-all scale-100">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ title || t('common.confirm') }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-300 mb-6">{{ message }}</p>

                <div class="flex justify-end gap-2">
                    <button type="button" @click="cancel" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">{{ t('common.cancel') }}</button>
                    <button type="button" @click="confirm" class="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700 transition">{{ t('common.confirm') }}</button>
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
    emits: ['confirm', 'cancel'],
    methods: {
        confirm() {
            this.$emit('confirm');
        },
        cancel() {
            this.$emit('cancel');
        }
    }
}
</script>
