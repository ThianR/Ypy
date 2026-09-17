<script setup lang="ts">
defineProps<{ for: string; required?: boolean }>()
</script>

<template>
  <label :for="$props.for" class="form-label">
    <span><slot /></span>
    <span v-if="required" class="form-label-requirement">Obligatorio</span>
  </label>
</template>

<style scoped>
.form-label { display: flex; align-items: baseline; flex-wrap: wrap; gap: 6px 10px; }
.form-label-requirement { color: var(--muted); font-size: .75rem; font-weight: 500; }
</style>
