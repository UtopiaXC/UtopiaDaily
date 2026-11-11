<template>
    <div>
        <div
            @click="state === 'idle' || state === 'error' ? verify() : null"
            :class="[
                'flex items-center justify-between px-3 py-2 rounded-md border text-xs font-medium transition select-none',
                state === 'verified'
                    ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-300 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300'
                    : state === 'error'
                        ? 'bg-rose-50 dark:bg-rose-950/40 border-rose-300 dark:border-rose-800 text-rose-700 dark:text-rose-300 cursor-pointer'
                        : 'bg-slate-50 dark:bg-slate-950 border-slate-300 dark:border-slate-700 text-slate-700 dark:text-slate-200 cursor-pointer hover:border-blue-500'
            ]"
        >
            <div class="flex items-center gap-2">
                <span
                    class="material-icons text-[16px]"
                    :class="[
                        state === 'solving' ? 'animate-spin text-blue-500' : '',
                        state === 'verified' ? 'text-emerald-600 dark:text-emerald-400' : '',
                        state === 'error' ? 'text-rose-500' : '',
                        state === 'idle' ? 'text-slate-400' : ''
                    ]"
                >
                    {{ iconName }}
                </span>
                <span>{{ statusLabel }}</span>
            </div>

            <button
                v-if="state === 'verified' || state === 'error'"
                type="button"
                @click.stop="resetAndVerify"
                class="p-0.5 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded transition"
                :title="t('common.refresh')"
            >
                <span class="material-icons text-[15px]">refresh</span>
            </button>
        </div>
    </div>
</template>

<script>
import http from '../utils/http.js';
import { solveAltchaChallenge } from '../utils/pow.js';

export default {
    props: {
        modelValue: {
            type: String,
            default: ''
        },
        t: {
            type: Function,
            required: true
        }
    },
    emits: ['update:modelValue'],
    data() {
        return {
            state: 'idle',
            abortController: null
        };
    },
    computed: {
        iconName() {
            switch (this.state) {
                case 'solving': return 'sync';
                case 'verified': return 'check_circle';
                case 'error': return 'error_outline';
                default: return 'verified_user';
            }
        },
        statusLabel() {
            switch (this.state) {
                case 'solving': return this.t('auth.pow.solving');
                case 'verified': return this.t('auth.pow.verified');
                case 'error': return this.t('auth.pow.failed');
                default: return this.t('auth.pow.idle');
            }
        }
    },
    beforeUnmount() {
        if (this.abortController) {
            this.abortController.abort();
        }
    },
    methods: {
        reset() {
            if (this.abortController) {
                this.abortController.abort();
                this.abortController = null;
            }
            this.state = 'idle';
            this.$emit('update:modelValue', '');
        },
        async resetAndVerify() {
            this.reset();
            await this.verify();
        },
        async verify() {
            if (this.state === 'solving') return;
            if (this.abortController) {
                this.abortController.abort();
            }
            this.abortController = new AbortController();
            this.state = 'solving';
            this.$emit('update:modelValue', '');

            try {
                const res = await http.get('/api/auth/captcha');
                const solved = await solveAltchaChallenge(res.data, this.abortController.signal);
                this.state = 'verified';
                this.$emit('update:modelValue', solved);
                return true;
            } catch (err) {
                if (err.name === 'AbortError') return false;
                this.state = 'error';
                this.$emit('update:modelValue', '');
                return false;
            }
        }
    }
};
</script>
