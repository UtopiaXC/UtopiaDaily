<template>
    <div class="font-sans antialiased text-gray-800 dark:text-gray-100">
        <ToastContainer ref="toast" />
        <div v-if="isInitializing" class="fixed inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-900 z-50">
            <div class="flex flex-col items-center">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                <p class="mt-4 text-gray-500 dark:text-gray-400 text-sm">{{ t('common.loading') }}</p>
            </div>
        </div>

        <template v-else>
            <Login
                v-if="!isLoggedIn"
                :server-name="serverName"
                :system-info="systemInfo"
                :is-dark-mode="isDarkMode"
                :current-locale="currentLocale"
                :ui-locale-options="uiLocaleOptions"
                :t="t"
                @toggle-dark-mode="toggleDarkMode"
                @update:locale="changeLocale"
                @login-success="handleLoginSuccess"
                @setup-completed="fetchSystemInfo"
            />

            <Dashboard
                v-else
                :server-name="serverName"
                :user="user"
                :menu="menu"
                :is-dark-mode="isDarkMode"
                :current-locale="currentLocale"
                :ui-locale-options="uiLocaleOptions"
                :all-locales="allLocales"
                :t="t"
                @logout="logout"
                @toggle-dark-mode="toggleDarkMode"
                @update:locale="changeLocale"
                @user-updated="handleUserUpdated"
            />
        </template>
    </div>
</template>

<script>
import ToastContainer from './components/Toast.vue';
import Login from './views/Login.vue';
import Dashboard from './views/Dashboard.vue';
import http, {
    getAccessToken,
    getRefreshToken,
    getStoredUser,
    saveAuthState,
    updateStoredUser,
    clearAuthState
} from './utils/http.js';

export default {
    components: { ToastContainer, Login, Dashboard },
    provide() {
        return {
            showToast: this.showToast
        };
    },
    data() {
        return {
            isInitializing: true,
            serverName: 'Utopia Daily',
            systemInfo: {
                name: 'Utopia Daily',
                version: '2.0.0',
                default_locale: 'en_US',
                allow_registration: false,
                setup_required: false
            },
            isDarkMode: false,
            systemMediaQuery: null,

            user: null,
            hasSession: false,
            menu: [],

            currentLocale: 'en_US',
            allLocales: [],
            messages: {}
        };
    },
    computed: {
        isLoggedIn() {
            return this.hasSession && !!this.user;
        },
        uiLocaleOptions() {
            return this.allLocales.filter(l => l.value !== 'auto');
        }
    },
    watch: {
        isDarkMode(newVal) {
            if (newVal) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        }
    },
    async created() {
        try {
            window.addEventListener('auth-logout', this.handleForceLogout);
            window.addEventListener('server-name-updated', this.onServerNameUpdated);

            this.initTheme();
            await Promise.all([this.fetchSystemInfo(), this.fetchLocales()]);

            const storedLocale = localStorage.getItem('locale');
            if (storedLocale && storedLocale !== 'auto') {
                this.currentLocale = storedLocale;
            } else {
                this.currentLocale = this.detectBrowserLocale();
            }
            await this.loadMessages();

            const token = getAccessToken();
            const storedUser = getStoredUser();
            if (token) {
                this.hasSession = true;
                if (storedUser) {
                    this.user = storedUser;
                }
                try {
                    await this.fetchCurrentUser();
                    await this.fetchMenu();
                } catch {
                    clearAuthState();
                    this.hasSession = false;
                    this.user = null;
                }
            }
        } finally {
            this.isInitializing = false;
            document.title = this.serverName;
        }
    },
    beforeUnmount() {
        if (this.systemMediaQuery) {
            this.systemMediaQuery.removeEventListener('change', this.onSystemThemeChange);
        }
        window.removeEventListener('auth-logout', this.handleForceLogout);
        window.removeEventListener('server-name-updated', this.onServerNameUpdated);
    },
    methods: {
        initTheme() {
            this.systemMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            this.systemMediaQuery.addEventListener('change', this.onSystemThemeChange);

            const systemIsDark = this.systemMediaQuery.matches;
            const systemThemeStr = systemIsDark ? 'dark' : 'light';
            const storedTheme = localStorage.getItem('theme');
            const storedSnapshot = localStorage.getItem('theme_snapshot');

            if (storedTheme) {
                if (storedSnapshot && storedSnapshot !== systemThemeStr) {
                    this.isDarkMode = systemIsDark;
                    localStorage.removeItem('theme');
                    localStorage.removeItem('theme_snapshot');
                } else {
                    this.isDarkMode = storedTheme === 'dark';
                }
            } else {
                this.isDarkMode = systemIsDark;
            }

            if (this.isDarkMode) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        },
        onSystemThemeChange(e) {
            this.isDarkMode = e.matches;
            localStorage.removeItem('theme');
            localStorage.removeItem('theme_snapshot');
        },
        onServerNameUpdated(e) {
            this.serverName = e.detail || 'Utopia Daily';
            document.title = this.serverName;
        },
        t(key, defaultVal) {
            return this.messages[key] || defaultVal || key;
        },
        showToast(msg, type = 'success') {
            if (this.$refs.toast) {
                this.$refs.toast.add(msg, type);
            }
        },
        toggleDarkMode() {
            this.isDarkMode = !this.isDarkMode;
            localStorage.setItem('theme', this.isDarkMode ? 'dark' : 'light');
            if (this.systemMediaQuery) {
                localStorage.setItem('theme_snapshot', this.systemMediaQuery.matches ? 'dark' : 'light');
            }
        },
        detectBrowserLocale() {
            const lang = (navigator.language || '').toLowerCase();
            if (lang.includes('zh')) return 'zh_CN';
            return this.systemInfo.default_locale || 'en_US';
        },
        async fetchSystemInfo() {
            try {
                const res = await http.get('/api/common/system-info');
                this.systemInfo = res.data;
                if (res.data.name) {
                    this.serverName = res.data.name;
                    document.title = this.serverName;
                }
            } catch {
                // fallback to defaults
            }
        },
        async fetchLocales() {
            try {
                const res = await http.get('/api/common/locales');
                const rawLangs = Array.isArray(res.data) ? res.data : (res.data?.languages || []);
                if (res.data?.default_locale && !this.systemInfo.default_locale) {
                    this.systemInfo.default_locale = res.data.default_locale;
                }
                const mapped = rawLangs.map(item => ({
                    label: item.name,
                    value: item.code
                }));
                if (mapped.length > 0) {
                    this.allLocales = [{ label: 'Auto', value: 'auto' }, ...mapped];
                    return;
                }
            } catch {
                // fallback below
            }
            this.allLocales = [
                { label: 'Auto', value: 'auto' },
                { label: 'English', value: 'en_US' },
                { label: '中文', value: 'zh_CN' }
            ];
        },
        async loadMessages() {
            let targetLocale = this.currentLocale;
            if (targetLocale === 'auto') {
                targetLocale = this.detectBrowserLocale();
            }
            document.documentElement.lang = targetLocale.replace('_', '-');
            try {
                const res = await http.get(`/api/common/i18n/${encodeURIComponent(targetLocale)}`);
                this.messages = res.data || {};
            } catch {
                // ignore offline failure
            }
        },
        async changeLocale(newVal) {
            this.currentLocale = newVal;
            localStorage.setItem('locale', newVal);
            await this.loadMessages();
        },
        async fetchMenu() {
            const res = await http.get('/api/dashboard/layout/menu');
            this.menu = res.data.menu || [];
        },
        async fetchCurrentUser() {
            const res = await http.get('/api/auth/me');
            this.user = res.data;
            updateStoredUser(res.data);
        },
        async handleLoginSuccess({ auth, rememberMe }) {
            saveAuthState(auth, rememberMe);
            this.user = auth.user;
            this.hasSession = true;
            await this.fetchMenu();
        },
        handleUserUpdated(updated) {
            this.user = updated;
            updateStoredUser(updated);
        },
        handleForceLogout() {
            clearAuthState();
            this.hasSession = false;
            this.user = null;
            this.menu = [];
        },
        async logout() {
            const refresh = getRefreshToken();
            try {
                await http.post('/api/auth/logout', { refresh_token: refresh });
            } catch {
                // ignore network failure during logout
            }
            this.handleForceLogout();
        }
    }
}
</script>
