<template>
    <div>
        <div
            @click="state === 'idle' || state === 'error' ? verify() : null"
            :class="[
                'flex items-center justify-between px-3 py-2.5 rounded-md border text-sm transition select-none',
                state === 'verified'
                    ? 'bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-700 text-green-700 dark:text-green-300'
                    : state === 'error'
                        ? 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-700 text-red-700 dark:text-red-300 cursor-pointer'
                        : 'bg-gray-50 dark:bg-gray-700/60 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 cursor-pointer hover:border-blue-500'
            ]"
        >
            <div class="flex items-center gap-2.5">
                <span
                    class="material-icons text-lg"
                    :class="[
                        state === 'solving' ? 'animate-spin text-blue-500' : '',
                        state === 'verified' ? 'text-green-600 dark:text-green-400' : '',
                        state === 'error' ? 'text-red-500' : '',
                        state === 'idle' ? 'text-gray-500 dark:text-gray-400' : ''
                    ]"
                >
                    {{ iconName }}
                </span>
                <span class="font-medium">{{ statusLabel }}</span>
            </div>

            <button
                v-if="state === 'verified' || state === 'error'"
                type="button"
                @click.stop="resetAndVerify"
                class="p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 rounded transition"
                :title="t('common.refresh')"
            >
                <span class="material-icons text-base">refresh</span>
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
}
</script>
