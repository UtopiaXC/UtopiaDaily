<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header pb-0 flex justify-between items-center">
            <div class="flex gap-6">
                <button
                    @click="switchTab('users')"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        activeTab === 'users'
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    {{ t('users.tab.users') }}
                </button>
                <button
                    @click="switchTab('roles')"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        activeTab === 'roles'
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    {{ t('users.tab.roles') }}
                </button>
            </div>
            <button @click="refreshAll" class="ui-btn-secondary mb-2">
                <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                <span>{{ t('common.refresh') }}</span>
            </button>
        </div>

        <div class="p-5 flex-1 overflow-y-auto">
            <div v-if="activeTab === 'users'">
                <div v-if="canEdit" class="flex justify-end mb-4">
                    <button @click="openUserModal()" class="ui-btn-primary text-xs py-1.5 px-3">
                        <span class="material-icons text-[16px]">add</span>
                        <span>{{ t('users.create_user') }}</span>
                    </button>
                </div>

                <div class="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-lg">
                    <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800">
                        <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                            <tr>
                                <th class="px-4 py-3 text-left font-semibold">{{ t('users.username') }}</th>
                                <th class="px-4 py-3 text-left font-semibold">{{ t('users.nickname') }}</th>
                                <th class="px-4 py-3 text-left font-semibold">{{ t('users.email') }}</th>
                                <th class="px-4 py-3 text-left font-semibold">{{ t('users.role') }}</th>
                                <th class="px-4 py-3 text-left font-semibold">{{ t('users.status') }}</th>
                                <th v-if="canEdit" class="px-4 py-3 text-right font-semibold">{{ t('common.actions') }}</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white dark:bg-slate-900 divide-y divide-slate-200 dark:divide-slate-800">
                            <tr v-if="users.length === 0">
                                <td :colspan="canEdit ? 6 : 5" class="px-4 py-8 text-center text-xs text-slate-400">
                                    {{ t('common.no_data') }}
                                </td>
                            </tr>
                            <tr v-for="u in users" :key="u.id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                                <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-slate-900 dark:text-white">{{ u.username }}</td>
                                <td class="px-4 py-3 whitespace-nowrap text-sm text-slate-600 dark:text-slate-300">{{ u.nickname || '-' }}</td>
                                <td class="px-4 py-3 whitespace-nowrap text-xs text-slate-500 dark:text-slate-400 font-mono">{{ u.email || '-' }}</td>
                                <td class="px-4 py-3 whitespace-nowrap text-xs">
                                    <span class="px-2 py-0.5 font-medium rounded bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200 dark:border-blue-900">
                                        {{ u.role_name }}
                                    </span>
                                </td>
                                <td class="px-4 py-3 whitespace-nowrap text-xs">
                                    <span :class="statusBadgeClass(u.status)" class="px-2 py-0.5 font-medium rounded border">
                                        {{ t('users.status.' + u.status, u.status) }}
                                    </span>
                                </td>
                                <td v-if="canEdit" class="px-4 py-3 whitespace-nowrap text-right text-xs font-medium space-x-3">
                                    <button
                                        v-if="u.id !== currentUser?.id && u.status !== 'banned'"
                                        @click="banUser(u)"
                                        class="text-amber-600 dark:text-amber-400 hover:underline"
                                    >
                                        {{ t('users.ban') }}
                                    </button>
                                    <button
                                        v-if="u.status === 'frozen' || u.status === 'banned'"
                                        @click="unfreezeUser(u)"
                                        class="text-emerald-600 dark:text-emerald-400 hover:underline"
                                    >
                                        {{ u.status === 'banned' ? t('users.unban') : t('users.unfreeze') }}
                                    </button>
                                    <button
                                        @click="openResetPwdModal(u)"
                                        class="text-slate-600 dark:text-slate-300 hover:underline"
                                    >
                                        {{ t('users.reset_password') }}
                                    </button>
                                    <button
                                        @click="openUserModal(u)"
                                        class="text-blue-600 dark:text-blue-400 hover:underline"
                                    >
                                        {{ t('common.edit') }}
                                    </button>
                                    <button
                                        v-if="u.id !== currentUser?.id"
                                        @click="confirmDeleteUser(u)"
                                        class="text-rose-600 dark:text-rose-400 hover:underline"
                                    >
                                        {{ t('common.delete') }}
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div v-else>
                <div v-if="canEdit" class="flex justify-end mb-4">
                    <button @click="openRoleModal()" class="ui-btn-primary text-xs py-1.5 px-3">
                        <span class="material-icons text-[16px]">add</span>
                        <span>{{ t('roles.create_role') }}</span>
                    </button>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    <div
                        v-for="role in roles"
                        :key="role.id"
                        class="ui-card p-4 flex flex-col justify-between hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
                    >
                        <div>
                            <div class="flex items-start justify-between gap-2 mb-2">
                                <div class="flex items-center gap-2 min-w-0">
                                    <h4 class="text-sm font-semibold text-slate-900 dark:text-white truncate">{{ role.name }}</h4>
                                    <span
                                        v-if="role.is_system"
                                        class="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 font-medium border border-slate-200 dark:border-slate-700"
                                    >
                                        {{ t('roles.system_badge') }}
                                    </span>
                                </div>
                                <span class="text-xs text-slate-500 dark:text-slate-400 whitespace-nowrap font-mono">
                                    {{ t('roles.user_count') }}: {{ role.user_count }}
                                </span>
                            </div>
                            <p class="text-xs text-slate-500 dark:text-slate-400 mb-3">
                                {{ roleDescription(role) }}
                            </p>
                            <div class="text-xs text-slate-500 dark:text-slate-400 font-mono">
                                {{ t('roles.permissions') }}: {{ (role.permissions || []).length }}
                            </div>
                        </div>

                        <div v-if="canEdit" class="flex justify-end gap-3 mt-4 pt-3 border-t border-slate-200 dark:border-slate-800">
                            <button
                                @click="openRoleModal(role)"
                                class="text-blue-600 dark:text-blue-400 hover:underline text-xs font-medium"
                            >
                                {{ t('common.edit') }}
                            </button>
                            <button
                                v-if="!role.is_system"
                                @click="confirmDeleteRole(role)"
                                class="text-rose-600 dark:text-rose-400 hover:underline text-xs font-medium"
                            >
                                {{ t('common.delete') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <transition name="modal">
            <div v-if="userModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg p-6 w-full max-w-md border border-slate-200 dark:border-slate-800">
                    <h3 class="text-base font-semibold mb-4 text-slate-900 dark:text-white">
                        {{ userModal.editing ? t('users.edit_user') : t('users.create_user') }}
                    </h3>
                    <form @submit.prevent="saveUser" class="space-y-4">
                        <div>
                            <div class="flex items-center gap-1.5 mb-1">
                                <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">{{ t('users.username') }}</label>
                                <div v-if="!userModal.editing" class="relative group flex items-center">
                                    <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-56 p-2 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('auth.register.username_hint') }}
                                    </div>
                                </div>
                            </div>
                            <input
                                v-model="userForm.username"
                                type="text"
                                :disabled="!!userModal.editing"
                                class="ui-input"
                            >
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('users.nickname') }}</label>
                            <input v-model="userForm.nickname" type="text" class="ui-input">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('users.email') }}</label>
                            <input v-model="userForm.email" type="email" class="ui-input">
                        </div>
                        <div v-if="!userModal.editing">
                            <div class="flex items-center gap-1.5 mb-1">
                                <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">{{ t('users.password') }}</label>
                                <div class="relative group flex items-center">
                                    <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-48 p-2 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('auth.register.password_hint') }}
                                    </div>
                                </div>
                            </div>
                            <input v-model="userForm.password" type="password" class="ui-input">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('users.role') }}</label>
                            <CustomSelect v-model="userForm.role_id" :options="roleSelectOptions" />
                        </div>
                        <div v-if="userModal.editing">
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('users.status') }}</label>
                            <CustomSelect v-model="userForm.status" :options="userStatusOptions" />
                        </div>
                        <div class="flex justify-end gap-2 pt-2">
                            <button type="button" @click="userModal.show = false" class="ui-btn-secondary">
                                {{ t('common.cancel') }}
                            </button>
                            <button type="submit" :disabled="submitting" class="ui-btn-primary">
                                {{ t('common.save') }}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </transition>

        <transition name="modal">
            <div v-if="pwdModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg p-6 w-full max-w-sm border border-slate-200 dark:border-slate-800">
                    <h3 class="text-base font-semibold mb-1 text-slate-900 dark:text-white">{{ t('users.reset_password') }}</h3>
                    <p class="text-xs text-slate-500 dark:text-slate-400 mb-4 font-mono">{{ pwdModal.user?.username }}</p>
                    <form @submit.prevent="submitResetPassword" class="space-y-4">
                        <div>
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('auth.reset.new_password') }}</label>
                            <input v-model="pwdModal.password" type="password" class="ui-input">
                        </div>
                        <div class="flex justify-end gap-2">
                            <button type="button" @click="pwdModal.show = false" class="ui-btn-secondary">
                                {{ t('common.cancel') }}
                            </button>
                            <button type="submit" :disabled="submitting" class="ui-btn-primary">
                                {{ t('common.confirm') }}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </transition>

        <transition name="modal">
            <div v-if="roleModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg p-6 w-full max-w-2xl max-h-[88vh] flex flex-col border border-slate-200 dark:border-slate-800">
                    <h3 class="text-base font-semibold mb-4 text-slate-900 dark:text-white">
                        {{ roleModal.editing ? t('roles.edit_role') : t('roles.create_role') }}
                    </h3>

                    <div v-if="isEditingAdminRole" class="mb-4 p-3 rounded-md bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900 text-xs text-amber-800 dark:text-amber-300">
                        {{ t('roles.admin_locked') }}
                    </div>

                    <form @submit.prevent="saveRole" class="flex-1 overflow-y-auto space-y-4 pr-1">
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div>
                                <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('roles.name') }}</label>
                                <input v-model="roleForm.name" type="text" :disabled="roleModal.editing?.is_system" class="ui-input">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('roles.description') }}</label>
                                <input v-model="roleForm.description" type="text" :disabled="isEditingAdminRole" class="ui-input">
                            </div>
                        </div>

                        <div>
                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-2">{{ t('roles.permissions') }}</label>
                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                <div
                                    v-for="(perms, grp) in groupedPermissions"
                                    :key="grp"
                                    class="border border-slate-200 dark:border-slate-800 rounded-md p-3 bg-slate-50/60 dark:bg-slate-950/40"
                                >
                                    <div class="text-[11px] font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
                                        {{ t('permission_group.' + grp, grp) }}
                                    </div>
                                    <div class="space-y-1.5">
                                        <label
                                            v-for="p in perms"
                                            :key="p.code"
                                            class="flex items-start gap-2 text-xs text-slate-700 dark:text-slate-300 cursor-pointer"
                                        >
                                            <input
                                                type="checkbox"
                                                :value="p.code"
                                                v-model="roleForm.permissions"
                                                :disabled="isEditingAdminRole"
                                                class="mt-0.5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500"
                                            >
                                            <span>{{ t('permission.' + p.code, p.code) }}</span>
                                        </label>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="flex justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
                            <button type="button" @click="roleModal.show = false" class="ui-btn-secondary">
                                {{ t('common.cancel') }}
                            </button>
                            <button v-if="!isEditingAdminRole" type="submit" :disabled="submitting" class="ui-btn-primary">
                                {{ t('common.save') }}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </transition>

        <ConfirmModal
            :show="confirmModal.show"
            :title="confirmModal.title"
            :message="confirmModal.message"
            :t="t"
            @confirm="executeConfirm"
            @cancel="confirmModal.show = false"
        />
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';

export default {
    components: { CustomSelect, ConfirmModal },
    props: ['t', 'currentUser', 'subtab'],
    emits: ['update:subtab'],
    inject: ['showToast'],
    data() {
        const initialTab = ['users', 'roles'].includes(this.subtab) ? this.subtab : 'users';
        return {
            activeTab: initialTab,
            loading: false,
            submitting: false,
            users: [],
            roles: [],
            permissions: [],
            userModal: { show: false, editing: null },
            userForm: { username: '', nickname: '', email: '', password: '', role_id: 2, status: 'active' },
            pwdModal: { show: false, user: null, password: '' },
            roleModal: { show: false, editing: null },
            roleForm: { name: '', description: '', permissions: [] },
            confirmModal: { show: false, title: '', message: '', onConfirm: null }
        };
    },
    watch: {
        subtab(next) {
            const target = ['users', 'roles'].includes(next) ? next : 'users';
            if (target !== this.activeTab) {
                this.activeTab = target;
            }
        }
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('user_manager.edit');
        },
        roleSelectOptions() {
            return this.roles.map(r => ({ label: r.name, value: r.id }));
        },
        userStatusOptions() {
            return ['active', 'inactive', 'frozen', 'banned', 'disabled'].map(s => ({
                value: s,
                label: this.t('users.status.' + s, s)
            }));
        },
        groupedPermissions() {
            const groups = {};
            this.permissions.forEach(p => {
                const grp = p.group_code || 'system';
                if (!groups[grp]) groups[grp] = [];
                groups[grp].push(p);
            });
            return groups;
        },
        isEditingAdminRole() {
            return this.roleModal.editing?.name === 'admin';
        }
    },
    mounted() {
        this.refreshAll();
    },
    methods: {
        switchTab(tab) {
            const target = ['users', 'roles'].includes(tab) ? tab : 'users';
            this.activeTab = target;
            this.$emit('update:subtab', target);
        },
        async refreshAll() {
            this.loading = true;
            try {
                const [usersRes, rolesRes, permsRes] = await Promise.all([
                    http.get('/api/dashboard/user-manager/users'),
                    http.get('/api/dashboard/user-manager/roles'),
                    http.get('/api/dashboard/user-manager/permissions')
                ]);
                this.users = usersRes.data || [];
                this.roles = rolesRes.data || [];
                this.permissions = permsRes.data || [];
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        statusBadgeClass(status) {
            switch (status) {
                case 'active':
                    return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900';
                case 'frozen':
                case 'inactive':
                    return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900';
                default:
                    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
            }
        },
        roleDescription(role) {
            if (!role.description) return this.t('common.no_data');
            return this.t(role.description, role.description);
        },
        openUserModal(user = null) {
            this.userModal = { show: true, editing: user };
            if (user) {
                this.userForm = {
                    username: user.username,
                    nickname: user.nickname || '',
                    email: user.email || '',
                    password: '',
                    role_id: user.role_id,
                    status: user.status || 'active'
                };
            } else {
                const defaultRole = this.roles.find(r => r.name === 'user') || this.roles[0];
                this.userForm = {
                    username: '',
                    nickname: '',
                    email: '',
                    password: '',
                    role_id: defaultRole ? defaultRole.id : 2,
                    status: 'active'
                };
            }
        },
        async saveUser() {
            this.submitting = true;
            try {
                if (this.userModal.editing) {
                    await http.put(`/api/dashboard/user-manager/users/${this.userModal.editing.id}`, {
                        nickname: this.userForm.nickname,
                        email: this.userForm.email,
                        role_id: this.userForm.role_id,
                        status: this.userForm.status
                    });
                } else {
                    await http.post('/api/dashboard/user-manager/users', {
                        username: this.userForm.username,
                        nickname: this.userForm.nickname,
                        email: this.userForm.email,
                        password: this.userForm.password,
                        role_id: this.userForm.role_id
                    });
                }
                this.userModal.show = false;
                this.showToast(this.t('common.saved_success'), 'success');
                await this.refreshAll();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.submitting = false;
            }
        },
        openResetPwdModal(user) {
            this.pwdModal = { show: true, user, password: '' };
        },
        async submitResetPassword() {
            if ((this.pwdModal.password || '').length < 8) {
                this.showToast(this.t('error.AUTH_PASSWORD_TOO_SHORT'), 'error');
                return;
            }
            this.submitting = true;
            try {
                await http.post(`/api/dashboard/user-manager/users/${this.pwdModal.user.id}/password`, {
                    password: this.pwdModal.password
                });
                this.pwdModal.show = false;
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.submitting = false;
            }
        },
        async banUser(user) {
            try {
                await http.post(`/api/dashboard/user-manager/users/${user.id}/ban`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.refreshAll();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async unfreezeUser(user) {
            try {
                await http.post(`/api/dashboard/user-manager/users/${user.id}/unfreeze`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.refreshAll();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        confirmDeleteUser(user) {
            this.confirmModal = {
                show: true,
                title: this.t('users.delete_user'),
                message: this.t('users.delete_confirm').replace('{username}', user.username),
                onConfirm: async () => {
                    await http.delete(`/api/dashboard/user-manager/users/${user.id}`);
                    this.showToast(this.t('common.saved_success'), 'success');
                    await this.refreshAll();
                }
            };
        },
        openRoleModal(role = null) {
            this.roleModal = { show: true, editing: role };
            if (role) {
                this.roleForm = {
                    name: role.name,
                    description: role.description || '',
                    permissions: Array.isArray(role.permissions) ? [...role.permissions] : []
                };
            } else {
                this.roleForm = { name: '', description: '', permissions: [] };
            }
        },
        async saveRole() {
            if (this.isEditingAdminRole) return;
            this.submitting = true;
            try {
                const payload = {
                    name: this.roleForm.name,
                    description: this.roleForm.description,
                    permissions: this.roleForm.permissions
                };
                if (this.roleModal.editing) {
                    await http.put(`/api/dashboard/user-manager/roles/${this.roleModal.editing.id}`, payload);
                } else {
                    await http.post('/api/dashboard/user-manager/roles', payload);
                }
                this.roleModal.show = false;
                this.showToast(this.t('common.saved_success'), 'success');
                await this.refreshAll();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.submitting = false;
            }
        },
        confirmDeleteRole(role) {
            this.confirmModal = {
                show: true,
                title: this.t('roles.delete_role'),
                message: this.t('roles.delete_confirm').replace('{name}', role.name),
                onConfirm: async () => {
                    await http.delete(`/api/dashboard/user-manager/roles/${role.id}`);
                    this.showToast(this.t('common.saved_success'), 'success');
                    await this.refreshAll();
                }
            };
        },
        async executeConfirm() {
            const fn = this.confirmModal.onConfirm;
            this.confirmModal.show = false;
            if (!fn) return;
            try {
                await fn();
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        }
    }
};
</script>
