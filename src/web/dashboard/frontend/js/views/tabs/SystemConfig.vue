<template>
    <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 rounded-t-xl">
            <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('menu.system_config') }}</h3>
            <button @click="fetchSystemConfigs" class="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 text-sm flex items-center transition-colors bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 px-3 py-1.5 rounded-md shadow-sm hover:shadow">
                <span class="material-icons text-sm mr-1">refresh</span> {{ t('common.refresh') }}
            </button>
        </div>

        <div v-if="loading" class="p-12 text-center text-gray-500 dark:text-gray-400">
            <span class="material-icons animate-spin text-3xl mb-2 text-blue-500">sync</span>
            <p>{{ t('common.loading') }}</p>
        </div>

        <div v-else class="p-6 space-y-6">
            <div v-for="conf in configs" :key="conf.key" class="flex flex-col md:flex-row md:items-center gap-6 border-b border-gray-100 dark:border-gray-700 pb-6 last:border-0 last:pb-0">
                <div class="md:w-1/3">
                    <label class="block text-sm font-bold text-gray-800 dark:text-gray-200">{{ t(conf.description, conf.description) }}</label>
                </div>

                <div class="md:w-2/3 flex gap-3 items-center">
                    <input v-if="conf.type === 'string'" v-model="conf.value" type="text"
                           :disabled="!conf.is_editable"
                           :class="['flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500 transition bg-white dark:bg-gray-700 dark:text-white', !conf.is_editable ? 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed' : '']">

                    <div v-else-if="conf.type === 'boolean'" class="flex-1">
                        <CustomSelect
                            v-if="conf.is_editable"
                            v-model="conf.value"
                            :options="[{label: 'True', value: 'true'}, {label: 'False', value: 'false'}]"
                        />
                        <input v-else type="text" :value="conf.value" disabled class="flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed w-full">
                    </div>

                    <div v-else-if="conf.type === 'select'" class="flex-1">
                        <CustomSelect
                            v-if="conf.is_editable"
                            v-model="conf.value"
                            :options="conf.options"
                        />
                        <input v-else type="text" :value="conf.value" disabled class="flex-1 shadow-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed w-full">
                    </div>

                    <button v-if="conf.is_editable" @click="saveConfig(conf)"
                            class="ml-2 bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700 transition flex items-center disabled:opacity-50 shadow-sm"
                            :disabled="savingConfig === conf.key">
                        <span v-if="savingConfig === conf.key" class="material-icons text-sm animate-spin mr-1">sync</span>
                        {{ savingConfig === conf.key ? t('common.saving') : t('common.save') }}
                    </button>
                </div>
            </div>
            <div v-if="configs.length === 0" class="text-center text-gray-500 dark:text-gray-400 py-8">
                {{ t('common.no_data') }}
            </div>
        </div>
    </div>
</template>

<script>
import CustomSelect from '../../components/CustomSelect.vue';
import http from '../../utils/http.js';

export default {
    components: { CustomSelect },
    props: ['t', 'allLocales'],
    inject: ['showToast'],
    data() {
        return {
            loading: false,
            configs: [],
            savingConfig: null
        }
    },
    mounted() {
        this.fetchSystemConfigs();
    },
    methods: {
        async fetchSystemConfigs() {
            this.loading = true;
            try {
                const res = await http.get('/api/dashboard/system-config/');
                this.configs = res.data.map(conf => {
                    if (conf.key === 'default_locale') {
                        conf.options = this.allLocales;
                    }
                    return conf;
                });
            } catch (err) {
                console.error("Failed to fetch configs", err);
            } finally {
                this.loading = false;
            }
        },
        async saveConfig(conf) {
            this.savingConfig = conf.key;
            try {
                await http.put(`/api/dashboard/system-config/${conf.key}`, {
                    value: conf.value
                });
                this.showToast(this.t('common.saved_success', 'Configuration saved'), 'success');
                if (conf.key === 'server_name') {
                    window.dispatchEvent(new CustomEvent('server-name-updated', { detail: conf.value }));
                }
            } catch (err) {
                const msg = err.response?.data?.detail || err.message;
                this.showToast(this.t('common.saved_failed', 'Failed to save: ') + msg, 'error');
            } finally {
                this.savingConfig = null;
            }
        }
    }
}
</script>
