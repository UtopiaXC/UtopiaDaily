<template>
    <div class="fixed bottom-4 right-4 z-[70] flex flex-col gap-2">
        <transition-group name="toast">
            <div v-for="toast in toasts" :key="toast.id"
                 :class="['px-4 py-3 rounded-lg shadow-lg text-white text-sm flex items-center min-w-[220px] max-w-sm',
                          toast.type === 'success' ? 'bg-green-600' : 'bg-red-600']">
                <span class="material-icons text-sm mr-2">{{ toast.type === 'success' ? 'check_circle' : 'error' }}</span>
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
        }
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
}
</script>
