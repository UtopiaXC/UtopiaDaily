<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-96 max-w-full border border-gray-200 dark:border-gray-700">
                <h3 class="text-lg font-bold mb-2 text-gray-800 dark:text-white">{{ title }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-300 mb-6">{{ message }}</p>
                <div class="flex justify-end gap-3">
                    <button @click="$emit('cancel')" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition">
                        {{ t('common.cancel') }}
                    </button>
                    <button @click="$emit('confirm')" class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 transition">
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
}
</script>
