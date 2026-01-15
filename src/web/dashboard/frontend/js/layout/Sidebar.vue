<template>
    <aside class="w-64 bg-gray-900 dark:bg-gray-950 text-white flex flex-col shadow-2xl z-20 border-r border-gray-800 dark:border-gray-900">
        <div class="h-16 flex items-center px-6 border-b border-gray-800 dark:border-gray-900 bg-gray-900 dark:bg-gray-950">
            <span class="text-lg font-bold tracking-wider text-white">{{ serverName }}</span>
        </div>

        <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
            <a href="#" @click.prevent="$emit('update:tab', 'dashboard')"
               :class="['group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors', currentTab === 'dashboard' ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white']">
                <span :class="['material-icons mr-3 flex-shrink-0 h-6 w-6 text-lg', currentTab === 'dashboard' ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-300']">dashboard</span>
                {{ t('menu.overview') }}
            </a>

            <a v-for="item in menu" :key="item.id" href="#" @click.prevent="$emit('update:tab', item.id)"
               :class="['group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors', currentTab === item.id ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white']">
                <span :class="['material-icons mr-3 flex-shrink-0 h-6 w-6 text-lg', currentTab === item.id ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-300']">{{ item.icon }}</span>
                {{ t('menu.' + item.id, item.label) }}
            </a>
        </nav>

        <div class="p-4 border-t border-gray-800 dark:border-gray-900 bg-gray-900 dark:bg-gray-950">
            <div class="flex items-center mb-4">
                <div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-sm font-bold shadow-md ring-2 ring-gray-800">
                    {{ user.username.charAt(0).toUpperCase() }}
                </div>
                <div class="ml-3 overflow-hidden">
                    <p class="text-sm font-medium text-white truncate">{{ user.username }}</p>
                    <p class="text-xs text-gray-400 truncate">{{ user.role_name }}</p>
                </div>
            </div>
            <button @click="$emit('logout')" class="w-full flex items-center px-2 py-2 text-sm font-medium text-red-400 rounded-md hover:bg-gray-800 hover:text-red-300 transition-colors">
                <span class="material-icons text-sm mr-2">logout</span> {{ t('common.logout') }}
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
        t: Function
    },
    emits: ['update:tab', 'logout']
}
</script>
