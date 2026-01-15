<template>
    <div class="flex h-screen overflow-hidden bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
        <Sidebar
            :server-name="serverName"
            :user="user"
            :menu="menu"
            :current-tab="currentTab"
            :t="t"
            @update:tab="updateTab"
            @logout="$emit('logout')"
        />
        <div class="flex-1 flex flex-col overflow-hidden">
            <Header
                :current-tab-label="currentTabLabel"
                :is-dark-mode="isDarkMode"
                :current-locale="currentLocale"
                :ui-locale-options="uiLocaleOptions"
                @toggle-dark-mode="$emit('toggle-dark-mode')"
                @update:locale="$emit('update:locale', $event)"
            />
            <main class="flex-1 overflow-x-hidden overflow-y-auto p-8 bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
                <component :is="currentTabComponent" :t="t" :all-locales="allLocales" />
            </main>
        </div>
    </div>
</template>

<script>
import Sidebar from '../layout/Sidebar.vue';
import Header from '../layout/Header.vue';
import Overview from './tabs/Overview.vue';
import SystemConfig from './tabs/SystemConfig.vue';

export default {
    components: { Sidebar, Header, Overview, SystemConfig },
    props: {
        serverName: String,
        user: Object,
        menu: Array,
        isDarkMode: Boolean,
        currentLocale: String,
        uiLocaleOptions: Array,
        allLocales: Array,
        t: Function
    },
    emits: ['logout', 'toggle-dark-mode', 'update:locale'],
    data() {
        return {
            currentTab: 'dashboard'
        }
    },
    computed: {
        currentTabLabel() {
            if (this.currentTab === 'dashboard') return this.t('menu.overview');
            const item = this.menu.find(i => i.id === this.currentTab);
            return item ? this.t('menu.' + item.id, item.label) : 'Dashboard';
        },
        currentTabComponent() {
            switch(this.currentTab) {
                case 'dashboard': return 'Overview';
                case 'system_config': return 'SystemConfig';
                default: return {
                    props: ['t'],
                    template: `
                        <div class="bg-white dark:bg-gray-800 shadow-sm rounded-xl border border-gray-200 dark:border-gray-700 p-12 text-center">
                            <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-700 mb-4">
                                <span class="material-icons text-gray-400 dark:text-gray-500 text-3xl">construction</span>
                            </div>
                            <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">{{ currentTabLabel }}</h3>
                            <p class="text-gray-500 dark:text-gray-400">This module is under construction.</p>
                        </div>
                    `,
                    computed: {
                        currentTabLabel() {
                            const item = this.$parent.menu.find(i => i.id === this.$parent.currentTab);
                            return item ? this.t('menu.' + item.id, item.label) : '';
                        }
                    }
                }
            }
        }
    },
    methods: {
        updateTab(tab) {
            this.currentTab = tab;
        }
    }
}
</script>
