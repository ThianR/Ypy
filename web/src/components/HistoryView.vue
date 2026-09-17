<script setup lang="ts">
import { computed, ref } from 'vue'
import { dateLabel, lineTotal, money, type Workspace } from '../workspace'
import type { PosOperation } from '../pos/adapters/dexie'
const props = defineProps<{ workspace: Workspace }>()
const query = ref('')
const date = ref('')
const selected = ref<PosOperation>()
const dialog = ref<HTMLDialogElement>()
const filtered = computed(() => props.workspace.state.sales.filter(sale => {
  const localDate = new Date(sale.createdAt)
  const day = `${localDate.getFullYear()}-${String(localDate.getMonth() + 1).padStart(2, '0')}-${String(localDate.getDate()).padStart(2, '0')}`
  return sale.number.toLowerCase().includes(query.value.toLowerCase()) && (!date.value || date.value === day)
}))
function detail(sale: PosOperation) { selected.value = sale; dialog.value?.showModal() }
</script>

<template>
  <div class="page-title"><div><h1>Historial de ventas</h1><p>Revisa las operaciones registradas en este espacio de trabajo.</p></div><button class="button primary" @click="workspace.navigate('venta')">Nueva venta</button></div>
  <p v-if="workspace.state.mode === 'central'" class="message info">El historial comercial central aún no está conectado a esta interfaz. No se muestran ventas de demostración en una sesión central.</p>
  <section class="panel"><div class="table-toolbar"><label class="search-label"><span class="sr-only">Buscar comprobante</span><input v-model="query" type="search" placeholder="Buscar comprobante" /></label><label class="inline-label">Fecha<input v-model="date" type="date" /></label><button v-if="query || date" class="text-button" @click="query = ''; date = ''">Limpiar filtros</button></div>
    <div class="table-scroll"><table><thead><tr><th>Comprobante</th><th>Fecha y hora</th><th>Cliente</th><th>Estado</th><th class="numeric">Total</th><th><span class="sr-only">Acciones</span></th></tr></thead><tbody><tr v-for="sale in filtered" :key="sale.operationId"><td><strong>{{ sale.number }}</strong></td><td>{{ dateLabel(sale.createdAt) }}</td><td>Consumidor final</td><td><span class="badge neutral">Práctica local</span></td><td class="numeric">{{ money(sale.total) }}</td><td><button class="text-button" :aria-label="`Ver detalle de ${sale.number}`" @click="detail(sale)">Ver detalle</button></td></tr></tbody></table></div>
    <div v-if="!filtered.length" class="empty-state"><h3>{{ query || date ? 'No hay resultados para estos filtros' : 'Todavía no hay ventas registradas' }}</h3><p>{{ query || date ? 'Prueba con otra fecha o número de comprobante.' : 'Las ventas aparecerán aquí después de confirmar el cobro.' }}</p><button v-if="workspace.state.mode === 'demo' && !query && !date" class="button secondary" @click="workspace.navigate('venta')">Crear una venta</button></div>
    <footer class="table-footer">{{ filtered.length }} operaciones. {{ workspace.state.mode === 'demo' ? 'Comprobantes de práctica, sin validez fiscal.' : 'Consulta central pendiente de integración.' }}</footer>
  </section>
  <dialog ref="dialog" class="dialog" aria-labelledby="receipt-title"><template v-if="selected"><div class="dialog-heading"><div><p class="muted small">Comprobante de práctica</p><h2 id="receipt-title">{{ selected.number }}</h2></div><button class="text-button" @click="dialog?.close()">Cerrar</button></div><p class="muted">{{ dateLabel(selected.createdAt) }}</p><p>Consumidor final</p><ul class="receipt-lines"><li v-for="(line, index) in selected.receiptLines" :key="index"><div><strong>{{ line.name }}</strong><p>{{ line.quantity }} unidades a {{ money(line.price) }}</p></div><span>{{ money(lineTotal(line.price, line.quantity)) }}</span></li></ul><div class="checkout-amount"><span>Total</span><strong>{{ money(selected.total) }}</strong></div><p class="message info">Guardado en este navegador. Sin envío al servidor ni validez fiscal.</p><div class="dialog-actions"><button class="button primary" @click="dialog?.close()">Volver al historial</button></div></template></dialog>
</template>
