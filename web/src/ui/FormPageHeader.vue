<script setup lang="ts">
defineProps<{ title: string; subtitle: string; help: boolean }>()
const emit = defineEmits<{ info: []; toggleHelp: [] }>()
</script>

<template>
  <div class="page-title">
    <div>
      <h1>{{ title }} <span class="title-actions"><button class="title-icon" aria-label="Ver datos del formulario" title="Ver datos del formulario" @click="emit('info')"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 10v6M12 7h.01" /></svg></button><button class="title-icon" :aria-label="help ? 'Ocultar ayuda' : 'Mostrar ayuda'" :title="help ? 'Ocultar ayuda' : 'Mostrar ayuda'" :aria-pressed="help" @click="emit('toggleHelp')"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M9.8 9a2.3 2.3 0 1 1 3.7 1.8c-.9.7-1.5 1-1.5 2.2M12 16h.01" /></svg></button></span></h1>
      <p>{{ subtitle }}</p>
    </div>
    <div class="page-actions"><slot name="actions" /></div>
  </div>
</template>

<style scoped>
.title-actions { display: inline-flex; align-items: center; gap: 6px; margin-left: 8px; vertical-align: middle; }
.title-icon { display: inline-grid; place-items: center; width: 36px; height: 36px; padding: 0; border: 1px solid #bfd0c7; border-radius: 9px; background: var(--paper); color: var(--green); box-shadow: 0 3px 0 #c8d6ce, 0 5px 10px #10282416; font: 600 .9rem/1 inherit; cursor: pointer; transition: transform .12s ease, box-shadow .12s ease, background .12s ease; }
.title-icon:hover, .title-icon:focus-visible { border-color: var(--green); background: #edf5f0; outline: 2px solid var(--orange); outline-offset: 2px; }
.title-icon:active, .title-icon[aria-pressed="true"] { transform: translateY(3px); background: #dcebe3; box-shadow: 0 0 0 #c8d6ce, inset 0 2px 4px #1028241f; }
.title-icon svg { width: 20px; height: 20px; fill: none; stroke: currentColor; stroke-width: 2.1; stroke-linecap: round; stroke-linejoin: round; }
.page-actions { display: flex; align-items: center; }
</style>
