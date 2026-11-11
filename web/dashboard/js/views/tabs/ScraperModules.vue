<template>
    <div class="flex flex-col md:flex-row h-full gap-4 md:gap-5 relative">
        <div
            class="ui-surface flex flex-col"
            :class="selectedModule ? 'hidden md:flex md:w-72 lg:w-80 flex-shrink-0' : 'w-full flex-1 md:w-72 lg:w-80 md:flex-none'"
        >
            <div class="ui-surface-header px-4 py-3 flex justify-between items-center">
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('scraper.modules') }}</h3>
                <button
                    v-if="canEdit"
                    @click="reloadModules"
                    :disabled="reloading"
                    class="ui-btn-secondary py-1 px-2.5 text-xs"
                >
                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': reloading }">refresh</span>
                    <span>{{ t('scraper.reload_modules') }}</span>
                </button>
            </div>

            <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
                <div v-if="loadingList" class="p-8 text-center text-slate-400 text-xs">
                    {{ t('common.loading') }}
                </div>
                <div v-else-if="modules.length === 0" class="p-8 text-center text-slate-400 text-xs">
                    {{ t('common.no_data') }}
                </div>
                <div
                    v-for="mod in modules"
                    :key="mod.module_id"
                    @click="selectModule(mod)"
                    class="p-3 rounded-md cursor-pointer transition-colors flex flex-col gap-2 border"
                    :class="selectedModule && selectedModule.module_id === mod.module_id
                        ? 'bg-blue-50/70 dark:bg-blue-950/30 border-blue-200 dark:border-blue-900'
                        : 'hover:bg-slate-50 dark:hover:bg-slate-800/50 border-transparent'"
                >
                    <div class="flex justify-between items-start gap-2">
                        <div class="font-medium text-sm text-slate-900 dark:text-white truncate">{{ t(mod.name, mod.module_id) }}</div>
                        <div class="text-[11px] text-slate-400 font-mono flex-shrink-0">{{ mod.version }}</div>
                    </div>

                    <div class="flex items-center gap-1.5 flex-wrap">
                        <span
                            class="text-[10px] px-1.5 py-0.5 rounded border font-medium uppercase tracking-wider"
                            :class="mod.source === 'external'
                                ? 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/40 dark:text-indigo-300 dark:border-indigo-900'
                                : 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900'"
                        >
                            {{ t('scraper.source.' + mod.source, mod.source) }}
                        </span>
                        <span
                            class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                            :class="mod.is_enable
                                ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900'
                                : 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900'"
                        >
                            {{ mod.is_enable ? t('scraper.status.enabled') : t('scraper.status.disabled') }}
                        </span>
                        <span
                            v-if="mod.source === 'external'"
                            class="text-[10px] px-1.5 py-0.5 rounded border font-medium"
                            :class="mod.env_ready
                                ? 'bg-teal-50 text-teal-700 border-teal-200 dark:bg-teal-950/40 dark:text-teal-300 dark:border-teal-900'
                                : 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900'"
                        >
                            {{ mod.env_ready ? t('scraper.env.ready') : t('scraper.env.pending') }}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <div
            class="ui-surface flex flex-col overflow-hidden"
            :class="selectedModule ? 'w-full flex-1' : 'hidden md:flex md:flex-1'"
        >
            <div v-if="selectedModule" class="flex flex-col h-full">
                <div class="p-4 md:p-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-900/40">
                    <div class="flex items-start gap-3 mb-3">
                        <button
                            @click="clearSelectedModule"
                            class="md:hidden p-1 -ml-1 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500"
                        >
                            <span class="material-icons">arrow_back</span>
                        </button>
                        <div class="flex-1 min-w-0">
                            <h2 class="text-lg font-semibold text-slate-900 dark:text-white break-words">
                                {{ t(selectedModule.name, selectedModule.module_id) }}
                            </h2>
                            <p class="text-slate-500 dark:text-slate-400 text-xs mt-1">
                                {{ t(selectedModule.description, '') }}
                            </p>
                        </div>
                    </div>

                    <div class="flex flex-wrap gap-x-5 gap-y-1.5 text-xs text-slate-500 dark:text-slate-400 mb-4">
                        <span class="flex items-center gap-1">
                            <span class="material-icons text-[14px]">fingerprint</span>
                            <span class="font-mono">{{ selectedModule.module_id }}</span>
                        </span>
                        <span class="flex items-center gap-1">
                            <span class="material-icons text-[14px]">person</span>
                            <span>{{ selectedModule.author }}</span>
                        </span>
                        <span class="flex items-center gap-1">
                            <span class="material-icons text-[14px]">tag</span>
                            <span class="font-mono">{{ selectedModule.version }}</span>
                        </span>
                        <span class="flex items-center gap-1">
                            <span class="material-icons text-[14px]">extension</span>
                            <span>{{ t('scraper.source.' + selectedModule.source, selectedModule.source) }}</span>
                        </span>
                    </div>

                    <div v-if="canEdit" class="flex flex-wrap gap-2.5 justify-end">
                        <button
                            v-if="selectedModule.module_id === 'bilibili_ranking_and_update'"
                            @click="openBilibiliQR"
                            class="ui-btn-secondary text-xs py-1.5 px-3 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60"
                        >
                            <span class="material-icons text-[16px]">qr_code_2</span>
                            <span>{{ t('scraper.bilibili_qr.button') }}</span>
                        </button>

                        <div v-if="selectedModule.source === 'external'" class="flex items-center gap-1.5">
                            <button
                                @click="prepareEnv"
                                :disabled="preparingEnv"
                                class="ui-btn-secondary text-xs py-1.5 px-3"
                            >
                                <span class="material-icons text-[15px]" :class="{ 'animate-spin': preparingEnv }">build</span>
                                <span>{{ preparingEnv ? t('scraper.env.rebuilding') : t('scraper.prepare_env') }}</span>
                            </button>
                            <div class="relative group flex items-center">
                                <span class="material-icons text-slate-400 text-[16px] cursor-help hover:text-blue-500">help_outline</span>
                                <div class="absolute right-0 bottom-full mb-2 w-60 p-2 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                    {{ t('scraper.prepare_env_hint') }}
                                </div>
                            </div>
                        </div>

                        <button
                            @click="requestToggleEnable"
                            :disabled="enabling"
                            class="text-xs py-1.5 px-3.5"
                            :class="selectedModule.is_enable ? 'ui-btn-danger' : 'ui-btn-emerald'"
                        >
                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': enabling }">
                                {{ enabling ? 'sync' : (selectedModule.is_enable ? 'power_settings_new' : 'play_arrow') }}
                            </span>
                            <span>{{ enabling ? t('common.enabling') : (selectedModule.is_enable ? t('common.disable') : t('common.enable')) }}</span>
                        </button>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto p-4 md:p-5 space-y-6">
                    <div v-if="loadingDetail" class="flex justify-center py-12">
                        <span class="material-icons animate-spin text-2xl text-blue-600">sync</span>
                    </div>

                    <template v-else-if="moduleDetail">
                        <div class="ui-card">
                            <div class="ui-card-header">
                                <div class="flex items-center gap-2">
                                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">tune</span>
                                    <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                        {{ t('scraper.config.settings') }}
                                    </span>
                                </div>
                                <div v-if="canEdit" class="flex items-center gap-2">
                                    <button
                                        @click="testModule"
                                        :disabled="testing"
                                        class="ui-btn-secondary text-xs py-1.5 px-3"
                                    >
                                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': testing }">
                                            {{ testing ? 'sync' : 'wifi_tethering' }}
                                        </span>
                                        <span>{{ testing ? t('common.testing') : t('scraper.config.test') }}</span>
                                    </button>
                                    <button
                                        @click="saveConfig"
                                        :disabled="saving"
                                        class="ui-btn-primary text-xs py-1.5 px-3"
                                    >
                                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': saving }">
                                            {{ saving ? 'sync' : 'save' }}
                                        </span>
                                        <span>{{ saving ? t('common.saving') : t('scraper.save_config') }}</span>
                                    </button>
                                </div>
                            </div>

                            <div class="p-4">
                                <div v-if="!moduleDetail.config || moduleDetail.config.length === 0" class="text-xs text-slate-400 py-4 text-center">
                                    {{ t('scraper.no_config') }}
                                </div>

                                <div v-else class="space-y-4">
                                    <div v-for="field in sortedConfigFields" :key="field.key" class="flex flex-col gap-1.5">
                                        <div class="flex items-center gap-1.5">
                                            <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">
                                                {{ t(field.description, field.key) }}
                                            </label>
                                            <div v-if="hasTranslation(field.hint)" class="relative group flex items-center">
                                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                                <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                                    {{ t(field.hint) }}
                                                </div>
                                            </div>
                                        </div>

                                        <div v-if="field.type === 'array'" class="space-y-2">
                                            <div class="flex flex-wrap gap-1.5">
                                                <span
                                                    v-for="(item, idx) in (formValues[field.key] || [])"
                                                    :key="idx"
                                                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200 dark:border-blue-900"
                                                >
                                                    <span>{{ item }}</span>
                                                    <button
                                                        v-if="canEdit"
                                                        type="button"
                                                        @click="removeArrayItem(field.key, idx)"
                                                        class="hover:text-rose-500 flex items-center"
                                                    >
                                                        <span class="material-icons text-[14px]">close</span>
                                                    </button>
                                                </span>
                                            </div>
                                            <div v-if="canEdit" class="flex gap-2">
                                                <input
                                                    v-model="arrayDrafts[field.key]"
                                                    @keyup.enter.prevent="addArrayItem(field.key)"
                                                    type="text"
                                                    :placeholder="t('scraper.array_placeholder')"
                                                    class="ui-input flex-1 py-1.5"
                                                >
                                                <button
                                                    type="button"
                                                    @click="addArrayItem(field.key)"
                                                    class="ui-btn-secondary py-1.5 px-3 text-xs"
                                                >
                                                    <span class="material-icons text-[15px]">add</span>
                                                    <span>{{ t('common.add') }}</span>
                                                </button>
                                            </div>
                                        </div>

                                        <CustomSelect
                                            v-else-if="field.type === 'select'"
                                            v-model="formValues[field.key]"
                                            :options="field.options || []"
                                            :disabled="!canEdit"
                                        />

                                        <CustomSelect
                                            v-else-if="field.type === 'bool' || field.type === 'boolean'"
                                            v-model="formValues[field.key]"
                                            :options="[
                                                { label: t('common.true'), value: true },
                                                { label: t('common.false'), value: false }
                                            ]"
                                            :disabled="!canEdit"
                                        />

                                        <input
                                            v-else-if="field.type === 'int'"
                                            v-model="formValues[field.key]"
                                            type="number"
                                            :disabled="!canEdit"
                                            class="ui-input"
                                        >

                                        <input
                                            v-else
                                            v-model="formValues[field.key]"
                                            :type="field.secret || field.type === 'password' ? 'password' : 'text'"
                                            :disabled="!canEdit"
                                            class="ui-input"
                                        >
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div v-if="canViewClassification" class="ui-card">
                            <div class="ui-card-header">
                                <div class="flex items-center gap-2">
                                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">auto_awesome</span>
                                    <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                        {{ t('scraper.cls.title') }}
                                    </span>
                                </div>
                                <button
                                    v-if="canEditClassification"
                                    @click="saveModuleClassification"
                                    :disabled="savingCls"
                                    class="ui-btn-primary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingCls }">
                                        {{ savingCls ? 'sync' : 'save' }}
                                    </span>
                                    <span>{{ savingCls ? t('common.saving') : t('scraper.cls.save') }}</span>
                                </button>
                            </div>

                            <div class="p-4 space-y-4">
                                <label class="inline-flex items-center gap-2.5 cursor-pointer select-none">
                                    <input
                                        type="checkbox"
                                        v-model="clsConfig.override_system"
                                        :disabled="!canEditClassification"
                                        class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                    >
                                    <span class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ t('scraper.cls.override') }}</span>
                                </label>

                                <div v-if="clsConfig.override_system" class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2 border-t border-slate-200 dark:border-slate-800">
                                    <div>
                                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                            {{ t('config.classification.provider.desc') }}
                                        </label>
                                        <CustomSelect
                                            v-model="clsConfig.provider"
                                            :disabled="!canEditClassification"
                                            :options="[
                                                { value: 'none', label: t('config.classification.provider.none') },
                                                { value: 'onnx', label: t('config.classification.provider.onnx') },
                                                { value: 'llm', label: t('config.classification.provider.llm') }
                                            ]"
                                        />
                                    </div>

                                    <div v-if="clsConfig.provider !== 'none'">
                                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                            {{ t('scraper.cls.profile') }}
                                        </label>
                                        <CustomSelect
                                            v-model="clsConfig.profile_id"
                                            :disabled="!canEditClassification"
                                            :options="profileSelectOptions"
                                        />
                                    </div>

                                    <div v-if="clsConfig.provider !== 'none'">
                                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                            {{ t('config.classification.confidence.desc') }}
                                        </label>
                                        <input
                                            v-model.number="clsConfig.confidence"
                                            type="number"
                                            min="1"
                                            max="100"
                                            :disabled="!canEditClassification"
                                            class="ui-input"
                                        >
                                    </div>

                                    <div v-if="clsConfig.provider !== 'none'">
                                        <div class="flex items-center gap-1.5 mb-1.5">
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                                                {{ t('config.classification.max_input_chars.desc') }}
                                            </label>
                                            <div class="relative group flex items-center">
                                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                                <div class="absolute left-0 top-full mt-1.5 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                                    {{ t('config.classification.max_input_chars.hint') }}
                                                </div>
                                            </div>
                                        </div>
                                        <input
                                            v-model="clsConfig.max_input_chars"
                                            type="number"
                                            min="0"
                                            :disabled="!canEditClassification"
                                            :placeholder="t('scraper.cls.use_system_limit')"
                                            class="ui-input"
                                        >
                                    </div>

                                    <div v-if="clsConfig.provider !== 'none'">
                                        <div class="flex items-center gap-1.5 mb-1.5">
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                                                {{ t('config.classification.truncate_strategy.desc') }}
                                            </label>
                                            <div class="relative group flex items-center">
                                                <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                                <div class="absolute left-0 top-full mt-1.5 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                                    {{ t('config.classification.truncate_strategy.hint') }}
                                                </div>
                                            </div>
                                        </div>
                                        <CustomSelect
                                            v-model="clsConfig.truncate_strategy"
                                            :disabled="!canEditClassification"
                                            :options="truncateStrategyOptions"
                                        />
                                    </div>

                                    <div v-if="clsConfig.provider === 'onnx'">
                                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                            {{ t('config.classification.onnx_model.desc') }}
                                        </label>
                                        <CustomSelect
                                            v-model="clsConfig.model_name"
                                            :disabled="!canEditClassification"
                                            :options="clsOnnxModelOptions"
                                        />
                                    </div>

                                    <template v-if="clsConfig.provider === 'llm'">
                                        <div>
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                                {{ t('config.classification.llm_protocol.desc') }}
                                            </label>
                                            <CustomSelect
                                                v-model="clsConfig.llm_protocol"
                                                :disabled="!canEditClassification"
                                                :options="[
                                                    { value: 'ollama', label: t('config.classification.llm_protocol.ollama') },
                                                    { value: 'openai', label: t('config.classification.llm_protocol.openai') }
                                                ]"
                                            />
                                        </div>

                                        <div>
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                                {{ t('config.classification.llm_url.desc') }}
                                            </label>
                                            <input
                                                v-model="clsConfig.llm_url"
                                                type="text"
                                                :disabled="!canEditClassification"
                                                class="ui-input font-mono text-xs"
                                            >
                                        </div>

                                        <div>
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                                {{ t('config.classification.llm_model.desc') }}
                                            </label>
                                            <input
                                                v-model="clsConfig.llm_model"
                                                type="text"
                                                :disabled="!canEditClassification"
                                                :placeholder="t('scraper.cls.use_system_model')"
                                                class="ui-input font-mono text-xs"
                                            >
                                        </div>

                                        <div>
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                                {{ t('config.classification.llm_api_key.desc') }}
                                            </label>
                                            <input
                                                v-model="clsConfig.llm_api_key"
                                                type="password"
                                                :disabled="!canEditClassification"
                                                class="ui-input"
                                            >
                                        </div>

                                        <div class="md:col-span-2">
                                            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                                {{ t('config.classification.llm_prompt.desc') }}
                                            </label>
                                            <textarea
                                                v-model="clsConfig.llm_prompt"
                                                rows="2"
                                                :disabled="!canEditClassification"
                                                class="ui-input"
                                            ></textarea>
                                        </div>
                                    </template>
                                </div>
                            </div>
                        </div>

                        <div class="ui-card">
                            <div class="ui-card-header">
                                <div class="flex items-center gap-2">
                                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">playlist_play</span>
                                    <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                        {{ t('scraper.config.tasks') }}
                                    </span>
                                </div>
                            </div>

                            <div class="divide-y divide-slate-200 dark:divide-slate-800">
                                <div
                                    v-for="task in (moduleDetail.tasks || [])"
                                    :key="task.task_key"
                                    class="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
                                >
                                    <div class="min-w-0">
                                        <div class="flex items-center gap-2 flex-wrap">
                                            <span class="font-medium text-sm text-slate-900 dark:text-white">
                                                {{ t(task.name, task.task_key) }}
                                            </span>
                                            <span class="text-xs font-mono px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                                                {{ task.cron_expr }}
                                            </span>
                                        </div>
                                        <p v-if="hasTranslation(task.description)" class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                                            {{ t(task.description) }}
                                        </p>
                                    </div>

                                    <button
                                        v-if="canEdit"
                                        @click="runTask(task.task_key)"
                                        :disabled="runningTask === task.task_key"
                                        class="ui-btn-secondary text-xs py-1.5 px-3 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60 flex-shrink-0"
                                    >
                                        <span class="material-icons text-[15px]" :class="{ 'animate-spin': runningTask === task.task_key }">
                                            {{ runningTask === task.task_key ? 'sync' : 'play_circle' }}
                                        </span>
                                        <span>{{ runningTask === task.task_key ? t('common.running') : t('scraper.run_task') }}</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </template>
                </div>
            </div>

            <div v-else class="flex-1 flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 p-8 text-center">
                <span class="material-icons text-4xl mb-2 opacity-40">cloud_download</span>
                <p class="text-xs">{{ t('scraper.select_module') }}</p>
            </div>
        </div>

        <transition name="modal">
            <div v-if="biliModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg p-6 w-full max-w-sm border border-slate-200 dark:border-slate-800">
                    <div class="flex items-center justify-between mb-3">
                        <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('scraper.bilibili_qr.title') }}</h3>
                        <button @click="closeBilibiliQR" class="p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded">
                            <span class="material-icons text-[18px]">close</span>
                        </button>
                    </div>
                    <p class="text-xs text-slate-500 dark:text-slate-400 mb-4 leading-relaxed">{{ t('scraper.bilibili_qr.desc') }}</p>

                    <div class="flex flex-col items-center justify-center p-4 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-lg mb-4 min-h-[252px]">
                        <div v-if="biliModal.loading" class="flex flex-col items-center text-slate-400">
                            <span class="material-icons animate-spin text-2xl text-blue-600 mb-2">sync</span>
                            <span class="text-xs">{{ t('common.loading') }}</span>
                        </div>
                        <canvas v-else ref="biliQrCanvas" class="rounded bg-white"></canvas>
                    </div>

                    <div
                        class="text-center text-xs font-medium px-3 py-2 rounded-md border mb-4"
                        :class="biliStatusClass"
                    >
                        {{ biliStatusText }}
                    </div>

                    <div class="flex justify-end gap-2">
                        <button
                            v-if="biliModal.status === 'expired' || biliModal.error"
                            @click="refreshBilibiliQR"
                            class="ui-btn-primary text-xs py-1.5 px-3"
                        >
                            <span class="material-icons text-[15px]">refresh</span>
                            <span>{{ t('scraper.bilibili_qr.refresh_qr') }}</span>
                        </button>
                        <button @click="closeBilibiliQR" class="ui-btn-secondary text-xs py-1.5 px-3">
                            {{ t('common.close') }}
                        </button>
                    </div>
                </div>
            </div>
        </transition>

        <ConfirmModal
            :show="externalConfirm.show"
            :title="t('scraper.external_warning.title')"
            :message="t('scraper.external_warning.content')"
            :t="t"
            @confirm="performEnableModule"
            @cancel="externalConfirm.show = false"
        />

        <InfoModal
            :show="outputModal.show"
            :title="outputModal.title"
            :message="outputModal.message"
            :t="t"
            @close="outputModal.show = false"
        />
    </div>
</template>

<script>
import QRCode from 'qrcode';
import http, { formatAPIError } from '../../utils/http.js';
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';
import InfoModal from '../../components/InfoModal.vue';

function defaultClsConfig() {
    return {
        override_system: false,
        provider: 'none',
        model_name: '',
        confidence: 60,
        max_input_chars: '',
        truncate_strategy: '',
        profile_id: 0,
        llm_protocol: 'ollama',
        llm_url: 'http://127.0.0.1:11434',
        llm_api_key: '',
        llm_model: '',
        llm_prompt: ''
    };
}

export default {
    components: { CustomSelect, ConfirmModal, InfoModal },
    props: ['t', 'currentUser', 'subtab'],
    emits: ['update:subtab'],
    inject: ['showToast'],
    data() {
        return {
            modules: [],
            selectedModule: null,
            moduleDetail: null,
            formValues: {},
            arrayDrafts: {},
            loadingList: false,
            loadingDetail: false,
            reloading: false,
            enabling: false,
            preparingEnv: false,
            saving: false,
            testing: false,
            runningTask: '',
            externalConfirm: { show: false },
            outputModal: { show: false, title: '', message: '' },

            clsConfig: defaultClsConfig(),
            clsProfiles: [],
            onnxLocalModels: [],
            savingCls: false,

            biliModal: {
                show: false,
                loading: false,
                qrcodeKey: '',
                status: 'waiting',
                error: ''
            },
            biliPollTimer: null,
            biliPolling: false
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('scraper.edit');
        },
        canViewClassification() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('classification.view');
        },
        canEditClassification() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('classification.edit');
        },
        sortedConfigFields() {
            const cfg = this.moduleDetail?.config || [];
            return cfg.slice().sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
        },
        profileSelectOptions() {
            const opts = [{ value: 0, label: this.t('scraper.cls.use_system_profile') }];
            (this.clsProfiles || []).forEach(p => {
                opts.push({
                    value: p.id,
                    label: p.is_default ? `${p.name} [${this.t('classification.profile_default')}]` : p.name
                });
            });
            return opts;
        },
        clsOnnxModelOptions() {
            const opts = [{ value: '', label: this.t('scraper.cls.use_system_model') }];
            const seen = new Set();
            (this.onnxLocalModels || []).forEach(m => {
                if (!m?.value || seen.has(m.value)) return;
                seen.add(m.value);
                const label = m.ready ? m.label : `${m.label} · ${this.t('classification.onnx_env_pending')}`;
                opts.push({ value: m.value, label });
            });
            const cur = String(this.clsConfig.model_name || '').trim();
            if (cur && !seen.has(cur)) {
                opts.push({ value: cur, label: cur });
            }
            return opts;
        },
        truncateStrategyOptions() {
            return [
                { value: '', label: this.t('scraper.cls.use_system_strategy') },
                { value: 'title_head_tail', label: this.t('config.classification.truncate_strategy.title_head_tail') },
                { value: 'head', label: this.t('config.classification.truncate_strategy.head') },
                { value: 'tail', label: this.t('config.classification.truncate_strategy.tail') }
            ];
        },
        biliStatusText() {
            if (this.biliModal.error) return this.biliModal.error;
            return this.t('scraper.bilibili_qr.status_' + this.biliModal.status, this.biliModal.status);
        },
        biliStatusClass() {
            if (this.biliModal.error || this.biliModal.status === 'expired') {
                return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
            }
            if (this.biliModal.status === 'confirmed') {
                return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900';
            }
            if (this.biliModal.status === 'scanned') {
                return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
            }
            return 'bg-slate-50 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
        }
    },
    watch: {
        async subtab(next) {
            const targetId = String(next || '').trim();
            if (!targetId) {
                this.stopBilibiliPoll();
                this.selectedModule = null;
                this.moduleDetail = null;
                return;
            }
            if (this.selectedModule && this.selectedModule.module_id === targetId) {
                return;
            }
            const found = (this.modules || []).find(m => m.module_id === targetId);
            if (found) {
                this.stopBilibiliPoll();
                this.selectedModule = found;
                await this.fetchModuleDetail(found.module_id);
            }
        }
    },
    mounted() {
        this.fetchModules();
    },
    unmounted() {
        this.stopBilibiliPoll();
    },
    methods: {
        hasTranslation(key) {
            if (!key) return false;
            const v = this.t(key, '');
            return Boolean(v && v !== key);
        },
        async fetchModules() {
            this.loadingList = true;
            try {
                const res = await http.get('/api/dashboard/scraper/modules');
                this.modules = res.data || [];
                const targetSubtab = String(this.subtab || '').trim();
                if (this.selectedModule) {
                    const updated = this.modules.find(m => m.module_id === this.selectedModule.module_id);
                    if (updated) {
                        this.selectedModule = updated;
                    }
                } else if (targetSubtab) {
                    const matched = this.modules.find(m => m.module_id === targetSubtab);
                    if (matched) {
                        this.selectedModule = matched;
                        await this.fetchModuleDetail(matched.module_id);
                    } else {
                        this.$emit('update:subtab', { subtab: '', replace: true });
                    }
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loadingList = false;
            }
        },
        async reloadModules() {
            this.reloading = true;
            try {
                await http.post('/api/dashboard/scraper/modules/reload');
                await this.fetchModules();
                if (this.selectedModule) {
                    await this.fetchModuleDetail(this.selectedModule.module_id);
                }
                this.showToast(this.t('common.saved_success'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.reloading = false;
            }
        },
        async selectModule(mod) {
            this.stopBilibiliPoll();
            this.selectedModule = mod;
            this.$emit('update:subtab', mod.module_id);
            await this.fetchModuleDetail(mod.module_id);
        },
        clearSelectedModule() {
            this.stopBilibiliPoll();
            this.selectedModule = null;
            this.moduleDetail = null;
            this.$emit('update:subtab', '');
        },
        async fetchModuleDetail(id) {
            this.loadingDetail = true;
            try {
                const reqs = [http.get(`/api/dashboard/scraper/modules/${encodeURIComponent(id)}`)];
                if (this.canViewClassification) {
                    reqs.push(http.get(`/api/dashboard/classification/modules/${encodeURIComponent(id)}`).catch(() => null));
                    reqs.push(http.get('/api/dashboard/classification/profiles').catch(() => null));
                    reqs.push(http.get('/api/dashboard/classification/overview').catch(() => null));
                }

                const [detailRes, clsRes, profRes, ovRes] = await Promise.all(reqs);
                this.moduleDetail = detailRes.data;
                const vals = {};
                const drafts = {};
                (detailRes.data?.config || []).forEach(f => {
                    if (f.type === 'array') {
                        vals[f.key] = Array.isArray(f.value) ? [...f.value] : [];
                        drafts[f.key] = '';
                    } else if (f.type === 'bool' || f.type === 'boolean') {
                        vals[f.key] = Boolean(f.value === true || f.value === 'true');
                    } else {
                        vals[f.key] = f.value ?? '';
                    }
                });
                this.formValues = vals;
                this.arrayDrafts = drafts;

                if (clsRes?.data) {
                    const c = clsRes.data;
                    this.clsConfig = {
                        ...defaultClsConfig(),
                        ...c,
                        max_input_chars: (c.max_input_chars === null || c.max_input_chars === undefined) ? '' : c.max_input_chars,
                        truncate_strategy: c.truncate_strategy || '',
                        profile_id: c.profile_id || 0
                    };
                } else {
                    this.clsConfig = defaultClsConfig();
                }
                if (Array.isArray(profRes?.data)) {
                    this.clsProfiles = profRes.data;
                }
                if (ovRes?.data) {
                    this.onnxLocalModels = ovRes.data.onnx_local_models || [];
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loadingDetail = false;
            }
        },
        addArrayItem(key) {
            const val = (this.arrayDrafts[key] || '').trim();
            if (!val) return;
            if (!Array.isArray(this.formValues[key])) {
                this.formValues[key] = [];
            }
            if (!this.formValues[key].includes(val)) {
                this.formValues[key].push(val);
            }
            this.arrayDrafts[key] = '';
        },
        removeArrayItem(key, idx) {
            if (Array.isArray(this.formValues[key])) {
                this.formValues[key].splice(idx, 1);
            }
        },
        requestToggleEnable() {
            if (!this.selectedModule) return;
            if (this.selectedModule.is_enable) {
                this.performDisableModule();
            } else if (this.selectedModule.source === 'external') {
                this.externalConfirm.show = true;
            } else {
                this.performEnableModule();
            }
        },
        async performEnableModule() {
            this.externalConfirm.show = false;
            this.enabling = true;
            try {
                await http.post(`/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/enable`);
                await this.fetchModules();
                await this.fetchModuleDetail(this.selectedModule.module_id);
                this.showToast(this.t('scraper.status.enabled'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.enabling = false;
            }
        },
        async performDisableModule() {
            this.enabling = true;
            try {
                await http.post(`/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/disable`);
                await this.fetchModules();
                await this.fetchModuleDetail(this.selectedModule.module_id);
                this.showToast(this.t('scraper.status.disabled'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.enabling = false;
            }
        },
        async prepareEnv() {
            if (!this.selectedModule) return;
            this.preparingEnv = true;
            try {
                const res = await http.post(`/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/prepare-env`);
                await this.fetchModules();
                this.showToast(this.t('scraper.env.ready'), 'success');
                if (res.data?.output && res.data.output.trim()) {
                    this.outputModal = {
                        show: true,
                        title: this.t('scraper.prepare_env'),
                        message: res.data.output.trim()
                    };
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.preparingEnv = false;
            }
        },
        buildConfigPayload() {
            const payload = {};
            (this.moduleDetail?.config || []).forEach(f => {
                const val = this.formValues[f.key];
                if (f.secret && val === '********') {
                    return;
                }
                if (f.type === 'int') {
                    payload[f.key] = String(val ?? '0');
                } else {
                    payload[f.key] = val;
                }
            });
            return payload;
        },
        async saveConfig() {
            if (!this.selectedModule) return;
            this.saving = true;
            try {
                await http.post(
                    `/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/config`,
                    this.buildConfigPayload()
                );
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchModuleDetail(this.selectedModule.module_id);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.saving = false;
            }
        },
        async testModule() {
            if (!this.selectedModule) return;
            this.testing = true;
            try {
                const res = await http.post(
                    `/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/test`,
                    this.buildConfigPayload()
                );
                const result = res.data || {};
                const baseMsg = this.t(result.message_key, result.success ? this.t('module.test.passed') : this.t('module.test.failed'));
                const msg = result.detail ? `${baseMsg}: ${result.detail}` : baseMsg;
                this.showToast(msg, result.success ? 'success' : 'error');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.testing = false;
            }
        },
        async saveModuleClassification() {
            if (!this.selectedModule) return;
            this.savingCls = true;
            try {
                const rawMaxChars = this.clsConfig.max_input_chars;
                let maxInputChars = null;
                if (rawMaxChars !== '' && rawMaxChars !== null && rawMaxChars !== undefined) {
                    const parsedMaxChars = Number(rawMaxChars);
                    if (!Number.isNaN(parsedMaxChars) && parsedMaxChars >= 0) {
                        maxInputChars = Math.floor(parsedMaxChars);
                    }
                }
                await http.put(`/api/dashboard/classification/modules/${encodeURIComponent(this.selectedModule.module_id)}`, {
                    module_id: this.selectedModule.module_id,
                    override_system: Boolean(this.clsConfig.override_system),
                    provider: this.clsConfig.provider || 'none',
                    model_name: String(this.clsConfig.model_name || '').trim(),
                    confidence: Number(this.clsConfig.confidence) || 60,
                    max_input_chars: maxInputChars,
                    truncate_strategy: String(this.clsConfig.truncate_strategy || '').trim(),
                    profile_id: this.clsConfig.profile_id > 0 ? Number(this.clsConfig.profile_id) : null,
                    llm_protocol: this.clsConfig.llm_protocol || 'ollama',
                    llm_url: String(this.clsConfig.llm_url || '').trim(),
                    llm_api_key: String(this.clsConfig.llm_api_key || '').trim(),
                    llm_model: String(this.clsConfig.llm_model || '').trim(),
                    llm_prompt: String(this.clsConfig.llm_prompt || '').trim()
                });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchModuleDetail(this.selectedModule.module_id);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingCls = false;
            }
        },
        async runTask(taskKey) {
            if (!this.selectedModule) return;
            this.runningTask = taskKey;
            try {
                const res = await http.post(
                    `/api/dashboard/scraper/modules/${encodeURIComponent(this.selectedModule.module_id)}/tasks/${encodeURIComponent(taskKey)}/run`
                );
                const inserted = res.data?.inserted ?? 0;
                this.showToast(this.t('scraper.task_inserted').replace('{inserted}', String(inserted)), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.runningTask = '';
            }
        },
        async openBilibiliQR() {
            this.biliModal = {
                show: true,
                loading: true,
                qrcodeKey: '',
                status: 'waiting',
                error: ''
            };
            await this.refreshBilibiliQR();
        },
        async refreshBilibiliQR() {
            this.stopBilibiliPoll();
            this.biliModal.loading = true;
            this.biliModal.status = 'waiting';
            this.biliModal.error = '';
            try {
                const res = await http.post('/api/dashboard/scraper/bilibili/qrcode/generate');
                const { url, qrcode_key } = res.data || {};
                this.biliModal.qrcodeKey = qrcode_key;
                this.biliModal.loading = false;
                await this.$nextTick();
                if (this.$refs.biliQrCanvas && url) {
                    await QRCode.toCanvas(this.$refs.biliQrCanvas, url, { width: 220, margin: 2 });
                }
                this.biliPollTimer = setInterval(() => this.pollBilibiliQR(), 2000);
            } catch (err) {
                this.biliModal.loading = false;
                this.biliModal.error = formatAPIError(err, this.t);
            }
        },
        async pollBilibiliQR() {
            if (!this.biliModal.show || !this.biliModal.qrcodeKey || this.biliPolling) return;
            this.biliPolling = true;
            try {
                const res = await http.get('/api/dashboard/scraper/bilibili/qrcode/poll', {
                    params: { qrcode_key: this.biliModal.qrcodeKey }
                });
                const status = res.data?.status || 'waiting';
                this.biliModal.status = status;
                if (status === 'expired') {
                    this.stopBilibiliPoll();
                } else if (status === 'confirmed') {
                    this.stopBilibiliPoll();
                    this.showToast(this.t('scraper.bilibili_qr.status_confirmed'), 'success');
                    await this.fetchModuleDetail('bilibili_ranking_and_update');
                    setTimeout(() => {
                        this.biliModal.show = false;
                    }, 600);
                }
            } catch (err) {
                this.biliModal.error = formatAPIError(err, this.t);
                this.stopBilibiliPoll();
            } finally {
                this.biliPolling = false;
            }
        },
        stopBilibiliPoll() {
            if (this.biliPollTimer) {
                clearInterval(this.biliPollTimer);
                this.biliPollTimer = null;
            }
            this.biliPolling = false;
        },
        closeBilibiliQR() {
            this.stopBilibiliPoll();
            this.biliModal.show = false;
        }
    }
};
</script>
