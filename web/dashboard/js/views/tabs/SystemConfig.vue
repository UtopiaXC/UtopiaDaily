<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 flex flex-col h-full">
        <div class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <div class="px-6 py-4 flex justify-between items-center">
                <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('menu.system_config') }}</h3>
                <button
                    @click="fetchSystemConfigs"
                    class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow"
                >
                    <span class="material-icons text-sm mr-1" :class="{'animate-spin': loading}">refresh</span>
                    {{ t('common.refresh') }}
                </button>
            </div>

            <div class="flex px-6 gap-6 overflow-x-auto">
                <button
                    v-for="tab in sortedTabs"
                    :key="tab"
                    @click="currentTab = tab"
                    :class="[
                        'pb-3 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        currentTab === tab
                            ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                    ]"
                >
                    {{ t('config.tab.' + tab, tab) }}
                </button>
            </div>
        </div>

        <div v-if="loading" class="p-12 text-center text-gray-500 dark:text-gray-400 flex-1">
            <span class="material-icons animate-spin text-3xl mb-2 text-blue-500">sync</span>
            <p class="text-sm">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-6 space-y-6 overflow-y-auto flex-1">
            <div v-if="currentTab === 'classification'" class="space-y-4">
                <div class="p-4 rounded-lg border border-blue-200 dark:border-blue-900 bg-blue-50/70 dark:bg-blue-900/20 flex items-center gap-3 text-sm text-blue-700 dark:text-blue-300">
                    <span class="material-icons text-blue-500">info</span>
                    <span>{{ t('config.classification.slot_notice') }}</span>
                </div>
                <div v-if="classificationSlot" class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50/60 dark:bg-gray-900/30">
                        <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('config.classification.provider_label') }}</div>
                        <div class="text-base font-bold text-gray-800 dark:text-white mt-1 font-mono">{{ classificationSlot.provider }}</div>
                    </div>
                    <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50/60 dark:bg-gray-900/30">
                        <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('config.classification.groups_label') }}</div>
                        <div class="text-base font-bold text-gray-800 dark:text-white mt-1">{{ classificationSlot.tag_group_count }}</div>
                    </div>
                    <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50/60 dark:bg-gray-900/30">
                        <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('config.classification.tags_label') }}</div>
                        <div class="text-base font-bold text-gray-800 dark:text-white mt-1">{{ classificationSlot.tag_count }}</div>
                    </div>
                </div>
            </div>

            <div v-if="Object.keys(currentTabGroups).length === 0" class="text-center text-gray-500 dark:text-gray-400 py-8">
                {{ t('common.no_data') }}
            </div>

            <div v-for="(items, group) in currentTabGroups" :key="group" class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm">
                <div class="px-4 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700 font-semibold text-gray-700 dark:text-gray-200 text-xs uppercase tracking-wide rounded-t-lg">
                    {{ t('config.group.' + group, group) }}
                </div>

                <div class="p-4 space-y-5">
                    <div v-for="conf in items" :key="conf.key" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5">
                            <label class="text-sm font-semibold text-gray-800 dark:text-gray-200">
                                {{ t(conf.description, conf.key) }}
                            </label>
                            <div v-if="conf.hint" class="relative group flex items-center">
                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500 transition-colors">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t(conf.hint) }}
                                </div>
                            </div>
                        </div>

                        <div class="flex flex-col sm:flex-row gap-3 sm:items-center">
                            <div class="flex-1">
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
                                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:text-gray-500"
                                >
                                <input
                                    v-else-if="conf.type === 'password' || conf.is_secret"
                                    v-model="conf.value"
                                    type="password"
                                    :disabled="!conf.is_editable || !canEdit"
                                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:text-gray-500"
                                >
                                <input
                                    v-else
                                    v-model="conf.value"
                                    type="text"
                                    :disabled="!conf.is_editable || !canEdit"
                                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:text-gray-500"
                                >
                            </div>

                            <button
                                v-if="conf.is_editable && canEdit"
                                @click="saveConfig(conf)"
                                :disabled="savingKey === conf.key"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-1.5 transition-colors shadow-sm disabled:opacity-50 flex-shrink-0"
                            >
                                <span class="material-icons text-sm" :class="{'animate-spin': savingKey === conf.key}">
                                    {{ savingKey === conf.key ? 'sync' : 'save' }}
                                </span>
                                {{ savingKey === conf.key ? t('common.saving') : t('common.save') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div v-if="currentTab === 'smtp' && canEdit" class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm">
                <div class="px-4 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700 font-semibold text-gray-700 dark:text-gray-200 text-xs uppercase tracking-wide rounded-t-lg">
                    {{ t('config.smtp.test_title') }}
                </div>
                <div class="p-4 flex flex-col sm:flex-row gap-3 sm:items-center">
                    <input
                        v-model="testRecipient"
                        type="email"
                        :placeholder="t('config.smtp.test_recipient')"
                        class="flex-1 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                    <button
                        @click="testSMTP"
                        :disabled="testingSMTP"
                        class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-1.5 transition-colors shadow-sm disabled:opacity-50 flex-shrink-0"
                    >
                        <span class="material-icons text-sm" :class="{'animate-spin': testingSMTP}">
                            {{ testingSMTP ? 'sync' : 'mail_outline' }}
                        </span>
                        {{ testingSMTP ? t('common.testing') : t('config.smtp.test_send') }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';

const TAB_ORDER = ['system', 'security', 'smtp', 'classification'];

export default {
    components: { CustomSelect },
    props: ['t', 'currentUser'],
    inject: ['showToast'],
    data() {
        return {
            configs: [],
            originalValues: {},
            loading: false,
            savingKey: '',
            currentTab: 'system',
            testRecipient: '',
            testingSMTP: false,
            classificationSlot: null
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('system.config.edit');
        },
        sortedTabs() {
            const found = Array.from(new Set(this.configs.map(c => c.tab)));
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
        async fetchSystemConfigs() {
            this.loading = true;
            try {
                const [cfgRes, slotRes] = await Promise.all([
                    http.get('/api/dashboard/system-config'),
                    http.get('/api/dashboard/system-config/classification-slot').catch(() => ({ data: null }))
                ]);
                this.configs = (cfgRes.data || []).map(item => ({
                    ...item,
                    value: item.value ?? ''
                }));
                this.classificationSlot = slotRes.data;
                const snap = {};
                this.configs.forEach(item => {
                    snap[item.key] = item.value;
                });
                this.originalValues = snap;
                if (this.sortedTabs.length > 0 && !this.sortedTabs.includes(this.currentTab)) {
                    this.currentTab = this.sortedTabs[0];
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loading = false;
            }
        },
        async saveConfig(conf) {
            if (conf.is_secret && conf.value === '********') {
                return;
            }
            this.savingKey = conf.key;
            try {
                await http.put(`/api/dashboard/system-config/${encodeURIComponent(conf.key)}`, {
                    value: String(conf.value)
                });
                this.originalValues[conf.key] = String(conf.value);
                if (conf.key === 'server_name') {
                    window.dispatchEvent(new CustomEvent('server-name-updated', { detail: String(conf.value) }));
                }
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingKey = '';
            }
        },
        async testSMTP() {
            this.testingSMTP = true;
            try {
                const smtpItems = this.configs.filter(c => c.tab === 'smtp' && c.is_editable);
                for (const item of smtpItems) {
                    const curr = String(item.value ?? '');
                    if (item.is_secret && curr === '********') continue;
                    if (curr !== this.originalValues[item.key]) {
                        await http.put(`/api/dashboard/system-config/${encodeURIComponent(item.key)}`, {
                            value: curr
                        });
                        this.originalValues[item.key] = curr;
                    }
                }
                await http.post('/api/dashboard/system-config/smtp/test', {
                    recipient: this.testRecipient.trim()
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
