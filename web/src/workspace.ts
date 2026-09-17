import { computed, reactive } from 'vue'
import { addDecimal } from './modules/maestros/money'
import { PosDatabase, recordSale, type CatalogProduct, type PosOperation } from './pos/adapters/dexie'

export type Page = 'inicio' | 'venta' | 'catalogo' | 'historial' | 'configuracion' | 'unidades' | 'categorias' | 'ayuda'
export type UnitOfMeasure = { id: string; codigo: string; nombre: string; dimension: string; admiteFraccion: boolean; createdAt: string; updatedAt: string; createdBy: string; updatedBy: string }
export type Category = { id: string; codigo: string; nombre: string; createdAt: string; updatedAt: string; createdBy: string; updatedBy: string }
export type Product = CatalogProduct & { stock: string }
export type CartLine = Product & { quantity: number }
export type Settings = { empresa: string; sucursal: string; terminal: string; lista: string }
export type ThemePreferences = {
  primary: string
  accent: string
  background: string
  surface: string
  ink: string
  muted: string
  line: string
  density: 'comfortable' | 'compact'
}
export type ProductForm = { itemId: string; barcode: string; name: string; category: string; price: string; stock: string }
export const categories = ['Almacén', 'Bebidas', 'Lácteos', 'Limpieza']

export const defaultTheme: ThemePreferences = {
  primary: '#154D47', accent: '#D87B45', background: '#F5F6F2',
  surface: '#FFFFFF', ink: '#1D2A2A', muted: '#687474', line: '#DDE3DE',
  density: 'comfortable',
}

const samples: Product[] = [
  { itemId: 'arroz', barcode: '779001', name: 'Arroz blanco 1 kg', category: 'Almacén', price: '6500', stock: '100' },
  { itemId: 'leche', barcode: '779002', name: 'Leche entera 1 L', category: 'Lácteos', price: '7500', stock: '60' },
  { itemId: 'aceite', barcode: '779003', name: 'Aceite de girasol 900 ml', category: 'Almacén', price: '12500', stock: '32' },
  { itemId: 'yerba', barcode: '779004', name: 'Yerba mate 500 g', category: 'Almacén', price: '18000', stock: '48' },
  { itemId: 'agua', barcode: '779005', name: 'Agua mineral 2 L', category: 'Bebidas', price: '5000', stock: '72' },
  { itemId: 'azucar', barcode: '779006', name: 'Azúcar blanca 1 kg', category: 'Almacén', price: '8000', stock: '45' },
  { itemId: 'yogur', barcode: '779007', name: 'Yogur natural 1 L', category: 'Lácteos', price: '11000', stock: '18' },
  { itemId: 'detergente', barcode: '779008', name: 'Detergente 750 ml', category: 'Limpieza', price: '9500', stock: '24' },
]

export function money(value: string) {
  const [integer, fraction = ''] = value.split('.')
  const decimals = fraction.replace(/0+$/, '')
  return `${new Intl.NumberFormat('es-PY').format(BigInt(integer || '0'))}${decimals ? ',' + decimals : ''} Gs.`
}

export function lineTotal(price: string, quantity: number) {
  return addDecimal(...Array.from({ length: quantity }, () => price))
}

export function cashChange(received: string, total: string) {
  if (!/^\d+(\.\d{1,6})?$/.test(received)) return null
  const change = addDecimal(received, `-${total}`)
  return change.startsWith('-') ? null : change
}

export function dateLabel(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Sin registro'
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export function useWorkspace() {
  const savedTheme = localStorage.getItem('ypy.theme')
  let parsedTheme: ThemePreferences = defaultTheme
  try {
    if (savedTheme) parsedTheme = { ...defaultTheme, ...JSON.parse(savedTheme) }
  } catch { parsedTheme = defaultTheme }
  const state = reactive({
    mode: '' as '' | 'demo' | 'central', page: 'inicio' as Page,
    checking: true, busy: false, saving: false, loading: false,
    user: '', session: sessionStorage.getItem('ypy.session') ?? '',
    error: '', notice: '', loginError: '', online: navigator.onLine,
    products: [] as Product[], cart: [] as CartLine[], sales: [] as PosOperation[],
    settings: {
      empresa: localStorage.getItem('ypy.empresaId') ?? '1',
      sucursal: localStorage.getItem('ypy.sucursalId') ?? '1',
      terminal: localStorage.getItem('ypy.terminalId') ?? '1',
      lista: localStorage.getItem('ypy.listaId') ?? '1',
    } as Settings,
    theme: parsedTheme,
    unitsOfMeasure: [] as UnitOfMeasure[],
    categories: [] as Category[],
  })
  let db: PosDatabase | undefined
  let catalogRequest = 0
  const total = computed(() => addDecimal(...state.cart.map(line => lineTotal(line.price, line.quantity))))
  const units = computed(() => state.cart.reduce((sum, line) => sum + line.quantity, 0))
  const todaySales = computed(() => state.sales.filter(sale => new Date(sale.createdAt).toDateString() === new Date().toDateString()))
  const todayTotal = computed(() => addDecimal(...todaySales.value.map(sale => sale.total)))

  function navigate(page: Page) {
    state.page = page
    window.location.hash = page
    state.error = ''
    state.notice = ''
  }

  async function loadUnits() {
    if (state.mode === 'central') {
      const response = await request('/api/v1/maestros/unidades-medida')
      if (!response.ok) throw new Error('No se pudieron consultar las unidades de medida.')
      const records = await response.json() as UnitOfMeasure[]
      state.unitsOfMeasure = records.map(item => ({ ...item, id: String(item.id), createdAt: item.createdAt || new Date().toISOString(), updatedAt: item.updatedAt || item.createdAt || new Date().toISOString(), createdBy: item.createdBy || 'Base de datos', updatedBy: item.updatedBy || 'Base de datos' }))
      return
    }
    try {
      const stored = localStorage.getItem('ypy.unidades')
      const now = new Date().toISOString()
      const parsed = stored ? JSON.parse(stored) : [{ id: 'un', codigo: 'UN', nombre: 'Unidad', dimension: 'UNIDAD', admiteFraccion: true, createdAt: now, updatedAt: now, createdBy: 'Sistema', updatedBy: 'Sistema' }]
      state.unitsOfMeasure = (Array.isArray(parsed) ? parsed : []).map(item => ({ ...item, createdAt: item.createdAt || now, updatedAt: item.updatedAt || item.createdAt || now, createdBy: item.createdBy || 'Usuario de demostración', updatedBy: item.updatedBy || item.createdBy || 'Usuario de demostración' }))
      localStorage.setItem('ypy.unidades', JSON.stringify(state.unitsOfMeasure))
    } catch { state.unitsOfMeasure = [] }
  }

  async function loadCategories() {
    if (state.mode === 'central') {
      const response = await request('/api/v1/maestros/categorias')
      if (!response.ok) throw new Error('No se pudieron consultar las categorías.')
      const records = await response.json() as Category[]
      state.categories = records.map(item => ({ ...item, id: String(item.id), createdBy: item.createdBy || 'Base de datos', updatedBy: item.updatedBy || 'Base de datos' }))
      return
    }
    try {
      const now = new Date().toISOString()
      const stored = localStorage.getItem('ypy.categorias')
      const parsed = stored ? JSON.parse(stored) : ['Almacén', 'Bebidas', 'Lácteos', 'Limpieza'].map((nombre, index) => ({ id: `cat-${index + 1}`, codigo: `CAT-${index + 1}`, nombre, createdAt: now, updatedAt: now, createdBy: 'Sistema', updatedBy: 'Sistema' }))
      state.categories = (Array.isArray(parsed) ? parsed : []).map(item => ({ ...item, createdAt: item.createdAt || now, updatedAt: item.updatedAt || item.createdAt || now, createdBy: item.createdBy || 'Usuario de demostración', updatedBy: item.updatedBy || item.createdBy || 'Usuario de demostración' }))
      localStorage.setItem('ypy.categorias', JSON.stringify(state.categories))
    } catch { state.categories = [] }
  }

  function saveCategory(category: Pick<Category, 'codigo' | 'nombre'> & { id?: string }) {
    const codigo = category.codigo.trim().toUpperCase()
    const nombre = category.nombre.trim()
    if (!/^[A-Z0-9_-]{1,20}$/.test(codigo)) throw new Error('El código debe tener entre 1 y 20 caracteres alfanuméricos.')
    if (!nombre) throw new Error('Ingresa el nombre de la categoría.')
    if (state.categories.some(item => item.codigo === codigo && item.id !== category.id)) throw new Error('Ya existe una categoría con ese código.')
    const now = new Date().toISOString(); const previous = state.categories.find(item => item.id === category.id); const actor = state.user || 'Usuario de demostración'
    const next = { id: category.id || crypto.randomUUID(), codigo, nombre, createdAt: previous?.createdAt || now, updatedAt: now, createdBy: previous?.createdBy || actor, updatedBy: actor }
    const index = state.categories.findIndex(item => item.id === next.id); const records = [...state.categories]
    if (index >= 0) records.splice(index, 1, next); else records.push(next)
    localStorage.setItem('ypy.categorias', JSON.stringify(records)); state.categories = records; state.notice = previous ? 'Categoría modificada correctamente.' : 'Categoría guardada correctamente.'
  }

  function deleteCategory(id: string) {
    if (!state.categories.some(item => item.id === id)) throw new Error('La categoría ya no existe.')
    const records = state.categories.filter(item => item.id !== id); localStorage.setItem('ypy.categorias', JSON.stringify(records)); state.categories = records; state.notice = 'Categoría eliminada correctamente.'
  }

  async function saveUnit(unit: Pick<UnitOfMeasure, 'codigo' | 'nombre' | 'dimension' | 'admiteFraccion'> & { id?: string }) {
    const codigo = unit.codigo.trim().toUpperCase()
    const nombre = unit.nombre.trim()
    const dimension = unit.dimension.trim().toUpperCase()
    if (!/^[A-Z0-9_-]{1,12}$/.test(codigo)) throw new Error('El código debe tener entre 1 y 12 caracteres alfanuméricos.')
    if (!nombre) throw new Error('Ingresa el nombre de la unidad.')
    if (!dimension) throw new Error('Ingresa la dimensión.')
    if (state.unitsOfMeasure.some(item => item.codigo === codigo && item.id !== unit.id)) throw new Error('Ya existe una unidad con ese código.')
    const now = new Date().toISOString()
    const previous = state.unitsOfMeasure.find(item => item.id === unit.id)
    if (state.mode === 'central') {
      const response = await request(unit.id ? '/api/v1/maestros/unidades-medida' : '/api/v1/maestros/unidades-medida', { method: unit.id ? 'PUT' : 'POST', body: JSON.stringify({ ...unit, id: unit.id ? Number(unit.id) : undefined }) })
      if (!response.ok) throw new Error('No se pudo guardar la unidad en la base de datos.')
      await loadUnits(); state.notice = previous ? 'Unidad de medida modificada correctamente.' : 'Unidad de medida guardada correctamente.'; return
    }
    const actor = state.user || 'Usuario de demostración'
    const next = { id: unit.id || crypto.randomUUID(), codigo, nombre, dimension, admiteFraccion: unit.admiteFraccion, createdAt: previous?.createdAt || now, updatedAt: now, createdBy: previous?.createdBy || actor, updatedBy: actor }
    const index = state.unitsOfMeasure.findIndex(item => item.id === next.id)
    const records = [...state.unitsOfMeasure]
    if (index >= 0) records.splice(index, 1, next); else records.push(next)
    localStorage.setItem('ypy.unidades', JSON.stringify(records))
    state.unitsOfMeasure = records
    state.notice = previous ? 'Unidad de medida modificada correctamente.' : 'Unidad de medida guardada correctamente.'
  }

  async function deleteUnit(id: string) {
    const index = state.unitsOfMeasure.findIndex(item => item.id === id)
    if (index < 0) throw new Error('La unidad ya no existe.')
    if (state.mode === 'central') { const response = await request(`/api/v1/maestros/unidades-medida/${id}`, { method: 'DELETE' }); if (!response.ok) throw new Error('No se pudo eliminar la unidad en la base de datos.'); await loadUnits(); state.notice = 'Unidad de medida eliminada correctamente.'; return }
    const records = state.unitsOfMeasure.filter(item => item.id !== id)
    localStorage.setItem('ypy.unidades', JSON.stringify(records))
    state.unitsOfMeasure = records
    state.notice = 'Unidad de medida eliminada correctamente.'
  }

  async function request(path: string, init: RequestInit = {}) {
    const response = await fetch(path, {
      ...init, signal: AbortSignal.timeout(10000),
      headers: { 'Content-Type': 'application/json', 'X-Session-ID': state.session, ...init.headers },
    })
    if (response.status === 401 && path !== '/api/v1/auth/login') {
      resetSession()
      state.loginError = 'La sesión venció. Inicia sesión nuevamente.'
      throw new Error(state.loginError)
    }
    return response
  }

  function resetSession() {
    state.mode = ''; state.session = ''; state.user = ''; state.cart = []
    state.products = []; state.sales = []
    sessionStorage.removeItem('ypy.session')
    sessionStorage.removeItem('ypy.mode')
    sessionStorage.removeItem('ypy.user')
    db?.close(); db = undefined
  }

  async function refresh() {
    if (!db) return
    const [catalog, stock, operations] = await Promise.all([db.catalog.toArray(), db.stock.toArray(), db.operations.orderBy('createdAt').reverse().toArray()])
    state.products = catalog.map(item => ({ ...item, stock: stock.find(row => row.itemId === item.itemId)?.quantity ?? '0' }))
    state.sales = operations
  }

  async function enterDemo() {
    state.busy = true; state.loginError = ''
    try {
      db?.close()
      db = new PosDatabase('ypy-demo-interfaz-v1')
      const demo = db
      await demo.transaction('rw', demo.catalog, demo.stock, async () => {
        if (await demo.catalog.count()) return
        await demo.catalog.bulkAdd(samples.map(({ stock, ...item }) => item))
        await demo.stock.bulkAdd(samples.map(item => ({ itemId: item.itemId, quantity: item.stock })))
      })
      await refresh()
      state.mode = 'demo'; state.user = 'Usuario de demostración'; state.session = ''
      sessionStorage.removeItem('ypy.session')
      sessionStorage.setItem('ypy.mode', 'demo')
      navigate('inicio')
    } catch {
      resetSession()
      state.loginError = 'No se pudo abrir el almacenamiento local. Comprueba que el navegador permita guardar datos.'
    } finally { state.busy = false }
  }

  async function login(user: string, password: string) {
    if (state.busy) return
    state.busy = true; state.loginError = ''
    try {
      const response = await request('/api/v1/auth/login', {
        method: 'POST', body: JSON.stringify({ login: user.trim(), password,
          empresa_id: Number(state.settings.empresa), sucursal_id: Number(state.settings.sucursal), terminal_id: Number(state.settings.terminal) }),
      })
      if (!response.ok) throw new Error(response.status === 401
        ? 'Usuario o contraseña incorrectos, o sin acceso a esta terminal.'
        : 'El servicio de acceso no está disponible. Revisa la conexión con el servidor.')
      const result = await response.json()
      if (!result.session_id) throw new Error('El servidor no devolvió una sesión válida.')
      state.session = result.session_id; state.user = user.trim(); state.mode = 'central'
      sessionStorage.setItem('ypy.session', state.session)
      sessionStorage.setItem('ypy.user', state.user)
      sessionStorage.setItem('ypy.mode', 'central')
      navigate('inicio')
    } catch (error) {
      state.loginError = error instanceof Error && !['TypeError', 'TimeoutError'].includes(error.name) ? error.message : 'No se pudo contactar al servidor. Comprueba la conexión e inténtalo de nuevo.'
    } finally { state.busy = false }
  }

  async function logout() {
    if (state.busy) return
    state.busy = true
    try {
      if (state.mode === 'central') {
        const response = await request('/api/v1/auth/session', { method: 'DELETE' })
        if (!response.ok) throw new Error('No se pudo cerrar la sesión en el servidor. Reintenta cuando haya conexión.')
      }
      resetSession(); navigate('inicio')
    } catch (error) { state.error = error instanceof Error ? error.message : 'No se pudo cerrar la sesión.' }
    finally { state.busy = false }
  }

  async function restore() {
    const requestedPage = window.location.hash.slice(1)
    try {
      if (sessionStorage.getItem('ypy.mode') === 'demo') await enterDemo()
      else if (state.session) {
        const response = await request('/api/v1/auth/session')
        if (!response.ok) { resetSession(); state.loginError = 'Inicia sesión para continuar.'; return }
        const scope = await response.json()
        state.settings.empresa = scope.empresa_id; state.settings.sucursal = scope.sucursal_id; state.settings.terminal = scope.terminal_id
        state.mode = 'central'; state.user = sessionStorage.getItem('ypy.user') ?? 'Usuario'
      }
      await loadUnits()
      await loadCategories()
      if (['inicio', 'venta', 'catalogo', 'historial', 'configuracion', 'unidades', 'categorias', 'ayuda'].includes(requestedPage)) state.page = requestedPage as Page
    } catch {
      resetSession(); state.loginError = 'No se pudo verificar la sesión. Conecta con el servidor o utiliza la demostración.'
    } finally { state.checking = false }
  }

  async function searchCatalog(query: string) {
    const generation = ++catalogRequest
    if (state.mode !== 'central') return
    if (!query.trim()) { state.products = []; state.loading = false; return }
    state.loading = true; state.error = ''
    try {
      const params = new URLSearchParams({ empresa_id: state.settings.empresa, lista_id: state.settings.lista, q: query.trim(), limite: '50', as_of: new Date().toISOString() })
      const response = await request(`/api/v1/catalog/items?${params}`)
      if (!response.ok) throw new Error('No se pudo consultar el catálogo. Comprueba la lista de precios y vuelve a buscar.')
      const result = await response.json()
      if (generation !== catalogRequest || state.mode !== 'central') return
      state.products = (result.items ?? []).map((item: { barcode?: string; articulo_id?: number; presentacion_id?: number; descripcion?: string; precio?: string; article_id?: number; presentation_id?: number; description?: string; price?: string; Barcode?: string; ArticleID?: number; PresentationID?: number; Description?: string; Price?: string }) => ({
        itemId: `${item.articulo_id ?? item.article_id ?? item.ArticleID}:${item.presentacion_id ?? item.presentation_id ?? item.PresentationID}`, barcode: item.barcode ?? item.Barcode ?? '', name: item.descripcion ?? item.description ?? item.Description ?? '', price: item.precio ?? item.price ?? item.Price ?? '', category: 'Catálogo central', stock: '',
      }))
    } catch (error) { if (generation === catalogRequest) { state.products = []; state.error = error instanceof Error ? error.message : 'No se pudo consultar el catálogo.' } }
    finally { if (generation === catalogRequest) state.loading = false }
  }

  function add(item: Product) {
    if (state.mode !== 'demo') { state.error = 'El cobro central todavía requiere la integración de caja, existencias y numeración. Puedes practicar el circuito completo en la demostración.'; return }
    const line = state.cart.find(row => row.itemId === item.itemId)
    if ((line?.quantity ?? 0) >= Math.min(Number(item.stock), 999)) { state.error = 'No hay más unidades disponibles para este artículo.'; return }
    if (line) line.quantity++
    else state.cart.push({ ...item, quantity: 1 })
    state.error = ''; state.notice = ''
  }

  function remove(itemId: string) {
    const line = state.cart.find(item => item.itemId === itemId)
    if (!line) return
    if (line.quantity > 1) line.quantity--
    else state.cart = state.cart.filter(item => item.itemId !== itemId)
  }

  async function checkout(method: string, received: string) {
    if (state.saving || !db || !state.cart.length || state.mode !== 'demo') return false
    if (!['CASH', 'CARD'].includes(method)) return false
    if (method === 'CASH' && cashChange(received, total.value) === null) { state.error = 'El importe recibido debe cubrir el total.'; return false }
    state.saving = true; state.error = ''
    try {
      const operationId = crypto.randomUUID()
      await recordSale({ operationId, number: `PR-${operationId.slice(0, 8).toUpperCase()}`, total: total.value, currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString(),
        lines: state.cart.map(line => ({ itemId: line.itemId, quantity: String(line.quantity) })),
        receiptLines: state.cart.map(line => ({ name: line.name, price: line.price, quantity: line.quantity })),
      }, [{ operationId, method, currency: 'PYG', amount: total.value }], db)
      state.cart = []
      state.notice = 'Venta de práctica guardada. Puedes consultar el comprobante en el historial.'
      await refresh()
      return true
    } catch (error) {
      const message = error instanceof Error ? error.message : ''
      state.error = message === 'INSUFFICIENT_STOCK' ? 'Las existencias cambiaron. Revisa las cantidades de la venta.'
        : message === 'STORAGE_QUOTA_EXCEEDED' ? 'El almacenamiento está lleno. No se guardó la venta.' : 'No se pudo guardar la venta. El carrito sigue disponible para reintentar.'
      return false
    } finally { state.saving = false }
  }

  async function saveProduct(form: ProductForm) {
    if (!db || state.mode !== 'demo') throw new Error('La edición del catálogo central aún no está disponible.')
    if (!form.name.trim() || !form.barcode.trim() || !categories.includes(form.category)) throw new Error('Completa el nombre, código y categoría.')
    if (!/^\d{1,12}$/.test(form.price) || BigInt(form.price) <= 0n) throw new Error('Ingresa un precio mayor que cero, en guaraníes enteros.')
    if (!/^\d{1,6}$/.test(form.stock)) throw new Error('Ingresa existencias entre 0 y 999999.')
    const currentDB = db
    const itemId = form.itemId || crypto.randomUUID()
    await currentDB.transaction('rw', currentDB.catalog, currentDB.stock, async () => {
      const duplicate = await currentDB.catalog.where('barcode').equals(form.barcode.trim()).first()
      if (duplicate && duplicate.itemId !== itemId) throw new Error('Ya existe un artículo con este código.')
      await currentDB.catalog.put({ itemId, barcode: form.barcode.trim(), name: form.name.trim(), category: form.category, price: form.price })
      await currentDB.stock.put({ itemId, quantity: form.stock })
    })
    state.cart = state.cart.filter(line => line.itemId !== itemId)
    await refresh()
    state.notice = 'Artículo guardado en el catálogo de demostración.'
  }

  function saveSettings(settings: Settings) {
    if (Object.values(settings).some(value => !/^[1-9]\d{0,8}$/.test(value))) throw new Error('Todos los identificadores deben ser números enteros positivos.')
    if (state.mode === 'central') throw new Error('Cierra sesión para cambiar la empresa, sucursal o terminal del acceso.')
    state.settings = { ...settings }
    localStorage.setItem('ypy.empresaId', settings.empresa)
    localStorage.setItem('ypy.sucursalId', settings.sucursal)
    localStorage.setItem('ypy.terminalId', settings.terminal)
    localStorage.setItem('ypy.listaId', settings.lista)
    state.notice = 'Configuración de acceso guardada.'
  }

  function saveTheme(theme: ThemePreferences) {
    const colors = [theme.primary, theme.accent, theme.background,
      theme.surface, theme.ink, theme.muted, theme.line]
    if (colors.some(color => !/^#[0-9a-fA-F]{6}$/.test(color))) {
      throw new Error('Cada color debe utilizar el formato hexadecimal de seis dígitos.')
    }
    state.theme = { ...defaultTheme, ...theme }
    localStorage.setItem('ypy.theme', JSON.stringify(state.theme))
    state.notice = 'Apariencia guardada en este navegador.'
  }

  function resetTheme() {
    state.theme = { ...defaultTheme }
    localStorage.removeItem('ypy.theme')
    state.notice = 'Apariencia predeterminada restaurada.'
  }

  loadUnits()
  loadCategories()
  return { state, total, units, todaySales, todayTotal, restore, navigate, login, logout, enterDemo, add, remove, checkout, searchCatalog, saveProduct, saveSettings, saveTheme, resetTheme, saveUnit, deleteUnit, saveCategory, deleteCategory }
}

export type Workspace = ReturnType<typeof useWorkspace>
