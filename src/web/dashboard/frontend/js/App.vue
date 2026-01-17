<template>
    <div class="font-sans antialiased text-gray-800 dark:text-gray-100">
        <ToastContainer ref="toast" />

        <!-- Loading Screen -->
        <div v-if="isInitializing" class="fixed inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-900 z-50">
            <div class="flex flex-col items-center">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                <p class="mt-4 text-gray-500 dark:text-gray-400 text-sm">Loading Utopia Daily...</p>
            </div>
        </div>

        <template v-else>
            <Login
                v-if="!isLoggedIn"
                :server-name="serverName"
                :is-dark-mode="isDarkMode"
                :current-locale="currentLocale"
                :ui-locale-options="uiLocaleOptions"
                :t="t"
                @toggle-dark-mode="toggleDarkMode"
                @update:locale="changeLocale"
                @login-success="handleLoginSuccess"
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
            />
        </template>
    </div>
</template>

<script>
import ToastContainer from './components/Toast.vue';
import Login from './views/Login.vue';
import Dashboard from './views/Dashboard.vue';
import http from './utils/http.js';

export default {
    components: { ToastContainer, Login, Dashboard },
    provide() {
        return {
            showToast: this.showToast
        }
    },
    data() {
        return {
            isInitializing: true,
            serverName: 'Utopia Daily',
            isDarkMode: false,

            user: null,
            token: null,
            menu: [],

            currentLocale: 'en_US',
            allLocales: [],
            messages: {},
        }
    },
    computed: {
        isLoggedIn() {
            return !!this.token && !!this.user;
        },
        uiLocaleOptions() {
            return this.allLocales.filter(l => l.value !== 'auto');
        }
    },
    watch: {
        isDarkMode(newVal) {
            if (newVal) {
                document.documentElement.classList.add('dark');
                localStorage.setItem('theme', 'dark');
            } else {
                document.documentElement.classList.remove('dark');
                localStorage.setItem('theme', 'light');
            }
        }
    },
    async created() {
        try {
            // Auth Logout Listener
            window.addEventListener('auth-logout', this.logout);
            window.addEventListener('server-name-updated', (e) => {
                this.serverName = e.detail;
                this.updatePageTitle();
            });

            // Initialize Theme
            const storedTheme = localStorage.getItem('theme');
            if (storedTheme === 'dark' || (!storedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
                this.isDarkMode = true;
            }

            // Initialize Data
            await this.fetchSystemInfo();
            await this.fetchLocales();

            const storedLocale = localStorage.getItem('locale');
            if (storedLocale) {
                this.currentLocale = storedLocale;
            } else {
                this.currentLocale = this.detectBrowserLocale();
            }
            await this.loadMessages();

            const storedToken = localStorage.getItem('token');
            const storedUser = localStorage.getItem('user');

            if (storedToken) {
                this.token = storedToken;
                // Optimistically set user from storage to avoid flicker if valid
                if (storedUser) {
                    try {
                        this.user = JSON.parse(storedUser);
                    } catch (e) {
                        console.error("Invalid stored user data");
                    }
                }

                // Verify token and refresh user data
                try {
                    await this.fetchCurrentUser();
                    await this.fetchMenu();
                } catch (e) {
                    // Token invalid or expired
                    this.logout();
                }
            }
        } catch (e) {
            console.error("Initialization error", e);
        } finally {
            this.isInitializing = false;
            this.updatePageTitle();
        }
    },
    methods: {
        t(key, defaultVal) {
            return this.messages[key] || defaultVal || key;
        },
        showToast(msg, type) {
            this.$refs.toast.add(msg, type);
        },
        toggleDarkMode() {
            this.isDarkMode = !this.isDarkMode;
        },
        detectBrowserLocale() {
            const lang = navigator.language;
            if (lang.includes('zh')) return 'zh_CN';
            return 'en_US';
        },
        updatePageTitle() {
            document.title = this.serverName;
        },
        async fetchSystemInfo() {
            try {
                const res = await http.get('/api/common/system-info');
                if (res.data.server_name) {
                    this.serverName = res.data.server_name;
                }
            } catch (err) {
                console.error("Failed to fetch system info", err);
            }
        },
        async fetchLocales() {
            try {
                const res = await http.get('/api/common/locales');
                this.allLocales = res.data;
            } catch (err) {
                console.error("Failed to fetch locales", err);
                this.allLocales = [
                    { label: 'Auto', value: 'auto' },
                    { label: 'English', value: 'en_US' },
                    { label: '中文', value: 'zh_CN' }
                ];
            }
        },
        async loadMessages() {
            let targetLocale = this.currentLocale;
            if (targetLocale === 'auto') {
                targetLocale = this.detectBrowserLocale();
            }
            try {
                const res = await http.get(`/api/common/i18n/${targetLocale}`);
                this.messages = res.data;
            } catch (err) {
                console.error("Failed to load i18n messages", err);
            }
        },
        async changeLocale(newVal) {
            this.currentLocale = newVal;
            localStorage.setItem('locale', this.currentLocale);
            await this.loadMessages();
        },
        async fetchMenu() {
            try {
                const res = await http.get('/api/dashboard/layout/menu');
                this.menu = res.data.menu;
            } catch (err) {
                console.error("Failed to fetch menu", err);
            }
        },
        async fetchCurrentUser() {
            try {
                const res = await http.get('/api/dashboard/auth/me');
                this.user = res.data;
                localStorage.setItem('user', JSON.stringify(this.user));
            } catch (err) {
                console.error("Failed to fetch current user", err);
                throw err; // Re-throw to handle in created()
            }
        },
        handleLoginSuccess(data) {
            this.token = data.token;
            this.user = data.user;
            localStorage.setItem('token', this.token);
            localStorage.setItem('user', JSON.stringify(this.user));
            this.fetchMenu();
        },
        logout() {
            this.token = null;
            this.user = null;
            this.menu = [];
            localStorage.removeItem('token');
            localStorage.removeItem('user');
        }
    }
}
</script>
