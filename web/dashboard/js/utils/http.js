import axios from 'axios';

const ACCESS_KEY = 'utd_access_token';
const REFRESH_KEY = 'utd_refresh_token';
const USER_KEY = 'utd_user';

export function getTokenStore() {
    if (localStorage.getItem(ACCESS_KEY) || localStorage.getItem(REFRESH_KEY)) {
        return localStorage;
    }
    if (sessionStorage.getItem(ACCESS_KEY) || sessionStorage.getItem(REFRESH_KEY)) {
        return sessionStorage;
    }
    return localStorage;
}

export function getAccessToken() {
    return localStorage.getItem(ACCESS_KEY) || sessionStorage.getItem(ACCESS_KEY) || '';
}

export function getRefreshToken() {
    return localStorage.getItem(REFRESH_KEY) || sessionStorage.getItem(REFRESH_KEY) || '';
}

export function saveAuthState(data, rememberMe = true) {
    clearAuthState();
    const store = rememberMe ? localStorage : sessionStorage;
    if (data.token) store.setItem(ACCESS_KEY, data.token);
    if (data.refresh_token) store.setItem(REFRESH_KEY, data.refresh_token);
    if (data.user) store.setItem(USER_KEY, JSON.stringify(data.user));
}

export function updateStoredUser(user) {
    const store = getTokenStore();
    if (user) {
        store.setItem(USER_KEY, JSON.stringify(user));
    }
}

export function getStoredUser() {
    const raw = localStorage.getItem(USER_KEY) || sessionStorage.getItem(USER_KEY);
    if (!raw) return null;
    try {
        return JSON.parse(raw);
    } catch {
        return null;
    }
}

export function clearAuthState() {
    [localStorage, sessionStorage].forEach(store => {
        store.removeItem(ACCESS_KEY);
        store.removeItem(REFRESH_KEY);
        store.removeItem(USER_KEY);
    });
}

const http = axios.create({
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json'
    }
});

http.interceptors.request.use(config => {
    const token = getAccessToken();
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

let refreshPromise = null;

async function performRefresh() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) {
        throw new Error('Missing refresh token');
    }
    const remember = !!localStorage.getItem(REFRESH_KEY);
    const res = await axios.post('/api/auth/refresh', { refresh_token: refreshToken });
    saveAuthState(res.data, remember);
    return res.data.token;
}

http.interceptors.response.use(
    response => response,
    async error => {
        const original = error.config;
        const status = error.response?.status;
        const url = original?.url || '';

        const isAuthPublic = url.includes('/api/common/') ||
            url.includes('/api/auth/login') ||
            url.includes('/api/auth/refresh') ||
            url.includes('/api/auth/register') ||
            url.includes('/api/auth/setup') ||
            url.includes('/api/auth/activate') ||
            url.includes('/api/auth/forgot-password') ||
            url.includes('/api/auth/reset-password') ||
            url.includes('/api/auth/captcha') ||
            url.includes('/api/auth/login-policy');

        if (status === 401 && original && !original._retry && !isAuthPublic) {
            original._retry = true;
            try {
                if (!refreshPromise) {
                    refreshPromise = performRefresh().finally(() => {
                        refreshPromise = null;
                    });
                }
                const newToken = await refreshPromise;
                original.headers.Authorization = `Bearer ${newToken}`;
                return http(original);
            } catch {
                clearAuthState();
                window.dispatchEvent(new CustomEvent('auth-logout'));
            }
        }
        return Promise.reject(error);
    }
);

export function formatAPIError(err, t) {
    const data = err?.response?.data;
    if (data && data.code) {
        let msg = t('error.' + data.code, data.code);
        if (data.params && typeof data.params === 'object') {
            if (data.params.detail) {
                msg = `${msg}: ${data.params.detail}`;
            } else {
                Object.keys(data.params).forEach(k => {
                    msg = msg.split(`{${k}}`).join(String(data.params[k]));
                });
            }
        }
        return msg;
    }
    return err?.message || t('error.OPERATION_FAILED');
}

export default http;
