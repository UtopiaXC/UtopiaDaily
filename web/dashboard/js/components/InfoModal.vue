<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-96 max-w-full border border-gray-200 dark:border-gray-700">
                <h3 class="text-lg font-bold mb-3 text-gray-800 dark:text-white">{{ title }}</h3>
                <div class="text-sm text-gray-600 dark:text-gray-300 mb-6 break-all whitespace-pre-wrap bg-gray-50 dark:bg-gray-900/50 p-3 rounded-lg border border-gray-200 dark:border-gray-700 max-h-64 overflow-y-auto">{{ message }}</div>
                <div class="flex justify-end">
                    <button @click="$emit('close')" class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition">
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
}
</script>
