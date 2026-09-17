import 'fake-indexeddb/auto'
import { describe, expect, it } from 'vitest'
import { PosDatabase, applyStockSnapshot } from './dexie'

describe('snapshot offline', () => {
  it('conserva el efecto de operaciones pendientes', async () => {
    const db = new PosDatabase(`snapshot-${crypto.randomUUID()}`)
    await db.operations.bulkAdd([
      { operationId: 'pending', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16', itemId: 'item', quantity: '2' },
      { operationId: 'applied', number: '2', total: '10', currency: 'PYG', state: 'APPLIED', createdAt: '2026-09-16', itemId: 'item', quantity: '3' },
    ])
    await applyStockSnapshot([{ itemId: 'item', quantity: '10' }], db)
    expect((await db.stock.get('item'))?.quantity).toBe('8')
    await db.delete()
  })

  it('rechaza snapshot insuficiente sin borrar el saldo anterior', async () => {
    const db = new PosDatabase(`snapshot-fail-${crypto.randomUUID()}`)
    await db.stock.put({ itemId: 'item', quantity: '5' })
    await db.operations.add({ operationId: 'pending', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16', itemId: 'item', quantity: '2' })
    await expect(applyStockSnapshot([{ itemId: 'item', quantity: '1' }], db)).rejects.toThrow('INSUFFICIENT_STOCK_SNAPSHOT')
    expect((await db.stock.get('item'))?.quantity).toBe('5')
    await db.delete()
  })
})
