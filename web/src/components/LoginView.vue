<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { Workspace } from '../workspace'
const props = defineProps<{ workspace: Workspace }>()
const user = ref('')
const password = ref('')
const visible = ref(false)
const settings = reactive({ ...props.workspace.state.settings })
const configError = ref('')
async function submit() {
  await props.workspace.login(user.value, password.value)
  if (props.workspace.state.mode) password.value = ''
}
function save() {
  configError.value = ''
  try { props.workspace.saveSettings(settings); configError.value = 'Configuración guardada.' }
  catch (error) { configError.value = error instanceof Error ? error.message : 'Revisa los datos.' }
}
</script>

<template>
  <main class="login-layout">
    <section class="login-intro" aria-label="Ypy, gestión comercial">
      <a class="brand brand-light" href="#">ypy<span>Gestión comercial</span></a>
      <div class="login-message">
        <p class="intro-label">Tu negocio, en orden.</p>
        <h1>Todo listo para<br />una nueva jornada.</h1>
        <p>Un espacio para vender, consultar tus artículos y llevar el control de cada operación.</p>
        <div class="login-preview" aria-hidden="true">
          <div class="preview-heading"><span>Tu espacio de trabajo</span><span>Ypy</span></div>
          <div class="preview-row"><span>Punto de venta</span><span>Buscar, agregar y cobrar</span></div>
          <div class="preview-row"><span>Catálogo</span><span>Artículos y existencias</span></div>
          <div class="preview-row"><span>Historial</span><span>Cada operación a mano</span></div>
        </div>
      </div>
      <p class="intro-footer">Gestión comercial para el trabajo de todos los días.</p>
    </section>
    <section class="login-side">
      <div class="login-form-wrap">
        <p class="subtle">Bienvenido a Ypy</p>
        <h2>Iniciar sesión</h2>
        <p class="muted">Ingresa con el usuario asignado a tu terminal.</p>
        <form class="login-form" @submit.prevent="submit">
          <label>Usuario<input v-model="user" name="username" autocomplete="username" required autofocus placeholder="Tu usuario" :disabled="workspace.state.busy" /></label>
          <label>Contraseña
            <span class="password-field"><input v-model="password" name="password" :type="visible ? 'text' : 'password'" autocomplete="current-password" required placeholder="Tu contraseña" :disabled="workspace.state.busy" /><button type="button" class="text-button" :aria-pressed="visible" @click="visible = !visible">{{ visible ? 'Ocultar' : 'Mostrar' }}</button></span>
          </label>
          <p v-if="workspace.state.loginError" class="message error" role="alert">{{ workspace.state.loginError }}</p>
          <button class="button primary full" :disabled="workspace.state.busy">{{ workspace.state.busy ? 'Verificando acceso...' : 'Iniciar sesión' }}</button>
        </form>
        <div class="demo-entry">
          <h3>Conoce el espacio de trabajo</h3>
          <p>Prueba ventas y formularios con datos de ejemplo. No modifica la información de tu empresa.</p>
          <button class="button secondary full" :disabled="workspace.state.busy" @click="workspace.enterDemo">Explorar demostración</button>
        </div>
        <details class="login-settings">
          <summary>Configuración de acceso</summary>
          <form class="form-grid compact" @submit.prevent="save">
            <label>Empresa<input v-model="settings.empresa" inputmode="numeric" pattern="[1-9][0-9]{0,8}" required /></label>
            <label>Sucursal<input v-model="settings.sucursal" inputmode="numeric" pattern="[1-9][0-9]{0,8}" required /></label>
            <label>Terminal<input v-model="settings.terminal" inputmode="numeric" pattern="[1-9][0-9]{0,8}" required /></label>
            <label>Lista de precios<input v-model="settings.lista" inputmode="numeric" pattern="[1-9][0-9]{0,8}" required /></label>
            <p v-if="configError" class="span-all small" role="status">{{ configError }}</p>
            <button class="button secondary span-all">Guardar configuración</button>
          </form>
        </details>
        <p class="login-help">Si no tienes un usuario, solicita acceso al responsable de tu empresa.</p>
      </div>
      <p class="login-footer">Ypy MVP <span>Entorno local</span></p>
    </section>
  </main>
</template>
