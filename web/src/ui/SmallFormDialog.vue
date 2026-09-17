<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{ open: boolean; title: string; titleId: string }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement>()
let opener: HTMLElement | null = null

watch(() => props.open, async open => {
  await nextTick()
  if (open && !dialog.value?.open) {
    opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
    dialog.value?.showModal()
    dialog.value?.querySelector<HTMLElement>('[data-initial-focus]')?.focus()
  } else if (!open && dialog.value?.open) {
    dialog.value.close()
    if (opener?.isConnected) opener.focus()
  }
}, { immediate: true })
onBeforeUnmount(() => dialog.value?.close())
</script>

<template>
  <dialog ref="dialog" class="small-form-dialog" :aria-labelledby="titleId" @cancel.prevent="emit('close')">
    <div class="small-form-shell">
      <header class="small-form-header">
        <h2 :id="titleId">{{ title }}</h2>
        <button type="button" class="button secondary" aria-label="Cerrar formulario" @click="emit('close')">Cerrar</button>
      </header>
      <div class="small-form-body"><slot /></div>
      <footer class="small-form-footer"><slot name="actions" /></footer>
    </div>
  </dialog>
</template>

<style scoped>
.small-form-dialog { position: fixed; inset: 0 0 0 auto; margin: 0; width: min(540px, 100%); max-width: 100%; height: 100dvh; max-height: 100dvh; padding: 0; border: 0; border-left: 1px solid var(--line); background: var(--paper); color: var(--ink); box-shadow: -12px 0 48px #10282426; }
.small-form-dialog::backdrop { background: #10282466; }
.small-form-shell { display: flex; flex-direction: column; height: 100%; }
.small-form-header, .small-form-footer { flex: 0 0 auto; display: flex; align-items: center; gap: 12px; padding: 20px 24px; background: var(--paper); }
.small-form-header { justify-content: space-between; border-bottom: 1px solid var(--line); }
.small-form-header h2 { margin: 0; font-size: 1.4rem; letter-spacing: -.03em; }
.small-form-body { flex: 1; min-height: 0; overflow-y: auto; padding: 24px; overscroll-behavior: contain; }
.small-form-footer { justify-content: flex-end; border-top: 1px solid var(--line); padding-bottom: max(20px, env(safe-area-inset-bottom)); }
@media (max-width: 480px) { .small-form-header, .small-form-body, .small-form-footer { padding-left: 16px; padding-right: 16px; } }
</style>
