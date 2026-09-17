import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import App from './App.vue'
import './styles.css'
import 'primeicons/primeicons.css'

const app = createApp(App)

// PrimeVue aporta comportamiento y accesibilidad; el aspecto sigue siendo el tema propio de Ypy.
app.use(PrimeVue, {
  unstyled: true,
  ripple: false,
})

app.mount('#app')
