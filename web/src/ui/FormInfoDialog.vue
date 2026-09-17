<script setup lang="ts">
import { ref } from 'vue'
import Dialog from 'primevue/dialog'
defineProps<{
  open: boolean
  title: string
  identifier: string
  module: string
  description: string
  permissions: string[]
  userPermissions: string[]
  specialPermissions?: string[]
}>()
const emit = defineEmits<{ close: [] }>()
const tab = ref<'access' | 'special'>('access')
</script>

<template>
  <Dialog :visible="open" modal :closable="false" :dismissable-mask="true" :show-header="false" :pt="{ mask: 'form-info-dialog-mask', root: 'form-info-dialog', content: 'form-info-dialog-content' }" aria-labelledby="form-info-title" @hide="emit('close')">
    <div class="form-info-card">
      <header><div><p class="eyebrow">Información del formulario</p><h2 id="form-info-title">{{ title }}</h2></div><button class="icon-close" aria-label="Cerrar información" title="Cerrar" @click="emit('close')">×</button></header>
      <div class="form-info-content">
        <dl class="form-metadata"><div><dt>Identificador</dt><dd><code>{{ identifier }}</code></dd></div><div><dt>Módulo</dt><dd>{{ module }}</dd></div></dl>
        <p class="form-description">{{ description }}</p>
        <nav class="info-tabs" aria-label="Información de acceso"><button :class="{ active: tab === 'access' }" @click="tab = 'access'">Acceso</button><button v-if="specialPermissions?.length" :class="{ active: tab === 'special' }" @click="tab = 'special'">Permisos especiales</button></nav>
        <div v-if="tab === 'access'" class="permission-grid"><section><h3>Permisos del formulario</h3><ul><li v-for="permission in permissions" :key="permission"><span class="permission-dot" aria-hidden="true">✓</span>{{ permission }}</li></ul></section><section><h3>Permisos de este usuario</h3><ul><li v-for="permission in userPermissions" :key="permission"><span class="permission-dot" aria-hidden="true">✓</span>{{ permission }}</li></ul></section></div>
        <section v-else class="special-list"><h3>Permisos especiales asignados</h3><ul><li v-for="permission in specialPermissions" :key="permission"><span class="permission-dot" aria-hidden="true">✓</span><code>{{ permission }}</code></li></ul></section>
      </div>
      <footer><button class="button primary" @click="emit('close')">Cerrar</button></footer>
    </div>
  </Dialog>
</template>

<style>
.form-info-dialog-mask { background: #10282488; backdrop-filter: blur(3px); }
.form-info-dialog { width: min(640px, calc(100vw - 32px)); max-height: min(720px, calc(100vh - 32px)); overflow: auto; padding: 0; border: 1px solid var(--line); border-radius: 14px; background: var(--paper); color: var(--ink); box-shadow: 0 24px 70px #10282433; animation: dialog-in .18s ease-out; }
.form-info-dialog-content { padding: 0; }
.form-info-card { width: min(640px, calc(100vw - 32px)); max-height: min(720px, calc(100vh - 32px)); overflow: auto; border: 1px solid var(--line); border-radius: 14px; background: var(--paper); box-shadow: 0 24px 70px #10282433; animation: dialog-in .18s ease-out; }
@keyframes dialog-in { from { opacity: 0; transform: translateY(8px) scale(.98); } to { opacity: 1; transform: translateY(0) scale(1); } }
.form-info-card header, .form-info-card footer { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 20px 24px; }
.form-info-card header { border-bottom: 1px solid var(--line); }
.eyebrow { margin: 0 0 5px; color: var(--orange); font-size: .7rem; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
.form-info-card h2 { margin: 0; font-size: 1.25rem; }
.icon-close { width: 34px; height: 34px; border: 1px solid var(--line); border-radius: 8px; background: var(--paper); color: var(--muted); font-size: 1.4rem; line-height: 1; cursor: pointer; }
.form-info-content { padding: 24px; }
.form-metadata { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin: 0 0 18px; }
.form-metadata dt, .permission-grid h3 { color: var(--muted); font-size: .75rem; font-weight: 600; }
.form-metadata dd { margin: 5px 0 0; font-size: .88rem; }
code { padding: 4px 7px; border-radius: 5px; background: #edf3ef; color: var(--green); font-size: .8rem; }
.form-description { margin: 0 0 22px; color: var(--muted); line-height: 1.55; }
.info-tabs { display: flex; gap: 20px; margin-bottom: 16px; border-bottom: 1px solid var(--line); }
.info-tabs button { padding: 0 0 10px; border: 0; border-bottom: 2px solid transparent; background: transparent; color: var(--muted); font: inherit; font-size: .84rem; cursor: pointer; }
.info-tabs button.active { border-color: var(--orange); color: var(--green); font-weight: 700; }
.permission-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
.permission-grid section { padding: 16px; border: 1px solid var(--line); border-radius: 10px; background: #fbfcfa; }
.permission-grid h3 { margin: 0 0 12px; color: var(--ink); font-size: .85rem; }
.permission-grid ul { display: grid; gap: 9px; margin: 0; padding: 0; list-style: none; color: var(--muted); font-size: .84rem; }
.permission-grid li { display: flex; align-items: center; gap: 8px; }
.permission-dot { color: var(--green); font-weight: 700; }
.special-list { padding: 16px; border: 1px solid var(--line); border-radius: 10px; background: #fbfcfa; }
.special-list h3 { margin: 0 0 12px; font-size: .85rem; }
.special-list ul { display: grid; gap: 10px; margin: 0; padding: 0; list-style: none; }
.form-info-card footer { justify-content: flex-end; border-top: 1px solid var(--line); background: #fbfcfa; }
@media (max-width: 600px) { .form-metadata, .permission-grid { grid-template-columns: 1fr; } .form-info-card header, .form-info-card footer, .form-info-content { padding: 18px; } }
</style>
