<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import type { UnitOfMeasure, Workspace } from '../workspace'
import SmallFormDialog from '../ui/SmallFormDialog.vue'
import FormLabel from '../ui/FormLabel.vue'
import FormInfoDialog from '../ui/FormInfoDialog.vue'
import FormPageHeader from '../ui/FormPageHeader.vue'
import { usePagination } from '../composables/usePagination'
import InputText from 'primevue/inputtext'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'

const props = defineProps<{ workspace: Workspace }>()
const editing = ref<UnitOfMeasure | null>(null)
const editorOpen = ref(false)
const query = ref('')
const showHelp = ref(localStorage.getItem('ypy.units.help') !== 'hidden')
const showFilters = ref(true)
const showFormInfo = ref(false)
const filters = reactive({ codigo: '', nombre: '', dimension: '', fraction: 'all' })
const fractionOptions = [{ label: 'Todas', value: 'all' }, { label: 'Admite fracciones', value: 'true' }, { label: 'Solo enteros', value: 'false' }]
const error = ref('')
const deleteError = ref('')
const pendingDelete = ref<UnitOfMeasure | null>(null)
const deleteDialog = ref(false)
const discardDialog = ref(false)
const formElement = ref<HTMLFormElement>()
const saving = ref(false)
const blank = () => ({ codigo: '', nombre: '', dimension: '', admiteFraccion: true })
const form = reactive(blank())
const baseline = ref(JSON.stringify(form))
const dirty = computed(() => JSON.stringify(form) !== baseline.value)
const includes = (value: string, text: string) => value.toLocaleLowerCase().includes(text.trim().toLocaleLowerCase())
const filtered = computed(() => props.workspace.state.unitsOfMeasure.filter(unit =>
  includes(unit.codigo + ' ' + unit.nombre, query.value) && includes(unit.codigo, filters.codigo) &&
  includes(unit.nombre, filters.nombre) && includes(unit.dimension, filters.dimension) &&
  (filters.fraction === 'all' || String(unit.admiteFraccion) === filters.fraction)
))
const { page, pageSize, pageCount, start, end, rows } = usePagination(filtered)
const activeFilters = computed(() => Number(Boolean(query.value.trim())) + Number(Boolean(filters.codigo.trim())) +
  Number(Boolean(filters.nombre.trim())) + Number(Boolean(filters.dimension)) + Number(filters.fraction !== 'all'))
watch([query, () => filters.codigo, () => filters.nombre, () => filters.dimension, () => filters.fraction], () => { page.value = 1 }, { flush: 'sync' })
watch(showHelp, value => { try { localStorage.setItem('ypy.units.help', value ? 'visible' : 'hidden') } catch { /* Available during this visit. */ } })

function dateLabel(value: string) {
  const date = new Date(value)
  if (!value || Number.isNaN(date.getTime())) return 'Sin registro'
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
function openEditor(unit: UnitOfMeasure | null = null) {
  editing.value = unit
  Object.assign(form, unit ? { codigo: unit.codigo, nombre: unit.nombre, dimension: unit.dimension, admiteFraccion: unit.admiteFraccion } : blank())
  baseline.value = JSON.stringify(form)
  error.value = ''
  editorOpen.value = true
}
function duplicate(unit: UnitOfMeasure) {
  const base = unit.codigo.toUpperCase()
  let index = 1
  let codigo = `${base.slice(0, 11)}${index}`
  while (props.workspace.state.unitsOfMeasure.some(item => item.codigo === codigo)) { index += 1; const suffix = String(index); codigo = `${base.slice(0, 12 - suffix.length)}${suffix}` }
  editing.value = null
  Object.assign(form, { codigo, nombre: unit.nombre, dimension: unit.dimension, admiteFraccion: unit.admiteFraccion })
  baseline.value = JSON.stringify(form)
  error.value = ''
  editorOpen.value = true
}
function requestClose() {
  if (saving.value) return
  if (dirty.value) discardDialog.value = true
  else editorOpen.value = false
}
function discard() { discardDialog.value = false; editorOpen.value = false }
async function save() {
  if (saving.value) return
  saving.value = true
  error.value = ''
  try {
    await props.workspace.saveUnit({ ...form, id: editing.value?.id })
    baseline.value = JSON.stringify(form)
    editorOpen.value = false
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'No se pudo guardar. Reintenta.'
    await nextTick()
    formElement.value?.querySelector<HTMLElement>('[role="alert"]')?.focus()
  } finally { saving.value = false }
}
function clearFilters() { query.value = ''; Object.assign(filters, { codigo: '', nombre: '', dimension: '', fraction: 'all' }) }
async function remove(unit: UnitOfMeasure) {
  pendingDelete.value = unit
  deleteError.value = ''
  await nextTick()
  deleteDialog.value = true
}
function closeDelete() { deleteDialog.value = false; pendingDelete.value = null }
async function confirmRemove() {
  if (!pendingDelete.value) return
  try { await props.workspace.deleteUnit(pendingDelete.value.id); closeDelete() }
  catch (cause) { deleteError.value = cause instanceof Error ? cause.message : 'No se pudo eliminar. Reintenta.' }
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (editorOpen.value && dirty.value) { event.preventDefault(); event.returnValue = '' }
}
onMounted(() => window.addEventListener('beforeunload', beforeUnload))
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
</script>

<template>
  <div class="units-page">
    <div class="page-title">
      <FormPageHeader title="Unidades de medida" subtitle="Define cómo se cuentan, pesan o miden tus artículos." :help="showHelp" @info="showFormInfo = true" @toggle-help="showHelp = !showHelp"><template #actions><button class="button primary" @click="openEditor()">Nuevo</button></template></FormPageHeader>
    </div>
    <p v-if="showHelp" class="help-copy">Una unidad indica cómo expresas una cantidad: por unidad, kilogramo o litro. Usa “Nuevo” para registrarla o “Editar” para cambiar sus datos.</p>

    <section class="panel unit-list" aria-labelledby="units-list-title">
      <header class="list-heading"><h2 id="units-list-title">Unidades registradas</h2><span class="muted">{{ workspace.state.unitsOfMeasure.length }} registros</span></header>
      <div class="list-tools">
        <div class="search-row">
          <label class="unit-search">Buscar<InputText v-model="query" type="search" placeholder="Código o nombre" /></label>
          <button class="button secondary" :aria-expanded="showFilters" @click="showFilters = !showFilters">{{ showFilters ? 'Ocultar filtros' : 'Filtros específicos' }}</button>
          <button v-if="activeFilters" class="text-button" @click="clearFilters">Limpiar filtros ({{ activeFilters }})</button>
        </div>
      </div>
      <div class="unit-table-scroll" tabindex="0" role="region" aria-label="Listado de unidades">
        <table class="unit-table">
          <caption class="sr-only">Unidades de medida, página {{ page }} de {{ pageCount }}</caption>
          <thead><tr><th scope="col">Código</th><th scope="col">Nombre</th><th scope="col">Dimensión</th><th scope="col">Fracciones</th><th scope="col" class="row-actions">Acciones</th></tr><tr v-if="showFilters" class="column-filters-row"><th><label class="sr-only" for="unit-code-filter">Código contiene</label><InputText id="unit-code-filter" v-model="filters.codigo" placeholder="Contiene…" /></th><th><label class="sr-only" for="unit-name-filter">Nombre contiene</label><InputText id="unit-name-filter" v-model="filters.nombre" placeholder="Contiene…" /></th><th><label class="sr-only" for="unit-dimension-filter">Dimensión contiene</label><InputText id="unit-dimension-filter" v-model="filters.dimension" placeholder="Ej. PESO o VOLUMEN" /></th><th><label class="sr-only" for="unit-fraction-filter">Fracciones</label><Select input-id="unit-fraction-filter" v-model="filters.fraction" :options="fractionOptions" option-label="label" option-value="value" /></th><th class="row-actions"><span class="sr-only">Sin filtro</span></th></tr></thead>
          <tbody><tr v-if="!filtered.length"><td colspan="5" class="table-empty"><strong>{{ activeFilters ? 'No hay coincidencias' : 'Todavía no hay unidades' }}</strong><span>{{ activeFilters ? 'Cambia o limpia los filtros para ver otros registros.' : 'Selecciona Nuevo para registrar la primera unidad.' }}</span><button v-if="activeFilters" class="button secondary" @click="clearFilters">Limpiar filtros</button></td></tr><tr v-for="unit in rows" :key="unit.id">
            <td><strong>{{ unit.codigo }}</strong></td><td>{{ unit.nombre }}</td><td>{{ unit.dimension }}</td><td>{{ unit.admiteFraccion ? 'Sí' : 'No' }}</td>
            <td class="row-actions"><div class="row-buttons">
              <button class="row-button" :aria-label="'Editar ' + unit.codigo" :title="'Editar ' + unit.codigo" @click="openEditor(unit)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 5 4 4M4 20l4-1L20 7a2.8 2.8 0 0 0-4-4L4 15Z" /></svg></button>
              <button class="row-button" :aria-label="'Duplicar ' + unit.codigo" :title="'Duplicar ' + unit.codigo" @click="duplicate(unit)"><svg viewBox="0 0 24 24" aria-hidden="true"><rect x="8" y="8" width="11" height="11" rx="2" /><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2" /></svg></button>
              <button class="row-button delete-button" :aria-label="'Eliminar ' + unit.codigo" :title="'Eliminar ' + unit.codigo" @click="remove(unit)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13M10 10v7M14 10v7" /></svg></button>
            </div></td>
          </tr></tbody>
        </table>
      </div>
      <footer class="unit-pagination">
        <label>Filas por página<Select v-model="pageSize" :options="[10, 25, 50]" /></label>
        <span role="status">{{ start }}–{{ end }} de {{ filtered.length }}</span>
        <div class="pager-actions"><button class="button secondary" :disabled="page === 1" @click="page--">Anterior</button><span>{{ page }} / {{ pageCount }}</span><button class="button secondary" :disabled="page >= pageCount" @click="page++">Siguiente</button></div>
      </footer>
    </section>
    <FormInfoDialog :open="showFormInfo" title="Unidades de medida" identifier="maestros.unidades_medida" module="Catálogos base" description="Define las unidades que el sistema utiliza para contar, pesar y medir artículos." :permissions="['Visualizar', 'Crear', 'Editar', 'Duplicar', 'Eliminar']" :user-permissions="['Visualizar', 'Crear', 'Editar', 'Duplicar', 'Eliminar']" @close="showFormInfo = false" />

    <SmallFormDialog :open="editorOpen" :title="editing ? 'Editar unidad de medida' : 'Nueva unidad de medida'" title-id="unit-editor-title" @close="requestClose">
      <button type="button" class="text-button help-control" :aria-expanded="showHelp" @click="showHelp = !showHelp">{{ showHelp ? 'Ocultar ayuda sobre esta pantalla' : 'Mostrar ayuda sobre esta pantalla' }}</button>
      <form id="unit-editor-form" ref="formElement" @submit.prevent="save">
        <fieldset :disabled="saving" class="unit-fields">
          <FormLabel for="unit-code" required>Código</FormLabel>
          <InputText id="unit-code" v-model="form.codigo" data-initial-focus maxlength="12" required :aria-describedby="showHelp ? 'unit-code-help' : undefined" placeholder="Ej. KG" />
          <p v-if="showHelp" id="unit-code-help" class="field-help">Identificador corto y único para buscar la unidad. Por ejemplo: UN, KG o L.</p>
          <FormLabel for="unit-name" required>Nombre</FormLabel>
          <InputText id="unit-name" v-model="form.nombre" maxlength="80" required :aria-describedby="showHelp ? 'unit-name-help' : undefined" placeholder="Ej. Kilogramo" />
          <p v-if="showHelp" id="unit-name-help" class="field-help">Nombre que verá tu equipo al seleccionar una unidad.</p>
          <FormLabel for="unit-dimension" required>Dimensión</FormLabel>
          <InputText id="unit-dimension" v-model="form.dimension" maxlength="30" required :aria-describedby="showHelp ? 'unit-dimension-help' : undefined" placeholder="Ej. PESO" />
          <p v-if="showHelp" id="unit-dimension-help" class="field-help">Qué se mide: PESO para kilogramos y gramos, VOLUMEN para litros y mililitros, UNIDAD para contar piezas. Agruparlas no realiza conversiones automáticamente.</p>
          <label class="fraction-field"><Checkbox v-model="form.admiteFraccion" binary input-id="unit-fraction" :aria-describedby="showHelp ? 'unit-fraction-help' : undefined" /><span>Permitir cantidades fraccionarias</span></label>
          <p v-if="showHelp" id="unit-fraction-help" class="field-help">Permite cantidades como 0,5 kg o 1,25 L. Desactívalo si solo se admiten cantidades enteras.</p>
        </fieldset>
        <section class="unit-audit" aria-label="Información del registro">
          <h3>Información del registro</h3>
          <dl v-if="editing"><div><dt>Creación</dt><dd>{{ dateLabel(editing.createdAt) }}<span>{{ editing.createdBy || 'Sin registro' }}</span></dd></div><div><dt>Última actualización</dt><dd>{{ dateLabel(editing.updatedAt) }}<span>{{ editing.updatedBy || 'Sin registro' }}</span></dd></div></dl>
          <p v-else>Sin guardar. La fecha y el usuario se registrarán al guardar.</p>
        </section>
        <p v-if="error" role="alert" tabindex="-1" class="message error">{{ error }}</p>
      </form>
      <template #actions><button type="button" class="button secondary" :disabled="saving" @click="requestClose">Cancelar</button><button type="submit" form="unit-editor-form" class="button primary" :disabled="saving">{{ saving ? 'Guardando…' : 'Guardar' }}</button></template>
    </SmallFormDialog>

    <Dialog :visible="deleteDialog" modal :show-header="false" :pt="{ root: 'dialog unit-confirm', content: 'confirm-dialog-content' }" @update:visible="deleteDialog = $event">
      <h2 id="unit-delete-title">¿Eliminar este registro?</h2><p><strong>{{ pendingDelete?.codigo }} · {{ pendingDelete?.nombre }}</strong></p><p>Se eliminará del catálogo local. Esta acción no se puede deshacer.</p>
      <p v-if="deleteError" class="message error" role="alert">{{ deleteError }}</p>
      <div class="dialog-actions"><button class="button secondary" autofocus @click="closeDelete">Cancelar</button><button class="button danger" @click="confirmRemove">Eliminar</button></div>
    </Dialog>
    <Dialog :visible="discardDialog" modal :show-header="false" :pt="{ root: 'dialog unit-confirm', content: 'confirm-dialog-content' }" @update:visible="discardDialog = $event">
      <h2 id="unit-discard-title">Hay cambios sin guardar</h2><p>Puedes continuar editando o cerrar y descartar los cambios.</p>
      <div class="dialog-actions"><button class="button secondary" autofocus @click="discardDialog = false">Continuar editando</button><button class="button danger" @click="discard">Descartar cambios</button></div>
    </Dialog>
  </div>
</template>

<style scoped>
.units-page { min-width: 0; }
.help-control { display: block; margin: 0 0 16px; text-align: left; }
.help-copy { margin: 0 0 22px; max-width: 75ch; color: var(--muted); font-size: .88rem; line-height: 1.6; }
.unit-list { overflow: hidden; }
.list-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 12px; padding: 20px; border-bottom: 1px solid var(--line); }
.list-heading h2 { margin: 0; font-size: 1.05rem; }
.list-heading > span { font-size: .8rem; }
.list-tools { padding: 20px; }
.search-row { display: flex; align-items: end; flex-wrap: wrap; gap: 12px; }
.unit-search { flex: 1 1 220px; }
.search-row > .text-button { min-height: 44px; }
.unit-filters { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; padding-top: 20px; }
.unit-filters p { grid-column: 1 / -1; margin: 0; color: var(--muted); font-size: .8rem; }
.unit-table-scroll { max-height: 52vh; overflow: auto; isolation: isolate; }
.unit-table th, .unit-table td { padding: 12px 16px; white-space: normal; overflow-wrap: anywhere; }
.unit-table th { position: sticky; top: 0; z-index: 1; background: var(--paper); border-bottom: 1px solid var(--line); }
.column-filters-row th { top: 38px; z-index: 1; padding-top: 8px; padding-bottom: 8px; background: var(--paper); }
.column-filters-row input, .column-filters-row select { width: 100%; min-width: 0; min-height: 36px; padding: 7px 9px; font-size: .78rem; }
.column-heading { display: inline-flex; align-items: center; gap: 5px; white-space: nowrap; }
.column-filter { position: relative; display: inline-block; }
.column-filter summary { display: grid; place-items: center; width: 22px; height: 22px; border-radius: 4px; color: var(--muted); cursor: pointer; list-style: none; }
.column-filter summary::-webkit-details-marker { display: none; }
.column-filter summary:hover, .column-filter[open] summary { background: #e7f0eb; color: var(--green); }
.column-filter input, .column-filter select { position: absolute; z-index: 4; top: 28px; left: 0; width: 160px; min-height: 38px; padding: 8px; border: 1px solid var(--line); box-shadow: 0 8px 20px #10282422; font-size: .76rem; }
.column-filter select { width: 135px; }
.unit-table td { font-size: .85rem; }
.table-empty { padding: 28px 16px !important; color: var(--muted); text-align: center !important; }
.table-empty strong, .table-empty span { display: block; }
.table-empty strong { color: var(--ink); margin-bottom: 6px; }
.table-empty button { margin-top: 14px; }
.unit-table tbody tr:hover { background: var(--page-bg); }
.unit-table .row-actions { position: sticky; right: 0; width: 150px; min-width: 150px; background: var(--paper); border-left: 1px solid var(--line); }
.unit-table th.row-actions { z-index: 2; }
.row-buttons { display: flex; gap: 8px; }
.row-button { display: inline-flex; align-items: center; justify-content: center; gap: 6px; min-height: 40px; min-width: 40px; padding: 7px 9px; border: 1px solid var(--line); border-radius: 6px; color: var(--green); background: var(--paper); font: inherit; }
.row-button:hover { border-color: currentColor; }
.row-button svg { width: 17px; height: 17px; fill: none; stroke: currentColor; stroke-width: 1.7; stroke-linecap: round; stroke-linejoin: round; }
.delete-button { color: #974d3c; }
.unit-pagination { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; padding: 16px 20px; border-top: 1px solid var(--line); font-size: .8rem; }
.unit-pagination label { display: flex; align-items: center; gap: 10px; font-weight: 500; }
.unit-pagination select { width: 72px; }
.pager-actions { display: flex; align-items: center; gap: 12px; }
.pager-actions > span { white-space: nowrap; }
.unit-fields { display: grid; gap: 8px; border: 0; padding: 0; margin: 0; min-width: 0; }
.unit-fields > label:not(:first-child) { margin-top: 14px; }
.field-help { margin: 0; color: var(--muted); font-size: .82rem; line-height: 1.5; }
.fraction-field { display: flex; align-items: center; gap: 10px; }
.fraction-field input { width: 18px; min-height: 18px; flex: 0 0 18px; accent-color: var(--green); }
.unit-audit { margin-top: 28px; padding-top: 20px; border-top: 1px solid var(--line); }
.unit-audit h3 { margin: 0 0 14px; font-size: .9rem; }
.unit-audit dl { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 0; }
.unit-audit dt { color: var(--muted); font-size: .76rem; margin-bottom: 6px; }
.unit-audit dd { margin: 0; font-size: .82rem; line-height: 1.5; overflow-wrap: anywhere; }
.unit-audit dd span { display: block; color: var(--muted); }
.unit-audit p { color: var(--muted); font-size: .82rem; line-height: 1.5; }
.unit-confirm { max-height: calc(100dvh - 32px); overflow-y: auto; background: var(--paper); }
.unit-confirm .dialog-actions { flex-wrap: wrap; }
button:focus-visible, .unit-table-scroll:focus-visible { outline: 3px solid var(--orange); outline-offset: 3px; }
@media (max-width: 600px) {
  .unit-filters { grid-template-columns: 1fr; }
  .unit-table { min-width: 520px; }
  .unit-table .row-actions { width: 140px; min-width: 140px; }
  .row-button span { display: none; }
  .unit-table th, .unit-table td { padding: 10px; }
  .unit-audit dl { grid-template-columns: 1fr; }
  .pager-actions { width: 100%; justify-content: space-between; }
}
</style>
