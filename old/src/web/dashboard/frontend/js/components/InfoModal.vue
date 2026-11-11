<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-96 max-w-full transform transition-all scale-100">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ title }}</h3>

                <div class="bg-gray-100 dark:bg-gray-700 p-4 rounded-md mb-6 break-all font-mono text-center text-lg text-gray-800 dark:text-gray-200 select-all">
                    {{ message }}
                </div>

                <div class="flex justify-end gap-2">
                    <button type="button" @click="copy" class="px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm rounded hover:bg-gray-300 dark:hover:bg-gray-500 transition">{{ t('common.copy') }}</button>
                    <button type="button" @click="close" class="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition">{{ t('common.confirm') }}</button>
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
    emits: ['close'],
    inject: ['showToast'],
    methods: {
        close() {
            this.$emit('close');
        },
        async copy() {
            try {
                await navigator.clipboard.writeText(this.message);
                this.showToast('Copied to clipboard', 'success');
            } catch (err) {
                this.showToast('Failed to copy', 'error');
            }
        }
    }
}
</script>
