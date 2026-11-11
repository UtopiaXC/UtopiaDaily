<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <!-- Tabs Header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <div class="flex space-x-4">
                <button
                    @click="activeTab = 'users'"
                    :class="['px-4 py-2 text-sm font-medium rounded-md transition-colors', activeTab === 'users' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700']">
                    {{ t('user_manager.users') }}
                </button>
                <button
                    @click="activeTab = 'roles'"
                    :class="['px-4 py-2 text-sm font-medium rounded-md transition-colors', activeTab === 'roles' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700']">
                    {{ t('user_manager.roles') }}
                </button>
            </div>
            <button @click="refresh" class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow">
                <span class="material-icons text-sm mr-1">refresh</span> {{ t('common.refresh') }}
            </button>
        </div>

        <div class="p-6 flex-1 overflow-y-auto">
            <div v-if="activeTab === 'users'">
                <div class="flex justify-end mb-4">
                    <button @click="openUserModal()" class="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700 transition flex items-center shadow-sm">
                        <span class="material-icons text-sm mr-1">add</span> {{ t('user_manager.add_user') }}
                    </button>
                </div>

                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                        <thead class="bg-gray-50 dark:bg-gray-700">
                            <tr>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('login.username') }}</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('login.nickname') }}</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('common.email') }}</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('user_manager.role') }}</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('common.status') }}</th>
                                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">{{ t('common.actions') }}</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                            <tr v-for="user in users" :key="user.id">
                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ user.username }}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ user.nickname || '-' }}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ user.email || '-' }}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                                        {{ user.role_name }}
                                    </span>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                    <span :class="['px-2 inline-flex text-xs leading-5 font-semibold rounded-full', user.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200']">
                                        {{ user.is_active ? t('common.active') : t('common.inactive') }}
                                    </span>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <button
                                        v-if="user.id !== currentUser.id"
                                        @click="confirmResetPassword(user)"
                                        class="text-yellow-600 dark:text-yellow-400 hover:text-yellow-900 dark:hover:text-yellow-300 mr-3">
                                        {{ t('common.reset_password') }}
                                    </button>
                                    <button @click="openUserModal(user)" class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 mr-3">{{ t('common.edit') }}</button>
                                    <button @click="confirmDeleteUser(user)" class="text-red-600 dark:text-red-400 hover:text-red-900 dark:hover:text-red-300">{{ t('common.delete') }}</button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div v-else-if="activeTab === 'roles'">
                <div class="flex justify-end mb-4">
                    <button @click="openRoleModal()" class="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700 transition flex items-center shadow-sm">
                        <span class="material-icons text-sm mr-1">add</span> {{ t('user_manager.add_role') }}
                    </button>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div v-for="role in roles" :key="role.id" class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 hover:shadow-md transition-shadow">
                        <div class="flex justify-between items-start mb-4">
                            <div>
                                <h4 class="text-lg font-bold text-gray-900 dark:text-white">{{ role.name }}</h4>
                                <p class="text-sm text-gray-500 dark:text-gray-400">{{ role.description || t('common.no_data') }}</p>
                            </div>
                            <span class="bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs px-2 py-1 rounded-full">
                                {{ t('user_manager.users_count').replace('{count}', role.user_count) }}
                            </span>
                        </div>
                        <div class="flex justify-end space-x-2 mt-4">
                            <button @click="openRoleModal(role)" class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 text-sm font-medium">{{ t('common.edit') }}</button>
                            <button @click="confirmDeleteRole(role)" class="text-red-600 dark:text-red-400 hover:text-red-800 text-sm font-medium">{{ t('common.delete') }}</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="showUserModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-96 max-w-full">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ editingUser ? t('user_manager.edit_user') : t('user_manager.add_user') }}</h3>

                <form @submit.prevent="saveUser">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('login.username') }}</label>
                        <input v-model="userForm.username" type="text" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('login.nickname') }}</label>
                        <input v-model="userForm.nickname" type="text" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.email') }}</label>
                        <input v-model="userForm.email" type="email" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('login.password') }}</label>
                        <input v-model="userForm.password" type="password" :placeholder="editingUser ? t('login.password_placeholder_edit') : ''" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('user_manager.role') }}</label>
                        <select v-model="userForm.role_id" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                            <option v-for="role in roles" :key="role.id" :value="role.id">{{ role.name }}</option>
                        </select>
                    </div>
                    <div class="mb-6 flex items-center">
                        <input v-model="userForm.is_active" type="checkbox" class="h-4 w-4 text-blue-600 border-gray-300 rounded">
                        <label class="ml-2 block text-sm text-gray-900 dark:text-gray-300">{{ t('common.active') }}</label>
                    </div>

                    <div class="flex justify-end gap-2">
                        <button type="button" @click="showUserModal = false" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">{{ t('common.cancel') }}</button>
                        <button type="submit" class="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition">{{ t('common.save') }}</button>
                    </div>
                </form>
            </div>
        </div>

        <div v-if="showRoleModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-[600px] max-w-full max-h-[90vh] flex flex-col">
                <h3 class="text-lg font-bold mb-4 text-gray-800 dark:text-white">{{ editingRole ? t('user_manager.edit_role') : t('user_manager.add_role') }}</h3>

                <div class="flex-1 overflow-y-auto pr-2">
                    <form @submit.prevent="saveRole" id="roleForm">
                        <div class="mb-4">
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('user_manager.role_name') }}</label>
                            <input v-model="roleForm.name" :disabled="roleForm.name === 'admin'" type="text" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed">
                        </div>
                        <div class="mb-4">
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('user_manager.role_desc') }}</label>
                            <input v-model="roleForm.description" type="text" class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 dark:text-white">
                        </div>

                        <div class="mb-4">
                            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('user_manager.permissions') }}</label>

                            <!-- Admin Readonly Message -->
                            <div v-if="roleForm.name === 'admin'" class="mb-3 p-3 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 text-sm rounded-md flex items-start">
                                <span class="material-icons text-sm mr-2 mt-0.5">info</span>
                                <span>{{ t('user_manager.admin_readonly') }}</span>
                            </div>

                            <div class="space-y-4">
                                <div v-for="(perms, group) in groupedPermissions" :key="group" class="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-md border border-gray-200 dark:border-gray-700">
                                    <h5 class="text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3 border-b border-gray-200 dark:border-gray-600 pb-2">
                                        {{ t('permission_group.' + group, group) }}
                                    </h5>
                                    <div class="grid grid-cols-2 gap-2">
                                        <div v-for="perm in perms" :key="perm" class="flex items-center">
                                            <input type="checkbox" :id="perm" :value="perm" v-model="roleForm.permissions"
                                                   :disabled="roleForm.name === 'admin'"
                                                   class="h-4 w-4 text-blue-600 border-gray-300 rounded disabled:opacity-50 disabled:cursor-not-allowed">
                                            <label :for="perm" class="ml-2 block text-xs text-gray-700 dark:text-gray-300 break-all cursor-pointer">{{ t('permission.' + perm, perm) }}</label>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>

                <div class="flex justify-end gap-2 mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
                    <button type="button" @click="showRoleModal = false" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">{{ t('common.cancel') }}</button>
                    <button type="submit" form="roleForm" class="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition">{{ t('common.save') }}</button>
                </div>
            </div>
        </div>

        <ConfirmModal
            :show="showConfirmModal"
            :title="confirmTitle"
            :message="confirmMessage"
            :t="t"
            @confirm="handleConfirm"
            @cancel="showConfirmModal = false"
        />

        <InfoModal
            :show="showInfoModal"
            :title="infoTitle"
            :message="infoMessage"
            :t="t"
            @close="showInfoModal = false"
        />
    </div>
</template>

<script>
import http from '../../utils/http.js';
import ConfirmModal from '../../components/ConfirmModal.vue';
import InfoModal from '../../components/InfoModal.vue';
import MD5 from 'crypto-js/md5';

export default {
    components: { ConfirmModal, InfoModal },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            activeTab: 'users',
            users: [],
            roles: [],
            allPermissions: [],

            // User Modal
            showUserModal: false,
            editingUser: null,
            userForm: {
                username: '',
                nickname: '',
                email: '',
                password: '',
                role_id: '',
                is_active: true
            },

            // Role Modal
            showRoleModal: false,
            editingRole: null,
            roleForm: {
                name: '',
                description: '',
                permissions: []
            },

            // Confirm Modal
            showConfirmModal: false,
            confirmTitle: '',
            confirmMessage: '',
            confirmAction: null,
            confirmData: null,

            // Info Modal
            showInfoModal: false,
            infoTitle: '',
            infoMessage: ''
        }
    },
    computed: {
        groupedPermissions() {
            const groups = {};
            this.allPermissions.forEach(perm => {
                const parts = perm.split('.');
                const group = parts[0];
                if (!groups[group]) {
                    groups[group] = [];
                }
                groups[group].push(perm);
            });
            return groups;
        }
    },
    mounted() {
        this.refresh();
    },
    methods: {
        async refresh() {
            await Promise.all([
                this.fetchUsers(),
                this.fetchRoles(),
                this.fetchPermissions()
            ]);
        },
        async fetchUsers() {
            try {
                const res = await http.get('/api/dashboard/user-manager/users');
                this.users = res.data;
            } catch (err) {
                console.error("Failed to fetch users", err);
            }
        },
        async fetchRoles() {
            try {
                const res = await http.get('/api/dashboard/user-manager/roles');
                this.roles = res.data;
            } catch (err) {
                console.error("Failed to fetch roles", err);
            }
        },
        async fetchPermissions() {
            try {
                const res = await http.get('/api/dashboard/user-manager/permissions');
                this.allPermissions = res.data;
            } catch (err) {
                console.error("Failed to fetch permissions", err);
            }
        },

        openUserModal(user = null) {
            this.editingUser = user;
            if (user) {
                this.userForm = {
                    username: user.username,
                    nickname: user.nickname || '',
                    email: user.email,
                    password: '', // Don't fill password
                    role_id: user.role_id,
                    is_active: user.is_active
                };
            } else {
                this.userForm = {
                    username: '',
                    nickname: '',
                    email: '',
                    password: '',
                    role_id: this.roles.length > 0 ? this.roles[0].id : '',
                    is_active: true
                };
            }
            this.showUserModal = true;
        },
        async saveUser() {
            try {
                if (!this.editingUser) {
                    const usernameRegex = /^[a-zA-Z0-9_\-\.@]+$/;
                    if (this.userForm.username.length < 3 || !usernameRegex.test(this.userForm.username)) {
                        this.showToast(this.t('login.username_invalid'), 'error');
                        return;
                    }
                }

                if (this.userForm.password && this.userForm.password.length < 5) {
                    this.showToast(this.t('login.password_too_short'), 'error');
                    return;
                }

                const payload = { ...this.userForm };

                if (payload.password) {
                    payload.password = MD5(payload.password).toString();
                } else {
                    delete payload.password;
                }

                if (this.editingUser) {
                    // Allow username update if changed
                    if (payload.username === this.editingUser.username) {
                        delete payload.username;
                    }
                    await http.put(`/api/dashboard/user-manager/users/${this.editingUser.id}`, payload);
                    this.showToast(this.t('common.saved_success'), 'success');
                } else {
                    await http.post('/api/dashboard/user-manager/users', payload);
                    this.showToast(this.t('common.saved_success'), 'success');
                }
                this.showUserModal = false;
                this.fetchUsers();
                this.fetchRoles(); // Refresh roles to update user count
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            }
        },

        confirmDeleteUser(user) {
            this.confirmTitle = this.t('common.delete');
            this.confirmMessage = this.t('common.confirm_delete').replace('{name}', user.username);
            this.confirmAction = 'deleteUser';
            this.confirmData = user;
            this.showConfirmModal = true;
        },
        confirmDeleteRole(role) {
            this.confirmTitle = this.t('common.delete');
            this.confirmMessage = this.t('common.confirm_delete').replace('{name}', role.name);
            this.confirmAction = 'deleteRole';
            this.confirmData = role;
            this.showConfirmModal = true;
        },
        confirmResetPassword(user) {
            this.confirmTitle = this.t('common.reset_password');
            this.confirmMessage = this.t('common.confirm_reset_password').replace('{name}', user.username);
            this.confirmAction = 'resetPassword';
            this.confirmData = user;
            this.showConfirmModal = true;
        },
        async handleConfirm() {
            this.showConfirmModal = false;
            if (this.confirmAction === 'deleteUser') {
                await this.deleteUser(this.confirmData);
            } else if (this.confirmAction === 'deleteRole') {
                await this.deleteRole(this.confirmData);
            } else if (this.confirmAction === 'resetPassword') {
                await this.resetPassword(this.confirmData);
            }
        },

        async deleteUser(user) {
            try {
                await http.delete(`/api/dashboard/user-manager/users/${user.id}`);
                this.showToast('User deleted', 'success');
                this.fetchUsers();
                this.fetchRoles(); // Refresh roles to update user count
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            }
        },
        async deleteRole(role) {
            try {
                await http.delete(`/api/dashboard/user-manager/roles/${role.id}`);
                this.showToast('Role deleted', 'success');
                this.fetchRoles();
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            }
        },
        async resetPassword(user) {
            try {
                const res = await http.post(`/api/dashboard/user-manager/users/${user.id}/reset-password`);
                const newPass = res.data.new_password;

                this.infoTitle = this.t('common.password_reset_success').replace('{password}', '');
                this.infoMessage = newPass;
                this.showInfoModal = true;

            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            }
        },

        openRoleModal(role = null) {
            this.editingRole = role;
            if (role) {
                const perms = Object.keys(role.permissions).filter(k => role.permissions[k]);
                this.roleForm = {
                    name: role.name,
                    description: role.description,
                    permissions: perms
                };
            } else {
                this.roleForm = {
                    name: '',
                    description: '',
                    permissions: []
                };
            }
            this.showRoleModal = true;
        },
        async saveRole() {
            try {
                const permDict = {};
                this.roleForm.permissions.forEach(p => permDict[p] = true);

                const payload = {
                    name: this.roleForm.name,
                    description: this.roleForm.description,
                    permissions: permDict
                };

                if (this.editingRole) {
                    await http.put(`/api/dashboard/user-manager/roles/${this.editingRole.id}`, payload);
                    this.showToast(this.t('common.saved_success'), 'success');
                } else {
                    await http.post('/api/dashboard/user-manager/roles', payload);
                    this.showToast(this.t('common.saved_success'), 'success');
                }
                this.showRoleModal = false;
                this.fetchRoles();
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.failed_op') + msg, 'error');
            }
        }
    }
}
</script>
