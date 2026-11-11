<template>
    <div class="fixed bottom-4 right-4 z-[70] flex flex-col gap-2">
        <transition-group name="toast">
            <div
                v-for="toast in toasts"
                :key="toast.id"
                :class="[
                    'px-3.5 py-2.5 rounded-md text-xs font-medium flex items-center min-w-[220px] max-w-sm border',
                    toast.type === 'success'
                        ? 'bg-slate-900 text-white border-slate-800 dark:bg-slate-800 dark:border-slate-700'
                        : 'bg-rose-600 text-white border-rose-700'
                ]"
            >
                <span
                    class="material-icons text-[16px] mr-2"
                    :class="toast.type === 'success' ? 'text-emerald-400' : 'text-white'"
                >
                    {{ toast.type === 'success' ? 'check_circle' : 'error_outline' }}
                </span>
                <span class="break-words">{{ toast.message }}</span>
            </div>
        </transition-group>
    </div>
</template>

<script>
export default {
    data() {
        return {
            toasts: [],
            counter: 0
        };
    },
    methods: {
        add(message, type = 'success') {
            const id = this.counter++;
            this.toasts.push({ id, message, type });
            setTimeout(() => {
                this.remove(id);
            }, 3500);
        },
        remove(id) {
            const index = this.toasts.findIndex(t => t.id === id);
            if (index !== -1) {
                this.toasts.splice(index, 1);
            }
        }
    }
};
</script>
