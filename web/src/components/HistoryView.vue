<script setup lang="ts">
import { computed, ref } from 'vue'
import { dateLabel, lineTotal, money, type Workspace } from '../workspace'
import type { PosOperation } from '../pos/adapters/dexie'
import InputText from 'primevue/inputtext'
import Dialog from 'primevue/dialog'
import DatePicker from 'primevue/datepicker'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
const props = defineProps<{ workspace: Workspace }>()
const query = ref('')
const dateFilter = ref<Date | null>(null)
const selected = ref<PosOperation>()
const dialogOpen = ref(false)
const filtered = computed(() => props.workspace.state.sales.filter(sale => {
  const localDate = new Date(sale.createdAt)
  return sale.number.toLowerCase().includes(query.value.toLowerCase()) && (!dateFilter.value || dateFilter.value.getFullYear() === localDate.getFullYear() && dateFilter.value.getMonth() === localDate.getMonth() && dateFilter.value.getDate() === localDate.getDate())
}))
function detail(sale: PosOperation) { selected.value = sale; dialogOpen.value = true }
</script>

<template>
  <div class="page-title"><div><h1>Historial de ventas</h1><p>Revisa las operaciones registradas en este espacio de trabajo.</p></div><button class="button primary" @click="workspace.navigate('venta')">Nueva venta</button></div>
  <p v-if="workspace.state.mode === 'central'" class="message info">El historial comercial central aún no está conectado a esta interfaz. No se muestran ventas de demostración en una sesión central.</p>
  <section class="panel"><div class="table-toolbar"><label class="search-label"><span class="sr-only">Buscar comprobante</span><InputText v-model="query" type="search" placeholder="Buscar comprobante" /></label><label class="inline-label">Fecha<DatePicker v-model="dateFilter" date-format="dd/mm/yy" show-icon /></label><button v-if="query || dateFilter" class="text-button" @click="query = ''; dateFilter = null">Limpiar filtros</button></div>
    <DataTable v-if="filtered.length" :value="filtered" data-key="operationId" paginator :rows="10" :rows-per-page-options="[10, 25, 50]" row-hover class="ypy-data-table"><Column field="number" header="Comprobante"><template #body="slotProps"><strong>{{ slotProps.data.number }}</strong></template></Column><Column header="Fecha y hora"><template #body="slotProps">{{ dateLabel(slotProps.data.createdAt) }}</template></Column><Column header="Cliente"><template #body>Consumidor final</template></Column><Column header="Estado"><template #body><span class="badge neutral">Práctica local</span></template></Column><Column header="Total" header-class="numeric" body-class="numeric"><template #body="slotProps">{{ money(slotProps.data.total) }}</template></Column><Column header="Acciones"><template #body="slotProps"><button class="text-button" :aria-label="`Ver detalle de ${slotProps.data.number}`" @click="detail(slotProps.data)">Ver detalle</button></template></Column></DataTable>
    <div v-if="!filtered.length" class="empty-state"><h3>{{ query || dateFilter ? 'No hay resultados para estos filtros' : 'Todavía no hay ventas registradas' }}</h3><p>{{ query || dateFilter ? 'Prueba con otra fecha o número de comprobante.' : 'Las ventas aparecerán aquí después de confirmar el cobro.' }}</p><button v-if="workspace.state.mode === 'demo' && !query && !dateFilter" class="button secondary" @click="workspace.navigate('venta')">Crear una venta</button></div>
    <footer class="table-footer">{{ filtered.length }} operaciones. {{ workspace.state.mode === 'demo' ? 'Comprobantes de práctica, sin validez fiscal.' : 'Consulta central pendiente de integración.' }}</footer>
  </section>
  <Dialog :visible="dialogOpen" modal :show-header="false" :pt="{ root: 'dialog', content: 'history-dialog-content' }" @update:visible="dialogOpen = $event"><template v-if="selected"><div class="dialog-heading"><div><p class="muted small">Comprobante de práctica</p><h2 id="receipt-title">{{ selected.number }}</h2></div><button class="text-button" @click="dialogOpen = false">Cerrar</button></div><p class="muted">{{ dateLabel(selected.createdAt) }}</p><p>Consumidor final</p><ul class="receipt-lines"><li v-for="(line, index) in selected.receiptLines" :key="index"><div><strong>{{ line.name }}</strong><p>{{ line.quantity }} unidades a {{ money(line.price) }}</p></div><span>{{ money(lineTotal(line.price, line.quantity)) }}</span></li></ul><div class="checkout-amount"><span>Total</span><strong>{{ money(selected.total) }}</strong></div><p class="message info">Guardado en este navegador. Sin envío al servidor ni validez fiscal.</p><div class="dialog-actions"><button class="button primary" @click="dialogOpen = false">Volver al historial</button></div></template></Dialog>
</template>
