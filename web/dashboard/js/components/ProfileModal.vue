<template>
    <transition name="modal">
        <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
            <div class="bg-white dark:bg-slate-900 rounded-lg w-full max-w-md border border-slate-200 dark:border-slate-800 overflow-hidden">
                <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center bg-slate-50 dark:bg-slate-900/60">
                    <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('auth.profile.title') }}</h3>
                    <button @click="$emit('close')" class="text-slate-400 hover:text-slate-700 dark:hover:text-slate-200">
                        <span class="material-icons text-[20px]">close</span>
                    </button>
                </div>

                <div class="p-5 space-y-5 max-h-[80vh] overflow-y-auto" v-if="user">
                    <form @submit.prevent="saveProfile" class="space-y-3.5">
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                            <div>
                                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.username') }}</label>
                                <input :value="user.username" disabled class="ui-input">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.role') }}</label>
                                <input :value="user.role_name" disabled class="ui-input">
                            </div>
                        </div>
                        <div>
                            <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.nickname') }}</label>
                            <input v-model="profileForm.nickname" type="text" class="ui-input">
                        </div>
                        <div>
                            <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.email') }}</label>
                            <input v-model="profileForm.email" type="email" class="ui-input">
                        </div>
                        <div class="flex justify-end">
                            <button type="submit" :disabled="savingProfile" class="ui-btn-primary">
                                <span v-if="savingProfile" class="material-icons text-[16px] animate-spin">sync</span>
                                {{ t('auth.profile.save') }}
                            </button>
                        </div>
                    </form>

                    <div class="border-t border-slate-200 dark:border-slate-800 pt-4">
                        <h4 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-3">{{ t('auth.profile.password_section') }}</h4>
                        <form @submit.prevent="changePassword" class="space-y-3">
                            <div>
                                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.old_password') }}</label>
                                <input v-model="pwdForm.old_password" type="password" class="ui-input">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.new_password') }}</label>
                                <input v-model="pwdForm.new_password" type="password" class="ui-input">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">{{ t('auth.profile.confirm_password') }}</label>
                                <input v-model="pwdForm.confirm_password" type="password" class="ui-input">
                            </div>
                            <div class="flex justify-between items-center pt-1">
                                <button type="button" @click="logoutAll" :disabled="revoking" class="px-3 py-1.5 text-xs font-medium text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 rounded-md transition">
                                    {{ t('auth.profile.logout_all') }}
                                </button>
                                <button type="submit" :disabled="savingPwd" class="ui-btn-primary">
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
};
</script>
