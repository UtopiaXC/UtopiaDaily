<template>
    <aside
        class="fixed md:static inset-y-0 left-0 z-30 bg-gray-900 dark:bg-gray-950 text-white flex flex-col shadow-2xl border-r border-gray-800 dark:border-gray-900 transition-all duration-300 transform md:transform-none"
        :class="[
            isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
            isCollapsed ? 'w-20' : 'w-64'
        ]"
    >
        <!-- Header -->
        <div class="h-16 flex items-center justify-between px-4 border-b border-gray-800 dark:border-gray-900 bg-gray-900 dark:bg-gray-950 overflow-hidden whitespace-nowrap">
            <div class="flex items-center overflow-hidden flex-1" :class="{'justify-center': isCollapsed}">
                <span v-if="!isCollapsed" class="text-lg font-bold tracking-wider text-white truncate">{{ serverName }}</span>
                <span v-else class="text-lg font-bold tracking-wider text-white">UD</span>
            </div>

            <button
                @click="$emit('toggle-collapse')"
                class="hidden md:flex items-center justify-center p-1 rounded-md text-gray-400 hover:text-white hover:bg-gray-800 transition-colors focus:outline-none ml-2"
            >
                <span class="material-icons">{{ isCollapsed ? 'menu' : 'menu_open' }}</span>
            </button>
        </div>

        <nav class="flex-1 overflow-y-auto py-4 px-2 space-y-1 overflow-x-hidden">
            <a href="#" @click.prevent="$emit('update:tab', 'dashboard')"
               :class="['group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors', currentTab === 'dashboard' ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white', isCollapsed ? 'justify-center' : '']"
               :title="isCollapsed ? t('menu.overview') : ''">
                <span :class="['material-icons flex-shrink-0 h-6 w-6 text-lg', currentTab === 'dashboard' ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-300', isCollapsed ? '' : 'mr-3']">dashboard</span>
                <span v-if="!isCollapsed" class="truncate">{{ t('menu.overview') }}</span>
            </a>

            <a v-for="item in menu" :key="item.id" href="#" @click.prevent="$emit('update:tab', item.id)"
               :class="['group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors', currentTab === item.id ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white', isCollapsed ? 'justify-center' : '']"
               :title="isCollapsed ? t('menu.' + item.id, item.label) : ''">
                <span :class="['material-icons flex-shrink-0 h-6 w-6 text-lg', currentTab === item.id ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-300', isCollapsed ? '' : 'mr-3']">{{ item.icon }}</span>
                <span v-if="!isCollapsed" class="truncate">{{ t('menu.' + item.id, item.label) }}</span>
            </a>
        </nav>

        <div class="p-4 border-t border-gray-800 dark:border-gray-900 bg-gray-900 dark:bg-gray-950">
            <div class="flex items-center mb-4" :class="{'justify-center': isCollapsed}">
                <div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-sm font-bold shadow-md ring-2 ring-gray-800 flex-shrink-0">
                    {{ (user.nickname || user.username).charAt(0).toUpperCase() }}
                </div>
                <div v-if="!isCollapsed" class="ml-3 overflow-hidden">
                    <p class="text-sm font-medium text-white truncate">{{ user.nickname || user.username }}</p>
                    <p class="text-xs text-gray-400 truncate">{{ user.role_name }}</p>
                </div>
            </div>

            <button @click="$emit('logout')" class="w-full flex items-center px-2 py-2 text-sm font-medium text-red-400 rounded-md hover:bg-gray-800 hover:text-red-300 transition-colors" :class="{'justify-center': isCollapsed}" :title="isCollapsed ? t('common.logout') : ''">
                <span class="material-icons text-sm" :class="{'mr-2': !isCollapsed}">logout</span>
                <span v-if="!isCollapsed">{{ t('common.logout') }}</span>
            </button>
        </div>
    </aside>
</template>

<script>
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
    emits: ['update:tab', 'logout', 'toggle-collapse', 'close-mobile']
}
</script>
