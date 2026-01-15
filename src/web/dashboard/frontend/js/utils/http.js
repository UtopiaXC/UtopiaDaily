import axios from 'axios';

const http = axios.create();

// Request Interceptor
http.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    
    // Add Accept-Language header based on current locale
    const locale = localStorage.getItem('locale') || navigator.language;
    config.headers['Accept-Language'] = locale;

    return config;
});

// Response Interceptor
http.interceptors.response.use(response => response, error => {
    if (error.response && error.response.status === 401) {
        const url = error.config.url || '';
        const isAuthRequest = url.includes('/auth/login') || url.includes('/auth/captcha');
        
        if (!isAuthRequest) {
            localStorage.removeItem('token');
            // Optional: Emit an event or redirect, handled by App.vue watching token
            window.dispatchEvent(new Event('auth-logout'));
        }
    }
    return Promise.reject(error);
});

export default http;
