import 'fake-indexeddb/auto'
import Dexie from 'dexie'
import { describe, expect, it } from 'vitest'
import { PosDatabase, getQuote, getScaleRules, recordSale, saveQuote, saveScaleRules } from './dexie'

describe('venta local Dexie', () => {
  it('persiste reglas de balanza para uso offline', async () => {
    const db = new PosDatabase(`scale-cache-${crypto.randomUUID()}`)
    const rules = [{ prefix: '99', length: 10 }]
    await saveScaleRules('4', rules, db)
    expect(await getScaleRules('4', db)).toEqual(rules)
    await db.delete()
  })
  it('persiste y recupera el snapshot de cotización offline', async () => {
    const db = new PosDatabase(`quote-cache-${crypto.randomUUID()}`)
    await saveQuote({ empresaId: '4', from: 'USD', to: 'PYG', rate: '7500.000000', asOf: '2026-09-16T12:00:00.000Z' }, db)
    await saveQuote({ empresaId: '4', from: 'USD', to: 'PYG', rate: '7600.000000', asOf: '2026-09-17T12:00:00.000Z' }, db)
    expect((await getQuote('4', 'USD', 'PYG', '2026-09-16T23:00:00.000Z', db))?.rate).toBe('7500.000000')
    await expect(saveQuote({ empresaId: '4', from: 'USD', to: 'PYG', rate: '0', asOf: '2026-09-16T12:00:00.000Z' }, db)).rejects.toThrow('INVALID_QUOTE')
    await db.delete()
  })
  it('persiste operación, pago y stock en una transacción', async () => {
    const db = new PosDatabase(`test-${crypto.randomUUID()}`)
    const operationId = 'sale-1'
    await db.stock.put({ itemId: 'item-1', quantity: '2' })
    await recordSale({ operationId, number: 'PENDING', total: '150000', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString(), itemId: 'item-1', quantity: '2' }, [{ operationId, method: 'CASH', currency: 'PYG', amount: '150000' }], db)
    expect(await db.operations.get(operationId)).toBeTruthy()
    expect(await db.paymentLines.where('operationId').equals(operationId).count()).toBe(1)
    expect((await db.stock.get('item-1'))?.quantity).toBe('0')
    await db.delete()
  })
  it('descuenta todas las líneas de una venta', async () => {
    const db = new PosDatabase(`multi-line-${crypto.randomUUID()}`)
    await db.stock.bulkPut([{ itemId: 'item-1', quantity: '5' }, { itemId: 'item-2', quantity: '7' }])
    await recordSale({ operationId: 'sale-multi', number: 'PENDING', total: '300', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString(), lines: [{ itemId: 'item-1', quantity: '2' }, { itemId: 'item-2', quantity: '3' }] }, [{ operationId: 'sale-multi', method: 'CASH', currency: 'PYG', amount: '300' }], db)
    expect((await db.stock.get('item-1'))?.quantity).toBe('3')
    expect((await db.stock.get('item-2'))?.quantity).toBe('4')
    await db.delete()
  })
  it('bloquea venta sin saldo disponible', async () => {
    const db = new PosDatabase(`stock-${crypto.randomUUID()}`)
    await expect(recordSale({ operationId: 'sale-stock', number: 'PENDING', total: '10', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString(), itemId: 'item-1', quantity: '1' }, [{ operationId: 'sale-stock', method: 'CASH', currency: 'PYG', amount: '10' }], db)).rejects.toThrow('INSUFFICIENT_STOCK')
    await db.delete()
  })
  it('rechaza pago que no cuadra con el total', async () => {
    const db = new PosDatabase(`payment-${crypto.randomUUID()}`)
    await expect(recordSale({ operationId: 'sale-bad', number: 'PENDING', total: '100', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'sale-bad', method: 'CASH', currency: 'PYG', amount: '99' }], db)).rejects.toThrow('PAYMENT_TOTAL_MISMATCH')
    expect(await db.operations.get('sale-bad')).toBeUndefined()
    await db.delete()
  })
  it('suma pagos decimales sin perder precisión', async () => {
    const db = new PosDatabase(`decimal-${crypto.randomUUID()}`)
    await recordSale({ operationId: 'sale-decimal', number: 'PENDING', total: '100.50', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'sale-decimal', method: 'CASH', currency: 'PYG', amount: '50.25' }, { operationId: 'sale-decimal', method: 'CARD', currency: 'PYG', amount: '50.25' }], db)
    expect(await db.operations.get('sale-decimal')).toBeTruthy()
    await db.delete()
  })

  it('migra pagos del esquema v1 sin perderlos', async () => {
    const name = `upgrade-${crypto.randomUUID()}`
    const legacy = new Dexie(name)
    legacy.version(1).stores({ operations: '&operationId, state, createdAt', payments: '&operationId', stock: '&itemId', sync: '&scope' })
    await legacy.open()
    await legacy.table('payments').add({ operationId: 'legacy-sale', method: 'CASH', currency: 'PYG', amount: '100' })
    await legacy.close()
    const upgraded = new PosDatabase(name)
    await upgraded.open()
    const payments = await upgraded.paymentLines.where('operationId').equals('legacy-sale').toArray()
    expect(payments).toHaveLength(1)
    expect(payments[0].paymentId).toBe('legacy-sale-1')
    await upgraded.delete()
  })

  it('rechaza cantidades cero o negativas antes de modificar stock', async () => {
    const db = new PosDatabase(`quantity-${crypto.randomUUID()}`)
    const operation = { operationId: 'sale-quantity', number: 'PENDING', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: new Date().toISOString(), itemId: 'item-1', quantity: '-1' }
    await expect(recordSale(operation, [{ operationId: operation.operationId, method: 'CASH', currency: 'PYG', amount: '10' }], db)).rejects.toThrow('INVALID_QUANTITY')
    expect(await db.stock.get('item-1')).toBeUndefined()
    await db.delete()
  })

  it('rechaza totales negativos y medios vacíos', async () => {
    const db = new PosDatabase(`money-invalid-${crypto.randomUUID()}`)
    await expect(recordSale({ operationId: 'negative-total', number: 'PENDING', total: '-10', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'negative-total', method: 'CASH', currency: 'PYG', amount: '10' }], db)).rejects.toThrow('INVALID_MONEY')
    await expect(recordSale({ operationId: 'empty-method', number: 'PENDING', total: '10', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'empty-method', method: '', currency: 'PYG', amount: '10' }], db)).rejects.toThrow('INVALID_MONEY')
    await db.delete()
  })
  it('normaliza medios y monedas antes de persistir', async () => {
    const db = new PosDatabase(`payment-normalize-${crypto.randomUUID()}`)
    await recordSale({ operationId: 'sale-normalize', number: 'PENDING', total: '10', currency: 'PYG', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'sale-normalize', method: ' card ', currency: 'PYG', amount: '10' }], db)
    expect((await db.paymentLines.get('sale-normalize-1'))?.method).toBe('CARD')
    await expect(recordSale({ operationId: 'sale-empty-currency', number: 'PENDING', total: '10', currency: '', state: 'PENDING', createdAt: new Date().toISOString() }, [{ operationId: 'sale-empty-currency', method: 'CASH', currency: '', amount: '10' }], db)).rejects.toThrow('PAYMENT_CURRENCY_MISMATCH')
    await db.delete()
  })
})
