<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useStandardForm } from '../composables/useStandardForm'
import { categories, money, type Product, type ProductForm, type Workspace } from '../workspace'
const props = defineProps<{ workspace: Workspace }>()
const query = ref('')
const category = ref('Todos')
const dialog = ref<HTMLDialogElement>()
const standard = useStandardForm<ProductForm>(() => ({ itemId: '', name: '', barcode: '', category: categories[0], price: '', stock: '0' }))
const form = standard.values
let timer: ReturnType<typeof setTimeout>
const filtered = computed(() => props.workspace.state.products.filter(item => (category.value === 'Todos' || item.category === category.value) && `${item.name} ${item.barcode}`.toLowerCase().includes(query.value.toLowerCase())))
watch(query, value => { clearTimeout(timer); timer = setTimeout(() => props.workspace.searchCatalog(value), 300) })
onBeforeUnmount(() => clearTimeout(timer))
function edit(product?: Product) {
  standard.reset(product ?? { itemId: '', name: '', barcode: '', category: categories[0], price: '', stock: '0' })
  dialog.value?.showModal()
}
async function save() {
  if (await standard.submit(values => props.workspace.saveProduct(values))) dialog.value?.close()
}
</script>

<template>
  <div class="page-title"><div><h1>Catálogo de artículos</h1><p>Consulta códigos, precios y existencias en un solo lugar.</p></div><button v-if="workspace.state.mode === 'demo'" class="button primary" @click="edit()">Nuevo artículo</button></div>
  <p v-if="workspace.state.mode === 'central'" class="message info">Consulta de precios del servidor. El alta y la edición central de artículos aún no están disponibles en esta interfaz.</p>
  <section class="panel">
    <div class="table-toolbar"><label class="search-label"><span class="sr-only">Buscar en el catálogo</span><input v-model="query" type="search" placeholder="Buscar por nombre o código" /></label><label v-if="workspace.state.mode === 'demo'"><span class="sr-only">Categoría</span><select v-model="category"><option>Todos</option><option v-for="item in categories" :key="item">{{ item }}</option></select></label><span class="muted small">{{ filtered.length }} artículos</span></div>
    <div class="table-scroll"><table><thead><tr><th>Artículo</th><th>Código</th><th>Categoría</th><th class="numeric">Precio de venta</th><th class="numeric">Disponibles</th><th><span class="sr-only">Acciones</span></th></tr></thead><tbody><tr v-for="product in filtered" :key="product.itemId"><td><strong>{{ product.name }}</strong></td><td class="muted">{{ product.barcode }}</td><td>{{ product.category }}</td><td class="numeric">{{ money(product.price) }}</td><td class="numeric"><span :class="['badge', product.stock === '0' ? 'warning' : 'neutral']">{{ product.stock === '' ? 'Sin datos' : product.stock }}</span></td><td><button v-if="workspace.state.mode === 'demo'" class="text-button" :aria-label="`Editar ${product.name}`" @click="edit(product)">Editar</button></td></tr></tbody></table></div>
    <div v-if="!filtered.length" class="empty-state"><h3>{{ workspace.state.loading ? 'Buscando artículos...' : 'Sin artículos para mostrar' }}</h3><p>{{ workspace.state.mode === 'central' ? 'Escribe un nombre o código para consultar los precios vigentes.' : 'Cambia la búsqueda o crea tu primer artículo.' }}</p></div>
    <footer class="table-footer">{{ workspace.state.mode === 'demo' ? 'Catálogo de demostración. Los cambios se guardan en este navegador.' : 'Los precios corresponden a la lista configurada y a la fecha de consulta.' }}</footer>
  </section>
  <dialog ref="dialog" class="dialog" aria-labelledby="product-title" @cancel="standard.saving && $event.preventDefault()">
    <form @submit.prevent="save">
      <div class="dialog-heading"><div><p class="muted small">Catálogo de demostración</p><h2 id="product-title">{{ form.itemId ? 'Editar artículo' : 'Nuevo artículo' }}</h2></div><button type="button" class="text-button" :disabled="standard.saving" @click="dialog?.close()">Cerrar</button></div>
      <div class="form-grid"><label class="span-all">Nombre del artículo<input v-model="form.name" required maxlength="120" placeholder="Ej.: Café molido 250 g" /></label><label>Código de barras<input v-model="form.barcode" required maxlength="48" placeholder="Código único" /></label><label>Categoría<select v-model="form.category"><option v-for="item in categories" :key="item">{{ item }}</option></select></label><label>Precio de venta (Gs.)<input v-model="form.price" required inputmode="numeric" pattern="[0-9]{1,12}" placeholder="0" /></label><label>Existencias de práctica<input v-model="form.stock" required inputmode="numeric" pattern="[0-9]{1,6}" /></label></div>
      <p class="muted small field-space">Los campos son obligatorios. Usa precios enteros en guaraníes. Al editar un artículo, se retira del carrito actual para revisar su precio.</p>
      <p v-if="standard.error" class="message error" role="alert">{{ standard.error }}</p>
      <div class="dialog-actions"><button type="button" class="button secondary" :disabled="standard.saving" @click="dialog?.close()">Cancelar</button><button class="button primary" :disabled="standard.saving">{{ standard.saving ? 'Guardando...' : 'Guardar artículo' }}</button></div>
    </form>
  </dialog>
</template>
