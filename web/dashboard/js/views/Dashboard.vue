<template>
    <div class="flex h-screen overflow-hidden bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
        <div
            v-if="isSidebarOpen"
            class="fixed inset-0 bg-black bg-opacity-50 z-20 md:hidden"
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

        <div class="flex-1 flex flex-col overflow-hidden transition-all duration-300">
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
            <main class="flex-1 overflow-x-hidden overflow-y-auto p-4 md:p-8 bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
                <component
                    :is="currentTabComponent"
                    :t="t"
                    :all-locales="allLocales"
                    :current-user="user"
                    :current-locale="currentLocale"
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
import PushModules from './tabs/PushModules.vue';
import Schedule from './tabs/Schedule.vue';
import FrontendConfig from './tabs/FrontendConfig.vue';
import UserPushConfig from './tabs/UserPushConfig.vue';

const TAB_COMPONENTS = {
    overview: 'Overview',
    system_config: 'SystemConfig',
    user_manager: 'UserManager',
    scraper_modules: 'ScraperModules',
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
        return {
            currentTab: '',
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
                if (!list.some(item => item.id === this.currentTab)) {
                    this.currentTab = list[0].id;
                }
            }
        }
    },
    created() {
        this.checkScreenSize();
        window.addEventListener('resize', this.checkScreenSize);
    },
    unmounted() {
        window.removeEventListener('resize', this.checkScreenSize);
    },
    computed: {
        currentTabLabel() {
            const item = (this.menu || []).find(i => i.id === this.currentTab);
            return item ? this.t(item.label, this.currentTab) : this.t('common.dashboard');
        },
        currentTabComponent() {
            return TAB_COMPONENTS[this.currentTab] || 'Overview';
        }
    },
    methods: {
        updateTab(tab) {
            this.currentTab = tab;
            if (window.innerWidth < 768) {
                this.isSidebarOpen = false;
            }
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
