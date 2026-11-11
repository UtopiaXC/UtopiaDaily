<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <div class="flex space-x-2">
                <button
                    @click="activeTab = 'users'"
                    :class="[
                        'px-4 py-2 text-sm font-medium rounded-md transition-colors',
                        activeTab === 'users'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                    ]"
                >
                    {{ t('users.tab.users') }}
                </button>
                <button
                    @click="activeTab = 'roles'"
                    :class="[
                        'px-4 py-2 text-sm font-medium rounded-md transition-colors',
                        activeTab === 'roles'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                    ]"
                >
                    {{ t('users.tab.roles') }}
                </button>
            </div>
            <button
                @click="refreshAll"
                class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow"
            >
                <span class="material-icons text-sm mr-1" :class="{'animate-spin': loading}">refresh</span>
                {{ t('common.refresh') }}
            </button>
        </div>

        <div class="p-6 flex-1 overflow-y-auto">
            <div v-if="activeTab === 'users'">
                <div v-if="canEdit" class="flex justify-end mb-4">
                    <button
                        @click="openUserModal()"
                        class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition flex items-center gap-1 shadow-sm"
                    >
                        <span class="material-icons text-sm">add</span>
                        {{ t('users.create_user') }}
                    </button>
                </div>

                <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
                    <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                        <thead class="bg-gray-50 dark:bg-gray-700/60">
                            <tr>
                                <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('users.username') }}</th>
                                <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('users.nickname') }}</th>
                                <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('users.email') }}</th>
                                <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('users.role') }}</th>
                                <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('users.status') }}</th>
                                <th v-if="canEdit" class="px-5 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('common.actions') }}</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                            <tr v-if="users.length === 0">
                                <td :colspan="canEdit ? 6 : 5" class="px-5 py-8 text-center text-sm text-gray-400">
                                    {{ t('common.no_data') }}
                                </td>
                            </tr>
                            <tr v-for="u in users" :key="u.id" class="hover:bg-gray-50/70 dark:hover:bg-gray-700/30">
                                <td class="px-5 py-3.5 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ u.username }}</td>
                                <td class="px-5 py-3.5 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">{{ u.nickname || '-' }}</td>
                                <td class="px-5 py-3.5 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ u.email || '-' }}</td>
                                <td class="px-5 py-3.5 whitespace-nowrap text-sm">
                                    <span class="px-2.5 py-0.5 text-xs font-semibold rounded-full bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 border border-blue-200 dark:border-blue-800">
                                        {{ u.role_name }}
                                    </span>
                                </td>
                                <td class="px-5 py-3.5 whitespace-nowrap text-sm">
                                    <span :class="statusBadgeClass(u.status)" class="px-2.5 py-0.5 text-xs font-semibold rounded-full border">
                                        {{ t('users.status.' + u.status, u.status) }}
                                    </span>
                                </td>
                                <td v-if="canEdit" class="px-5 py-3.5 whitespace-nowrap text-right text-xs font-medium space-x-3">
                                    <button
                                        v-if="u.id !== currentUser?.id && u.status !== 'banned'"
                                        @click="banUser(u)"
                                        class="text-orange-600 dark:text-orange-400 hover:underline"
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
                                        class="text-amber-600 dark:text-amber-400 hover:underline"
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
                                        class="text-red-600 dark:text-red-400 hover:underline"
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
                    <button
                        @click="openRoleModal()"
                        class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition flex items-center gap-1 shadow-sm"
                    >
                        <span class="material-icons text-sm">add</span>
                        {{ t('roles.create_role') }}
                    </button>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                    <div
                        v-for="role in roles"
                        :key="role.id"
                        class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-5 flex flex-col justify-between hover:shadow-md transition-shadow"
                    >
                        <div>
                            <div class="flex items-start justify-between gap-2 mb-2">
                                <div class="flex items-center gap-2 min-w-0">
                                    <h4 class="text-base font-bold text-gray-900 dark:text-white truncate">{{ role.name }}</h4>
                                    <span
                                        v-if="role.is_system"
                                        class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 font-medium"
                                    >
                                        {{ t('roles.system_badge') }}
                                    </span>
                                </div>
                                <span class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                                    {{ t('roles.user_count') }}: {{ role.user_count }}
                                </span>
                            </div>
                            <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
                                {{ roleDescription(role) }}
                            </p>
                            <div class="text-xs text-gray-500 dark:text-gray-400">
                                {{ t('roles.permissions') }}: {{ (role.permissions || []).length }}
                            </div>
                        </div>

                        <div v-if="canEdit" class="flex justify-end gap-3 mt-4 pt-3 border-t border-gray-100 dark:border-gray-700">
                            <button
                                @click="openRoleModal(role)"
                                class="text-blue-600 dark:text-blue-400 hover:underline text-xs font-medium"
                            >
                                {{ t('common.edit') }}
                            </button>
                            <button
                                v-if="!role.is_system"
                                @click="confirmDeleteRole(role)"
                                class="text-red-600 dark:text-red-400 hover:underline text-xs font-medium"
                            >
                                {{ t('common.delete') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <transition name="modal">
            <div v-if="userModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
                <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md border border-gray-200 dark:border-gray-700">
                    <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">
                        {{ userModal.editing ? t('users.edit_user') : t('users.create_user') }}
                    </h3>
                    <form @submit.prevent="saveUser" class="space-y-4">
                        <div>
                            <div class="flex items-center gap-1.5 mb-1">
                                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('users.username') }}</label>
                                <div v-if="!userModal.editing" class="relative group flex items-center">
                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-56 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('auth.register.username_hint') }}
                                    </div>
                                </div>
                            </div>
                            <input
                                v-model="userForm.username"
                                type="text"
                                :disabled="!!userModal.editing"
                                class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-900"
                            >
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('users.nickname') }}</label>
                            <input
                                v-model="userForm.nickname"
                                type="text"
                                class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white"
                            >
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('users.email') }}</label>
                            <input
                                v-model="userForm.email"
                                type="email"
                                class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white"
                            >
                        </div>
                        <div v-if="!userModal.editing">
                            <div class="flex items-center gap-1.5 mb-1">
                                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('users.password') }}</label>
                                <div class="relative group flex items-center">
                                    <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-48 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t('auth.register.password_hint') }}
                                    </div>
                                </div>
                            </div>
                            <input
                                v-model="userForm.password"
                                type="password"
                                class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white"
                            >
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('users.role') }}</label>
                            <CustomSelect v-model="userForm.role_id" :options="roleSelectOptions" />
                        </div>
                        <div v-if="userModal.editing">
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('users.status') }}</label>
                            <CustomSelect v-model="userForm.status" :options="userStatusOptions" />
                        </div>
                        <div class="flex justify-end gap-3 pt-2">
                            <button
                                type="button"
                                @click="userModal.show = false"
                                class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                            >
                                {{ t('common.cancel') }}
                            </button>
                            <button
                                type="submit"
                                :disabled="submitting"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg disabled:opacity-50"
                            >
                                {{ t('common.save') }}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </transition>

        <transition name="modal">
            <div v-if="pwdModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
                <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm border border-gray-200 dark:border-gray-700">
                    <h3 class="text-lg font-bold mb-1 text-gray-800 dark:text-white">{{ t('users.reset_password') }}</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">{{ pwdModal.user?.username }}</p>
                    <form @submit.prevent="submitResetPassword" class="space-y-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('auth.reset.new_password') }}</label>
                            <input
                                v-model="pwdModal.password"
                                type="password"
                                class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white"
                            >
                        </div>
                        <div class="flex justify-end gap-3">
                            <button
                                type="button"
                                @click="pwdModal.show = false"
                                class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                            >
                                {{ t('common.cancel') }}
                            </button>
                            <button
                                type="submit"
                                :disabled="submitting"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg disabled:opacity-50"
                            >
                                {{ t('common.confirm') }}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </transition>

        <transition name="modal">
            <div v-if="roleModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm p-4">
                <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-2xl max-h-[88vh] flex flex-col border border-gray-200 dark:border-gray-700">
                    <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">
                        {{ roleModal.editing ? t('roles.edit_role') : t('roles.create_role') }}
                    </h3>

                    <div v-if="isEditingAdminRole" class="mb-4 p-3 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 text-xs text-amber-700 dark:text-amber-300">
                        {{ t('roles.admin_locked') }}
                    </div>

                    <form @submit.prevent="saveRole" class="flex-1 overflow-y-auto space-y-4 pr-1">
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div>
                                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('roles.name') }}</label>
                                <input
                                    v-model="roleForm.name"
                                    type="text"
                                    :disabled="roleModal.editing?.is_system"
                                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-900"
                                >
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('roles.description') }}</label>
                                <input
                                    v-model="roleForm.description"
                                    type="text"
                                    :disabled="isEditingAdminRole"
                                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-900"
                                >
                            </div>
                        </div>

                        <div>
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('roles.permissions') }}</label>
                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                <div
                                    v-for="(perms, grp) in groupedPermissions"
                                    :key="grp"
                                    class="border border-gray-200 dark:border-gray-700 rounded-lg p-3 bg-gray-50/50 dark:bg-gray-900/30"
                                >
                                    <div class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">
                                        {{ t('permission_group.' + grp, grp) }}
                                    </div>
                                    <div class="space-y-1.5">
                                        <label
                                            v-for="p in perms"
                                            :key="p.code"
                                            class="flex items-start gap-2 text-xs text-gray-700 dark:text-gray-300 cursor-pointer"
                                        >
                                            <input
                                                type="checkbox"
                                                :value="p.code"
                                                v-model="roleForm.permissions"
                                                :disabled="isEditingAdminRole"
                                                class="mt-0.5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                            >
                                            <span>{{ t('permission.' + p.code, p.code) }}</span>
                                        </label>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="flex justify-end gap-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                            <button
                                type="button"
                                @click="roleModal.show = false"
                                class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                            >
                                {{ t('common.cancel') }}
                            </button>
                            <button
                                v-if="!isEditingAdminRole"
                                type="submit"
                                :disabled="submitting"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg disabled:opacity-50"
                            >
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
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            activeTab: 'users',
            loading: false,
            submitting: false,
            users: [],
            roles: [],
            permissions: [],
            userModal: {
                show: false,
                editing: null
            },
            userForm: {
                username: '',
                nickname: '',
                email: '',
                password: '',
                role_id: 2,
                status: 'active'
            },
            pwdModal: {
                show: false,
                user: null,
                password: ''
            },
            roleModal: {
                show: false,
                editing: null
            },
            roleForm: {
                name: '',
                description: '',
                permissions: []
            },
            confirmModal: {
                show: false,
                title: '',
                message: '',
                onConfirm: null
            }
        };
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
                    return 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800';
                case 'frozen':
                case 'inactive':
                    return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:border-amber-800';
                default:
                    return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800';
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
                this.roleForm = {
                    name: '',
                    description: '',
                    permissions: []
                };
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
