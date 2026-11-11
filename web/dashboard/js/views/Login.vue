<template>
    <div class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950 transition-colors duration-150 p-4">
        <div
            class="bg-white dark:bg-slate-900 p-7 rounded-lg w-full relative border border-slate-200 dark:border-slate-800 transition-colors duration-150"
            :class="mode === 'setup' ? 'max-w-lg' : 'max-w-md'"
        >
            <div class="absolute top-4 right-4 flex items-center gap-1.5">
                <button @click="$emit('toggle-dark-mode')" type="button" class="p-1.5 rounded-md border border-slate-200 dark:border-slate-800 text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition focus:outline-none">
                    <span class="material-icons text-[18px] leading-none block">{{ isDarkMode ? 'light_mode' : 'dark_mode' }}</span>
                </button>
                <CustomSelect
                    :model-value="currentLocale"
                    @update:model-value="$emit('update:locale', $event)"
                    :options="uiLocaleOptions"
                    variant="icon"
                />
            </div>

            <div class="mb-6 mt-1">
                <div class="inline-flex items-center justify-center w-9 h-9 rounded-md bg-slate-900 dark:bg-blue-600 text-white text-xs font-bold mb-3">UD</div>
                <h2 class="text-xl font-bold text-slate-900 dark:text-white tracking-tight">{{ mode === 'setup' ? t('setup.title') : serverName }}</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">{{ modeSubtitle }}</p>
            </div>

            <div v-if="error" class="bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-400 px-3 py-2.5 rounded-md mb-4 text-xs flex items-start">
                <span class="material-icons text-[16px] mr-2 mt-0.5">error_outline</span>
                <span>{{ error }}</span>
            </div>

            <div v-if="notice" class="bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900 text-emerald-700 dark:text-emerald-400 px-3 py-2.5 rounded-md mb-4 text-xs flex items-start">
                <span class="material-icons text-[16px] mr-2 mt-0.5">check_circle</span>
                <span>{{ notice }}</span>
            </div>

            <div v-if="mode === 'activate'" class="text-center py-4 space-y-4">
                <div v-if="loading" class="flex flex-col items-center gap-2 text-gray-500 dark:text-gray-400">
                    <span class="material-icons animate-spin text-3xl text-blue-500">sync</span>
                    <span class="text-sm">{{ t('auth.activate.verifying') }}</span>
                </div>
                <button @click="switchMode('login')" class="w-full py-2.5 px-4 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition">
                    {{ t('auth.login.back_to_login') }}
                </button>
            </div>

            <form v-else-if="mode === 'setup'" @submit.prevent="handleSetupSubmit" class="space-y-5">
                <div class="grid grid-cols-3 gap-2 text-xs font-medium">
                    <button
                        v-for="(label, idx) in setupSteps"
                        :key="idx"
                        type="button"
                        @click="goToSetupStep(idx + 1)"
                        :class="[
                            'py-2 px-2 rounded-lg border text-center transition-colors truncate',
                            setupStep === idx + 1
                                ? 'bg-blue-50 dark:bg-blue-900/30 border-blue-500 text-blue-700 dark:text-blue-300'
                                : setupStep > idx + 1
                                    ? 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-400 text-emerald-700 dark:text-emerald-300'
                                    : 'bg-gray-50 dark:bg-gray-700/40 border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400'
                        ]"
                    >
                        {{ idx + 1 }}. {{ label }}
                    </button>
                </div>

                <div v-if="setupStep === 1" class="space-y-4">
                    <div>
                        <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.site_name') }}</label>
                        <input v-model="setupForm.site_name" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                    <div>
                        <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.site_desc') }}</label>
                        <input v-model="setupForm.site_description" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                    <div>
                        <div class="flex items-center gap-1.5 mb-1.5">
                            <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('setup.base_url') }}</label>
                            <div class="relative group flex items-center">
                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-60 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t('setup.base_url_hint') }}
                                </div>
                            </div>
                        </div>
                        <input v-model="setupForm.base_url" type="url" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                    <div>
                        <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.default_locale') }}</label>
                        <CustomSelect
                            v-model="setupForm.default_locale"
                            @update:model-value="$emit('update:locale', $event)"
                            :options="uiLocaleOptions"
                        />
                    </div>
                </div>

                <div v-else-if="setupStep === 2" class="space-y-4">
                    <label class="flex items-center gap-2 cursor-pointer py-1">
                        <input v-model="setupForm.smtp_enabled" type="checkbox" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded">
                        <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t('setup.smtp_enable') }}</span>
                    </label>

                    <div v-if="setupForm.smtp_enabled" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.host.desc') }}</label>
                            <input v-model="setupForm.smtp_host" type="text" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.port.desc') }}</label>
                            <input v-model.number="setupForm.smtp_port" type="number" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.encryption.desc') }}</label>
                            <CustomSelect v-model="setupForm.smtp_encryption" :options="smtpEncryptionOptions" />
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.user.desc') }}</label>
                            <input v-model="setupForm.smtp_user" type="text" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.password.desc') }}</label>
                            <input v-model="setupForm.smtp_password" type="password" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.from.desc') }}</label>
                            <input v-model="setupForm.smtp_from" type="email" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                        <div class="sm:col-span-2">
                            <label class="block text-gray-700 dark:text-gray-300 text-xs font-medium mb-1">{{ t('config.smtp.from_name.desc') }}</label>
                            <input v-model="setupForm.smtp_from_name" type="text" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                        </div>
                    </div>
                </div>

                <div v-else class="space-y-4">
                    <div>
                        <div class="flex items-center gap-1.5 mb-1.5">
                            <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('setup.admin_username') }}</label>
                            <div class="relative group flex items-center">
                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-56 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t('setup.admin_username_hint') }}
                                </div>
                            </div>
                        </div>
                        <input v-model="setupForm.username" type="text" required class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.admin_nickname') }}</label>
                            <input v-model="setupForm.nickname" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                        </div>
                        <div>
                            <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.admin_email') }}</label>
                            <input v-model="setupForm.email" type="email" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                        </div>
                    </div>
                    <div>
                        <div class="flex items-center gap-1.5 mb-1.5">
                            <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('setup.admin_password') }}</label>
                            <div class="relative group flex items-center">
                                <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-48 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t('auth.register.password_hint') }}
                                </div>
                            </div>
                        </div>
                        <input v-model="setupForm.password" type="password" required class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                    <div>
                        <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('setup.confirm_password') }}</label>
                        <input v-model="setupForm.confirmPassword" type="password" required class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:ring-blue-500/20 focus:border-blue-500">
                    </div>
                </div>

                <div class="flex items-center justify-between gap-3 pt-2">
                    <button
                        v-if="setupStep > 1"
                        type="button"
                        @click="setupStep--"
                        class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                    >
                        {{ t('setup.prev') }}
                    </button>
                    <span v-else></span>

                    <button
                        v-if="setupStep < 3"
                        type="button"
                        @click="setupStep++"
                        class="px-5 py-2 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition"
                    >
                        {{ t('setup.next') }}
                    </button>
                    <button
                        v-else
                        :disabled="loading"
                        type="submit"
                        class="flex items-center justify-center py-2 px-5 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition"
                    >
                        <span v-if="loading" class="material-icons animate-spin text-sm mr-2">sync</span>
                        {{ t('setup.finish') }}
                    </button>
                </div>
            </form>

            <form v-else-if="mode === 'register'" @submit.prevent="submitRegister" class="space-y-4">
                <div>
                    <div class="flex items-center gap-1.5 mb-1.5">
                        <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('auth.register.username') }}</label>
                        <div class="relative group flex items-center">
                            <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-56 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                {{ t('auth.register.username_hint') }}
                            </div>
                        </div>
                    </div>
                    <input v-model="registerForm.username" type="text" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.register.nickname') }}</label>
                    <input v-model="registerForm.nickname" type="text" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.register.email') }}</label>
                    <input v-model="registerForm.email" type="email" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <div class="flex items-center gap-1.5 mb-1.5">
                        <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('auth.register.password') }}</label>
                        <div class="relative group flex items-center">
                            <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-48 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                {{ t('auth.register.password_hint') }}
                            </div>
                        </div>
                    </div>
                    <input v-model="registerForm.password" type="password" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.register.confirm_password') }}</label>
                    <input v-model="registerForm.confirmPassword" type="password" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.pow.label') }}</label>
                    <PowWidget ref="powWidget" v-model="powPayload" :t="t" />
                </div>
                <button :disabled="loading" type="submit" class="w-full flex justify-center py-2.5 px-4 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition">
                    <span v-if="loading" class="material-icons animate-spin text-sm mr-2">sync</span>
                    {{ t('auth.register.submit') }}
                </button>
                <div class="text-center pt-1">
                    <button type="button" @click="switchMode('login')" class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                        {{ t('auth.login.back_to_login') }}
                    </button>
                </div>
            </form>

            <form v-else-if="mode === 'forgot'" @submit.prevent="submitForgot" class="space-y-4">
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.forgot.email') }}</label>
                    <input v-model="forgotEmail" type="email" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.pow.label') }}</label>
                    <PowWidget ref="powWidget" v-model="powPayload" :t="t" />
                </div>
                <button :disabled="loading" type="submit" class="w-full flex justify-center py-2.5 px-4 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition">
                    <span v-if="loading" class="material-icons animate-spin text-sm mr-2">sync</span>
                    {{ t('auth.forgot.submit') }}
                </button>
                <div class="text-center pt-1">
                    <button type="button" @click="switchMode('login')" class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                        {{ t('auth.login.back_to_login') }}
                    </button>
                </div>
            </form>

            <form v-else-if="mode === 'reset'" @submit.prevent="submitReset" class="space-y-4">
                <div>
                    <div class="flex items-center gap-1.5 mb-1.5">
                        <label class="text-gray-700 dark:text-gray-300 text-sm font-medium">{{ t('auth.reset.new_password') }}</label>
                        <div class="relative group flex items-center">
                            <span class="material-icons text-gray-400 text-sm cursor-help hover:text-blue-500">help_outline</span>
                            <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-48 p-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                {{ t('auth.register.password_hint') }}
                            </div>
                        </div>
                    </div>
                    <input v-model="resetForm.password" type="password" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.reset.confirm_password') }}</label>
                    <input v-model="resetForm.confirmPassword" type="password" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md text-sm bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                </div>
                <button :disabled="loading" type="submit" class="w-full flex justify-center py-2.5 px-4 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition">
                    <span v-if="loading" class="material-icons animate-spin text-sm mr-2">sync</span>
                    {{ t('auth.reset.submit') }}
                </button>
                <div class="text-center pt-1">
                    <button type="button" @click="switchMode('login')" class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                        {{ t('auth.login.back_to_login') }}
                    </button>
                </div>
            </form>

            <form v-else @submit.prevent="submitLogin" class="space-y-4">
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.login.account') }}</label>
                    <div class="relative">
                        <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <span class="material-icons text-gray-400 text-sm">person</span>
                        </span>
                        <input v-model="loginForm.account" type="text" class="pl-9 block w-full border border-slate-300 dark:border-slate-700 rounded-md focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-sm py-2 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                    </div>
                </div>
                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.login.password') }}</label>
                    <div class="relative">
                        <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <span class="material-icons text-gray-400 text-sm">lock</span>
                        </span>
                        <input v-model="loginForm.password" type="password" class="pl-9 block w-full border border-slate-300 dark:border-slate-700 rounded-md focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-sm py-2 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100">
                    </div>
                </div>

                <div>
                    <label class="block text-gray-700 dark:text-gray-300 text-sm font-medium mb-1.5">{{ t('auth.pow.label') }}</label>
                    <PowWidget ref="powWidget" v-model="powPayload" :t="t" />
                </div>

                <div class="flex items-center justify-between pt-1">
                    <label class="flex items-center cursor-pointer">
                        <input id="remember_me" v-model="loginForm.rememberMe" type="checkbox" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700">
                        <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">{{ t('auth.login.remember_me') }}</span>
                    </label>
                    <button type="button" @click="switchMode('forgot')" class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                        {{ t('auth.login.forgot_link') }}
                    </button>
                </div>

                <button :disabled="loading" type="submit" class="w-full ui-btn-primary py-2.5">
                    <span v-if="loading" class="material-icons animate-spin text-[16px]">sync</span>
                    {{ loading ? t('auth.login.signing_in') : t('auth.login.submit') }}
                </button>

                <div v-if="systemInfo.allow_registration" class="text-center pt-1">
                    <button type="button" @click="switchMode('register')" class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                        {{ t('auth.login.register_link') }}
                    </button>
                </div>
            </form>
        </div>
    </div>
</template>

<script>
import CustomSelect from '../components/CustomSelect.vue';
import PowWidget from '../components/PowWidget.vue';
import http, { formatAPIError } from '../utils/http.js';

export default {
    components: { CustomSelect, PowWidget },
    props: {
        serverName: String,
        systemInfo: {
            type: Object,
            default: () => ({ allow_registration: false, setup_required: false })
        },
        isDarkMode: Boolean,
        currentLocale: String,
        uiLocaleOptions: Array,
        t: Function
    },
    emits: ['toggle-dark-mode', 'update:locale', 'login-success', 'setup-completed'],
    data() {
        return {
            mode: 'login',
            setupStep: 1,
            loading: false,
            error: null,
            notice: null,
            powPayload: '',
            actionToken: '',

            loginForm: {
                account: '',
                password: '',
                rememberMe: true
            },
            setupForm: {
                site_name: 'Utopia Daily',
                site_description: 'Modular news aggregation platform',
                base_url: typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8000',
                default_locale: 'en_US',
                smtp_enabled: false,
                smtp_host: '',
                smtp_port: 465,
                smtp_encryption: 'ssl',
                smtp_user: '',
                smtp_password: '',
                smtp_from: '',
                smtp_from_name: 'Utopia Daily',
                username: '',
                nickname: '',
                email: '',
                password: '',
                confirmPassword: ''
            },
            registerForm: {
                username: '',
                nickname: '',
                email: '',
                password: '',
                confirmPassword: ''
            },
            forgotEmail: '',
            resetForm: {
                password: '',
                confirmPassword: ''
            }
        };
    },
    computed: {
        setupSteps() {
            return [
                this.t('setup.step_site'),
                this.t('setup.step_smtp'),
                this.t('setup.step_admin')
            ];
        },
        smtpEncryptionOptions() {
            return [
                { value: 'ssl', label: 'SSL/TLS' },
                { value: 'tls', label: 'STARTTLS' },
                { value: 'none', label: 'None' }
            ];
        },
        modeSubtitle() {
            switch (this.mode) {
                case 'setup': return this.t('setup.subtitle');
                case 'register': return this.t('auth.register.title');
                case 'forgot': return this.t('auth.forgot.title');
                case 'reset': return this.t('auth.reset.title');
                case 'activate': return this.t('auth.activate.title');
                default: return this.t('common.dashboard');
            }
        }
    },
    watch: {
        currentLocale: {
            immediate: true,
            handler(val) {
                if (val && val !== 'auto' && this.setupStep === 1) {
                    this.setupForm.default_locale = val;
                }
            }
        },
        'systemInfo.setup_required': {
            immediate: true,
            handler(val) {
                if (val && this.mode === 'login') {
                    this.mode = 'setup';
                }
            }
        }
    },
    async mounted() {
        const params = new URLSearchParams(window.location.search);
        const queryMode = params.get('mode');
        const token = params.get('token');
        if (queryMode === 'activate' && token) {
            this.mode = 'activate';
            this.actionToken = token;
            await this.runActivation();
        } else if (queryMode === 'reset' && token) {
            this.mode = 'reset';
            this.actionToken = token;
        } else if (this.systemInfo.setup_required) {
            this.mode = 'setup';
        }
    },
    methods: {
        switchMode(target) {
            this.error = null;
            this.notice = null;
            this.mode = target;
            this.powPayload = '';
            if (window.history && window.location.search) {
                const params = new URLSearchParams(window.location.search);
                params.delete('mode');
                params.delete('token');
                const remaining = params.toString();
                const nextUrl = remaining ? `${window.location.pathname}?${remaining}` : window.location.pathname;
                window.history.replaceState({}, '', nextUrl);
            }
        },
        goToSetupStep(step) {
            this.error = null;
            this.setupStep = step;
        },
        handleSetupSubmit() {
            if (this.setupStep < 3) {
                this.setupStep++;
                return;
            }
            this.submitSetup();
        },
        async ensurePow() {
            if (this.powPayload) return true;
            if (this.$refs.powWidget) {
                return await this.$refs.powWidget.verify();
            }
            return false;
        },
        resetPow() {
            if (this.$refs.powWidget) {
                this.$refs.powWidget.reset();
            }
            this.powPayload = '';
        },
        async runActivation() {
            this.loading = true;
            this.error = null;
            try {
                await http.post('/api/auth/activate', { token: this.actionToken });
                this.notice = this.t('auth.activate.success');
            } catch (err) {
                this.error = formatAPIError(err, this.t);
            } finally {
                this.loading = false;
            }
        },
        async submitSetup() {
            if (this.setupForm.password !== this.setupForm.confirmPassword) {
                this.error = this.t('error.AUTH_PASSWORD_MISMATCH');
                return;
            }
            this.loading = true;
            this.error = null;
            try {
                const res = await http.post('/api/auth/setup', {
                    site_name: this.setupForm.site_name,
                    site_description: this.setupForm.site_description,
                    base_url: this.setupForm.base_url,
                    default_locale: this.setupForm.default_locale,
                    smtp_enabled: Boolean(this.setupForm.smtp_enabled),
                    smtp_host: this.setupForm.smtp_host,
                    smtp_port: Number(this.setupForm.smtp_port) || 465,
                    smtp_encryption: this.setupForm.smtp_encryption,
                    smtp_user: this.setupForm.smtp_user,
                    smtp_password: this.setupForm.smtp_password,
                    smtp_from: this.setupForm.smtp_from,
                    smtp_from_name: this.setupForm.smtp_from_name,
                    username: this.setupForm.username,
                    nickname: this.setupForm.nickname,
                    email: this.setupForm.email,
                    password: this.setupForm.password
                });
                this.$emit('setup-completed');
                this.$emit('login-success', { auth: res.data, rememberMe: true });
            } catch (err) {
                this.error = formatAPIError(err, this.t);
            } finally {
                this.loading = false;
            }
        },
        async submitLogin() {
            if (!this.loginForm.account || !this.loginForm.password) {
                this.error = this.t('auth.login.input_required');
                return;
            }
            this.loading = true;
            this.error = null;
            this.notice = null;

            const verified = await this.ensurePow();
            if (!verified || !this.powPayload) {
                this.error = this.t('auth.pow.required');
                this.loading = false;
                return;
            }

            try {
                const res = await http.post('/api/auth/login', {
                    account: this.loginForm.account,
                    password: this.loginForm.password,
                    pow_payload: this.powPayload,
                    remember_me: this.loginForm.rememberMe
                });
                this.$emit('login-success', { auth: res.data, rememberMe: this.loginForm.rememberMe });
            } catch (err) {
                this.error = formatAPIError(err, this.t);
                this.resetPow();
            } finally {
                this.loading = false;
            }
        },
        async submitRegister() {
            if (this.registerForm.password !== this.registerForm.confirmPassword) {
                this.error = this.t('error.AUTH_PASSWORD_MISMATCH');
                return;
            }
            this.loading = true;
            this.error = null;
            this.notice = null;

            const verified = await this.ensurePow();
            if (!verified || !this.powPayload) {
                this.error = this.t('auth.pow.required');
                this.loading = false;
                return;
            }

            try {
                await http.post('/api/auth/register', {
                    username: this.registerForm.username,
                    nickname: this.registerForm.nickname,
                    email: this.registerForm.email,
                    password: this.registerForm.password,
                    pow_payload: this.powPayload
                });
                this.notice = this.t('auth.register.success');
                this.registerForm = { username: '', nickname: '', email: '', password: '', confirmPassword: '' };
                this.resetPow();
            } catch (err) {
                this.error = formatAPIError(err, this.t);
                this.resetPow();
            } finally {
                this.loading = false;
            }
        },
        async submitForgot() {
            this.loading = true;
            this.error = null;
            this.notice = null;

            const verified = await this.ensurePow();
            if (!verified || !this.powPayload) {
                this.error = this.t('auth.pow.required');
                this.loading = false;
                return;
            }

            try {
                await http.post('/api/auth/forgot-password', {
                    email: this.forgotEmail,
                    pow_payload: this.powPayload
                });
                this.notice = this.t('auth.forgot.sent');
                this.resetPow();
            } catch (err) {
                this.error = formatAPIError(err, this.t);
                this.resetPow();
            } finally {
                this.loading = false;
            }
        },
        async submitReset() {
            if (this.resetForm.password !== this.resetForm.confirmPassword) {
                this.error = this.t('error.AUTH_PASSWORD_MISMATCH');
                return;
            }
            this.loading = true;
            this.error = null;
            this.notice = null;
            try {
                await http.post('/api/auth/reset-password', {
                    token: this.actionToken,
                    new_password: this.resetForm.password
                });
                this.notice = this.t('auth.reset.success');
                this.resetForm = { password: '', confirmPassword: '' };
            } catch (err) {
                this.error = formatAPIError(err, this.t);
            } finally {
                this.loading = false;
            }
        }
    }
}
</script>
