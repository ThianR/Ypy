<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useStandardForm } from '../composables/useStandardForm'
import { categories, money, type Product, type ProductForm, type Workspace } from '../workspace'
import InputText from 'primevue/inputtext'
import SmallFormDialog from '../ui/SmallFormDialog.vue'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
const props = defineProps<{ workspace: Workspace }>()
const query = ref('')
const category = ref('Todos')
const dialogOpen = ref(false)
const standard = useStandardForm<ProductForm>(() => ({ itemId: '', name: '', barcode: '', category: categories[0], price: '', stock: '0' }))
const form = standard.values
const categoryOptions = ['Todos', ...categories]
let timer: ReturnType<typeof setTimeout>
const filtered = computed(() => props.workspace.state.products.filter(item => (category.value === 'Todos' || item.category === category.value) && `${item.name} ${item.barcode}`.toLowerCase().includes(query.value.toLowerCase())))
watch(query, value => { clearTimeout(timer); timer = setTimeout(() => props.workspace.searchCatalog(value), 300) })
onBeforeUnmount(() => clearTimeout(timer))
function edit(product?: Product) {
  standard.reset(product ?? { itemId: '', name: '', barcode: '', category: categories[0], price: '', stock: '0' })
  dialogOpen.value = true
}
async function save() {
  if (await standard.submit(values => props.workspace.saveProduct(values))) dialogOpen.value = false
}
</script>

<template>
  <div class="page-title"><div><h1>Catálogo de artículos</h1><p>Consulta códigos, precios y existencias en un solo lugar.</p></div><button v-if="workspace.state.mode === 'demo'" class="button primary" @click="edit()">Nuevo artículo</button></div>
  <p v-if="workspace.state.mode === 'central'" class="message info">Consulta de precios del servidor. El alta y la edición central de artículos aún no están disponibles en esta interfaz.</p>
  <section class="panel">
    <div class="table-toolbar"><label class="search-label"><span class="sr-only">Buscar en el catálogo</span><InputText v-model="query" type="search" placeholder="Buscar por nombre o código" /></label><label v-if="workspace.state.mode === 'demo'"><span class="sr-only">Categoría</span><Select v-model="category" :options="categoryOptions" /></label><span class="muted small">{{ filtered.length }} artículos</span></div>
    <DataTable v-if="filtered.length" :value="filtered" data-key="itemId" paginator :rows="10" :rows-per-page-options="[10, 25, 50]" row-hover class="ypy-data-table"><Column field="name" header="Artículo"><template #body="slotProps"><strong>{{ slotProps.data.name }}</strong></template></Column><Column field="barcode" header="Código" body-class="muted" /><Column field="category" header="Categoría" /><Column header="Precio de venta" header-class="numeric" body-class="numeric"><template #body="slotProps">{{ money(slotProps.data.price) }}</template></Column><Column header="Disponibles" header-class="numeric" body-class="numeric"><template #body="slotProps"><span :class="['badge', slotProps.data.stock === '0' ? 'warning' : 'neutral']">{{ slotProps.data.stock === '' ? 'Sin datos' : slotProps.data.stock }}</span></template></Column><Column header="Acciones"><template #body="slotProps"><button v-if="workspace.state.mode === 'demo'" class="text-button" :aria-label="`Editar ${slotProps.data.name}`" @click="edit(slotProps.data)">Editar</button></template></Column></DataTable>
    <div v-if="!filtered.length" class="empty-state"><h3>{{ workspace.state.loading ? 'Buscando artículos...' : 'Sin artículos para mostrar' }}</h3><p>{{ workspace.state.mode === 'central' ? 'Escribe un nombre o código para consultar los precios vigentes.' : 'Cambia la búsqueda o crea tu primer artículo.' }}</p></div>
    <footer class="table-footer">{{ workspace.state.mode === 'demo' ? 'Catálogo de demostración. Los cambios se guardan en este navegador.' : 'Los precios corresponden a la lista configurada y a la fecha de consulta.' }}</footer>
  </section>
  <SmallFormDialog :open="dialogOpen" :title="form.itemId ? 'Editar artículo' : 'Nuevo artículo'" title-id="product-title" @close="dialogOpen = false">
    <form id="product-editor-form" @submit.prevent="save">
      <p class="muted small">Catálogo de demostración</p>
      <div class="form-grid"><label class="span-all">Nombre del artículo<InputText v-model="form.name" data-initial-focus required maxlength="120" placeholder="Ej.: Café molido 250 g" /></label><label>Código de barras<InputText v-model="form.barcode" required maxlength="48" placeholder="Código único" /></label><label>Categoría<Select v-model="form.category" :options="categories" /></label><label>Precio de venta (Gs.)<InputText v-model="form.price" required inputmode="numeric" pattern="[0-9]{1,12}" placeholder="0" /></label><label>Existencias de práctica<InputText v-model="form.stock" required inputmode="numeric" pattern="[0-9]{1,6}" /></label></div>
      <p class="muted small field-space">Los campos son obligatorios. Usa precios enteros en guaraníes. Al editar un artículo, se retira del carrito actual para revisar su precio.</p>
      <p v-if="standard.error" class="message error" role="alert">{{ standard.error }}</p>
    </form>
    <template #actions><button type="button" class="button secondary" :disabled="standard.saving" @click="dialogOpen = false">Cancelar</button><button type="submit" form="product-editor-form" class="button primary" :disabled="standard.saving">{{ standard.saving ? 'Guardando...' : 'Guardar artículo' }}</button></template>
  </SmallFormDialog>
</template>
