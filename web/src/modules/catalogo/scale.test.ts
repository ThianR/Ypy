import 'fake-indexeddb/auto'
import { describe, expect, it } from 'vitest'
import { bootstrapScaleRules, fetchScaleRules, parseScaleCode } from './scale'
import { PosDatabase, getScaleRules } from '../../pos/adapters/dexie'

describe('regla de balanza', () => {
  const rule = { code: 'MEAT', prefix: '99', weightDigits: 3, priceDigits: 5, priority: 1 }
  it('produce el mismo formato offline', () => expect(parseScaleCode('9912301234', [rule])).toEqual({ itemCode: 'MEAT', quantity: '123', price: '01234' }))
  it('rechaza precio cero', () => expect(() => parseScaleCode('9912300000', [rule])).toThrow())
  it('ignora reglas malformadas', () => expect(parseScaleCode('9912301234', [{ ...rule, weightDigits: 0 }, rule])).toEqual({ itemCode: 'MEAT', quantity: '123', price: '01234' }))
  it('descarga reglas V4 por empresa', async () => {
    const fetcher = async () => new Response(JSON.stringify({ rules: [{ prefix: '99', length: 10, product_start: 3, product_length: 3, value_start: 6, value_length: 5, decimals: 2, content: 'PRECIO' }] }), { status: 200 })
    await expect(fetchScaleRules('http://central', '4', fetcher)).resolves.toHaveLength(1)
  })
  it('descarga y persiste reglas para modo offline', async () => {
    const db = new PosDatabase(`scale-bootstrap-${crypto.randomUUID()}`)
    const fetcher = async () => new Response(JSON.stringify({ rules: [{ prefix: '98', length: 10, product_start: 3, product_length: 3, value_start: 6, value_length: 5, decimals: 2, content: 'PESO' }] }), { status: 200 })
    await bootstrapScaleRules('http://central', '4', db, fetcher)
    expect(await getScaleRules('4', db)).toHaveLength(1)
    await db.delete()
  })
  it('rechaza reglas V4 con posiciones inválidas', async () => {
    const fetcher = async () => new Response(JSON.stringify({ rules: [{ prefix: '99', length: 4, product_start: 3, product_length: 3, value_start: 1, value_length: 1, decimals: 2, content: 'PRECIO' }] }), { status: 200 })
    await expect(fetchScaleRules('http://central', '4', fetcher)).rejects.toThrow('INVALID_SCALE_RULES')
  })
  it('rechaza campos numéricos con tipo incorrecto', async () => {
    const fetcher = async () => new Response(JSON.stringify({ rules: [{ prefix: '99', length: '10', product_start: 3, product_length: 3, value_start: 6, value_length: 5, decimals: 2, content: 'PESO' }] }), { status: 200 })
    await expect(fetchScaleRules('http://central', '4', fetcher)).rejects.toThrow('INVALID_SCALE_RULES')
  })
  it('usa caché cuando el servidor no está disponible', async () => {
    const db = new PosDatabase(`scale-offline-${crypto.randomUUID()}`)
    const rules = [{ prefix: '97', length: 10, product_start: 3, product_length: 3, value_start: 6, value_length: 5, decimals: 2, content: 'PESO' as const }]
    const firstFetcher = async () => new Response(JSON.stringify({ rules }), { status: 200 })
    await bootstrapScaleRules('http://central', '4', db, firstFetcher)
    const offlineFetcher = async () => { throw new Error('offline') }
    await expect(bootstrapScaleRules('http://central', '4', db, offlineFetcher)).resolves.toEqual(rules)
    await db.delete()
  })
})
