<template>
    <aside
        class="fixed md:static inset-y-0 left-0 z-30 bg-white dark:bg-slate-950 text-slate-700 dark:text-slate-200 flex flex-col border-r border-slate-200 dark:border-slate-800 transition-all duration-200 transform md:transform-none"
        :class="[
            isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
            isCollapsed ? 'w-16' : 'w-60'
        ]"
    >
        <div class="h-14 flex items-center justify-between px-3.5 border-b border-slate-200 dark:border-slate-800 overflow-hidden whitespace-nowrap">
            <div class="flex items-center gap-2.5 overflow-hidden flex-1" :class="{'justify-center': isCollapsed}">
                <div class="w-7 h-7 rounded-md bg-slate-900 dark:bg-blue-600 text-white flex items-center justify-center text-xs font-bold tracking-tight flex-shrink-0">
                    UD
                </div>
                <span v-if="!isCollapsed" class="text-sm font-bold tracking-tight text-slate-900 dark:text-white truncate">{{ serverName }}</span>
            </div>

            <button
                @click="$emit('toggle-collapse')"
                class="hidden md:flex items-center justify-center p-1 rounded-md text-slate-400 hover:text-slate-700 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus:outline-none ml-1"
            >
                <span class="material-icons text-[18px]">{{ isCollapsed ? 'chevron_right' : 'chevron_left' }}</span>
            </button>
        </div>

        <nav class="flex-1 overflow-y-auto py-3 px-2.5 space-y-0.5 overflow-x-hidden">
            <a
                v-for="item in menu"
                :key="item.id"
                :href="tabHref(item.id)"
                @click="onNavClick($event, item.id)"
                :class="[
                    'group flex items-center px-2.5 py-2 text-[13px] font-medium rounded-md transition-colors',
                    currentTab === item.id
                        ? 'bg-slate-900 text-white dark:bg-blue-600/15 dark:text-blue-400 dark:ring-1 dark:ring-inset dark:ring-blue-500/30'
                        : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900 hover:text-slate-900 dark:hover:text-slate-100',
                    isCollapsed ? 'justify-center px-0' : ''
                ]"
                :title="isCollapsed ? t(item.label) : ''"
            >
                <span
                    :class="[
                        'material-icons flex-shrink-0 text-[18px] leading-none',
                        currentTab === item.id
                            ? 'text-white dark:text-blue-400'
                            : 'text-slate-400 dark:text-slate-500 group-hover:text-slate-700 dark:group-hover:text-slate-300',
                        isCollapsed ? '' : 'mr-2.5'
                    ]"
                >
                    {{ item.icon }}
                </span>
                <span v-if="!isCollapsed" class="truncate">{{ t(item.label) }}</span>
            </a>
        </nav>

        <div class="p-2.5 border-t border-slate-200 dark:border-slate-800 space-y-1">
            <button
                @click="$emit('open-profile')"
                class="w-full flex items-center p-2 rounded-md hover:bg-slate-100 dark:hover:bg-slate-900 transition-colors text-left"
                :class="{'justify-center': isCollapsed}"
                :title="t('auth.profile.title')"
            >
                <div class="w-8 h-8 rounded-md bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 flex items-center justify-center text-xs font-bold flex-shrink-0">
                    {{ ((user.nickname || user.username || '?').charAt(0)).toUpperCase() }}
                </div>
                <div v-if="!isCollapsed" class="ml-2.5 overflow-hidden flex-1">
                    <p class="text-xs font-semibold text-slate-900 dark:text-white truncate">{{ user.nickname || user.username }}</p>
                    <p class="text-[11px] font-mono text-slate-400 truncate">{{ user.role_name }}</p>
                </div>
                <span v-if="!isCollapsed" class="material-icons text-slate-400 text-[16px]">tune</span>
            </button>

            <button
                @click="$emit('logout')"
                class="w-full flex items-center px-2.5 py-2 text-xs font-medium text-rose-600 dark:text-rose-400 rounded-md hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors"
                :class="{'justify-center': isCollapsed}"
                :title="isCollapsed ? t('common.logout') : ''"
            >
                <span class="material-icons text-[16px]" :class="{'mr-2': !isCollapsed}">logout</span>
                <span v-if="!isCollapsed">{{ t('common.logout') }}</span>
            </button>
        </div>
    </aside>
</template>

<script>
import { buildDashboardPath, defaultSubtabFor } from '../utils/router.js';

export default {
    props: {
        serverName: String,
        user: Object,
        menu: Array,
        currentTab: String,
        t: Function,
        isOpen: Boolean,
        isCollapsed: Boolean
    },
    emits: ['update:tab', 'logout', 'toggle-collapse', 'close-mobile', 'open-profile'],
    methods: {
        tabHref(tabId) {
            return buildDashboardPath(tabId, defaultSubtabFor(tabId));
        },
        onNavClick(event, tabId) {
            if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) {
                return;
            }
            event.preventDefault();
            this.$emit('update:tab', tabId);
        }
    }
};
</script>
