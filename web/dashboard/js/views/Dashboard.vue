<template>
    <div class="flex h-screen overflow-hidden bg-slate-50 dark:bg-slate-950 transition-colors duration-150">
        <div
            v-if="isSidebarOpen"
            class="fixed inset-0 bg-slate-950/50 z-20 md:hidden"
            @click="isSidebarOpen = false"
        ></div>

        <Sidebar
            :server-name="serverName"
            :user="user"
            :menu="menu"
            :current-tab="currentTab"
            :t="t"
            :is-open="isSidebarOpen"
            :is-collapsed="isSidebarCollapsed"
            @update:tab="updateTab"
            @logout="$emit('logout')"
            @toggle-collapse="toggleSidebar"
            @close-mobile="isSidebarOpen = false"
            @open-profile="profileModalOpen = true"
        />

        <div class="flex-1 flex flex-col overflow-hidden min-w-0">
            <Header
                :current-tab-label="currentTabLabel"
                :is-dark-mode="isDarkMode"
                :current-locale="currentLocale"
                :ui-locale-options="uiLocaleOptions"
                :t="t"
                @toggle-dark-mode="$emit('toggle-dark-mode')"
                @update:locale="$emit('update:locale', $event)"
                @toggle-sidebar="toggleSidebar"
            />
            <main class="flex-1 overflow-x-hidden overflow-y-auto p-4 md:p-6 bg-slate-50 dark:bg-slate-950 transition-colors duration-150">
                <component
                    :is="currentTabComponent"
                    :t="t"
                    :all-locales="allLocales"
                    :current-user="user"
                    :current-locale="currentLocale"
                    v-bind="currentTabProps"
                    @navigate-tab="updateTab"
                    v-on="currentTabListeners"
                />
            </main>
        </div>

        <ProfileModal
            :show="profileModalOpen"
            :user="user"
            :t="t"
            @close="profileModalOpen = false"
            @user-updated="$emit('user-updated', $event)"
            @force-logout="$emit('logout')"
        />
    </div>
</template>

<script>
import Sidebar from '../layout/Sidebar.vue';
import Header from '../layout/Header.vue';
import ProfileModal from '../components/ProfileModal.vue';
import Overview from './tabs/Overview.vue';
import SystemConfig from './tabs/SystemConfig.vue';
import UserManager from './tabs/UserManager.vue';
import ScraperModules from './tabs/ScraperModules.vue';
import Classification from './tabs/Classification.vue';
import DataManager from './tabs/DataManager.vue';
import PushModules from './tabs/PushModules.vue';
import Schedule from './tabs/Schedule.vue';
import FrontendConfig from './tabs/FrontendConfig.vue';
import UserPushConfig from './tabs/UserPushConfig.vue';
import {
    parseDashboardPath,
    buildDashboardPath,
    normalizeSubtab,
    defaultSubtabFor,
    tabSupportsSubtab
} from '../utils/router.js';

const TAB_COMPONENTS = {
    overview: 'Overview',
    system_config: 'SystemConfig',
    user_manager: 'UserManager',
    scraper_modules: 'ScraperModules',
    classification: 'Classification',
    data_manager: 'DataManager',
    push_modules: 'PushModules',
    schedule: 'Schedule',
    frontend_config: 'FrontendConfig',
    user_push: 'UserPushConfig'
};

export default {
    components: {
        Sidebar,
        Header,
        ProfileModal,
        Overview,
        SystemConfig,
        UserManager,
        ScraperModules,
        Classification,
        DataManager,
        PushModules,
        Schedule,
        FrontendConfig,
        UserPushConfig
    },
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
    emits: ['logout', 'toggle-dark-mode', 'update:locale', 'user-updated'],
    data() {
        const initial = parseDashboardPath(window.location.pathname);
        return {
            currentTab: initial.tab || '',
            currentSubtab: initial.subtab || '',
            isSidebarOpen: false,
            isSidebarCollapsed: false,
            profileModalOpen: false
        };
    },
    watch: {
        menu: {
            immediate: true,
            handler(list) {
                if (!Array.isArray(list) || list.length === 0) return;
                const parsed = parseDashboardPath(window.location.pathname);
                let targetTab = parsed.tab || this.currentTab;
                let targetSubtab = parsed.tab ? parsed.subtab : this.currentSubtab;
                if (!list.some(item => item.id === targetTab)) {
                    targetTab = list[0].id;
                    targetSubtab = defaultSubtabFor(targetTab);
                } else {
                    targetSubtab = normalizeSubtab(targetTab, targetSubtab);
                }
                this.currentTab = targetTab;
                this.currentSubtab = targetSubtab;
                this.syncBrowserUrl(targetTab, targetSubtab, true);
            }
        }
    },
    created() {
        this.checkScreenSize();
        window.addEventListener('resize', this.checkScreenSize);
        window.addEventListener('popstate', this.onPopState);
    },
    unmounted() {
        window.removeEventListener('resize', this.checkScreenSize);
        window.removeEventListener('popstate', this.onPopState);
    },
    computed: {
        currentTabLabel() {
            const item = (this.menu || []).find(i => i.id === this.currentTab);
            return item ? this.t(item.label, this.currentTab) : this.t('common.dashboard');
        },
        currentTabComponent() {
            return TAB_COMPONENTS[this.currentTab] || 'Overview';
        },
        currentTabSupportsSubtab() {
            return tabSupportsSubtab(this.currentTab);
        },
        currentTabProps() {
            return this.currentTabSupportsSubtab ? { subtab: this.currentSubtab } : {};
        },
        currentTabListeners() {
            return this.currentTabSupportsSubtab ? { 'update:subtab': this.updateSubtab } : {};
        }
    },
    methods: {
        syncBrowserUrl(tab, subtab, replace = false) {
            if (!tab) return;
            const target = buildDashboardPath(tab, subtab);
            if (target === window.location.pathname) return;
            const state = { tab, subtab };
            if (replace) {
                window.history.replaceState(state, '', target);
            } else {
                window.history.pushState(state, '', target);
            }
        },
        onPopState() {
            const parsed = parseDashboardPath(window.location.pathname);
            const list = this.menu || [];
            if (list.length === 0) {
                this.currentTab = parsed.tab || '';
                this.currentSubtab = parsed.subtab || '';
                return;
            }
            if (!list.some(item => item.id === parsed.tab)) {
                const fallbackTab = list[0].id;
                const fallbackSubtab = defaultSubtabFor(fallbackTab);
                this.currentTab = fallbackTab;
                this.currentSubtab = fallbackSubtab;
                this.syncBrowserUrl(fallbackTab, fallbackSubtab, true);
                return;
            }
            this.currentTab = parsed.tab;
            this.currentSubtab = normalizeSubtab(parsed.tab, parsed.subtab);
        },
        updateTab(tab, subtab) {
            let nextTab = tab;
            let explicitSubtab = subtab;
            if (tab && typeof tab === 'object') {
                nextTab = tab.tab;
                explicitSubtab = tab.subtab;
            }
            if ((this.menu || []).some(item => item.id === nextTab)) {
                const nextSubtab = explicitSubtab !== undefined
                    ? normalizeSubtab(nextTab, explicitSubtab)
                    : defaultSubtabFor(nextTab);
                this.currentTab = nextTab;
                this.currentSubtab = nextSubtab;
                this.syncBrowserUrl(nextTab, nextSubtab, false);
            }
            if (window.innerWidth < 768) {
                this.isSidebarOpen = false;
            }
        },
        updateSubtab(payload, replace = false) {
            let next = payload;
            let useReplace = replace;
            if (payload && typeof payload === 'object') {
                next = payload.subtab ?? '';
                useReplace = Boolean(payload.replace);
            }
            if (!tabSupportsSubtab(this.currentTab)) return;
            const normalized = normalizeSubtab(this.currentTab, next);
            this.currentSubtab = normalized;
            this.syncBrowserUrl(this.currentTab, normalized, useReplace);
        },
        toggleSidebar() {
            if (window.innerWidth < 768) {
                this.isSidebarOpen = !this.isSidebarOpen;
            } else {
                this.isSidebarCollapsed = !this.isSidebarCollapsed;
            }
        },
        checkScreenSize() {
            if (window.innerWidth < 1024 && window.innerWidth >= 768) {
                this.isSidebarCollapsed = true;
            } else if (window.innerWidth >= 1024) {
                this.isSidebarCollapsed = false;
            }
            if (window.innerWidth >= 768) {
                this.isSidebarOpen = false;
            }
        }
    }
};
</script>
