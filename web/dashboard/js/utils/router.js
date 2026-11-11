export const DASHBOARD_BASE = '/dashboard';

const SUBTAB_DEFAULTS = {
    system_config: 'system',
    user_manager: 'users',
    classification: 'engine'
};

const KNOWN_SUBTABS = {
    user_manager: ['users', 'roles'],
    classification: ['engine', 'profiles', 'queue']
};

export function tabHasDefaultSubtab(tab) {
    return Object.prototype.hasOwnProperty.call(SUBTAB_DEFAULTS, tab);
}

export function tabSupportsSubtab(tab) {
    return (
        tabHasDefaultSubtab(tab) ||
        tab === 'scraper_modules' ||
        tab === 'push_modules'
    );
}

export function defaultSubtabFor(tab) {
    return SUBTAB_DEFAULTS[tab] || '';
}

export function normalizeSubtab(tab, subtab) {
    const cleaned = String(subtab || '').trim();
    if (!tabSupportsSubtab(tab)) {
        return '';
    }
    if (tab === 'scraper_modules' || tab === 'push_modules') {
        return cleaned;
    }
    const known = KNOWN_SUBTABS[tab];
    if (known) {
        return known.includes(cleaned) ? cleaned : SUBTAB_DEFAULTS[tab];
    }
    return cleaned || SUBTAB_DEFAULTS[tab] || '';
}

export function parseDashboardPath(pathname = window.location.pathname) {
    const raw = String(pathname || '');
    let rel = '';
    if (raw === DASHBOARD_BASE || raw === `${DASHBOARD_BASE}/`) {
        rel = '';
    } else if (raw.startsWith(`${DASHBOARD_BASE}/`)) {
        rel = raw.slice(DASHBOARD_BASE.length + 1);
    } else {
        return { tab: '', subtab: '' };
    }

    const parts = rel
        .split('/')
        .map(seg => {
            try {
                return decodeURIComponent(seg).trim();
            } catch {
                return seg.trim();
            }
        })
        .filter(Boolean);

    const tab = parts[0] || '';
    const subtab = parts[1] ? normalizeSubtab(tab, parts[1]) : defaultSubtabFor(tab);
    return { tab, subtab };
}

export function buildDashboardPath(tab, subtab = '') {
    if (!tab) return DASHBOARD_BASE;
    const base = `${DASHBOARD_BASE}/${encodeURIComponent(tab)}`;
    const normalized = normalizeSubtab(tab, subtab);
    if (!normalized) return base;
    return `${base}/${encodeURIComponent(normalized)}`;
}

