<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-80 transform transition-all scale-100">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ t('captcha.title') }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">{{ t('captcha.desc') }}</p>
                <div v-if="error" class="bg-red-50 dark:bg-red-900/20 border-l-4 border-red-500 text-red-700 dark:text-red-400 p-2 rounded mb-4 text-xs flex items-start">
                    <span class="material-icons text-xs mr-1 mt-0.5">error</span>
                    <span>{{ error }}</span>
                </div>

                <div class="flex gap-2 mb-4">
                    <div class="border dark:border-gray-600 rounded overflow-hidden flex-1 bg-gray-50 dark:bg-gray-700 h-16 flex items-center justify-center relative">
                        <img v-if="svg" :src="svg" alt="Captcha" class="h-full w-full object-cover">
                        <div v-if="loading" class="absolute inset-0 bg-white dark:bg-gray-800 bg-opacity-80 flex items-center justify-center">
                            <span class="material-icons animate-spin text-gray-500 dark:text-gray-400">sync</span>
                        </div>
                    </div>
                    <button type="button" @click="refresh" class="p-2 bg-gray-100 dark:bg-gray-700 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 transition h-16 w-10 flex items-center justify-center">
                        <span :class="['material-icons text-sm', loading ? 'animate-spin' : '']">refresh</span>
                    </button>
                </div>

                <input v-model="code"
                       @keyup.enter.prevent="submit"
                       ref="input"
                       class="block w-full shadow-sm border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm py-2 border px-3 mb-4 bg-white dark:bg-gray-700 dark:text-white"
                       type="text"
                       :placeholder="t('captcha.hint')">

                <div class="flex justify-end gap-2">
                    <button type="button" @click="close" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">{{ t('common.cancel') }}</button>
                    <button type="button" @click="submit" class="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition">{{ t('common.verify') }}</button>
                </div>
            </div>
        </div>
    </transition>
</template>

<script>
import http from '../utils/http.js';

export default {
    props: {
        show: Boolean,
        t: Function
    },
    emits: ['close', 'verify'],
    data() {
        return {
            loading: false,
            svg: '',
            id: '',
            code: '',
            error: null
        }
    },
    watch: {
        show(val) {
            if (val) {
                this.refresh();
                this.$nextTick(() => {
                    if (this.$refs.input) this.$refs.input.focus();
                });
            } else {
                this.code = '';
                this.error = null;
            }
        }
    },
    methods: {
        async refresh(preserveError = false) {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/auth/captcha');
                this.id = res.data.captcha_id;
                this.svg = res.data.svg;
                this.code = '';
                if (!preserveError) {
                    this.error = null;
                }
                this.$nextTick(() => {
                    if (this.$refs.input) this.$refs.input.focus();
                });
            } catch (err) {
                console.error("Failed to load captcha", err);
                this.error = "Failed to load CAPTCHA.";
            } finally {
                this.loading = false;
            }
        },
        close() {
            this.$emit('close');
        },
        submit() {
            if (!this.code) return;
            this.$emit('verify', { id: this.id, code: this.code });
        },
        setError(msg) {
            this.error = msg;
        }
    }
}
</script>
