<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import Message from 'primevue/message'

const props = defineProps<{ notice: string; error: string }>()
const emit = defineEmits<{ closeNotice: []; closeError: [] }>()
let timer: number | undefined
watch(() => props.notice, value => {
  if (timer) window.clearTimeout(timer)
  if (value) timer = window.setTimeout(() => emit('closeNotice'), 5000)
})
onBeforeUnmount(() => { if (timer) window.clearTimeout(timer) })
</script>

<template>
  <div class="message-center" aria-live="polite">
    <Message v-if="notice" severity="success" closable class="toast success" @close="emit('closeNotice')">{{ notice }}</Message>
    <Message v-if="error" severity="error" closable class="toast error" @close="emit('closeError')">{{ error }}</Message>
  </div>
</template>

<style>
.message-center { position: fixed; z-index: 1200; top: 20px; right: 24px; display: grid; gap: 10px; width: min(390px, calc(100vw - 32px)); pointer-events: none; }
.toast { display: grid; grid-template-columns: 26px 1fr 24px; align-items: start; gap: 10px; padding: 13px 14px; border: 1px solid; border-radius: 10px; box-shadow: 0 12px 30px #10282420; pointer-events: auto; animation: toast-in .18s ease-out; }
.toast.success { border-color: #b8d7c2; background: #f2faf4; color: #245d38; }
.toast.error { border-color: #e5b9aa; background: #fff7f4; color: #873d2e; }
.toast .p-message-icon { display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: currentColor; color: var(--paper); font-size: .78rem; }
.toast .p-message-text { padding-top: 3px; font-size: .82rem; line-height: 1.45; }
.toast .p-message-close { width: 24px; height: 24px; padding: 0; border: 0; background: transparent; color: currentColor; font-size: 1.25rem; line-height: 1; cursor: pointer; }
@keyframes toast-in { from { opacity: 0; transform: translateY(-6px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 600px) { .message-center { top: 12px; right: 16px; } }
</style>
