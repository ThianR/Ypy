<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import type { Category, Workspace } from '../workspace'
import SmallFormDialog from '../ui/SmallFormDialog.vue'
import FormLabel from '../ui/FormLabel.vue'
import FormInfoDialog from '../ui/FormInfoDialog.vue'
import FormPageHeader from '../ui/FormPageHeader.vue'
import { usePagination } from '../composables/usePagination'
import InputText from 'primevue/inputtext'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'

const props = defineProps<{ workspace: Workspace }>()
const editing = ref<Category | null>(null); const open = ref(false); const help = ref(true); const query = ref(''); const showFilters = ref(true); const showFormInfo = ref(false); const error = ref(''); const pending = ref<Category | null>(null); const deleteDialog = ref(false); const discardDialog = ref(false); const formElement = ref<HTMLFormElement>(); const saving = ref(false)
const form = reactive({ codigo: '', nombre: '' }); const filters = reactive({ codigo: '', nombre: '' }); const baseline = ref(JSON.stringify(form)); const dirty = computed(() => JSON.stringify(form) !== baseline.value)
const filtered = computed(() => props.workspace.state.categories.filter(item => `${item.codigo} ${item.nombre}`.toLowerCase().includes(query.value.toLowerCase().trim()) && item.codigo.toLowerCase().includes(filters.codigo.toLowerCase().trim()) && item.nombre.toLowerCase().includes(filters.nombre.toLowerCase().trim())))
const activeFilters = computed(() => Number(Boolean(query.value.trim())) + Number(Boolean(filters.codigo.trim())) + Number(Boolean(filters.nombre.trim())))
const { page, pageSize, pageCount, start, end, rows } = usePagination(filtered)
watch([query, () => filters.codigo, () => filters.nombre], () => { page.value = 1 });
function dateLabel(value: string) { const date = new Date(value); if (Number.isNaN(date.getTime())) return 'Sin registro'; const pad = (part: number) => String(part).padStart(2, '0'); return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}` }
function edit(item: Category | null = null) { editing.value = item; Object.assign(form, item ? { codigo: item.codigo, nombre: item.nombre } : { codigo: '', nombre: '' }); baseline.value = JSON.stringify(form); error.value = ''; open.value = true }
function duplicate(item: Category) {
  const base = item.codigo.toUpperCase()
  let index = 1
  let codigo = `${base.slice(0, 19)}${index}`
  while (props.workspace.state.categories.some(category => category.codigo === codigo)) { index += 1; const suffix = String(index); codigo = `${base.slice(0, 20 - suffix.length)}${suffix}` }
  editing.value = null
  Object.assign(form, { codigo, nombre: item.nombre })
  baseline.value = JSON.stringify(form)
  error.value = ''
  open.value = true
}
function closeEditor() { if (dirty.value) discardDialog.value = true; else open.value = false }
function discard() { discardDialog.value = false; open.value = false }
async function save() { if (saving.value) return; saving.value = true; error.value = ''; try { props.workspace.saveCategory({ ...form, id: editing.value?.id }); baseline.value = JSON.stringify(form); open.value = false } catch (cause) { error.value = cause instanceof Error ? cause.message : 'No se pudo guardar.'; await nextTick(); formElement.value?.querySelector<HTMLElement>('[role=alert]')?.focus() } finally { saving.value = false } }
async function remove(item: Category) { pending.value = item; await nextTick(); deleteDialog.value = true }
function confirmRemove() { if (!pending.value) return; try { props.workspace.deleteCategory(pending.value.id); deleteDialog.value = false; pending.value = null } catch (cause) { error.value = cause instanceof Error ? cause.message : 'No se pudo eliminar.' } }
function closeDelete() { deleteDialog.value = false; pending.value = null }
function clearFilters() { query.value = ''; Object.assign(filters, { codigo: '', nombre: '' }) }
function beforeUnload(event: BeforeUnloadEvent) { if (open.value && dirty.value) { event.preventDefault(); event.returnValue = '' } }
onMounted(() => window.addEventListener('beforeunload', beforeUnload)); onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
</script>

<template>
  <div class="units-page">
    <FormPageHeader title="Categorías" subtitle="Organiza los artículos para encontrarlos y analizarlos con facilidad." :help="help" @info="showFormInfo = true" @toggle-help="help = !help"><template #actions><button class="button primary" @click="edit()">Nuevo</button></template></FormPageHeader>
    <p v-if="help" class="help-copy">Una categoría agrupa artículos relacionados, como Almacén, Bebidas o Limpieza. Un artículo puede pertenecer a una categoría.</p>
    <section class="panel unit-list" aria-labelledby="categories-title">
      <header class="list-heading"><h2 id="categories-title">Categorías registradas</h2><span class="muted">{{ filtered.length }} registros</span></header>
      <div class="list-tools">
        <div class="search-row">
          <label class="unit-search">Buscar<InputText v-model="query" type="search" placeholder="Código o nombre" /></label>
          <button class="button secondary" :aria-expanded="showFilters" @click="showFilters = !showFilters">{{ showFilters ? 'Ocultar filtros' : 'Filtros específicos' }}</button>
          <button v-if="activeFilters" class="text-button" @click="clearFilters">Limpiar filtros ({{ activeFilters }})</button>
        </div>
        <div v-if="showFilters" class="unit-filters"><label>Código<InputText v-model="filters.codigo" placeholder="Contiene…" /></label><label>Nombre<InputText v-model="filters.nombre" placeholder="Contiene…" /></label></div>
      </div>
      <DataTable :value="rows" data-key="id" class="ypy-data-table" scrollable scroll-height="52vh"><Column field="codigo" header="Código"><template #body="slotProps"><strong>{{ slotProps.data.codigo }}</strong></template></Column><Column field="nombre" header="Nombre" /><Column header="Creada"><template #body="slotProps">{{ dateLabel(slotProps.data.createdAt) }}</template></Column><Column header="Actualizada"><template #body="slotProps">{{ dateLabel(slotProps.data.updatedAt) }}</template></Column><Column header="Acciones"><template #body="slotProps"><div class="row-buttons"><button class="row-button" :aria-label="`Editar ${slotProps.data.codigo}`" @click="edit(slotProps.data)">Editar</button><button class="row-button" :aria-label="`Duplicar ${slotProps.data.codigo}`" @click="duplicate(slotProps.data)">Duplicar</button><button class="row-button delete-button" :aria-label="`Eliminar ${slotProps.data.codigo}`" @click="remove(slotProps.data)">Eliminar</button></div></template></Column><template #empty><div class="table-empty"><strong>{{ activeFilters ? 'No hay coincidencias' : 'No hay categorías' }}</strong><span>{{ activeFilters ? 'Cambia o limpia los filtros para ver otros registros.' : 'Selecciona Nuevo para registrar la primera categoría.' }}</span><button v-if="activeFilters" class="button secondary" @click="clearFilters">Limpiar filtros</button></div></template></DataTable>
      <footer class="unit-pagination"><label>Filas por página<Select v-model="pageSize" :options="[10, 25, 50]" /></label><span role="status">{{ start }}–{{ end }} de {{ filtered.length }}</span><div class="pager-actions"><button class="button secondary" :disabled="page === 1" @click="page--">Anterior</button><span>{{ page }} / {{ pageCount }}</span><button class="button secondary" :disabled="page >= pageCount" @click="page++">Siguiente</button></div></footer>
    </section>
    <FormInfoDialog :open="showFormInfo" title="Categorías" identifier="maestros.categorias" module="Catálogos base" description="Organiza los artículos en grupos para facilitar su búsqueda y análisis." :permissions="['Visualizar', 'Crear', 'Editar', 'Duplicar', 'Eliminar']" :user-permissions="['Visualizar', 'Crear', 'Editar', 'Duplicar', 'Eliminar']" @close="showFormInfo = false" />
    <SmallFormDialog :open="open" :title="editing ? 'Editar categoría' : 'Nueva categoría'" title-id="category-editor-title" @close="closeEditor"><button type="button" class="text-button help-control" @click="help = !help">{{ help ? 'Ocultar ayuda sobre esta pantalla' : 'Mostrar ayuda sobre esta pantalla' }}</button><form id="category-form" ref="formElement" @submit.prevent="save"><fieldset class="unit-fields" :disabled="saving"><FormLabel for="category-code" required>Código</FormLabel><InputText id="category-code" v-model="form.codigo" data-initial-focus required maxlength="20" placeholder="Ej. BEB" /><p v-if="help" class="field-help">Código corto para buscar y referirse a la categoría.</p><FormLabel for="category-name" required>Nombre</FormLabel><InputText id="category-name" v-model="form.nombre" required maxlength="80" placeholder="Ej. Bebidas" /><p v-if="help" class="field-help">Nombre visible para el equipo y los filtros del catálogo.</p></fieldset><section class="unit-audit"><h3>Información del registro</h3><dl v-if="editing"><div><dt>Creación</dt><dd>{{ dateLabel(editing.createdAt) }}<span>{{ editing.createdBy }}</span></dd></div><div><dt>Última actualización</dt><dd>{{ dateLabel(editing.updatedAt) }}<span>{{ editing.updatedBy }}</span></dd></div></dl><p v-else>Sin guardar. La fecha y el usuario se registrarán al guardar.</p></section><p v-if="error" role="alert" tabindex="-1" class="message error">{{ error }}</p></form><template #actions><button type="button" class="button secondary" :disabled="saving" @click="closeEditor">Cancelar</button><button type="submit" form="category-form" class="button primary" :disabled="saving">{{ saving ? 'Guardando…' : 'Guardar' }}</button></template></SmallFormDialog>
    <Dialog :visible="deleteDialog" modal :show-header="false" :pt="{ root: 'dialog unit-confirm', content: 'confirm-dialog-content' }" @update:visible="deleteDialog = $event"><h2 id="category-delete-title">¿Eliminar este registro?</h2><p><strong>{{ pending?.codigo }} · {{ pending?.nombre }}</strong></p><p>Se quitará del catálogo local.</p><div class="dialog-actions"><button class="button secondary" autofocus @click="closeDelete">Cancelar</button><button class="button danger" @click="confirmRemove">Eliminar</button></div></Dialog><Dialog :visible="discardDialog" modal :show-header="false" :pt="{ root: 'dialog unit-confirm', content: 'confirm-dialog-content' }" @update:visible="discardDialog = $event"><h2>Hay cambios sin guardar</h2><p>¿Deseas descartar los cambios?</p><div class="dialog-actions"><button class="button secondary" autofocus @click="discardDialog = false">Continuar</button><button class="button danger" @click="discard">Descartar</button></div></Dialog>
  </div>
</template>

<style scoped>
.units-page { min-width: 0; }
.help-control { display: block; margin: 0 0 16px; }
.help-copy { margin: 0 0 22px; max-width: 75ch; color: var(--muted); line-height: 1.6; }
.unit-list { overflow: hidden; }
.list-heading { display: flex; justify-content: space-between; gap: 12px; padding: 20px; border-bottom: 1px solid var(--line); }
.list-heading h2 { margin: 0; font-size: 1.05rem; }
.list-tools { padding: 20px; }
.search-row { display: flex; align-items: end; flex-wrap: wrap; gap: 12px; }
.unit-search { flex: 1 1 220px; }
.search-row > .text-button { min-height: 44px; }
.unit-filters { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; padding-top: 20px; }
.unit-filters p { grid-column: 1 / -1; margin: 0; color: var(--muted); font-size: .8rem; }
.unit-table-scroll { max-height: 52vh; overflow: auto; isolation: isolate; }
.unit-table { width: 100%; border-collapse: collapse; }
.unit-table th, .unit-table td { padding: 12px 16px; border-bottom: 1px solid #edf0ed; text-align: left; white-space: normal; overflow-wrap: anywhere; }
.unit-table th { position: sticky; top: 0; z-index: 1; background: var(--paper); color: var(--muted); font-size: .7rem; border-bottom: 1px solid var(--line); }
.table-empty { padding: 28px 16px !important; color: var(--muted); text-align: center !important; }
.table-empty strong, .table-empty span { display: block; }
.table-empty strong { color: var(--ink); margin-bottom: 6px; }
.table-empty button { margin-top: 14px; }
.column-filters-row th { top: 38px; z-index: 1; padding-top: 8px; padding-bottom: 8px; background: var(--paper); }
.column-filters-row input { width: 100%; min-width: 0; min-height: 36px; padding: 7px 9px; font-size: .78rem; }
.column-heading { display: inline-flex; align-items: center; gap: 5px; }
.column-filter { position: relative; display: inline-block; }
.column-filter summary { display: grid; place-items: center; width: 22px; height: 22px; border-radius: 4px; color: var(--muted); cursor: pointer; list-style: none; }
.column-filter summary::-webkit-details-marker { display: none; }
.column-filter[open] summary, .column-filter summary:hover { background: #e7f0eb; color: var(--green); }
.column-filter input { position: absolute; z-index: 4; top: 28px; left: 0; width: 160px; min-height: 38px; padding: 8px; border: 1px solid var(--line); box-shadow: 0 8px 20px #10282422; font-size: .76rem; }
.column-filter input:focus { outline: 2px solid var(--orange); outline-offset: 1px; }
.unit-table .row-actions { position: sticky; right: 0; width: 150px; min-width: 150px; background: var(--paper); border-left: 1px solid var(--line); }
.unit-table th.row-actions { z-index: 2; }
.row-buttons { display: flex; gap: 8px; }
.row-button { display: inline-flex; align-items: center; justify-content: center; gap: 6px; min-height: 40px; min-width: 40px; padding: 7px 9px; border: 1px solid var(--line); border-radius: 6px; background: var(--paper); color: var(--green); font: inherit; }
.row-button:hover { border-color: currentColor; }
.row-button svg { width: 17px; height: 17px; fill: none; stroke: currentColor; stroke-width: 1.7; stroke-linecap: round; stroke-linejoin: round; }
.delete-button { color: #974d3c; }
.unit-pagination { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; padding: 16px 20px; border-top: 1px solid var(--line); font-size: .8rem; }
.unit-pagination label { display: flex; align-items: center; gap: 8px; font-weight: 500; }
.unit-pagination select { width: 72px; }
.pager-actions { display: flex; align-items: center; gap: 12px; }
.unit-fields { display: grid; gap: 8px; border: 0; padding: 0; margin: 0; min-width: 0; }
.unit-fields > label:not(:first-child) { margin-top: 14px; }
.field-help { margin: 0; color: var(--muted); font-size: .82rem; line-height: 1.5; }
.unit-audit { margin-top: 28px; padding-top: 20px; border-top: 1px solid var(--line); }
.unit-audit h3 { margin: 0 0 14px; font-size: .9rem; }
.unit-audit dl { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 0; }
.unit-audit dt { color: var(--muted); font-size: .76rem; }
.unit-audit dd { margin: 0; font-size: .82rem; }
.unit-audit dd span { display: block; color: var(--muted); }
.unit-audit p { color: var(--muted); font-size: .82rem; }
.unit-confirm { max-height: calc(100dvh - 32px); overflow-y: auto; background: var(--paper); }
button:focus-visible, .unit-table-scroll:focus-visible { outline: 3px solid var(--orange); outline-offset: 3px; }
@media (max-width: 600px) { .unit-filters { grid-template-columns: 1fr; } .unit-table { min-width: 620px; } .unit-table .row-actions { width: 140px; min-width: 140px; } .row-button span { display: none; } .unit-table th, .unit-table td { padding: 10px; } .unit-audit dl { grid-template-columns: 1fr; } .pager-actions { width: 100%; justify-content: space-between; } }
</style>
