<script setup lang="ts">
import { reactive, ref } from 'vue'
import { defaultTheme, type ThemePreferences, type Workspace } from '../workspace'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import ColorPicker from 'primevue/colorpicker'
const props = defineProps<{ workspace: Workspace }>()
const form = reactive({ ...props.workspace.state.settings })
const theme = reactive<ThemePreferences>({ ...props.workspace.state.theme })
const error = ref('')
const densityOptions = [{ label: 'Cómoda', value: 'comfortable' }, { label: 'Compacta', value: 'compact' }]
function save() {
  error.value = ''
  try { props.workspace.saveSettings(form) }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Revisa la configuración.' }
}
function saveAppearance() {
  error.value = ''
  try { props.workspace.saveTheme(theme) }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Revisa los colores.' }
}
function resetAppearance() {
  Object.assign(theme, defaultTheme)
  props.workspace.resetTheme()
}
</script>

<template>
  <div class="page-title"><div><h1>Configuración</h1><p>Datos de acceso y funcionamiento de esta terminal.</p></div></div>
  <div class="settings-layout"><section class="panel settings-panel"><h2>Acceso a la empresa</h2><p class="muted">Estos datos los proporciona el administrador de tu instalación.</p><p v-if="workspace.state.mode === 'central'" class="message info">El ámbito pertenece a tu sesión. Para cambiarlo, cierra sesión y utiliza la configuración de acceso.</p><form @submit.prevent="save"><fieldset :disabled="workspace.state.mode === 'central'"><div class="form-grid"><label>Identificador de empresa<InputText v-model="form.empresa" required inputmode="numeric" pattern="[1-9][0-9]{0,8}" /></label><label>Identificador de sucursal<InputText v-model="form.sucursal" required inputmode="numeric" pattern="[1-9][0-9]{0,8}" /></label><label>Identificador de terminal<InputText v-model="form.terminal" required inputmode="numeric" pattern="[1-9][0-9]{0,8}" /></label><label>Lista de precios<InputText v-model="form.lista" required inputmode="numeric" pattern="[1-9][0-9]{0,8}" /></label></div><p v-if="error" role="alert" class="message error">{{ error }}</p><div class="form-actions"><button class="button primary">Guardar configuración</button></div></fieldset></form></section>
    <section class="panel settings-panel appearance-panel"><h2>Apariencia</h2><p class="muted">Personaliza los colores y la densidad de las pantallas de este navegador.</p><form @submit.prevent="saveAppearance"><div class="appearance-grid"><label>Color principal<ColorPicker v-model="theme.primary" format="hex" /></label><label>Color de acento<ColorPicker v-model="theme.accent" format="hex" /></label><label>Fondo general<ColorPicker v-model="theme.background" format="hex" /></label><label>Superficie<ColorPicker v-model="theme.surface" format="hex" /></label><label>Texto principal<ColorPicker v-model="theme.ink" format="hex" /></label><label>Texto secundario<ColorPicker v-model="theme.muted" format="hex" /></label><label>Bordes<ColorPicker v-model="theme.line" format="hex" /></label></div><label class="field-space">Densidad de la interfaz<Select v-model="theme.density" :options="densityOptions" option-label="label" option-value="value" /></label><p class="muted small">Los colores se aplican a los componentes comunes. La versión predeterminada conserva la identidad verde y terracota de Ypy.</p><div class="form-actions"><button type="button" class="button secondary" @click="resetAppearance">Restaurar predeterminada</button><button class="button primary">Guardar apariencia</button></div></form></section>
    <aside class="settings-aside"><h2>Acerca de este espacio</h2><dl><div><dt>Modalidad</dt><dd>{{ workspace.state.mode === 'demo' ? 'Demostración local' : 'Conexión central' }}</dd></div><div><dt>Moneda</dt><dd>Guaraní paraguayo (PYG)</dd></div><div><dt>Almacenamiento</dt><dd>{{ workspace.state.mode === 'demo' ? 'En este navegador' : 'Servidor central' }}</dd></div></dl><p>La demostración conserva artículos y ventas entre visitas. Borrar los datos del navegador elimina esa información.</p><p>La sesión central se verifica con el servidor. Cerrar sesión revoca el acceso actual.</p><button class="text-button" @click="workspace.navigate('ayuda')">Consultar guía de uso</button></aside></div>
</template>
