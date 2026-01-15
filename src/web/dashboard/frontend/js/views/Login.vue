<template>
    <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
        <div class="bg-white dark:bg-gray-800 p-8 rounded-xl shadow-xl w-96 relative border border-gray-100 dark:border-gray-700 transition-colors duration-200">
            <!-- Top Right Controls -->
            <div class="absolute top-4 right-4 flex items-center space-x-2">

                <!-- Language Switcher (Icon Style) -->
                <CustomSelect
                    :model-value="currentLocale"
                    @update:model-value="$emit('update:locale', $event)"
                    :options="uiLocaleOptions"
                    variant="icon"
                />
            </div>

            <div class="text-center mb-8 mt-4">
                <h2 class="text-3xl font-extrabold text-gray-900 dark:text-white tracking-tight">{{ serverName }}</h2>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">{{ t('common.dashboard') }}</p>
            </div>

            <div v-if="error" class="bg-red-50 dark:bg-red-900/20 border-l-4 border-red-500 text-red-700 dark:text-red-400 p-4 rounded mb-6 text-sm flex items-start">
                <span class="material-icons text-sm mr-2 mt-0.5">error</span>
                <span>{{ error }}</span>
            </div>

            <form @submit.prevent="initiateLogin">
                <div class="mb-5">
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-2">{{ t('login.username') }}</label>
                    <div class="relative">
                        <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <span class="material-icons text-gray-400 text-sm">person</span>
                        </span>
                        <input v-model="username" class="pl-10 block w-full shadow-sm border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm py-2 border bg-white dark:bg-gray-700 dark:text-white" type="text">
                    </div>
                </div>
                <div class="mb-6">
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-2">{{ t('login.password') }}</label>
                    <div class="relative">
                        <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <span class="material-icons text-gray-400 text-sm">lock</span>
                        </span>
                        <input v-model="password" class="pl-10 block w-full shadow-sm border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm py-2 border bg-white dark:bg-gray-700 dark:text-white" type="password">
                    </div>
                </div>

                <div class="flex items-center justify-between mb-6">
                    <div class="flex items-center">
                        <input id="remember_me" v-model="rememberMe" type="checkbox" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700">
                        <label for="remember_me" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                            {{ t('login.remember_me') }}
                        </label>
                    </div>
                </div>

                <button :disabled="loading" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 transition duration-150 ease-in-out" type="submit">
                    <span v-if="loading" class="material-icons animate-spin text-sm mr-2">sync</span>
                    {{ loading ? t('login.signing_in') : t('login.sign_in') }}
                </button>
            </form>
        </div>

        <CaptchaModal
            ref="captchaModal"
            :show="showCaptchaModal"
            :t="t"
            @close="showCaptchaModal = false"
            @verify="handleCaptchaVerify"
        />
    </div>
</template>

<script>
import CustomSelect from '../components/CustomSelect.vue';
import CaptchaModal from '../components/CaptchaModal.vue';
import http from '../utils/http.js';
import MD5 from 'crypto-js/md5';

export default {
    components: { CustomSelect, CaptchaModal },
    props: {
        serverName: String,
        isDarkMode: Boolean,
        currentLocale: String,
        uiLocaleOptions: Array,
        t: Function
    },
    emits: ['toggle-dark-mode', 'update:locale', 'login-success'],
    data() {
        return {
            username: '',
            password: '',
            rememberMe: true,
            loading: false,
            error: null,
            showCaptchaModal: false
        }
    },
    methods: {
        initiateLogin() {
            if (!this.username || !this.password) {
                this.error = this.t('login.input_required', "Please enter username and password");
                return;
            }
            this.error = null;
            this.showCaptchaModal = true;
        },
        async handleCaptchaVerify({ id, code }) {
            this.loading = true;
            this.error = null;

            try {
                // Hash password with MD5 before sending
                const hashedPassword = MD5(this.password).toString();

                const payload = {
                    username: this.username,
                    password: hashedPassword,
                    remember_me: this.rememberMe,
                    captcha_id: id,
                    captcha_code: code
                };

                const response = await http.post('/api/dashboard/auth/login', payload);
                this.$emit('login-success', response.data);
                this.showCaptchaModal = false;
            } catch (err) {
                const status = err.response?.status;
                let detail = '';
                if (err.response && err.response.data && err.response.data.detail) {
                    detail = err.response.data.detail;
                }

                console.log("Login error:", status, detail);

                if (status === 400 || status === 403) {
                    // Pass error to modal
                    this.$refs.captchaModal.setError(detail);
                    this.$refs.captchaModal.refresh(true);
                } else {
                    this.error = detail || this.t('login.failed', 'Login failed');
                    this.showCaptchaModal = false;
                }
            } finally {
                this.loading = false;
            }
        }
    }
}
</script>
