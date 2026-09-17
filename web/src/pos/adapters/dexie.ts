import Dexie, { type Table } from 'dexie'
import { addDecimal } from '../../modules/maestros/money'

export type QueueState = 'PENDING' | 'SENDING' | 'APPLIED' | 'CONFLICT'
export type PosOperation = {
  operationId: string
  number: string
  total: string
  currency: string
  state: QueueState
  createdAt: string
  itemId?: string
  quantity?: string
  lines?: { itemId: string; quantity: string }[]
  receiptLines?: { name: string; price: string; quantity: number }[]
}
export type PosPayment = { paymentId?: string; operationId: string; method: string; currency: string; amount: string }

export const STORAGE_QUOTA_ERROR = 'STORAGE_QUOTA_EXCEEDED'
function normalizeStorageError(error: unknown): never {
  const name = error instanceof DOMException ? error.name : ''
  if (name === 'QuotaExceededError' || (error instanceof Error && /quota|storage.?full/i.test(error.message))) throw new Error(STORAGE_QUOTA_ERROR)
  throw error
}

export function stableStringify(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(stableStringify).join(',')}]`
  if (value && typeof value === 'object') {
    return `{${Object.keys(value as Record<string, unknown>).sort().map(key => `${JSON.stringify(key)}:${stableStringify((value as Record<string, unknown>)[key])}`).join(',')}}`
  }
  return JSON.stringify(value)
}
export type SyncCursor = { scope: string; cursor: string; updatedAt: string }
export type ScaleRuleCache = { key: string; empresaId: string; rules: unknown[]; updatedAt: string }
export type QuoteCache = { key: string; empresaId: string; from: string; to: string; rate: string; asOf: string; updatedAt: string }
export type CatalogProduct = { itemId: string; barcode: string; name: string; category: string; price: string }

class PosDatabase extends Dexie {
  operations!: Table<PosOperation, string>
  paymentLines!: Table<PosPayment, string>
  stock!: Table<{ itemId: string; quantity: string }, string>
  sync!: Table<SyncCursor, string>
  scaleRules!: Table<ScaleRuleCache, string>
  quotes!: Table<QuoteCache, string>
  catalog!: Table<CatalogProduct, string>
  constructor(name = 'ypy-pos') {
    super(name)
    this.version(1).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', stock: '&itemId', sync: '&scope' })
    this.version(2).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', paymentLines: '&paymentId, operationId', stock: '&itemId', sync: '&scope' }).upgrade(async tx => {
      const legacy = tx.table('payments')
      const rows = await legacy.toArray()
      const paymentLines = tx.table('paymentLines')
      await paymentLines.bulkAdd(rows.map((payment, index) => ({ ...payment, paymentId: payment.paymentId ?? `${payment.operationId}-${index + 1}` })))
    })
    this.version(3).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', paymentLines: '&paymentId, operationId', stock: '&itemId', sync: '&scope', scaleRules: '&key, empresaId' })
    this.version(4).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', paymentLines: '&paymentId, operationId', stock: '&itemId', sync: '&scope', scaleRules: '&key, empresaId', quotes: '&key, [empresaId+from+to]' })
    this.version(5).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', paymentLines: '&paymentId, operationId', stock: '&itemId', sync: '&scope', scaleRules: '&key, empresaId', quotes: '&key, [empresaId+from+to]', catalog: '&itemId, &barcode, name, category' })
  }
}

export async function saveScaleRules(empresaId: string, rules: unknown[], db = new PosDatabase()) {
  if (!/^\d+$/.test(empresaId) || BigInt(empresaId) <= 0n || !Array.isArray(rules)) throw new Error('INVALID_SCALE_RULES')
  await db.scaleRules.put({ key: empresaId, empresaId, rules, updatedAt: new Date().toISOString() })
}

export async function getScaleRules(empresaId: string, db = new PosDatabase()) {
  if (!/^\d+$/.test(empresaId) || BigInt(empresaId) <= 0n) throw new Error('INVALID_SCOPE')
  return (await db.scaleRules.get(empresaId))?.rules ?? []
}

export async function saveQuote(quote: Omit<QuoteCache, 'key' | 'updatedAt'>, db = new PosDatabase()) {
  if (!/^\d+$/.test(quote.empresaId) || BigInt(quote.empresaId) <= 0n || !quote.from.trim() || !quote.to.trim() || quote.from === quote.to || !/^\d+(\.\d+)?$/.test(quote.rate) || quote.rate === '0' || !quote.asOf || Number.isNaN(Date.parse(quote.asOf))) throw new Error('INVALID_QUOTE')
  await db.quotes.put({ ...quote, key: `${quote.empresaId}/${quote.from}/${quote.to}/${quote.asOf}`, updatedAt: new Date().toISOString() })
}

export async function getQuote(empresaId: string, from: string, to: string, asOf = new Date().toISOString(), db = new PosDatabase()) {
  if (!/^\d+$/.test(empresaId) || BigInt(empresaId) <= 0n || !from.trim() || !to.trim()) throw new Error('INVALID_SCOPE')
  if (Number.isNaN(Date.parse(asOf))) throw new Error('INVALID_QUOTE_DATE')
  const rows = (await db.quotes.where('[empresaId+from+to]').equals([empresaId, from, to]).toArray()).filter(row => row.asOf <= asOf)
  return rows.sort((a, b) => b.asOf.localeCompare(a.asOf))[0]
}

export async function recordSale(operation: PosOperation, payments: PosPayment[], db = new PosDatabase()) {
  if (!operation.currency.trim() || payments.length === 0) throw new Error('PAYMENT_CURRENCY_MISMATCH')
  const normalizedPayments = payments.map(payment => ({ ...payment, method: payment.method.trim().toUpperCase(), currency: payment.currency.trim() }))
  if (normalizedPayments.some(payment => payment.currency !== operation.currency.trim())) throw new Error('PAYMENT_CURRENCY_MISMATCH')
  if (!/^\d+(\.\d+)?$/.test(operation.total) || normalizedPayments.some(payment => payment.method === '' || !/^\d+(\.\d+)?$/.test(payment.amount))) throw new Error('INVALID_MONEY')
  const lines = operation.lines ?? (operation.itemId && operation.quantity ? [{ itemId: operation.itemId, quantity: operation.quantity }] : [])
  if (lines.some(line => !line.itemId || !/^\d+$/.test(line.quantity) || line.quantity === '0')) throw new Error('INVALID_QUANTITY')
  const paid = addDecimal(...normalizedPayments.map(payment => payment.amount))
  if (!/^0(?:\.0+)?$/.test(addDecimal(paid, `-${operation.total}`))) throw new Error('PAYMENT_TOTAL_MISMATCH')
  try { await db.transaction('rw', db.operations, db.paymentLines, db.stock, async () => {
    const previous = await db.operations.get(operation.operationId)
    if (previous) { if (JSON.stringify(previous) !== JSON.stringify(operation)) throw new Error('OPERATION_ID_CONFLICT'); return }
    await db.operations.add(operation)
    await db.paymentLines.bulkAdd(normalizedPayments.map((payment, index) => ({ ...payment, paymentId: payment.paymentId ?? `${operation.operationId}-${index + 1}` })))
    for (const line of lines) {
      const current = await db.stock.get(line.itemId)
      const available = BigInt(current?.quantity ?? '0')
      const requested = BigInt(line.quantity)
      if (available < requested) throw new Error('INSUFFICIENT_STOCK')
      const quantity = available - requested
      await db.stock.put({ itemId: line.itemId, quantity: quantity.toString() })
    }
  }) } catch (error) { normalizeStorageError(error) }
}

export async function recordOperation(operation: PosOperation, db = new PosDatabase()) {
  try { await db.transaction('rw', db.operations, async () => {
    const previous = await db.operations.get(operation.operationId)
    if (previous) {
      if (JSON.stringify(previous) !== JSON.stringify(operation)) throw new Error('OPERATION_ID_CONFLICT')
      return
    }
    await db.operations.add(operation)
  }) } catch (error) { normalizeStorageError(error) }
}

export async function markApplied(operationId: string, db = new PosDatabase()) {
  await db.operations.update(operationId, { state: 'APPLIED' })
}

export async function pendingOperations(db = new PosDatabase()) {
  return db.operations.where('state').anyOf('PENDING', 'SENDING').sortBy('createdAt')
}

export async function getSyncCursor(scope: string, db = new PosDatabase()) {
  return (await db.sync.get(scope))?.cursor ?? '0'
}

export async function saveSyncCursor(scope: string, cursor: string, db = new PosDatabase()) {
  if (!/^\d+$/.test(cursor)) throw new Error('INVALID_SYNC_CURSOR')
  const current = await getSyncCursor(scope, db)
  if (BigInt(cursor) < BigInt(current)) throw new Error('SYNC_CURSOR_REGRESSION')
  await db.sync.put({ scope, cursor, updatedAt: new Date().toISOString() })
}

export async function applyAck(operationId: string, accepted: boolean, contentHash: string, db = new PosDatabase()) {
  const operation = await db.operations.get(operationId)
  if (!operation) throw new Error('OPERATION_NOT_FOUND')
  const expectedHash = await hashOperation(operation)
  let conflict = false
  await db.transaction('rw', db.operations, async () => {
    if (expectedHash !== contentHash) { await db.operations.update(operationId, { state: 'CONFLICT' }); conflict = true; return }
    await db.operations.update(operationId, { state: accepted ? 'APPLIED' : 'CONFLICT' })
  })
  if (conflict) throw new Error('OPERATION_CONTENT_CONFLICT')
}

export async function hashOperation(operation: PosOperation): Promise<string> {
  const bytes = new TextEncoder().encode(stableStringify(operation))
  const digest = await crypto.subtle.digest('SHA-256', bytes)
  return [...new Uint8Array(digest)].map(byte => byte.toString(16).padStart(2, '0')).join('')
}

export async function applyStockSnapshot(snapshot: { itemId: string; quantity: string }[], db = new PosDatabase()) {
  await db.transaction('rw', db.stock, db.operations, async () => {
    const pending = await db.operations.where('state').anyOf('PENDING', 'SENDING').toArray()
    if (snapshot.some(row => !row.itemId || !/^\d+$/.test(row.quantity))) throw new Error('INVALID_STOCK_SNAPSHOT')
    const balances = new Map(snapshot.map(row => [row.itemId, BigInt(row.quantity)]))
    if (pending.some(operation => (operation.lines ?? (operation.itemId && operation.quantity ? [{ itemId: operation.itemId, quantity: operation.quantity }] : [])).some(line => (balances.get(line.itemId) ?? 0n) < BigInt(line.quantity)))) {
      throw new Error('INSUFFICIENT_STOCK_SNAPSHOT')
    }
    await db.stock.clear()
    await db.stock.bulkPut(snapshot)
    for (const operation of pending) {
      for (const line of operation.lines ?? (operation.itemId && operation.quantity ? [{ itemId: operation.itemId, quantity: operation.quantity }] : [])) {
        const row = await db.stock.get(line.itemId)
        const quantity = BigInt(row?.quantity ?? '0') - BigInt(line.quantity)
        await db.stock.put({ itemId: line.itemId, quantity: quantity.toString() })
      }
    }
  })
}

export { PosDatabase }
