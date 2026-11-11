<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md border border-gray-200 dark:border-gray-700 overflow-hidden">
                <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50">
                    <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('auth.profile.title') }}</h3>
                    <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
                        <span class="material-icons text-xl">close</span>
                    </button>
                </div>

                <div class="p-6 space-y-6 max-h-[80vh] overflow-y-auto" v-if="user">
                    <form @submit.prevent="saveProfile" class="space-y-4">
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                            <div>
                                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('auth.profile.username') }}</label>
                                <input :value="user.username" disabled class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-900 text-gray-500 cursor-not-allowed">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('auth.profile.role') }}</label>
                                <input :value="user.role_name" disabled class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-900 text-gray-500 cursor-not-allowed">
                            </div>
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('auth.profile.nickname') }}</label>
                            <input v-model="profileForm.nickname" type="text" class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('auth.profile.email') }}</label>
                            <input v-model="profileForm.email" type="email" class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                        </div>
                        <div class="flex justify-end">
                            <button type="submit" :disabled="savingProfile" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition disabled:opacity-50 flex items-center gap-1">
                                <span v-if="savingProfile" class="material-icons text-sm animate-spin">sync</span>
                                {{ t('auth.profile.save') }}
                            </button>
                        </div>
                    </form>

                    <div class="border-t border-gray-200 dark:border-gray-700 pt-5">
                        <h4 class="text-sm font-bold text-gray-800 dark:text-white mb-3">{{ t('auth.profile.password_section') }}</h4>
                        <form @submit.prevent="changePassword" class="space-y-3">
                            <div>
                                <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">{{ t('auth.profile.old_password') }}</label>
                                <input v-model="pwdForm.old_password" type="password" class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 dark:text-white">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">{{ t('auth.profile.new_password') }}</label>
                                <input v-model="pwdForm.new_password" type="password" class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 dark:text-white">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">{{ t('auth.profile.confirm_password') }}</label>
                                <input v-model="pwdForm.confirm_password" type="password" class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 dark:text-white">
                            </div>
                            <div class="flex justify-between items-center pt-2">
                                <button type="button" @click="logoutAll" :disabled="revoking" class="px-3 py-2 text-xs font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition">
                                    {{ t('auth.profile.logout_all') }}
                                </button>
                                <button type="submit" :disabled="savingPwd" class="px-4 py-2 bg-gray-800 dark:bg-gray-600 hover:bg-gray-900 dark:hover:bg-gray-500 text-white text-sm font-medium rounded-lg transition disabled:opacity-50">
                                    {{ t('auth.profile.change_password') }}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </transition>
</template>

<script>
import http, { formatAPIError } from '../utils/http.js';

export default {
    props: {
        show: Boolean,
        user: Object,
        t: Function
    },
    emits: ['close', 'user-updated', 'force-logout'],
    inject: ['showToast'],
    data() {
        return {
            profileForm: {
                nickname: '',
                email: ''
            },
            pwdForm: {
                old_password: '',
                new_password: '',
                confirm_password: ''
            },
            savingProfile: false,
            savingPwd: false,
            revoking: false
        };
    },
    watch: {
        show(val) {
            if (val && this.user) {
                this.profileForm.nickname = this.user.nickname || '';
                this.profileForm.email = this.user.email || '';
                this.pwdForm = { old_password: '', new_password: '', confirm_password: '' };
            }
        }
    },
    methods: {
        async saveProfile() {
            this.savingProfile = true;
            try {
                const res = await http.patch('/api/auth/me', this.profileForm);
                this.$emit('user-updated', res.data);
                this.showToast(this.t('auth.profile.updated'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingProfile = false;
            }
        },
        async changePassword() {
            if (this.pwdForm.new_password.length < 8) {
                this.showToast(this.t('error.AUTH_PASSWORD_TOO_SHORT'), 'error');
                return;
            }
            if (this.pwdForm.new_password !== this.pwdForm.confirm_password) {
                this.showToast(this.t('error.AUTH_PASSWORD_MISMATCH'), 'error');
                return;
            }
            this.savingPwd = true;
            try {
                await http.post('/api/auth/me/password', {
                    old_password: this.pwdForm.old_password,
                    new_password: this.pwdForm.new_password
                });
                this.showToast(this.t('auth.profile.password_changed'), 'success');
                this.$emit('force-logout');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingPwd = false;
            }
        },
        async logoutAll() {
            this.revoking = true;
            try {
                await http.post('/api/auth/me/logout-all');
                this.showToast(this.t('auth.profile.sessions_revoked'), 'success');
                this.$emit('force-logout');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.revoking = false;
            }
        }
    }
}
</script>
