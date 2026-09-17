<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { cashChange, categories, lineTotal, money, type Workspace } from '../workspace'
const props = defineProps<{ workspace: Workspace }>()
const query = ref('')
const category = ref('Todos')
const payment = ref('CASH')
const received = ref('')
const checkoutDialog = ref<HTMLDialogElement>()
const checkoutButton = ref<HTMLButtonElement>()
const localError = ref('')
let timer: ReturnType<typeof setTimeout>
const products = computed(() => props.workspace.state.products.filter(item =>
  (category.value === 'Todos' || item.category === category.value) &&
  (!query.value || `${item.name} ${item.barcode}`.toLocaleLowerCase('es').includes(query.value.toLocaleLowerCase('es')))))
const change = computed(() => cashChange(received.value, props.workspace.total.value))
watch(query, value => { clearTimeout(timer); timer = setTimeout(() => props.workspace.searchCatalog(value), 300) })
onBeforeUnmount(() => clearTimeout(timer))
function scan() {
  const exact = products.value.find(product => product.barcode === query.value.trim())
  if (exact) { props.workspace.add(exact); query.value = '' }
}
function review() {
  received.value = props.workspace.total.value.split('.')[0]
  localError.value = ''
  checkoutDialog.value?.showModal()
}
async function confirm() {
  localError.value = ''
  if (await props.workspace.checkout(payment.value, received.value)) checkoutDialog.value?.close()
  else localError.value = props.workspace.state.error
}
</script>

<template>
  <div class="page-title"><div><h1>Punto de venta</h1><p>Busca un artículo, agrégalo a la venta y registra el cobro.</p></div><span class="badge neutral">Venta contado</span></div>
  <p v-if="workspace.state.mode === 'central'" class="message info">Puedes consultar los precios centrales. El cobro central aún requiere integrar caja, stock y numeración en esta interfaz. El circuito completo está disponible en la demostración.</p>
  <div class="pos-layout">
    <section class="product-area" aria-label="Selección de artículos">
      <form class="search-field" role="search" @submit.prevent="scan"><label for="pos-search" class="sr-only">Buscar artículo o código de barras</label><input id="pos-search" v-model="query" type="search" placeholder="Buscar artículo o escanear código de barras" autocomplete="off" /><span class="search-hint">Enter para agregar</span></form>
      <div v-if="workspace.state.mode === 'demo'" class="filter-tabs" aria-label="Filtrar por categoría"><button v-for="item in ['Todos', ...categories]" :key="item" :class="{ active: category === item }" :aria-pressed="category === item" @click="category = item">{{ item }}</button></div>
      <div class="section-caption"><span>{{ products.length }} artículos</span><span>Precios en guaraníes</span></div>
      <p v-if="workspace.state.loading" class="empty-state" role="status">Buscando artículos...</p>
      <div v-else-if="products.length" class="product-grid">
        <button v-for="product in products" :key="product.itemId" class="product-card" :disabled="workspace.state.mode !== 'demo' || product.stock === '0'" :aria-label="`Agregar ${product.name}`" @click="workspace.add(product)">
          <span class="product-category" :data-category="product.category">{{ product.category }}</span>
          <span class="product-name">{{ product.name }}</span>
          <span class="product-code">{{ product.barcode }}</span>
          <span class="product-bottom"><strong>{{ money(product.price) }}</strong><span>{{ product.stock === '' ? 'Consulta' : product.stock === '0' ? 'Sin stock' : product.stock + ' disponibles' }}</span></span>
        </button>
      </div>
      <div v-else class="empty-state"><h3>{{ query ? 'No encontramos ese artículo' : 'Consulta el catálogo central' }}</h3><p>{{ query ? 'Prueba con otro nombre o código de barras.' : 'Escribe el nombre o código del artículo que buscas.' }}</p><button v-if="query" class="button secondary" @click="query = ''; category = 'Todos'">Limpiar búsqueda</button></div>
    </section>
    <aside class="sale-panel" aria-label="Venta actual">
      <header><div><h2>Venta actual</h2><p>Consumidor final</p></div><span class="badge neutral">{{ workspace.units.value }} unidades</span></header>
      <div v-if="!workspace.state.cart.length" class="cart-empty"><span class="empty-receipt" aria-hidden="true"></span><h3>Comienza una nueva venta</h3><p>Selecciona los artículos del catálogo.<br />Aquí verás el detalle del cobro.</p></div>
      <ul v-else class="cart-lines"><li v-for="line in workspace.state.cart" :key="line.itemId"><div class="cart-line-heading"><strong>{{ line.name }}</strong><span>{{ money(lineTotal(line.price, line.quantity)) }}</span></div><div class="cart-line-footer"><span>{{ money(line.price) }} por unidad</span><div class="quantity-control"><button :aria-label="`Quitar una unidad de ${line.name}`" @click="workspace.remove(line.itemId)">Restar</button><output :aria-label="`Cantidad de ${line.name}`">{{ line.quantity }}</output><button :aria-label="`Agregar una unidad de ${line.name}`" @click="workspace.add(line)">Sumar</button></div></div></li></ul>
      <footer class="sale-total"><div><span>Artículos</span><span>{{ workspace.units.value }}</span></div><div class="total-line"><span>Total</span><strong>{{ money(workspace.total.value) }}</strong></div><button ref="checkoutButton" class="button primary full" :disabled="!workspace.state.cart.length || workspace.state.saving" @click="review">Revisar y cobrar</button><p>Revisa el importe antes de confirmar.</p></footer>
    </aside>
  </div>
  <dialog ref="checkoutDialog" class="dialog" aria-labelledby="checkout-title" @cancel="workspace.state.saving && $event.preventDefault()" @close="checkoutButton?.focus()">
    <form @submit.prevent="confirm">
      <div class="dialog-heading"><div><p class="muted small">Último paso</p><h2 id="checkout-title">Confirmar cobro</h2></div><button type="button" class="text-button" :disabled="workspace.state.saving" @click="checkoutDialog?.close()">Volver</button></div>
      <p class="message info">Venta de práctica. No emite un comprobante fiscal ni se envía al servidor.</p>
      <div class="checkout-amount"><span>Total a cobrar</span><strong>{{ money(workspace.total.value) }}</strong></div>
      <label>Medio de pago<select v-model="payment" :disabled="workspace.state.saving"><option value="CASH">Efectivo</option><option value="CARD">Tarjeta</option></select></label>
      <template v-if="payment === 'CASH'"><label class="field-space">Importe recibido (Gs.)<input v-model="received" inputmode="numeric" pattern="[0-9]{1,15}" required :disabled="workspace.state.saving" /></label><p class="change-line"><span>Vuelto</span><strong>{{ change === null ? 'Importe insuficiente' : money(change) }}</strong></p></template>
      <p v-else class="muted small field-space">Confirma el pago en el dispositivo de tarjetas antes de registrar la venta.</p>
      <p v-if="localError" role="alert" class="message error">{{ localError }}</p>
      <div class="dialog-actions"><button type="button" class="button secondary" :disabled="workspace.state.saving" @click="checkoutDialog?.close()">Volver a la venta</button><button class="button primary" :disabled="workspace.state.saving || (payment === 'CASH' && change === null)">{{ workspace.state.saving ? 'Guardando...' : 'Confirmar venta de práctica' }}</button></div>
    </form>
  </dialog>
</template>
