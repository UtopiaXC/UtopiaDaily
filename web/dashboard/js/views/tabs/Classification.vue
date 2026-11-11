<template>
    <div class="ui-surface flex flex-col h-full">
        <div class="ui-surface-header pb-0">
            <div class="flex flex-wrap justify-between items-center gap-3 pb-3">
                <h3 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">{{ t('classification.title') }}</h3>
                <button @click="refreshCurrent" class="ui-btn-secondary">
                    <span class="material-icons text-[16px]" :class="{ 'animate-spin': loading }">refresh</span>
                    <span>{{ t('common.refresh') }}</span>
                </button>
            </div>

            <div class="flex gap-6 overflow-x-auto">
                <button
                    @click="switchSubtab('engine')"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        activeSubtab === 'engine'
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    {{ t('classification.tab.engine') }}
                </button>
                <button
                    @click="switchSubtab('profiles')"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                        activeSubtab === 'profiles'
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    {{ t('classification.tab.profiles') }}
                </button>
                <button
                    @click="switchSubtab('queue')"
                    :class="[
                        'pb-2.5 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap inline-flex items-center gap-2',
                        activeSubtab === 'queue'
                            ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                            : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                    ]"
                >
                    <span>{{ t('classification.tab.queue') }}</span>
                    <span
                        v-if="queuePendingTotal > 0"
                        class="text-[11px] font-mono px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900"
                    >
                        {{ queuePendingTotal }}
                    </span>
                </button>
            </div>
        </div>

        <div v-if="loading && !initialized" class="p-12 text-center text-slate-500 dark:text-slate-400 flex-1">
            <span class="material-icons animate-spin text-2xl mb-2 text-blue-600">sync</span>
            <p class="text-xs">{{ t('common.loading') }}</p>
        </div>

        <div v-else class="flex-1 overflow-y-auto p-5">
            <!-- Subtab 1: Engine & Models -->
            <div v-if="activeSubtab === 'engine'">
                <div class="ui-card">
                    <div class="ui-card-header">
                        <div class="flex items-center gap-2">
                            <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">tune</span>
                            <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                                {{ t('classification.global_settings') }}
                            </span>
                        </div>
                        <button
                            v-if="canEditSystem && engineConfigs.length > 0"
                            @click="saveEngineSettings"
                            :disabled="savingEngine"
                            class="ui-btn-primary text-xs py-1.5 px-3"
                        >
                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingEngine }">
                                {{ savingEngine ? 'sync' : 'save' }}
                            </span>
                            <span>{{ savingEngine ? t('common.saving') : t('classification.save_settings') }}</span>
                        </button>
                    </div>

                    <div v-if="engineConfigs.length > 0" class="p-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div
                            v-for="conf in visibleEngineConfigs"
                            :key="conf.key"
                            :class="conf.key === 'classification_llm_prompt' ? 'md:col-span-2 flex flex-col gap-1.5' : 'flex flex-col gap-1.5'"
                        >
                            <div class="flex items-center gap-1.5">
                                <label class="text-xs font-semibold text-slate-700 dark:text-slate-300">
                                    {{ t(conf.description, conf.key) }}
                                </label>
                                <div v-if="conf.hint" class="relative group flex items-center">
                                    <span class="material-icons text-slate-400 text-[15px] cursor-help hover:text-blue-500">help_outline</span>
                                    <div class="absolute left-full ml-2 top-1/2 -translate-y-1/2 w-64 p-2.5 bg-slate-900 text-white text-xs rounded-md border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-20">
                                        {{ t(conf.hint) }}
                                    </div>
                                </div>
                            </div>

                            <CustomSelect
                                v-if="conf.type === 'select'"
                                v-model="engineForm[conf.key]"
                                :disabled="!canEditSystem"
                                :options="formatOptions(conf.options)"
                            />
                            <CustomSelect
                                v-else-if="conf.key === 'classification_onnx_model'"
                                v-model="engineForm.classification_onnx_model"
                                :disabled="!canEditSystem"
                                :options="onnxModelSelectOptions"
                            />
                            <input
                                v-else-if="conf.key === 'classification_llm_model'"
                                v-model="engineForm.classification_llm_model"
                                type="text"
                                :disabled="!canEditSystem"
                                :placeholder="t('classification.llm_model_placeholder')"
                                class="ui-input font-mono text-xs"
                            >
                            <input
                                v-else-if="conf.type === 'int'"
                                v-model="engineForm[conf.key]"
                                type="number"
                                :min="conf.key === 'classification_confidence' ? 1 : 0"
                                :max="conf.key === 'classification_confidence' ? 100 : null"
                                :disabled="!canEditSystem"
                                class="ui-input"
                            >
                            <textarea
                                v-else-if="conf.key === 'classification_llm_prompt'"
                                v-model="engineForm[conf.key]"
                                rows="2"
                                :disabled="!canEditSystem"
                                class="ui-input"
                            ></textarea>
                            <input
                                v-else
                                v-model="engineForm[conf.key]"
                                :type="conf.is_secret ? 'password' : 'text'"
                                :disabled="!canEditSystem"
                                class="ui-input"
                                :class="conf.key.includes('url') ? 'font-mono text-xs' : ''"
                            >
                        </div>

                        <div
                            v-if="selectedProvider === 'none'"
                            class="md:col-span-2 p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 text-xs text-slate-600 dark:text-slate-300 flex items-center gap-2"
                        >
                            <span class="material-icons text-slate-400 text-[16px]">info</span>
                            <span>{{ t('classification.provider_disabled_hint') }}</span>
                        </div>
                    </div>

                    <div v-else class="p-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] font-medium text-slate-500 uppercase">{{ t('config.classification.provider_label') }}</div>
                            <div class="text-sm font-semibold text-slate-900 dark:text-white mt-1">
                                {{ t('config.classification.provider.' + (overview.system_provider || 'none'), overview.system_provider || 'none') }}
                            </div>
                        </div>
                        <div class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] font-medium text-slate-500 uppercase">{{ t('config.classification.confidence.desc') }}</div>
                            <div class="text-sm font-semibold font-mono text-slate-900 dark:text-white mt-1">
                                {{ overview.system_confidence || 60 }}%
                            </div>
                        </div>
                    </div>

                    <!-- ONNX Subsections (shown only when engine is ONNX) -->
                    <template v-if="selectedProvider === 'onnx'">
                        <div class="border-t border-slate-200 dark:border-slate-800 p-4 space-y-3">
                            <div class="flex flex-wrap items-center justify-between gap-3">
                                <div class="flex items-center gap-2.5">
                                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">memory</span>
                                    <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                                        {{ t('classification.onnx_env_title') }}
                                    </span>
                                    <span
                                        class="text-[11px] px-2 py-0.5 rounded border font-medium"
                                        :class="overview.onnx_env?.ready
                                            ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900'
                                            : 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900'"
                                    >
                                        {{ overview.onnx_env?.ready ? t('classification.onnx_env_ready') : t('classification.onnx_env_pending') }}
                                    </span>
                                </div>
                                <button
                                    v-if="canEdit"
                                    @click="prepareONNXEnv"
                                    :disabled="startingONNXEnv || activeONNXEnvRunning"
                                    class="ui-btn-secondary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': startingONNXEnv || activeONNXEnvRunning }">
                                        {{ (startingONNXEnv || activeONNXEnvRunning) ? 'sync' : 'build' }}
                                    </span>
                                    <span>{{ (startingONNXEnv || activeONNXEnvRunning) ? t('classification.onnx_preparing') : t('classification.onnx_prepare_env') }}</span>
                                </button>
                            </div>

                            <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('classification.onnx_env_hint') }}</p>

                            <div v-if="onnxEnvJobs.length > 0" class="space-y-2">
                                <div
                                    v-for="job in onnxEnvJobs"
                                    :key="job.id"
                                    class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 space-y-2"
                                >
                                    <div class="flex flex-wrap items-center justify-between gap-2">
                                        <div class="flex items-center gap-2 min-w-0">
                                            <span class="text-[11px] px-1.5 py-0.5 rounded border font-medium" :class="jobStageBadgeClass(job)">
                                                {{ t('classification.job_stage.' + (job.stage || 'preparing_env'), job.stage || 'preparing_env') }}
                                            </span>
                                            <span v-if="job.message" class="text-xs font-mono text-slate-600 dark:text-slate-300 truncate max-w-md">
                                                {{ job.message }}
                                            </span>
                                        </div>
                                        <div class="flex items-center gap-2">
                                            <span class="text-xs font-mono text-slate-600 dark:text-slate-300">
                                                {{ formatJobProgressSummary(job) }}
                                            </span>
                                            <button
                                                v-if="canEdit && job.status === 'running'"
                                                type="button"
                                                @click="cancelModelJob(job.id)"
                                                class="ui-btn-secondary text-[11px] py-0.5 px-2 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                            >
                                                {{ t('classification.job_cancel') }}
                                            </button>
                                            <button
                                                v-else-if="canEdit"
                                                type="button"
                                                @click="dismissModelJob(job.id)"
                                                :title="t('classification.job_dismiss')"
                                                class="p-0.5 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded"
                                            >
                                                <span class="material-icons text-[15px]">close</span>
                                            </button>
                                        </div>
                                    </div>
                                    <div class="h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                                        <div
                                            v-if="isJobIndeterminate(job)"
                                            class="h-full bg-blue-600 dark:bg-blue-500 rounded-full ui-progress-indeterminate"
                                        ></div>
                                        <div
                                            v-else
                                            class="h-full rounded-full transition-all duration-200"
                                            :class="jobBarColorClass(job)"
                                            :style="{ width: jobBarWidth(job) }"
                                        ></div>
                                    </div>
                                    <div v-if="job.error" class="text-xs font-mono text-rose-600 dark:text-rose-400 whitespace-pre-wrap break-all max-h-24 overflow-y-auto">
                                        {{ job.error }}
                                    </div>
                                </div>
                            </div>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block mb-0.5">Runtime</span>
                                    <span class="font-mono text-slate-700 dark:text-slate-300 break-all">{{ overview.onnx_env?.runtime_dir || '-' }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block mb-0.5">Models</span>
                                    <span class="font-mono text-slate-700 dark:text-slate-300 break-all">{{ overview.onnx_env?.models_dir || '-' }}</span>
                                </div>
                            </div>
                        </div>

                        <div class="border-t border-slate-200 dark:border-slate-800">
                            <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 bg-slate-50/40 dark:bg-slate-900/40">
                                <div class="flex items-center gap-2">
                                    <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">inventory_2</span>
                                    <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                                        {{ t('classification.onnx_models_title') }}
                                    </span>
                                </div>
                                <button
                                    v-if="canEdit"
                                    @click="openONNXDownloadModal"
                                    class="ui-btn-primary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]">download</span>
                                    <span>{{ t('classification.onnx_download_btn') }}</span>
                                </button>
                            </div>

                            <div v-if="onnxDownloadJobs.length > 0" class="p-4 border-b border-slate-200 dark:border-slate-800 space-y-2.5 bg-slate-50/40 dark:bg-slate-950/30">
                                <div
                                    v-for="job in onnxDownloadJobs"
                                    :key="job.id"
                                    class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-2"
                                >
                                    <div class="flex flex-wrap items-center justify-between gap-2">
                                        <div class="flex items-center gap-2 min-w-0">
                                            <span class="text-[11px] px-1.5 py-0.5 rounded border font-medium" :class="jobStageBadgeClass(job)">
                                                {{ t('classification.job_stage.' + (job.stage || 'resolving'), job.stage || 'resolving') }}
                                            </span>
                                            <span class="font-mono text-xs font-semibold text-slate-900 dark:text-white truncate">{{ job.target }}</span>
                                            <span v-if="job.current_item && job.current_item !== job.target" class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-xs">
                                                {{ job.current_item }}
                                            </span>
                                        </div>
                                        <div class="flex items-center gap-2">
                                            <span class="text-xs font-mono text-slate-600 dark:text-slate-300">
                                                {{ formatJobProgressSummary(job) }}
                                            </span>
                                            <button
                                                v-if="canEdit && job.status === 'running'"
                                                type="button"
                                                @click="cancelModelJob(job.id)"
                                                class="ui-btn-secondary text-[11px] py-0.5 px-2 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                            >
                                                {{ t('classification.job_cancel') }}
                                            </button>
                                            <button
                                                v-else-if="canEdit"
                                                type="button"
                                                @click="dismissModelJob(job.id)"
                                                :title="t('classification.job_dismiss')"
                                                class="p-0.5 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded"
                                            >
                                                <span class="material-icons text-[15px]">close</span>
                                            </button>
                                        </div>
                                    </div>
                                    <div class="h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                                        <div
                                            v-if="isJobIndeterminate(job)"
                                            class="h-full bg-blue-600 dark:bg-blue-500 rounded-full ui-progress-indeterminate"
                                        ></div>
                                        <div
                                            v-else
                                            class="h-full rounded-full transition-all duration-200"
                                            :class="jobBarColorClass(job)"
                                            :style="{ width: jobBarWidth(job) }"
                                        ></div>
                                    </div>
                                    <div v-if="job.error" class="text-xs font-mono text-rose-600 dark:text-rose-400 break-all">
                                        {{ job.error }}
                                    </div>
                                </div>
                            </div>

                            <div class="overflow-x-auto rounded-b-lg">
                                <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                                    <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                                        <tr>
                                            <th class="px-4 py-2.5 text-left font-semibold">Model</th>
                                            <th class="px-4 py-2.5 text-left font-semibold w-28">Size</th>
                                            <th class="px-4 py-2.5 text-left font-semibold w-28">{{ t('common.status') }}</th>
                                            <th class="px-4 py-2.5 text-right font-semibold w-52">{{ t('common.actions') }}</th>
                                        </tr>
                                    </thead>
                                    <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                                        <tr v-if="!overview.onnx_local_models || overview.onnx_local_models.length === 0">
                                            <td colspan="4" class="px-4 py-8 text-center text-xs text-slate-400">
                                                {{ t('classification.onnx_empty') }}
                                            </td>
                                        </tr>
                                        <tr v-for="m in (overview.onnx_local_models || [])" :key="m.value" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                                            <td class="px-4 py-3">
                                                <div class="font-medium text-slate-900 dark:text-white flex items-center gap-2 flex-wrap">
                                                    <span>{{ m.label }}</span>
                                                    <span
                                                        v-if="activeONNXModel === m.value"
                                                        class="text-[10px] px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900 font-semibold"
                                                    >
                                                        {{ t('classification.onnx_active') }}
                                                    </span>
                                                </div>
                                                <div class="text-[11px] font-mono text-slate-400 mt-0.5">{{ m.hf_repo || m.value }}</div>
                                                <div v-if="m.desc" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{{ t(m.desc, m.desc) }}</div>
                                            </td>
                                            <td class="px-4 py-3 text-xs font-mono text-slate-600 dark:text-slate-300">
                                                {{ formatBytes(m.size_bytes, m.size_about) }}
                                            </td>
                                            <td class="px-4 py-3">
                                                <span
                                                    class="text-[11px] px-2 py-0.5 rounded border font-medium"
                                                    :class="m.ready
                                                        ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900'
                                                        : 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900'"
                                                >
                                                    {{ m.ready ? t('classification.onnx_env_ready') : t('classification.onnx_env_pending') }}
                                                </span>
                                            </td>
                                            <td class="px-4 py-3 text-right">
                                                <div class="inline-flex items-center gap-2">
                                                    <button
                                                        v-if="canEditSystem && m.ready && activeONNXModel !== m.value"
                                                        @click="setDefaultONNXModel(m.value)"
                                                        class="ui-btn-secondary text-xs py-1 px-2.5 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60"
                                                    >
                                                        {{ t('classification.onnx_use_model') }}
                                                    </button>
                                                    <button
                                                        v-if="canEdit"
                                                        @click="requestDeleteONNXModel(m.value)"
                                                        class="ui-btn-secondary text-xs py-1 px-2.5 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                                    >
                                                        {{ t('common.delete') }}
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </template>

                    <!-- Ollama & LLM Management Subsection (shown only when engine is LLM) -->
                    <div v-else-if="selectedProvider === 'llm'" class="border-t border-slate-200 dark:border-slate-800">
                        <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 bg-slate-50/40 dark:bg-slate-900/40">
                            <div class="flex items-center gap-2">
                                <span class="material-icons text-slate-500 dark:text-slate-400 text-[18px]">smart_toy</span>
                                <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                                    {{ t('classification.llm_title') }}
                                </span>
                            </div>
                            <div class="flex items-center gap-2">
                                <button
                                    @click="detectOllamaModels(false)"
                                    :disabled="detectingOllama"
                                    class="ui-btn-secondary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': detectingOllama }">
                                        {{ detectingOllama ? 'sync' : 'search' }}
                                    </span>
                                    <span>{{ t('classification.llm_refresh_ollama') }}</span>
                                </button>
                                <button
                                    v-if="canEdit"
                                    @click="openOllamaPullModal"
                                    class="ui-btn-primary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]">cloud_download</span>
                                    <span>{{ t('classification.llm_pull_btn') }}</span>
                                </button>
                            </div>
                        </div>

                        <div class="p-4 space-y-4">
                            <div v-if="ollamaPullJobs.length > 0" class="space-y-2.5">
                                <div
                                    v-for="job in ollamaPullJobs"
                                    :key="job.id"
                                    class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 space-y-2"
                                >
                                    <div class="flex flex-wrap items-center justify-between gap-2">
                                        <div class="flex items-center gap-2 min-w-0">
                                            <span class="text-[11px] px-1.5 py-0.5 rounded border font-medium" :class="jobStageBadgeClass(job)">
                                                {{ t('classification.job_stage.' + (job.stage || 'resolving'), job.stage || 'resolving') }}
                                            </span>
                                            <span class="font-mono text-xs font-semibold text-slate-900 dark:text-white truncate">{{ job.target }}</span>
                                            <span v-if="job.current_item && job.current_item !== job.target" class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-xs">
                                                {{ job.current_item }}
                                            </span>
                                        </div>
                                        <div class="flex items-center gap-2">
                                            <span class="text-xs font-mono text-slate-600 dark:text-slate-300">
                                                {{ formatJobProgressSummary(job) }}
                                            </span>
                                            <button
                                                v-if="canEdit && job.status === 'running'"
                                                type="button"
                                                @click="cancelModelJob(job.id)"
                                                class="ui-btn-secondary text-[11px] py-0.5 px-2 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                            >
                                                {{ t('classification.job_cancel') }}
                                            </button>
                                            <button
                                                v-else-if="canEdit"
                                                type="button"
                                                @click="dismissModelJob(job.id)"
                                                :title="t('classification.job_dismiss')"
                                                class="p-0.5 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded"
                                            >
                                                <span class="material-icons text-[15px]">close</span>
                                            </button>
                                        </div>
                                    </div>
                                    <div class="h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                                        <div
                                            v-if="isJobIndeterminate(job)"
                                            class="h-full bg-blue-600 dark:bg-blue-500 rounded-full ui-progress-indeterminate"
                                        ></div>
                                        <div
                                            v-else
                                            class="h-full rounded-full transition-all duration-200"
                                            :class="jobBarColorClass(job)"
                                            :style="{ width: jobBarWidth(job) }"
                                        ></div>
                                    </div>
                                    <div v-if="job.error" class="text-xs font-mono text-rose-600 dark:text-rose-400 break-all">
                                        {{ job.error }}
                                    </div>
                                </div>
                            </div>

                            <div>
                                <div class="text-xs font-semibold text-slate-700 dark:text-slate-300 mb-2">
                                    {{ t('classification.llm_installed') }}
                                </div>
                                <div v-if="ollamaModels.length === 0" class="text-xs text-slate-400 py-2">
                                    {{ t('common.no_data') }}
                                </div>
                                <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
                                    <button
                                        v-for="m in ollamaModels"
                                        :key="m.name"
                                        type="button"
                                        @click="selectOllamaModel(m.name)"
                                        class="text-left p-2.5 rounded-md border transition-colors"
                                        :class="engineForm.classification_llm_model === m.name
                                            ? 'bg-blue-50/70 border-blue-300 dark:bg-blue-950/30 dark:border-blue-800'
                                            : 'bg-slate-50/60 border-slate-200 hover:bg-slate-100/70 dark:bg-slate-950/40 dark:border-slate-800 dark:hover:bg-slate-800/60'"
                                    >
                                        <div class="font-mono text-xs font-semibold text-slate-900 dark:text-white truncate">{{ m.name }}</div>
                                        <div class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5 flex items-center gap-2">
                                            <span v-if="m.parameter_size">{{ m.parameter_size }}</span>
                                            <span v-if="m.quantization_level">{{ m.quantization_level }}</span>
                                            <span v-if="m.size">{{ formatBytes(m.size) }}</span>
                                        </div>
                                    </button>
                                </div>
                            </div>

                            <div v-if="(overview.llm_presets || []).length > 0" class="pt-3 border-t border-slate-200 dark:border-slate-800">
                                <div class="text-xs font-semibold text-slate-700 dark:text-slate-300 mb-2">
                                    {{ t('classification.llm_presets') }}
                                </div>
                                <div class="flex flex-wrap gap-1.5">
                                    <button
                                        v-for="preset in overview.llm_presets"
                                        :key="preset"
                                        type="button"
                                        @click="openOllamaPullWithPreset(preset)"
                                        class="px-2.5 py-1 rounded text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:border-blue-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                                    >
                                        {{ preset }}
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Subtab 2: Tag Profiles -->
            <div v-else-if="activeSubtab === 'profiles'" class="flex flex-col lg:flex-row gap-5 min-h-[520px]">
                <div class="w-full lg:w-72 flex-shrink-0 ui-card flex flex-col">
                    <div class="ui-card-header">
                        <span class="font-semibold text-slate-700 dark:text-slate-200 text-xs uppercase tracking-wider">
                            {{ t('classification.profiles_list') }}
                        </span>
                        <button
                            v-if="canEdit"
                            @click="openCreateProfileModal"
                            class="ui-btn-primary text-xs py-1 px-2.5"
                        >
                            <span class="material-icons text-[14px]">add</span>
                            <span>{{ t('classification.create_profile') }}</span>
                        </button>
                    </div>

                    <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
                        <div v-if="profiles.length === 0" class="p-6 text-center text-xs text-slate-400">
                            {{ t('classification.no_profiles') }}
                        </div>
                        <div
                            v-for="p in profiles"
                            :key="p.id"
                            @click="selectProfile(p)"
                            class="p-3 rounded-md border cursor-pointer transition-colors"
                            :class="profileDraft && profileDraft.id === p.id
                                ? 'bg-blue-50/70 border-blue-200 dark:bg-blue-950/30 dark:border-blue-900'
                                : 'border-transparent hover:bg-slate-50 dark:hover:bg-slate-800/50'"
                        >
                            <div class="flex items-center justify-between gap-2">
                                <span class="text-sm font-medium text-slate-900 dark:text-white truncate">{{ p.name }}</span>
                                <span
                                    v-if="p.is_default"
                                    class="text-[10px] px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900 font-medium flex-shrink-0"
                                >
                                    {{ t('classification.profile_default') }}
                                </span>
                            </div>
                            <p v-if="p.description" class="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">
                                {{ p.description }}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="flex-1 ui-card flex flex-col overflow-hidden">
                    <div v-if="!profileDraft" class="flex-1 flex flex-col items-center justify-center p-8 text-slate-400 text-xs">
                        {{ t('classification.no_profiles') }}
                    </div>

                    <template v-else>
                        <div class="ui-card-header flex-wrap">
                            <div class="flex items-center gap-2 min-w-0">
                                <span class="font-semibold text-slate-900 dark:text-white text-sm truncate">{{ profileDraft.name }}</span>
                                <span
                                    v-if="profileDraft.is_default"
                                    class="text-[10px] px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900 font-medium"
                                >
                                    {{ t('classification.profile_default') }}
                                </span>
                            </div>

                            <div v-if="canEdit" class="flex flex-wrap items-center gap-2">
                                <button
                                    @click="loadProfileTemplate('zh_CN')"
                                    :disabled="loadingTemplate"
                                    class="ui-btn-secondary text-xs py-1 px-2.5"
                                >
                                    {{ t('classification.profile_load_zh') }}
                                </button>
                                <button
                                    @click="loadProfileTemplate('en_US')"
                                    :disabled="loadingTemplate"
                                    class="ui-btn-secondary text-xs py-1 px-2.5"
                                >
                                    {{ t('classification.profile_load_en') }}
                                </button>
                                <button
                                    v-if="profiles.length > 1"
                                    @click="deleteProfileConfirm = true"
                                    class="ui-btn-secondary text-xs py-1 px-2.5 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                >
                                    {{ t('common.delete') }}
                                </button>
                                <button
                                    @click="saveActiveProfile"
                                    :disabled="savingProfile"
                                    class="ui-btn-primary text-xs py-1.5 px-3"
                                >
                                    <span class="material-icons text-[15px]" :class="{ 'animate-spin': savingProfile }">
                                        {{ savingProfile ? 'sync' : 'save' }}
                                    </span>
                                    <span>{{ savingProfile ? t('common.saving') : t('classification.profile_save') }}</span>
                                </button>
                            </div>
                        </div>

                        <div class="p-4 space-y-5 overflow-y-auto flex-1">
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                        {{ t('classification.profile_name') }}
                                    </label>
                                    <input
                                        v-model="profileDraft.name"
                                        type="text"
                                        :disabled="!canEdit"
                                        class="ui-input"
                                    >
                                </div>
                                <div>
                                    <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                                        {{ t('classification.profile_desc') }}
                                    </label>
                                    <input
                                        v-model="profileDraft.description"
                                        type="text"
                                        :disabled="!canEdit"
                                        class="ui-input"
                                    >
                                </div>
                            </div>

                            <label class="inline-flex items-center gap-2 cursor-pointer select-none">
                                <input
                                    type="checkbox"
                                    v-model="profileDraft.is_default"
                                    :disabled="!canEdit"
                                    class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                >
                                <span class="text-xs font-medium text-slate-700 dark:text-slate-300">
                                    {{ t('classification.profile_set_default') }}
                                </span>
                            </label>

                            <div class="flex items-center justify-between pt-2 border-t border-slate-200 dark:border-slate-800">
                                <span class="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-200">
                                    {{ t('classification.groups_title') }}
                                </span>
                                <button
                                    v-if="canEdit"
                                    @click="addGroupToDraft"
                                    class="ui-btn-secondary text-xs py-1 px-2.5"
                                >
                                    <span class="material-icons text-[14px]">add</span>
                                    <span>{{ t('classification.add_group') }}</span>
                                </button>
                            </div>

                            <div class="space-y-4">
                                <div
                                    v-for="(grp, gIdx) in profileDraft.groups"
                                    :key="gIdx"
                                    class="border border-slate-200 dark:border-slate-800 rounded-lg bg-slate-50/40 dark:bg-slate-950/30 p-3.5 space-y-3"
                                >
                                    <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end">
                                        <div class="md:col-span-3">
                                            <label class="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                                                {{ t('classification.group_name') }}
                                            </label>
                                            <input v-model="grp.name" type="text" :disabled="!canEdit" class="ui-input py-1.5 text-xs">
                                        </div>
                                        <div class="md:col-span-4">
                                            <label class="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                                                {{ t('classification.group_desc') }}
                                            </label>
                                            <input v-model="grp.description" type="text" :disabled="!canEdit" class="ui-input py-1.5 text-xs">
                                        </div>
                                        <div class="md:col-span-2">
                                            <label class="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                                                {{ t('classification.selection_mode') }}
                                            </label>
                                            <CustomSelect
                                                v-model="grp.selection_mode"
                                                :disabled="!canEdit"
                                                :options="[
                                                    { value: 'single', label: t('classification.mode.single') },
                                                    { value: 'limit', label: t('classification.mode.limit') },
                                                    { value: 'multi', label: t('classification.mode.multi') }
                                                ]"
                                            />
                                        </div>
                                        <div class="md:col-span-2">
                                            <label class="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                                                {{ t('classification.max_tags') }}
                                            </label>
                                            <input
                                                v-model.number="grp.max_results"
                                                type="number"
                                                min="1"
                                                :disabled="!canEdit || grp.selection_mode === 'single'"
                                                class="ui-input py-1.5 text-xs font-mono"
                                            >
                                        </div>
                                        <div v-if="canEdit" class="md:col-span-1 flex justify-end">
                                            <button
                                                type="button"
                                                @click="removeGroupFromDraft(gIdx)"
                                                :title="t('classification.delete_group')"
                                                class="p-1.5 rounded border border-rose-200 dark:border-rose-900/60 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30"
                                            >
                                                <span class="material-icons text-[16px]">delete_outline</span>
                                            </button>
                                        </div>
                                    </div>

                                    <div class="space-y-2 pt-2 border-t border-slate-200/80 dark:border-slate-800">
                                        <div
                                            v-for="(tag, tIdx) in grp.tags"
                                            :key="tIdx"
                                            class="grid grid-cols-1 md:grid-cols-12 gap-2 items-center bg-white dark:bg-slate-900 p-2 rounded border border-slate-200 dark:border-slate-800"
                                        >
                                            <div class="md:col-span-2">
                                                <input
                                                    v-model="tag.name"
                                                    type="text"
                                                    :disabled="!canEdit"
                                                    :placeholder="t('classification.tag_name')"
                                                    class="ui-input py-1 text-xs"
                                                >
                                            </div>
                                            <div class="md:col-span-4">
                                                <input
                                                    v-model="tag.description"
                                                    type="text"
                                                    :disabled="!canEdit"
                                                    :placeholder="t('classification.tag_desc')"
                                                    class="ui-input py-1 text-xs"
                                                >
                                            </div>
                                            <div class="md:col-span-4">
                                                <input
                                                    v-model="tag.keywordsText"
                                                    type="text"
                                                    :disabled="!canEdit"
                                                    :placeholder="t('classification.tag_keywords')"
                                                    class="ui-input py-1 text-xs"
                                                >
                                            </div>
                                            <div class="md:col-span-1">
                                                <input
                                                    v-model.number="tag.priority"
                                                    type="number"
                                                    :disabled="!canEdit"
                                                    :placeholder="t('classification.tag_priority')"
                                                    class="ui-input py-1 text-xs font-mono"
                                                >
                                            </div>
                                            <div v-if="canEdit" class="md:col-span-1 flex justify-end">
                                                <button
                                                    type="button"
                                                    @click="removeTagFromGroup(gIdx, tIdx)"
                                                    class="p-1 text-slate-400 hover:text-rose-500 rounded"
                                                >
                                                    <span class="material-icons text-[16px]">close</span>
                                                </button>
                                            </div>
                                        </div>

                                        <button
                                            v-if="canEdit"
                                            type="button"
                                            @click="addTagToGroup(gIdx)"
                                            class="ui-btn-secondary text-xs py-1 px-2.5"
                                        >
                                            <span class="material-icons text-[14px]">add</span>
                                            <span>{{ t('classification.add_tag') }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </template>
                </div>
            </div>

            <!-- Subtab 3: Serial Queue -->
            <div v-else-if="activeSubtab === 'queue'" class="space-y-4">
                <div class="ui-card p-4 space-y-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                        <div class="flex items-center gap-2.5">
                            <span class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('classification.queue_title') }}</span>
                            <span
                                class="text-[11px] px-2 py-0.5 rounded border font-medium"
                                :class="workerStateBadgeClass"
                            >
                                {{ t('classification.queue_state.' + (queueStatus.worker_state || 'idle'), queueStatus.worker_state || 'idle') }}
                            </span>
                        </div>

                        <div v-if="canEdit" class="flex flex-wrap items-center gap-2">
                            <button
                                @click="togglePauseQueue"
                                :disabled="queueBusy"
                                class="ui-btn-secondary text-xs py-1.5 px-3"
                            >
                                <span class="material-icons text-[15px]">
                                    {{ queueStatus.paused ? 'play_arrow' : 'pause' }}
                                </span>
                                <span>{{ queueStatus.paused ? t('classification.queue_resume') : t('classification.queue_pause') }}</span>
                            </button>
                            <button
                                @click="enqueueUntagged"
                                :disabled="queueBusy"
                                class="ui-btn-primary text-xs py-1.5 px-3"
                            >
                                <span class="material-icons text-[15px]">playlist_add</span>
                                <span>{{ t('classification.queue_enqueue_untagged') }}</span>
                            </button>
                            <button
                                @click="queueBulkAction('retry-failed')"
                                :disabled="queueBusy || !queueStatus.stats?.failed"
                                class="ui-btn-secondary text-xs py-1.5 px-3"
                            >
                                <span>{{ t('classification.queue_retry_failed') }}</span>
                            </button>
                            <button
                                @click="queueBulkAction('cancel-pending')"
                                :disabled="queueBusy || !queueStatus.stats?.pending"
                                class="ui-btn-secondary text-xs py-1.5 px-3 text-rose-600 dark:text-rose-400"
                            >
                                <span>{{ t('classification.queue_cancel_pending') }}</span>
                            </button>
                            <button
                                @click="queueBulkAction('clear-finished')"
                                :disabled="queueBusy"
                                class="ui-btn-secondary text-xs py-1.5 px-3"
                            >
                                <span>{{ t('classification.queue_clear_finished') }}</span>
                            </button>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2.5">
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_pending') }}</div>
                            <div class="text-lg font-semibold font-mono text-slate-900 dark:text-white">{{ queueStatus.stats?.pending || 0 }}</div>
                        </div>
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_running') }}</div>
                            <div class="text-lg font-semibold font-mono text-blue-600 dark:text-blue-400">{{ queueStatus.stats?.running || 0 }}</div>
                        </div>
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_completed') }}</div>
                            <div class="text-lg font-semibold font-mono text-emerald-600 dark:text-emerald-400">{{ queueStatus.stats?.completed || 0 }}</div>
                        </div>
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_failed') }}</div>
                            <div class="text-lg font-semibold font-mono text-rose-600 dark:text-rose-400">{{ queueStatus.stats?.failed || 0 }}</div>
                        </div>
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_skipped') }}</div>
                            <div class="text-lg font-semibold font-mono text-slate-600 dark:text-slate-300">{{ queueStatus.stats?.skipped || 0 }}</div>
                        </div>
                        <div class="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                            <div class="text-[11px] text-slate-500">{{ t('classification.queue_cancelled') }}</div>
                            <div class="text-lg font-semibold font-mono text-slate-500">{{ queueStatus.stats?.cancelled || 0 }}</div>
                        </div>
                    </div>
                </div>

                <div class="ui-card overflow-hidden">
                    <div class="px-4 py-2.5 border-b border-slate-200 dark:border-slate-800 flex flex-wrap gap-2 bg-slate-50/70 dark:bg-slate-900/60">
                        <button
                            v-for="st in queueStatusFilters"
                            :key="st.value"
                            @click="changeQueueFilter(st.value)"
                            class="px-2.5 py-1 rounded text-xs font-medium border transition-colors"
                            :class="queueFilter === st.value
                                ? 'bg-slate-900 text-white border-slate-900 dark:bg-blue-600 dark:border-blue-600'
                                : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800'"
                        >
                            {{ st.label }}
                        </button>
                    </div>

                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                            <thead class="bg-slate-50 dark:bg-slate-900/80 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
                                <tr>
                                    <th class="px-4 py-2.5 text-left font-semibold">{{ t('classification.queue_col_article') }}</th>
                                    <th class="px-4 py-2.5 text-left font-semibold w-28">{{ t('classification.queue_col_status') }}</th>
                                    <th class="px-4 py-2.5 text-left font-semibold w-52">{{ t('classification.queue_col_engine') }}</th>
                                    <th class="px-4 py-2.5 text-left font-semibold w-40">{{ t('classification.queue_col_time') }}</th>
                                    <th class="px-4 py-2.5 text-right font-semibold w-36">{{ t('common.actions') }}</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
                                <tr v-if="queueTasks.length === 0">
                                    <td colspan="5" class="px-4 py-8 text-center text-xs text-slate-400">
                                        {{ t('classification.no_tasks') }}
                                    </td>
                                </tr>
                                <tr v-for="task in queueTasks" :key="task.id" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                                    <td class="px-4 py-3 max-w-md">
                                        <button
                                            type="button"
                                            @click="openTaskDetail(task.id)"
                                            class="text-left font-medium text-slate-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors truncate block w-full"
                                        >
                                            {{ plainText(task.article_title) || `#${task.result_id}` }}
                                        </button>
                                        <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono mt-0.5">
                                            {{ task.article_source || task.module_id }}
                                        </div>
                                        <div v-if="task.error_message" class="text-xs text-rose-600 dark:text-rose-400 mt-1 break-all line-clamp-2">
                                            {{ task.error_message }}
                                        </div>
                                    </td>
                                    <td class="px-4 py-3">
                                        <span class="text-[11px] px-2 py-0.5 rounded border font-medium" :class="taskStatusClass(task.status)">
                                            {{ t('classification.queue_' + task.status, task.status) }}
                                        </span>
                                    </td>
                                    <td class="px-4 py-3 text-xs text-slate-600 dark:text-slate-300">
                                        <div class="font-mono">{{ task.provider || '-' }} <span v-if="task.model_name">/ {{ task.model_name }}</span></div>
                                        <div class="text-[11px] text-slate-400 mt-0.5">Tags: {{ task.assigned_tags_count || 0 }}</div>
                                    </td>
                                    <td class="px-4 py-3 text-xs font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap">
                                        {{ formatTime(task.finished_at || task.started_at || task.created_at) }}
                                    </td>
                                    <td class="px-4 py-3 text-right whitespace-nowrap">
                                        <div class="inline-flex items-center gap-1.5">
                                            <button
                                                type="button"
                                                @click="openTaskDetail(task.id)"
                                                class="ui-btn-secondary text-xs py-1 px-2 text-slate-700 dark:text-slate-200"
                                            >
                                                {{ t('classification.queue_details') }}
                                            </button>
                                            <button
                                                v-if="canEdit && task.status === 'pending'"
                                                @click="cancelTask(task.id)"
                                                class="ui-btn-secondary text-xs py-1 px-2 text-rose-600 dark:text-rose-400"
                                            >
                                                {{ t('classification.queue_cancel') }}
                                            </button>
                                            <button
                                                v-else-if="canEdit && (task.status === 'failed' || task.status === 'cancelled' || task.status === 'skipped')"
                                                @click="retryTask(task.id)"
                                                class="ui-btn-secondary text-xs py-1 px-2 text-blue-600 dark:text-blue-400"
                                            >
                                                {{ t('classification.queue_retry') }}
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <div v-if="queueTotalPages > 1" class="px-4 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500">
                        <span>{{ t('common.page_info').replace('{page}', String(queuePage)).replace('{total}', String(queueTotalPages)) }}</span>
                        <div class="flex gap-2">
                            <button @click="changeQueuePage(queuePage - 1)" :disabled="queuePage <= 1" class="ui-btn-secondary text-xs py-1 px-2.5">
                                {{ t('common.prev') }}
                            </button>
                            <button @click="changeQueuePage(queuePage + 1)" :disabled="queuePage >= queueTotalPages" class="ui-btn-secondary text-xs py-1 px-2.5">
                                {{ t('common.next') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Download ONNX Model Modal -->
        <transition name="modal">
            <div v-if="onnxModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden">
                    <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
                        <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('classification.onnx_download_btn') }}</h3>
                        <button @click="onnxModal.show = false" class="p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded">
                            <span class="material-icons text-[18px]">close</span>
                        </button>
                    </div>

                    <div class="p-5 overflow-y-auto space-y-5 flex-1">
                        <div class="flex flex-col sm:flex-row gap-2">
                            <input
                                v-model="onnxModal.customRepo"
                                type="text"
                                :placeholder="t('classification.onnx_custom_repo')"
                                class="ui-input flex-1 font-mono text-xs"
                            >
                            <button
                                @click="downloadONNXRepo(onnxModal.customRepo)"
                                :disabled="onnxModal.submitting || !onnxModal.customRepo.trim() || isONNXRepoRunning(onnxModal.customRepo)"
                                class="ui-btn-primary text-xs py-2 px-3.5 flex-shrink-0"
                            >
                                <span class="material-icons text-[15px]" :class="{ 'animate-spin': onnxModal.submitting || isONNXRepoRunning(onnxModal.customRepo) }">
                                    {{ (onnxModal.submitting || isONNXRepoRunning(onnxModal.customRepo)) ? 'sync' : 'download' }}
                                </span>
                                <span>{{ isONNXRepoRunning(onnxModal.customRepo) ? t('classification.onnx_downloading') : t('classification.onnx_download_btn') }}</span>
                            </button>
                        </div>

                        <div
                            v-if="activeONNXModalJob"
                            class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 space-y-2"
                        >
                            <div class="flex flex-wrap items-center justify-between gap-2">
                                <div class="flex items-center gap-2 min-w-0">
                                    <span class="text-[11px] px-1.5 py-0.5 rounded border font-medium" :class="jobStageBadgeClass(activeONNXModalJob)">
                                        {{ t('classification.job_stage.' + (activeONNXModalJob.stage || 'resolving'), activeONNXModalJob.stage || 'resolving') }}
                                    </span>
                                    <span class="font-mono text-xs font-semibold text-slate-900 dark:text-white truncate">{{ activeONNXModalJob.target }}</span>
                                    <span v-if="activeONNXModalJob.current_item && activeONNXModalJob.current_item !== activeONNXModalJob.target" class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-[180px]">
                                        {{ activeONNXModalJob.current_item }}
                                    </span>
                                </div>
                                <div class="flex items-center gap-2">
                                    <span class="text-xs font-mono text-slate-600 dark:text-slate-300">
                                        {{ formatJobProgressSummary(activeONNXModalJob) }}
                                    </span>
                                    <button
                                        v-if="canEdit && activeONNXModalJob.status === 'running'"
                                        type="button"
                                        @click="cancelModelJob(activeONNXModalJob.id)"
                                        class="ui-btn-secondary text-[11px] py-0.5 px-2 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                    >
                                        {{ t('classification.job_cancel') }}
                                    </button>
                                </div>
                            </div>
                            <div class="h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                                <div
                                    v-if="isJobIndeterminate(activeONNXModalJob)"
                                    class="h-full bg-blue-600 dark:bg-blue-500 rounded-full ui-progress-indeterminate"
                                ></div>
                                <div
                                    v-else
                                    class="h-full rounded-full transition-all duration-200"
                                    :class="jobBarColorClass(activeONNXModalJob)"
                                    :style="{ width: jobBarWidth(activeONNXModalJob) }"
                                ></div>
                            </div>
                            <div v-if="activeONNXModalJob.error" class="text-xs font-mono text-rose-600 dark:text-rose-400 break-all">
                                {{ activeONNXModalJob.error }}
                            </div>
                        </div>

                        <div v-for="grp in (overview.onnx_presets || [])" :key="grp.group" class="space-y-2">
                            <div class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                                {{ t(grp.group, grp.group) }}
                            </div>
                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                                <div
                                    v-for="item in (grp.models || [])"
                                    :key="item.hf_repo"
                                    class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 flex flex-col justify-between gap-2.5"
                                >
                                    <div>
                                        <div class="flex items-center justify-between gap-2">
                                            <span class="text-xs font-semibold text-slate-900 dark:text-white">{{ item.name }}</span>
                                            <span class="text-[11px] font-mono text-slate-400">{{ item.size_about }}</span>
                                        </div>
                                        <div class="text-[11px] font-mono text-slate-400 mt-0.5 truncate">{{ item.hf_repo }}</div>
                                        <p class="text-xs text-slate-600 dark:text-slate-300 mt-1">{{ t(item.desc, item.desc) }}</p>
                                    </div>
                                    <div class="flex justify-end">
                                        <button
                                            @click="downloadONNXRepo(item.hf_repo)"
                                            :disabled="onnxModal.submitting || isONNXRepoRunning(item.hf_repo)"
                                            class="ui-btn-secondary text-xs py-1 px-2.5 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-900/60"
                                        >
                                            <span>{{ isONNXRepoRunning(item.hf_repo) ? t('classification.onnx_downloading') : t('classification.onnx_download_btn') }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex justify-end bg-slate-50/60 dark:bg-slate-950/30">
                        <button @click="onnxModal.show = false" class="ui-btn-secondary text-xs">{{ t('common.close') }}</button>
                    </div>
                </div>
            </div>
        </transition>

        <!-- Pull Ollama Model Modal -->
        <transition name="modal">
            <div v-if="ollamaPullModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 w-full max-w-md p-5 space-y-4">
                    <div class="flex items-center justify-between">
                        <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('classification.llm_pull_modal') }}</h3>
                        <button @click="ollamaPullModal.show = false" class="p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded">
                            <span class="material-icons text-[18px]">close</span>
                        </button>
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                            {{ t('config.classification.llm_url.desc') }}
                        </label>
                        <input v-model="ollamaPullModal.url" type="text" class="ui-input font-mono text-xs">
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                            {{ t('classification.llm_pull_custom') }}
                        </label>
                        <input v-model="ollamaPullModal.model" type="text" class="ui-input font-mono text-xs">
                    </div>

                    <div v-if="(overview.llm_presets || []).length > 0" class="flex flex-wrap gap-1.5">
                        <button
                            v-for="m in overview.llm_presets"
                            :key="m"
                            type="button"
                            @click="ollamaPullModal.model = m"
                            class="px-2 py-0.5 text-[11px] font-mono rounded border transition-colors"
                            :class="ollamaPullModal.model === m
                                ? 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900'
                                : 'bg-slate-50 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700'"
                        >
                            {{ m }}
                        </button>
                    </div>

                    <div
                        v-if="activeOllamaModalJob"
                        class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 space-y-2"
                    >
                        <div class="flex flex-wrap items-center justify-between gap-2">
                            <div class="flex items-center gap-2 min-w-0">
                                <span class="text-[11px] px-1.5 py-0.5 rounded border font-medium" :class="jobStageBadgeClass(activeOllamaModalJob)">
                                    {{ t('classification.job_stage.' + (activeOllamaModalJob.stage || 'resolving'), activeOllamaModalJob.stage || 'resolving') }}
                                </span>
                                <span class="font-mono text-xs font-semibold text-slate-900 dark:text-white truncate">{{ activeOllamaModalJob.target }}</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <span class="text-xs font-mono text-slate-600 dark:text-slate-300">
                                    {{ formatJobProgressSummary(activeOllamaModalJob) }}
                                </span>
                                <button
                                    v-if="canEdit && activeOllamaModalJob.status === 'running'"
                                    type="button"
                                    @click="cancelModelJob(activeOllamaModalJob.id)"
                                    class="ui-btn-secondary text-[11px] py-0.5 px-2 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-900/60"
                                >
                                    {{ t('classification.job_cancel') }}
                                </button>
                            </div>
                        </div>
                        <div v-if="activeOllamaModalJob.current_item && activeOllamaModalJob.current_item !== activeOllamaModalJob.target" class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate">
                            {{ activeOllamaModalJob.current_item }}
                        </div>
                        <div class="h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                            <div
                                v-if="isJobIndeterminate(activeOllamaModalJob)"
                                class="h-full bg-blue-600 dark:bg-blue-500 rounded-full ui-progress-indeterminate"
                            ></div>
                            <div
                                v-else
                                class="h-full rounded-full transition-all duration-200"
                                :class="jobBarColorClass(activeOllamaModalJob)"
                                :style="{ width: jobBarWidth(activeOllamaModalJob) }"
                            ></div>
                        </div>
                        <div v-if="activeOllamaModalJob.error" class="text-xs font-mono text-rose-600 dark:text-rose-400 break-all">
                            {{ activeOllamaModalJob.error }}
                        </div>
                    </div>

                    <div class="flex justify-end gap-2 pt-2">
                        <button @click="ollamaPullModal.show = false" class="ui-btn-secondary text-xs">{{ t('common.close') }}</button>
                        <button
                            @click="pullOllamaModel"
                            :disabled="ollamaPullModal.submitting || !ollamaPullModal.model.trim() || isOllamaModelRunning(ollamaPullModal.model)"
                            class="ui-btn-primary text-xs"
                        >
                            <span class="material-icons text-[15px]" :class="{ 'animate-spin': ollamaPullModal.submitting || isOllamaModelRunning(ollamaPullModal.model) }">
                                {{ (ollamaPullModal.submitting || isOllamaModelRunning(ollamaPullModal.model)) ? 'sync' : 'cloud_download' }}
                            </span>
                            <span>{{ isOllamaModelRunning(ollamaPullModal.model) ? t('classification.llm_pulling') : t('classification.llm_pull_btn') }}</span>
                        </button>
                    </div>
                </div>
            </div>
        </transition>

        <!-- Create Profile Modal -->
        <transition name="modal">
            <div v-if="createProfileModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 w-full max-w-md p-5 space-y-4">
                    <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('classification.create_profile') }}</h3>
                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                            {{ t('classification.profile_name') }}
                        </label>
                        <input v-model="createProfileModal.name" type="text" class="ui-input">
                    </div>
                    <div>
                        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                            {{ t('classification.profile_desc') }}
                        </label>
                        <input v-model="createProfileModal.description" type="text" class="ui-input">
                    </div>
                    <label class="inline-flex items-center gap-2 cursor-pointer">
                        <input
                            type="checkbox"
                            v-model="createProfileModal.is_default"
                            class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 h-4 w-4"
                        >
                        <span class="text-xs text-slate-700 dark:text-slate-300">{{ t('classification.profile_set_default') }}</span>
                    </label>
                    <div class="flex justify-end gap-2 pt-2">
                        <button @click="createProfileModal.show = false" class="ui-btn-secondary text-xs">{{ t('common.cancel') }}</button>
                        <button @click="createProfile" :disabled="creatingProfile" class="ui-btn-primary text-xs">
                            {{ creatingProfile ? t('common.saving') : t('common.create') }}
                        </button>
                    </div>
                </div>
            </div>
        </transition>

        <ConfirmModal
            :show="deleteProfileConfirm"
            :title="t('common.delete')"
            :message="t('classification.profile_delete_confirm')"
            :t="t"
            @confirm="deleteSelectedProfile"
            @cancel="deleteProfileConfirm = false"
        />

        <ConfirmModal
            :show="deleteONNXConfirm.show"
            :title="t('common.delete')"
            :message="deleteONNXConfirm.name"
            :t="t"
            @confirm="performDeleteONNXModel"
            @cancel="deleteONNXConfirm.show = false"
        />

        <!-- Classification Queue Task Detail Modal -->
        <transition name="modal">
            <div v-if="taskDetailModal.show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4">
                <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 w-full max-w-3xl max-h-[90vh] flex flex-col overflow-hidden">
                    <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-3">
                        <div class="flex items-center gap-2.5 min-w-0">
                            <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ t('classification.task_detail_title') }}</h3>
                            <span
                                v-if="taskDetailModal.task"
                                class="text-[11px] px-2 py-0.5 rounded border font-medium"
                                :class="taskStatusClass(taskDetailModal.task.status)"
                            >
                                {{ t('classification.queue_' + taskDetailModal.task.status, taskDetailModal.task.status) }}
                            </span>
                        </div>
                        <button @click="closeTaskDetail" class="p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded">
                            <span class="material-icons text-[18px]">close</span>
                        </button>
                    </div>

                    <div v-if="taskDetailModal.loading" class="p-12 text-center text-slate-400 text-xs flex-1">
                        <span class="material-icons animate-spin text-2xl text-blue-600 mb-2 block">sync</span>
                        <span>{{ t('common.loading') }}</span>
                    </div>

                    <div v-else-if="taskDetailModal.task" class="p-5 overflow-y-auto space-y-5 flex-1">
                        <div>
                            <h2 class="text-base font-semibold text-slate-900 dark:text-white leading-snug">
                                {{ plainText(taskDetailModal.task.article?.title || taskDetailModal.task.article_title) || `#${taskDetailModal.task.result_id}` }}
                            </h2>
                            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500 dark:text-slate-400 mt-1.5">
                                <span>{{ taskDetailModal.task.article?.source || taskDetailModal.task.article_source || '-' }}</span>
                                <span class="font-mono">{{ taskDetailModal.task.module_id }}</span>
                                <span v-if="taskDetailModal.task.article?.author">{{ taskDetailModal.task.article.author }}</span>
                                <a
                                    v-if="safeUrl(taskDetailModal.task.article?.from_url)"
                                    :href="safeUrl(taskDetailModal.task.article.from_url)"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:underline"
                                >
                                    <span>{{ t('data.open_original') }}</span>
                                    <span class="material-icons text-[13px]">open_in_new</span>
                                </a>
                            </div>
                        </div>

                        <!-- Execution Metadata -->
                        <div class="space-y-2">
                            <div class="text-xs font-semibold uppercase tracking-wider text-slate-600 dark:text-slate-300">
                                {{ t('classification.task_detail_meta') }}
                            </div>
                            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-xs">
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_id') }}</span>
                                    <span class="font-mono font-semibold text-slate-800 dark:text-slate-200">#{{ taskDetailModal.task.id }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.queue_col_engine') }}</span>
                                    <span class="font-mono font-semibold text-slate-800 dark:text-slate-200 break-all">
                                        {{ taskDetailModal.task.provider || '-' }}<template v-if="taskDetailModal.task.model_name"> / {{ taskDetailModal.task.model_name }}</template>
                                    </span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_attempts') }}</span>
                                    <span class="font-mono font-semibold text-slate-800 dark:text-slate-200">{{ taskDetailModal.task.attempts || 0 }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_source') }}</span>
                                    <span class="font-mono font-semibold text-slate-800 dark:text-slate-200 truncate block">{{ taskDetailModal.task.module_id || '-' }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 sm:col-span-2">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_created') }}</span>
                                    <span class="font-mono text-slate-700 dark:text-slate-300">{{ formatTime(taskDetailModal.task.created_at) }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_started') }}</span>
                                    <span class="font-mono text-slate-700 dark:text-slate-300">{{ formatTime(taskDetailModal.task.started_at) }}</span>
                                </div>
                                <div class="p-2.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40">
                                    <span class="text-slate-400 block text-[11px]">{{ t('classification.task_detail_finished') }}</span>
                                    <span class="font-mono text-slate-700 dark:text-slate-300">{{ formatTime(taskDetailModal.task.finished_at) }}</span>
                                </div>
                            </div>
                        </div>

                        <!-- Error Diagnostic -->
                        <div
                            v-if="taskDetailModal.task.error_message"
                            class="p-3.5 rounded-lg border border-rose-200 dark:border-rose-900/70 bg-rose-50/70 dark:bg-rose-950/30 space-y-1.5"
                        >
                            <div class="text-xs font-semibold text-rose-700 dark:text-rose-300 flex items-center gap-1.5">
                                <span class="material-icons text-[16px]">error_outline</span>
                                <span>{{ t('classification.task_detail_error') }}</span>
                            </div>
                            <pre class="text-xs font-mono text-rose-700 dark:text-rose-300 whitespace-pre-wrap break-all">{{ taskDetailModal.task.error_message }}</pre>
                        </div>

                        <!-- Assigned Tags -->
                        <div class="space-y-2">
                            <div class="text-xs font-semibold uppercase tracking-wider text-slate-600 dark:text-slate-300">
                                {{ t('classification.task_detail_tags') }}
                            </div>
                            <div
                                v-if="!taskDetailModal.task.assigned_tags || taskDetailModal.task.assigned_tags.length === 0"
                                class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 text-xs text-slate-500 dark:text-slate-400"
                            >
                                {{ t('classification.task_detail_tags_empty') }}
                            </div>
                            <div v-else class="space-y-2">
                                <div
                                    v-for="tag in taskDetailModal.task.assigned_tags"
                                    :key="tag.tag_id"
                                    class="p-3 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 space-y-1"
                                >
                                    <div class="flex flex-wrap items-center justify-between gap-2">
                                        <div class="flex items-center gap-2">
                                            <span class="text-xs font-semibold text-slate-900 dark:text-white">
                                                <span v-if="tag.group_name" class="text-slate-400 font-normal">{{ tag.group_name }} / </span>{{ tag.tag_name || `#${tag.tag_id}` }}
                                            </span>
                                            <span class="text-[11px] font-mono px-1.5 py-0.5 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300">
                                                {{ tag.provider || 'auto' }}
                                            </span>
                                        </div>
                                        <span class="text-xs font-mono font-semibold text-blue-600 dark:text-blue-400">
                                            {{ t('classification.task_detail_confidence') }}: {{ Math.round((Number(tag.confidence) || 0) * 100) }}%
                                        </span>
                                    </div>
                                    <div v-if="tag.reason" class="text-xs text-slate-600 dark:text-slate-300">
                                        <span class="text-slate-400 mr-1">{{ t('classification.task_detail_reason') }}:</span>{{ tag.reason }}
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Article Body Preview -->
                        <div v-if="taskDetailModal.task.article" class="space-y-2">
                            <div class="text-xs font-semibold uppercase tracking-wider text-slate-600 dark:text-slate-300">
                                {{ t('classification.task_detail_article') }}
                            </div>
                            <div class="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-950/30 max-h-72 overflow-y-auto text-sm text-slate-700 dark:text-slate-200">
                                <div
                                    v-if="isTaskArticleHtml(taskDetailModal.task.article)"
                                    class="leading-relaxed space-y-2 overflow-x-auto"
                                    v-html="cleanHtml(taskDetailModal.task.article.content)"
                                ></div>
                                <div
                                    v-else-if="isTaskArticleMarkdown(taskDetailModal.task.article)"
                                    class="leading-relaxed space-y-2"
                                    v-html="renderMd(taskDetailModal.task.article.content || taskDetailModal.task.article.summary || '')"
                                ></div>
                                <div v-else class="leading-relaxed whitespace-pre-wrap break-words">
                                    {{ taskDetailModal.task.article.content || taskDetailModal.task.article.summary || '' }}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 flex justify-end gap-2">
                        <button
                            v-if="canEdit && taskDetailModal.task && (taskDetailModal.task.status === 'failed' || taskDetailModal.task.status === 'cancelled' || taskDetailModal.task.status === 'skipped')"
                            @click="retryFromDetailModal(taskDetailModal.task.id)"
                            class="ui-btn-primary text-xs py-1.5 px-3"
                        >
                            {{ t('classification.queue_retry') }}
                        </button>
                        <button @click="closeTaskDetail" class="ui-btn-secondary text-xs">{{ t('common.close') }}</button>
                    </div>
                </div>
            </div>
        </transition>
    </div>
</template>

<script>
import http, { formatAPIError } from '../../utils/http.js';
import {
    stripMarkdownPlain,
    renderMarkdown,
    sanitizeRichHtml,
    sanitizeHttpUrl
} from '../../utils/markdown.js';
import CustomSelect from '../../components/CustomSelect.vue';
import ConfirmModal from '../../components/ConfirmModal.vue';

function cloneProfileForDraft(profile) {
    if (!profile) return null;
    return {
        id: profile.id,
        name: profile.name || '',
        description: profile.description || '',
        is_default: Boolean(profile.is_default),
        groups: (profile.groups || []).map(g => ({
            id: g.id || 0,
            name: g.name || '',
            description: g.description || '',
            selection_mode: g.selection_mode || 'limit',
            max_results: Number(g.max_results) > 0 ? Number(g.max_results) : 1,
            tags: (g.tags || []).map(tag => ({
                id: tag.id || 0,
                name: tag.name || '',
                description: tag.description || '',
                priority: Number(tag.priority) || 0,
                keywordsText: (tag.keywords || [])
                    .map(k => (typeof k === 'string' ? k : k?.keyword || ''))
                    .filter(Boolean)
                    .join(', ')
            }))
        }))
    };
}

function serializeDraftToSaveInput(draft) {
    return {
        name: String(draft.name || '').trim(),
        description: String(draft.description || '').trim(),
        is_default: Boolean(draft.is_default),
        groups: (draft.groups || []).map(g => {
            const mode = ['single', 'limit', 'multi'].includes(g.selection_mode) ? g.selection_mode : 'limit';
            const maxRes = mode === 'single' ? 1 : (Math.max(1, parseInt(g.max_results, 10) || 1));
            return {
                id: g.id || 0,
                name: String(g.name || '').trim(),
                description: String(g.description || '').trim(),
                selection_mode: mode,
                max_results: maxRes,
                tags: (g.tags || []).map(tag => ({
                    id: tag.id || 0,
                    name: String(tag.name || '').trim(),
                    description: String(tag.description || '').trim(),
                    priority: parseInt(tag.priority, 10) || 0,
                    keywords: String(tag.keywordsText || '')
                        .split(/[,，\n]/)
                        .map(s => s.trim())
                        .filter(Boolean)
                }))
            };
        })
    };
}

export default {
    components: { CustomSelect, ConfirmModal },
    props: ['t', 'currentUser', 'currentLocale', 'subtab'],
    emits: ['update:subtab'],
    inject: ['showToast'],
    data() {
        const initialSubtab = ['engine', 'profiles', 'queue'].includes(this.subtab) ? this.subtab : 'engine';
        return {
            activeSubtab: initialSubtab,
            loading: false,
            initialized: false,

            overview: {},
            engineConfigs: [],
            engineForm: {},
            savingEngine: false,
            startingONNXEnv: false,

            modelJobs: [],
            knownJobStatus: {},
            jobPollTimer: null,
            pollingJobs: false,

            onnxModal: {
                show: false,
                customRepo: '',
                submitting: false,
                activeJobId: ''
            },
            deleteONNXConfirm: {
                show: false,
                name: ''
            },

            ollamaModels: [],
            detectingOllama: false,
            ollamaPullModal: {
                show: false,
                url: 'http://127.0.0.1:11434',
                model: '',
                submitting: false,
                activeJobId: ''
            },

            profiles: [],
            profileDraft: null,
            savingProfile: false,
            loadingTemplate: false,
            creatingProfile: false,
            deleteProfileConfirm: false,
            createProfileModal: {
                show: false,
                name: '',
                description: '',
                is_default: false
            },

            queueStatus: {
                worker_state: 'idle',
                paused: false,
                current_task_id: 0,
                stats: {}
            },
            queueTasks: [],
            queuePage: 1,
            queuePageSize: 20,
            queueTotal: 0,
            queueFilter: 'all',
            queueBusy: false,
            queueTimer: null,
            taskDetailModal: {
                show: false,
                loading: false,
                task: null
            }
        };
    },
    computed: {
        canEdit() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('classification.edit');
        },
        canViewSystem() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('system.config.view');
        },
        canEditSystem() {
            const perms = this.currentUser?.permissions || [];
            return this.currentUser?.role_name === 'admin' || perms.includes('system.config.edit');
        },
        selectedProvider() {
            if (this.engineConfigs.length > 0 && this.engineForm.classification_provider) {
                return String(this.engineForm.classification_provider);
            }
            return String(this.overview.system_provider || 'none');
        },
        visibleEngineConfigs() {
            const provider = this.selectedProvider;
            return (this.engineConfigs || []).filter(conf => {
                if (conf.key === 'classification_provider') return true;
                if (provider === 'none') return false;
                if (
                    conf.key === 'classification_confidence' ||
                    conf.key === 'classification_max_input_chars' ||
                    conf.key === 'classification_truncate_strategy'
                ) {
                    return true;
                }
                if (provider === 'onnx') {
                    return conf.key === 'classification_onnx_model';
                }
                if (provider === 'llm') {
                    return conf.key.startsWith('classification_llm_');
                }
                return false;
            });
        },
        onnxModelSelectOptions() {
            const opts = [{ value: '', label: this.t('classification.select_onnx_placeholder') }];
            const seen = new Set();
            (this.overview.onnx_local_models || []).forEach(m => {
                if (!m?.value || seen.has(m.value)) return;
                seen.add(m.value);
                const label = m.ready ? m.label : `${m.label} · ${this.t('classification.onnx_env_pending')}`;
                opts.push({ value: m.value, label });
            });
            const cur = String(this.engineForm.classification_onnx_model || '').trim();
            if (cur && !seen.has(cur)) {
                opts.push({ value: cur, label: cur });
            }
            return opts;
        },
        activeONNXModel() {
            return String(this.engineForm.classification_onnx_model || '').trim();
        },
        onnxEnvJobs() {
            return (this.modelJobs || []).filter(j => j.kind === 'onnx_env');
        },
        activeONNXEnvRunning() {
            return this.onnxEnvJobs.some(j => j.status === 'running');
        },
        onnxDownloadJobs() {
            return (this.modelJobs || []).filter(j => j.kind === 'onnx_download');
        },
        ollamaPullJobs() {
            return (this.modelJobs || []).filter(j => j.kind === 'ollama_pull');
        },
        activeONNXModalJob() {
            if (this.onnxModal.activeJobId) {
                const found = this.onnxDownloadJobs.find(j => j.id === this.onnxModal.activeJobId);
                if (found) return found;
            }
            return this.onnxDownloadJobs.find(j => j.status === 'running') || null;
        },
        activeOllamaModalJob() {
            if (this.ollamaPullModal.activeJobId) {
                const found = this.ollamaPullJobs.find(j => j.id === this.ollamaPullModal.activeJobId);
                if (found) return found;
            }
            const target = String(this.ollamaPullModal.model || '').trim();
            if (target) {
                const byTarget = this.ollamaPullJobs.find(j => j.target === target && j.status === 'running');
                if (byTarget) return byTarget;
            }
            return this.ollamaPullJobs.find(j => j.status === 'running') || null;
        },
        queuePendingTotal() {
            const st = this.queueStatus?.stats || {};
            return (st.pending || 0) + (st.running || 0);
        },
        queueTotalPages() {
            return Math.max(1, Math.ceil((this.queueTotal || 0) / this.queuePageSize));
        },
        workerStateBadgeClass() {
            const st = this.queueStatus?.worker_state || 'idle';
            if (st === 'running') {
                return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
            }
            if (st === 'paused') {
                return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900';
            }
            return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
        },
        queueStatusFilters() {
            return [
                { value: 'all', label: this.t('common.all') },
                { value: 'pending', label: this.t('classification.queue_pending') },
                { value: 'running', label: this.t('classification.queue_running') },
                { value: 'completed', label: this.t('classification.queue_completed') },
                { value: 'failed', label: this.t('classification.queue_failed') },
                { value: 'skipped', label: this.t('classification.queue_skipped') },
                { value: 'cancelled', label: this.t('classification.queue_cancelled') }
            ];
        }
    },
    watch: {
        subtab(next) {
            const target = ['engine', 'profiles', 'queue'].includes(next) ? next : 'engine';
            if (target !== this.activeSubtab) {
                this.applySubtab(target);
            }
        },
        'engineForm.classification_provider'(next) {
            if (this.initialized && next === 'llm' && String(this.engineForm.classification_llm_protocol || 'ollama') !== 'openai') {
                this.detectOllamaModels(true);
            }
        },
        'engineForm.classification_llm_protocol'(next) {
            if (this.initialized && next !== 'openai' && this.selectedProvider === 'llm') {
                this.detectOllamaModels(true);
            }
        }
    },
    async mounted() {
        await this.loadInitialData();
    },
    unmounted() {
        this.stopQueuePolling();
        this.stopJobPolling();
    },
    methods: {
        formatOptions(options) {
            return (options || []).map(opt => {
                if (typeof opt === 'object' && opt !== null) {
                    return {
                        value: opt.value,
                        label: opt.label_key ? this.t(opt.label_key, opt.label || opt.value) : (opt.label || opt.value)
                    };
                }
                return { value: opt, label: String(opt) };
            });
        },
        formatBytes(bytes, fallbackText) {
            const b = Number(bytes) || 0;
            if (b <= 0) return fallbackText || '-';
            if (b >= 1024 * 1024 * 1024) {
                return (b / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
            }
            if (b >= 1024 * 1024) {
                return (b / (1024 * 1024)).toFixed(1) + ' MB';
            }
            if (b >= 1024) {
                return (b / 1024).toFixed(0) + ' KB';
            }
            return b + ' B';
        },
        formatTime(iso) {
            if (!iso) return '-';
            const d = new Date(iso);
            if (Number.isNaN(d.getTime())) return iso;
            return d.toLocaleString();
        },
        isONNXRepoRunning(repoID) {
            const target = String(repoID || '').trim();
            if (!target) return false;
            return this.onnxDownloadJobs.some(j => j.target === target && j.status === 'running');
        },
        isOllamaModelRunning(modelName) {
            const target = String(modelName || '').trim();
            if (!target) return false;
            return this.ollamaPullJobs.some(j => j.target === target && j.status === 'running');
        },
        isJobIndeterminate(job) {
            if (!job) return false;
            if (job.status !== 'running') return false;
            return Boolean(job.indeterminate) || Number(job.percent) < 0;
        },
        jobBarWidth(job) {
            if (!job) return '0%';
            if (job.status === 'completed') return '100%';
            const pct = Number(job.percent);
            if (Number.isNaN(pct) || pct < 0) return '0%';
            return Math.min(100, Math.max(0, pct)).toFixed(1) + '%';
        },
        jobBarColorClass(job) {
            switch (job?.status) {
                case 'completed':
                    return 'bg-emerald-600 dark:bg-emerald-500';
                case 'failed':
                    return 'bg-rose-600 dark:bg-rose-500';
                case 'cancelled':
                    return 'bg-slate-400 dark:bg-slate-600';
                default:
                    return 'bg-blue-600 dark:bg-blue-500';
            }
        },
        jobStageBadgeClass(job) {
            switch (job?.status) {
                case 'completed':
                    return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900';
                case 'failed':
                    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
                case 'cancelled':
                    return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
                default:
                    return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
            }
        },
        formatJobProgressSummary(job) {
            if (!job) return '';
            const parts = [];
            const pct = Number(job.percent);
            if (job.status === 'completed') {
                parts.push('100%');
            } else if (!this.isJobIndeterminate(job) && !Number.isNaN(pct) && pct >= 0) {
                parts.push(Math.min(100, Math.max(0, pct)).toFixed(1) + '%');
            }
            if (job.total_bytes > 0) {
                parts.push(`${this.formatBytes(job.completed_bytes)} / ${this.formatBytes(job.total_bytes)}`);
            } else if (job.completed_bytes > 0) {
                parts.push(this.formatBytes(job.completed_bytes));
            }
            if (parts.length === 0 && job.status === 'running') {
                return this.t('classification.job_calculating');
            }
            return parts.join(' · ');
        },
        syncJobs(nextJobs, detectTransitions = true) {
            const list = Array.isArray(nextJobs) ? nextJobs : [];
            const nextStatusMap = {};
            let shouldRefreshOverview = false;
            let shouldRefreshOllama = false;

            list.forEach(job => {
                if (!job || !job.id) return;
                nextStatusMap[job.id] = job.status;
                if (!detectTransitions) return;
                const prev = this.knownJobStatus[job.id];
                if (prev === 'running' && job.status !== 'running') {
                    if (job.status === 'completed') {
                        const label = job.target || this.t('classification.job_kind.' + job.kind, job.kind);
                        this.showToast(this.t('classification.job_completed').replace('{target}', label), 'success');
                        shouldRefreshOverview = true;
                        if (job.kind === 'ollama_pull') {
                            shouldRefreshOllama = true;
                        }
                    } else if (job.status === 'failed') {
                        const fallback = this.t('classification.job_failed').replace('{target}', job.target || job.kind);
                        this.showToast(job.error || fallback, 'error');
                    } else if (job.status === 'cancelled') {
                        const label = job.target || this.t('classification.job_kind.' + job.kind, job.kind);
                        this.showToast(this.t('classification.job_cancelled').replace('{target}', label), 'success');
                    }
                }
            });

            this.knownJobStatus = nextStatusMap;
            this.modelJobs = list;

            if (list.some(j => j.status === 'running')) {
                this.startJobPolling();
            } else {
                this.stopJobPolling();
            }

            if (shouldRefreshOverview) {
                this.fetchEngineData(false, true);
            }
            if (shouldRefreshOllama) {
                this.detectOllamaModels(true);
            }
        },
        upsertJob(job) {
            if (!job || !job.id) return;
            const filtered = (this.modelJobs || []).filter(j => j.id !== job.id);
            const nextList = [job, ...filtered];
            this.knownJobStatus = {
                ...this.knownJobStatus,
                [job.id]: job.status
            };
            this.modelJobs = nextList;
            if (nextList.some(j => j.status === 'running')) {
                this.startJobPolling();
            }
        },
        startJobPolling() {
            if (this.jobPollTimer) return;
            this.jobPollTimer = setInterval(() => {
                this.pollModelJobs();
            }, 700);
        },
        stopJobPolling() {
            if (this.jobPollTimer) {
                clearInterval(this.jobPollTimer);
                this.jobPollTimer = null;
            }
        },
        async pollModelJobs() {
            if (this.pollingJobs) return;
            this.pollingJobs = true;
            try {
                const res = await http.get('/api/dashboard/classification/jobs');
                if (Array.isArray(res.data)) {
                    this.syncJobs(res.data, true);
                }
            } catch (_) {
                // Keep polling silently across transient errors while task is running
            } finally {
                this.pollingJobs = false;
            }
        },
        async cancelModelJob(id) {
            if (!id) return;
            try {
                const res = await http.post(`/api/dashboard/classification/jobs/${encodeURIComponent(id)}/cancel`);
                if (res.data) {
                    this.upsertJob(res.data);
                }
                const target = res.data?.target || id;
                this.showToast(this.t('classification.job_cancelled').replace('{target}', target), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async dismissModelJob(id) {
            if (!id) return;
            try {
                await http.delete(`/api/dashboard/classification/jobs/${encodeURIComponent(id)}`);
                this.modelJobs = (this.modelJobs || []).filter(j => j.id !== id);
                const nextMap = { ...this.knownJobStatus };
                delete nextMap[id];
                this.knownJobStatus = nextMap;
                if (this.onnxModal.activeJobId === id) {
                    this.onnxModal.activeJobId = '';
                }
                if (this.ollamaPullModal.activeJobId === id) {
                    this.ollamaPullModal.activeJobId = '';
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        taskStatusClass(status) {
            switch (status) {
                case 'completed':
                    return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900';
                case 'running':
                    return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
                case 'failed':
                    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
                case 'pending':
                    return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900';
                default:
                    return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
            }
        },
        switchSubtab(tab) {
            const target = ['engine', 'profiles', 'queue'].includes(tab) ? tab : 'engine';
            this.$emit('update:subtab', target);
            this.applySubtab(target);
        },
        applySubtab(tab) {
            this.activeSubtab = tab;
            if (tab === 'queue') {
                this.fetchQueueData();
                this.startQueuePolling();
            } else {
                this.stopQueuePolling();
                if (tab === 'profiles' && this.profiles.length === 0) {
                    this.fetchProfiles();
                } else if (tab === 'engine') {
                    this.fetchEngineData();
                }
            }
        },
        async refreshCurrent() {
            if (this.activeSubtab === 'queue') {
                await this.fetchQueueData();
            } else if (this.activeSubtab === 'profiles') {
                await this.fetchProfiles();
            } else {
                await this.fetchEngineData();
            }
        },
        async loadInitialData() {
            this.loading = true;
            try {
                const tasks = [
                    this.fetchEngineData(false),
                    this.fetchProfiles(false)
                ];
                if (this.activeSubtab === 'queue') {
                    tasks.push(this.fetchQueueData(true));
                }
                await Promise.all(tasks);
                this.initialized = true;
                if (this.activeSubtab === 'queue') {
                    this.startQueuePolling();
                }
            } finally {
                this.loading = false;
            }
        },
        async fetchEngineData(withSpinner = true, preserveForm = false) {
            if (withSpinner) this.loading = true;
            try {
                const reqs = [http.get('/api/dashboard/classification/overview')];
                if (!preserveForm && this.canViewSystem) {
                    reqs.push(http.get('/api/dashboard/system-config').catch(() => ({ data: [] })));
                } else {
                    reqs.push(Promise.resolve({ data: null }));
                }
                const [ovRes, sysRes] = await Promise.all(reqs);
                this.overview = ovRes.data || {};
                if (this.overview.queue) {
                    this.queueStatus = this.overview.queue;
                }
                if (Array.isArray(this.overview.jobs)) {
                    this.syncJobs(this.overview.jobs, this.initialized);
                }
                if (Array.isArray(sysRes.data)) {
                    const clsFields = sysRes.data
                        .filter(c => c.tab === 'classification')
                        .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
                    this.engineConfigs = clsFields;
                    const form = {};
                    clsFields.forEach(c => {
                        form[c.key] = c.value ?? '';
                    });
                    this.engineForm = form;
                }
                if (this.selectedProvider === 'llm' && String(this.engineForm.classification_llm_protocol || 'ollama') !== 'openai') {
                    this.detectOllamaModels(true);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                if (withSpinner) this.loading = false;
            }
        },
        async saveEngineSettings() {
            const configs = {};
            this.engineConfigs.forEach(conf => {
                if (!conf.is_editable) return;
                const val = String(this.engineForm[conf.key] ?? '');
                if (conf.is_secret && val === '********') return;
                configs[conf.key] = val;
            });
            this.savingEngine = true;
            try {
                await http.put('/api/dashboard/system-config', { configs });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchEngineData(false);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingEngine = false;
            }
        },
        async setDefaultONNXModel(modelName) {
            try {
                await http.put('/api/dashboard/system-config', {
                    configs: { classification_onnx_model: modelName }
                });
                this.engineForm.classification_onnx_model = modelName;
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchEngineData(false, true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async prepareONNXEnv() {
            this.startingONNXEnv = true;
            try {
                const res = await http.post('/api/dashboard/classification/onnx/prepare-env');
                if (res.data?.job) {
                    this.upsertJob(res.data.job);
                }
                this.showToast(this.t('classification.job_started'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.startingONNXEnv = false;
            }
        },
        openONNXDownloadModal() {
            const running = this.onnxDownloadJobs.find(j => j.status === 'running');
            this.onnxModal.customRepo = '';
            this.onnxModal.activeJobId = running ? running.id : '';
            this.onnxModal.show = true;
        },
        async downloadONNXRepo(repoID) {
            const target = String(repoID || '').trim();
            if (!target) return;
            this.onnxModal.submitting = true;
            try {
                const res = await http.post('/api/dashboard/classification/onnx/models/download', { repo_id: target });
                if (res.data?.job) {
                    this.upsertJob(res.data.job);
                    this.onnxModal.activeJobId = res.data.job.id;
                }
                this.showToast(this.t('classification.job_started'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.onnxModal.submitting = false;
            }
        },
        requestDeleteONNXModel(name) {
            this.deleteONNXConfirm = { show: true, name };
        },
        async performDeleteONNXModel() {
            const name = this.deleteONNXConfirm.name;
            this.deleteONNXConfirm.show = false;
            if (!name) return;
            try {
                await http.delete(`/api/dashboard/classification/onnx/models/${encodeURIComponent(name)}`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchEngineData(false, true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async detectOllamaModels(silent = false) {
            this.detectingOllama = true;
            try {
                const url = String(this.engineForm.classification_llm_url || '').trim();
                const res = await http.get('/api/dashboard/classification/llm/ollama/models', {
                    params: url ? { url } : {}
                });
                this.ollamaModels = Array.isArray(res.data) ? res.data : [];
            } catch (err) {
                if (!silent) {
                    this.showToast(formatAPIError(err, this.t), 'error');
                }
            } finally {
                this.detectingOllama = false;
            }
        },
        selectOllamaModel(name) {
            if (this.canEditSystem && this.engineConfigs.length > 0) {
                this.engineForm.classification_llm_model = name;
            }
        },
        openOllamaPullModal() {
            const running = this.ollamaPullJobs.find(j => j.status === 'running');
            this.ollamaPullModal = {
                show: true,
                url: String(this.engineForm.classification_llm_url || 'http://127.0.0.1:11434'),
                model: String(running?.target || this.engineForm.classification_llm_model || this.overview.llm_presets?.[0] || 'qwen3.5:0.8b'),
                submitting: false,
                activeJobId: running ? running.id : ''
            };
        },
        openOllamaPullWithPreset(modelName) {
            const matchRunning = this.ollamaPullJobs.find(j => j.target === modelName && j.status === 'running');
            this.ollamaPullModal = {
                show: true,
                url: String(this.engineForm.classification_llm_url || 'http://127.0.0.1:11434'),
                model: modelName,
                submitting: false,
                activeJobId: matchRunning ? matchRunning.id : ''
            };
        },
        async pullOllamaModel() {
            const model = String(this.ollamaPullModal.model || '').trim();
            if (!model) return;
            this.ollamaPullModal.submitting = true;
            try {
                const res = await http.post('/api/dashboard/classification/llm/ollama/pull', {
                    url: String(this.ollamaPullModal.url || '').trim(),
                    model
                });
                if (res.data?.job) {
                    this.upsertJob(res.data.job);
                    this.ollamaPullModal.activeJobId = res.data.job.id;
                }
                this.showToast(this.t('classification.job_started'), 'success');
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.ollamaPullModal.submitting = false;
            }
        },
        async fetchProfiles(withSpinner = true) {
            if (withSpinner) this.loading = true;
            try {
                const res = await http.get('/api/dashboard/classification/profiles');
                this.profiles = Array.isArray(res.data) ? res.data : [];
                if (this.profiles.length > 0) {
                    const currentId = this.profileDraft?.id;
                    const found = (currentId && this.profiles.find(p => p.id === currentId)) || this.profiles[0];
                    this.profileDraft = cloneProfileForDraft(found);
                } else {
                    this.profileDraft = null;
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                if (withSpinner) this.loading = false;
            }
        },
        selectProfile(p) {
            this.profileDraft = cloneProfileForDraft(p);
        },
        openCreateProfileModal() {
            this.createProfileModal = {
                show: true,
                name: '',
                description: '',
                is_default: false
            };
        },
        async createProfile() {
            const name = String(this.createProfileModal.name || '').trim();
            if (!name) {
                this.showToast(this.t('classification.profile_name_required'), 'warning');
                return;
            }
            this.creatingProfile = true;
            try {
                const res = await http.post('/api/dashboard/classification/profiles', {
                    name,
                    description: String(this.createProfileModal.description || '').trim(),
                    is_default: Boolean(this.createProfileModal.is_default),
                    groups: []
                });
                this.createProfileModal.show = false;
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchProfiles(false);
                if (res.data) {
                    this.profileDraft = cloneProfileForDraft(res.data);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.creatingProfile = false;
            }
        },
        async saveActiveProfile() {
            if (!this.profileDraft) return;
            const payload = serializeDraftToSaveInput(this.profileDraft);
            if (!payload.name) {
                this.showToast(this.t('classification.profile_name_required'), 'warning');
                return;
            }
            this.savingProfile = true;
            try {
                const res = await http.put(`/api/dashboard/classification/profiles/${this.profileDraft.id}`, payload);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchProfiles(false);
                if (res.data) {
                    this.profileDraft = cloneProfileForDraft(res.data);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.savingProfile = false;
            }
        },
        async loadProfileTemplate(locale) {
            if (!this.profileDraft) return;
            this.loadingTemplate = true;
            try {
                const res = await http.post(`/api/dashboard/classification/profiles/${this.profileDraft.id}/load-default`, { locale });
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchProfiles(false);
                if (res.data) {
                    this.profileDraft = cloneProfileForDraft(res.data);
                }
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.loadingTemplate = false;
            }
        },
        async deleteSelectedProfile() {
            this.deleteProfileConfirm = false;
            if (!this.profileDraft) return;
            try {
                await http.delete(`/api/dashboard/classification/profiles/${this.profileDraft.id}`);
                this.showToast(this.t('common.saved_success'), 'success');
                this.profileDraft = null;
                await this.fetchProfiles(false);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        addGroupToDraft() {
            if (!this.profileDraft) return;
            this.profileDraft.groups.push({
                id: 0,
                name: '',
                description: '',
                selection_mode: 'limit',
                max_results: 2,
                tags: []
            });
        },
        removeGroupFromDraft(gIdx) {
            if (!this.profileDraft) return;
            this.profileDraft.groups.splice(gIdx, 1);
        },
        addTagToGroup(gIdx) {
            const grp = this.profileDraft?.groups?.[gIdx];
            if (!grp) return;
            grp.tags.push({
                id: 0,
                name: '',
                description: '',
                priority: 0,
                keywordsText: ''
            });
        },
        removeTagFromGroup(gIdx, tIdx) {
            const grp = this.profileDraft?.groups?.[gIdx];
            if (!grp) return;
            grp.tags.splice(tIdx, 1);
        },
        async fetchQueueData(silent = false) {
            if (!silent && !this.initialized) this.loading = true;
            try {
                const [statusRes, tasksRes] = await Promise.all([
                    http.get('/api/dashboard/classification/queue/status'),
                    http.get('/api/dashboard/classification/queue/tasks', {
                        params: {
                            page: this.queuePage,
                            page_size: this.queuePageSize,
                            status: this.queueFilter
                        }
                    })
                ]);
                if (statusRes.data) {
                    this.queueStatus = statusRes.data;
                }
                const data = tasksRes.data || {};
                this.queueTasks = Array.isArray(data.items) ? data.items : [];
                this.queueTotal = Number(data.total) || 0;
            } catch (err) {
                if (!silent) {
                    this.showToast(formatAPIError(err, this.t), 'error');
                }
            } finally {
                if (!silent && !this.initialized) this.loading = false;
            }
        },
        startQueuePolling() {
            this.stopQueuePolling();
            this.queueTimer = setInterval(() => {
                if (this.activeSubtab === 'queue') {
                    this.fetchQueueData(true);
                }
            }, 4000);
        },
        stopQueuePolling() {
            if (this.queueTimer) {
                clearInterval(this.queueTimer);
                this.queueTimer = null;
            }
        },
        changeQueueFilter(filter) {
            this.queueFilter = filter;
            this.queuePage = 1;
            this.fetchQueueData();
        },
        changeQueuePage(page) {
            if (page < 1 || page > this.queueTotalPages) return;
            this.queuePage = page;
            this.fetchQueueData();
        },
        async togglePauseQueue() {
            this.queueBusy = true;
            try {
                const ep = this.queueStatus.paused ? 'resume' : 'pause';
                await http.post(`/api/dashboard/classification/queue/${ep}`);
                await this.fetchQueueData(true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.queueBusy = false;
            }
        },
        async enqueueUntagged() {
            this.queueBusy = true;
            try {
                const res = await http.post('/api/dashboard/classification/queue/enqueue-untagged', { limit: 500 });
                const added = res.data?.added ?? 0;
                this.showToast(this.t('classification.queue_enqueued').replace('{count}', String(added)), 'success');
                await this.fetchQueueData(true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.queueBusy = false;
            }
        },
        async queueBulkAction(action) {
            this.queueBusy = true;
            try {
                await http.post(`/api/dashboard/classification/queue/${action}`);
                this.showToast(this.t('common.saved_success'), 'success');
                await this.fetchQueueData(true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            } finally {
                this.queueBusy = false;
            }
        },
        async cancelTask(id) {
            try {
                await http.post(`/api/dashboard/classification/queue/tasks/${id}/cancel`);
                await this.fetchQueueData(true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async retryTask(id) {
            try {
                await http.post(`/api/dashboard/classification/queue/tasks/${id}/retry`);
                await this.fetchQueueData(true);
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
            }
        },
        async openTaskDetail(id) {
            if (!id) return;
            this.taskDetailModal = {
                show: true,
                loading: true,
                task: null
            };
            try {
                const res = await http.get(`/api/dashboard/classification/queue/tasks/${id}`);
                this.taskDetailModal.task = res.data || null;
            } catch (err) {
                this.showToast(formatAPIError(err, this.t), 'error');
                this.taskDetailModal.show = false;
            } finally {
                this.taskDetailModal.loading = false;
            }
        },
        closeTaskDetail() {
            this.taskDetailModal.show = false;
            this.taskDetailModal.task = null;
        },
        async retryFromDetailModal(id) {
            await this.retryTask(id);
            await this.openTaskDetail(id);
        },
        plainText: stripMarkdownPlain,
        renderMd: renderMarkdown,
        cleanHtml: sanitizeRichHtml,
        safeUrl: sanitizeHttpUrl,
        isTaskArticleHtml(article) {
            return article?.render_type === 'html' || article?.content_type === 'html';
        },
        isTaskArticleMarkdown(article) {
            return !this.isTaskArticleHtml(article) && (article?.content_type === 'markdown' || article?.render_type === 'markdown');
        }
    }
};
</script>

