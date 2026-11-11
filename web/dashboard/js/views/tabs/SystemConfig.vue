<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header pb-0">
            <div class="flex justify-between items-center pb-3">
                <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('menu.system_config') }}</h3>
                <button @click="fetchSystemConfigs" class="ui-btn-secondary">
                    <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                    <span>{{ t('common.refresh') }}</span>
                </button>
            </div>

            <div class="flex gap-6 overflow-x-auto">
                <button
                    v-for="tab in sortedTabs"
                    :key="tab"
                    @click="switchTab(tab)"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        currentTab === tab
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    {{ t('config.tab.' + tab, tab) }}
                </button>
            </div>
        </div>

        <div v-if="loading" class="p-12 text-center text-slate-500 dark:text-slate-400 flex-1">
            <span class="material-icons animate-spin text-2xl mb-2 text-blue-600">sync</span>
            <p class="text-xs">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-5 space-y-5 overflow-y-auto flex-1">
            <div v-if="Object.keys(currentTabGroups).length === 0" class="text-center text-slate-400 py-8 text-sm">
                {{ t('common.no_data') }}
            </div>

            <div v-for="(items, group) in currentTabGroups" :key="group" class="ui-card">
                <div class="ui-card-header">
                    <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                        {{ t('config.group.' + group, group) }}
                    </span>
                    <button
                        v-if="canEdit && groupHasEditable(items)"
                        @click="saveGroup(group, items)"
                        :disabled="savingGroup === group"
                        class="ui-btn-primary text-xs py-1.5 px-3"
                    >
                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingGroup === group }">
                            {{ savingGroup === group ? 'sync' : 'save' }}
                        </span>
                        <span>{{ savingGroup === group ? t('common.saving') : t('common.save') }}</span>
                    </button>
                </div>

                <div class="p-4 grid grid-cols-1 md:grid-cols-2 gap-5">
                    <div v-for="conf in items" :key="conf.key" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5">
                            <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">
                                {{ t(conf.description, conf.key) }}
                            </label>
                            <div v-if="conf.hint" class="relative group flex items-center">
                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t(conf.hint) }}
                                </div>
                            </div>
                        </div>

                        <div>
                            <CustomSelect
                                v-if="conf.type === 'select' && conf.is_editable && canEdit"
                                v-model="conf.value"
                                :options="formatOptions(conf.options)"
                            />
                            <CustomSelect
                                v-else-if="conf.type === 'boolean' && conf.is_editable && canEdit"
                                v-model="conf.value"
                                :options="[
                                    { label: t('common.true'), value: 'true' },
                                    { label: t('common.false'), value: 'false' }
                                ]"
                            />
                            <input
                                v-else-if="conf.type === 'int'"
                                v-model="conf.value"
                                type="number"
                                :disabled="!conf.is_editable || !canEdit"
                                class="ui-input"
                            >
                            <input
                                v-else-if="conf.type === 'password' || conf.is_secret"
                                v-model="conf.value"
                                type="password"
                                :disabled="!conf.is_editable || !canEdit"
                                class="ui-input"
                            >
                            <input
                                v-else
                                v-model="conf.value"
                                type="text"
                                :disabled="!conf.is_editable || !canEdit"
                                class="ui-input"
                            >
                        </div>
                    </div>
                </div>
            </div>

            <div v-if="currentTab === 'smtp' && canEdit" class="ui-card">
                <div class="ui-card-header">
                    <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                        {{ t('config.smtp.test_title') }}
                    </span>
                </div>
                <div class="p-4 flex flex-col sm:flex-row gap-3 sm:items-center">
                    <input
                        v-model="testRecipient"
                        type="email"
                        :placeholder="t('config.smtp.test_recipient')"
                        class="ui-input flex-1"
                    >
                    <button
                        @click="testSMTP"
                        :disabled="testingSMTP"
                        class="ui-btn-emerald flex-shrink-0"
                    >
                        <span class="material-icons text-[16px]" :class="{ 'animate-spin': testingSMTP }">
                            {{ testingSMTP ? 'sync' : 'mail_outline' }}
                        </span>
                        <span>{{ testingSMTP ? t('common.testing') : t('config.smtp.test_send') }}</span>
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';

const TAB_ORDER = ['system', 'security', 'smtp'];

export default {
    components: { CustomSelect },
    props: ['t', 'currentUser', 'subtab'],
    emits: ['navigate-tab', 'update:subtab'],
    inject: ['showToast'],
    data() {
        return {
            configs: [],
            originalValues: {},
            loading: false,
            savingGroup: '',
            currentTab: String(this.subtab || 'system').trim() || 'system',
            testRecipient: '',
            testingSMTP: false
        };
    },
    watch: {
        subtab(next) {
            const target = String(next || '').trim() || 'system';
            if (target !== this.currentTab && (this.sortedTabs.length === 0 || this.sortedTabs.includes(target))) {
                this.currentTab = target;
            }
        }
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('system.config.edit');
        },
        sortedTabs() {
            const found = Array.from(new Set(this.configs.map(c => c.tab).filter(t => t && t !== 'classification')));
            return TAB_ORDER.filter(t => found.includes(t)).concat(found.filter(t => !TAB_ORDER.includes(t)));
        },
        currentTabGroups() {
            const groups = {};
            const filtered = this.configs
                .filter(c => c.tab === this.currentTab)
                .slice()
                .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));

            filtered.forEach(item => {
                const grp = item.group || 'basic';
                if (!groups[grp]) groups[grp] = [];
                groups[grp].push(item);
            });
            return groups;
        }
    },
    mounted() {
        this.fetchSystemConfigs();
    },
    methods: {
        switchTab(tab) {
            this.currentTab = tab;
            this.$emit('update:subtab', tab);
        },
        formatOptions(options) {
            return (options || []).map(opt => {
                if (typeof opt === 'object' && opt !== null) {
                    return {
                        value: opt.value,
                        label: opt.label_key ? this.t(opt.label_key, opt.label || opt.value) : (opt.label || opt.value)
                    };
                }
                return { value: opt, label: String(opt) };
            });
        },
        groupHasEditable(items) {
            return Array.isArray(items) && items.some(i => i.is_editable);
        },
        async fetchSystemConfigs() {
            this.loading = true;
            try {
                const cfgRes = await http.get('/api/dashboard/system-config');
                this.configs = (cfgRes.data || [])
                    .filter(item => item.tab !== 'classification')
                    .map(item => ({
                        ...item,
                        value: item.value ?? ''
                    }));
                const snap = {};
                this.configs.forEach(item => {
                    snap[item.key] = item.value;
                });
                this.originalValues = snap;
                if (this.sortedTabs.length > 0 && !this.sortedTabs.includes(this.currentTab)) {
                    this.currentTab = this.sortedTabs[0];
                    this.$emit('update:subtab', { subtab: this.currentTab, replace: true });
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        async saveGroup(group, items) {
            const payload = {};
            (items || []).forEach(conf => {
                if (!conf.is_editable) return;
                const val = String(conf.value ?? '');
                if (conf.is_secret && val === '********') return;
                payload[conf.key] = val;
            });
            this.savingGroup = group;
            try {
                await http.put('/api/dashboard/system-config', { configs: payload });
                Object.keys(payload).forEach(k => {
                    this.originalValues[k] = payload[k];
                });
                if (Object.prototype.hasOwnProperty.call(payload, 'server_name')) {
                    window.dispatchEvent(new CustomEvent('server-name-updated', { detail: payload.server_name }));
                }
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingGroup = '';
            }
        },
        async testSMTP() {
            this.testingSMTP = true;
            try {
                const smtpVal = key => {
                    const item = this.configs.find(c => c.key === key);
                    return item ? String(item.value ?? '') : '';
                };
                await http.post('/api/dashboard/system-config/smtp/test', {
                    recipient: this.testRecipient.trim(),
                    smtp_host: smtpVal('smtp_host'),
                    smtp_port: smtpVal('smtp_port'),
                    smtp_encryption: smtpVal('smtp_encryption'),
                    smtp_user: smtpVal('smtp_user'),
                    smtp_password: smtpVal('smtp_password'),
                    smtp_from: smtpVal('smtp_from'),
                    smtp_from_name: smtpVal('smtp_from_name')
                });
                this.showToast(this.t('config.smtp.test_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.testingSMTP = false;
            }
        }
    }
};
</script>
