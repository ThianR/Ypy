<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import LoginView from './components/LoginView.vue'
import SalesDesk from './components/SalesDesk.vue'
import CatalogView from './components/CatalogView.vue'
import HistoryView from './components/HistoryView.vue'
import SettingsView from './components/SettingsView.vue'
import UnitsView from './components/UnitsView.vue'
import CategoriesView from './components/CategoriesView.vue'
import MessageCenter from './ui/MessageCenter.vue'
import { dateLabel, money, useWorkspace, type Page } from './workspace'

const workspace = useWorkspace()
const state = workspace.state
const menuOpen = ref(false)
const logoutDialog = ref<HTMLDialogElement>()
const pages: { id: Page; label: string; section: string }[] = [
  { id: 'inicio', label: 'Inicio', section: 'Mi espacio' },
  { id: 'venta', label: 'Punto de venta', section: 'Operación' },
  { id: 'catalogo', label: 'Catálogo de artículos', section: 'Operación' },
  { id: 'historial', label: 'Historial de ventas', section: 'Operación' },
  { id: 'configuracion', label: 'Configuración', section: 'Administración' },
  { id: 'unidades', label: 'Unidades de medida', section: 'Administración' },
  { id: 'categorias', label: 'Categorías', section: 'Administración' },
  { id: 'ayuda', label: 'Guía de uso', section: 'Administración' },
]
watch(() => state.theme, theme => {
  const root = document.documentElement
  root.style.setProperty('--green', theme.primary)
  root.style.setProperty('--green-dark', theme.primary)
  root.style.setProperty('--orange', theme.accent)
  root.style.setProperty('--page-bg', theme.background)
  root.style.setProperty('--paper', theme.surface)
  root.style.setProperty('--ink', theme.ink)
  root.style.setProperty('--muted', theme.muted)
  root.style.setProperty('--line', theme.line)
  root.dataset.density = theme.density
}, { deep: true, immediate: true })
function navigate(page: Page) { workspace.navigate(page); menuOpen.value = false }
function onHash() {
  const page = window.location.hash.slice(1) as Page
  if (pages.some(item => item.id === page)) { state.page = page; menuOpen.value = false }
}
function onConnection() { state.online = navigator.onLine }
function onUnload(event: BeforeUnloadEvent) {
  if (state.cart.length) { event.preventDefault(); event.returnValue = '' }
}
function askLogout() { if (state.cart.length) logoutDialog.value?.showModal(); else void workspace.logout() }
async function confirmLogout() { await workspace.logout(); logoutDialog.value?.close() }
onMounted(() => {
  void workspace.restore()
  window.addEventListener('hashchange', onHash)
  window.addEventListener('online', onConnection)
  window.addEventListener('offline', onConnection)
  window.addEventListener('beforeunload', onUnload)
})
onBeforeUnmount(() => {
  window.removeEventListener('hashchange', onHash)
  window.removeEventListener('online', onConnection)
  window.removeEventListener('offline', onConnection)
  window.removeEventListener('beforeunload', onUnload)
})
</script>

<template>
  <div v-if="state.checking" class="boot-screen" role="status"><span class="brand">ypy</span><p>Preparando tu espacio de trabajo...</p></div>
  <LoginView v-else-if="!state.mode" :workspace="workspace" />
  <div v-else class="app-shell">
    <a class="skip-link" href="#main-content">Ir al contenido</a>
    <button v-if="menuOpen" class="menu-backdrop" aria-label="Cerrar menú" @click="menuOpen = false"></button>
    <aside id="main-menu" :class="['sidebar', { open: menuOpen }]" aria-label="Menú principal">
      <a class="brand" href="#inicio" @click.prevent="navigate('inicio')">ypy<span>Gestión comercial</span></a>
      <div class="workspace-label"><span class="workspace-avatar">{{ state.mode === 'demo' ? 'D' : 'E' }}</span><div><strong>{{ state.mode === 'demo' ? 'Empresa de demostración' : 'Empresa ' + state.settings.empresa }}</strong><span>{{ state.mode === 'demo' ? 'Espacio de práctica' : 'Sucursal ' + state.settings.sucursal }}</span></div></div>
      <nav aria-label="Secciones del espacio"><div v-for="section in ['Mi espacio', 'Operación', 'Administración']" :key="section" class="nav-section"><p>{{ section }}</p><a v-for="page in pages.filter(item => item.section === section)" :key="page.id" :href="'#' + page.id" :aria-current="state.page === page.id ? 'page' : undefined" @click.prevent="navigate(page.id)"><span class="nav-link-label"><span :class="['nav-icon', 'icon-' + page.id]" aria-hidden="true"></span>{{ page.label }}</span><span v-if="page.id === 'venta' && workspace.units.value" class="nav-count">{{ workspace.units.value }}</span></a></div></nav>
      <div class="sidebar-help"><strong>Un paso a la vez</strong><p>Consulta cómo registrar tu primera venta.</p><button class="text-button" @click="navigate('ayuda')">Ver guía de uso</button></div>
      <footer class="sidebar-footer"><span class="user-avatar">{{ state.user.slice(0, 1).toUpperCase() }}</span><div><strong>{{ state.mode === 'demo' ? 'Demostración' : state.user }}</strong><button class="text-button" :disabled="state.busy" @click="askLogout">{{ state.mode === 'demo' ? 'Salir de la demostración' : 'Cerrar sesión' }}</button></div></footer>
    </aside>
    <div class="main-shell">
      <header class="topbar"><div class="breadcrumb"><button class="button secondary menu-toggle" :aria-expanded="menuOpen" aria-controls="main-menu" @click="menuOpen = !menuOpen">Menú</button><span>Mi negocio</span><span class="breadcrumb-separator">/</span><strong>{{ pages.find(item => item.id === state.page)?.label }}</strong></div><div class="topbar-right"><span :class="['connection', { offline: !state.online }]">{{ state.online ? (state.mode === 'demo' ? 'Datos locales' : 'Sesión central') : 'Sin conexión' }}</span><span class="terminal-label">{{ state.mode === 'demo' ? 'Terminal de práctica' : 'Terminal ' + state.settings.terminal }}</span></div></header>
      <div v-if="state.mode === 'demo'" class="demo-banner"><strong>Modo demostración</strong><span>Practica con datos de ejemplo. Tus cambios se guardan solo en este navegador.</span><button class="text-button" @click="askLogout">Volver al acceso</button></div>
      <main id="main-content" class="page-content" tabindex="-1">
        <MessageCenter :notice="state.notice" :error="state.error" @close-notice="state.notice = ''" @close-error="state.error = ''" />
        <template v-if="state.page === 'inicio'">
          <div class="page-title"><div><h1>Tu jornada, de un vistazo</h1><p>{{ new Intl.DateTimeFormat('es-PY', { dateStyle: 'full' }).format(new Date()) }}</p></div><button class="button primary" @click="navigate('venta')">Nueva venta</button></div>
          <section class="welcome-panel"><div><span class="badge welcome-badge">{{ state.mode === 'demo' ? 'Tu espacio de práctica' : 'Bienvenido, ' + state.user }}</span><h2>Todo empieza con<br />una buena atención.</h2><p>Encuentra un artículo, prepara la venta y revisa el cobro. Tu próxima operación empieza aquí.</p><button class="button welcome-button" @click="navigate('venta')">Ir al punto de venta</button></div><div class="welcome-guide"><div><span>Buscar</span><p>Por nombre o código de barras</p></div><div><span>Preparar</span><p>Revisa artículos y cantidades</p></div><div><span>Cobrar</span><p>Confirma el medio de pago</p></div></div></section>
          <div class="stats-row"><section><p>Ventas de hoy</p><strong>{{ state.mode === 'demo' ? money(workspace.todayTotal.value) : 'Sin conexión al reporte' }}</strong><span>{{ state.mode === 'demo' ? 'Importe de práctica registrado' : 'Reporte central pendiente de integración' }}</span></section><section><p>Operaciones de hoy</p><strong>{{ state.mode === 'demo' ? workspace.todaySales.value.length : 'No disponible' }}</strong><span>{{ state.mode === 'demo' ? 'Ventas en este navegador' : 'No incluye datos de demostración' }}</span></section><section><p>Artículos disponibles</p><strong>{{ state.mode === 'demo' ? state.products.length : 'Consultar catálogo' }}</strong><span>{{ state.mode === 'demo' ? 'En tu catálogo de práctica' : 'Búsqueda de precios en el servidor' }}</span></section></div>
          <div class="home-bottom"><section class="panel"><div class="panel-heading"><h2>Actividad reciente</h2><button class="text-button" @click="navigate('historial')">Ver historial</button></div><div v-if="!state.sales.length" class="empty-state compact-empty"><h3>{{ state.mode === 'demo' ? 'Tu primera venta está por llegar' : 'Sesión iniciada correctamente' }}</h3><p>{{ state.mode === 'demo' ? 'Cuando confirmes una venta, podrás revisarla desde aquí.' : 'Consulta los precios del catálogo central o explora el circuito de ventas en la demostración.' }}</p><button class="button secondary" @click="navigate(state.mode === 'demo' ? 'venta' : 'catalogo')">{{ state.mode === 'demo' ? 'Comenzar una venta' : 'Consultar catálogo' }}</button></div><ul v-else class="activity-list"><li v-for="sale in state.sales.slice(0, 4)" :key="sale.operationId"><div><strong>{{ sale.number }}</strong><span>{{ dateLabel(sale.createdAt) }}</span></div><strong>{{ money(sale.total) }}</strong></li></ul></section><section class="quick-start"><h2>A mano para tu jornada</h2><button @click="navigate('catalogo')"><strong>Catálogo de artículos</strong><span>Consulta precios y existencias</span></button><button @click="navigate('configuracion')"><strong>Configurar el acceso</strong><span>Empresa, sucursal y terminal</span></button><button @click="navigate('ayuda')"><strong>Guía para comenzar</strong><span>Conoce el recorrido de una venta</span></button></section></div>
        </template>
        <SalesDesk v-else-if="state.page === 'venta'" :workspace="workspace" />
        <CatalogView v-else-if="state.page === 'catalogo'" :workspace="workspace" />
        <HistoryView v-else-if="state.page === 'historial'" :workspace="workspace" />
        <SettingsView v-else-if="state.page === 'configuracion'" :workspace="workspace" />
        <UnitsView v-else-if="state.page === 'unidades'" :workspace="workspace" />
        <CategoriesView v-else-if="state.page === 'categorias'" :workspace="workspace" />
        <template v-else><div class="page-title"><div><h1>Guía para comenzar</h1><p>El recorrido básico de una venta en Ypy.</p></div></div><section class="panel guide-panel"><h2>Tu primera venta de práctica</h2><ol class="steps-list"><li><h3>Revisa el catálogo</h3><p>Consulta los artículos de ejemplo. Usa Nuevo artículo para crear uno o Editar para cambiar su precio y existencias.</p></li><li><h3>Prepara la venta</h3><p>En Punto de venta, busca por nombre o código. Selecciona un artículo para agregarlo y ajusta las cantidades.</p></li><li><h3>Revisa y cobra</h3><p>Selecciona Revisar y cobrar. Elige efectivo o tarjeta. Si cobras en efectivo, ingresa el importe recibido para calcular el vuelto.</p></li><li><h3>Consulta el comprobante</h3><p>Confirma la venta y abre Historial de ventas. El detalle se conserva después de recargar la página.</p></li></ol><button class="button primary" @click="navigate('venta')">Ir al punto de venta</button></section><section class="panel guide-panel field-space"><h2>Alcance de esta versión</h2><p>La demostración guarda operaciones en este navegador y descuenta existencias de práctica. No sincroniza operaciones ni emite comprobantes fiscales.</p><p>La conexión central permite iniciar y cerrar sesión y consultar precios. La venta central, el alta central de artículos, caja, compras e inventario administrativo requieren completar sus pantallas e integración.</p></section></template>
        <footer class="page-footer"><span>Ypy Gestión comercial</span><span>{{ state.mode === 'demo' ? 'Entorno de demostración' : 'Sesión central verificada' }}</span></footer>
      </main>
    </div>
    <dialog ref="logoutDialog" class="dialog" aria-labelledby="logout-title"><h2 id="logout-title">Hay una venta sin confirmar</h2><p>Si sales ahora, se descartará el carrito actual. Las ventas ya confirmadas se conservan.</p><div class="dialog-actions"><button class="button secondary" :disabled="state.busy" @click="logoutDialog?.close()">Continuar trabajando</button><button class="button primary" :disabled="state.busy" @click="confirmLogout">Salir y descartar carrito</button></div></dialog>
  </div>
</template>
