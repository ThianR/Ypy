import 'fake-indexeddb/auto'
import { describe, expect, it } from 'vitest'
import { PosDatabase, applyAck, getSyncCursor, hashOperation, saveSyncCursor, stableStringify } from './dexie'

it('serializa JSON con claves ordenadas recursivamente', () => {
  expect(stableStringify({ z: 1, a: { y: 2, b: 3 } })).toBe('{"a":{"b":3,"y":2},"z":1}')
})

describe('ACK POS', () => {
  it('aplica ACK solo con el payload original', async () => {
    const db = new PosDatabase(`ack-${crypto.randomUUID()}`)
    const operation = { operationId: 'op-ack', number: '1001', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: '2026-09-16T00:00:00Z' }
    await db.operations.add(operation); const hash = await hashOperation(operation)
    await applyAck(operation.operationId, true, hash, db)
    expect((await db.operations.get(operation.operationId))?.state).toBe('APPLIED')
    await db.delete()
  })
  it('marca conflicto si cambia el contenido', async () => {
    const db = new PosDatabase(`conflict-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-conflict', number: '1001', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    await expect(applyAck('op-conflict', true, 'bad-hash', db)).rejects.toThrow('OPERATION_CONTENT_CONFLICT')
    expect((await db.operations.get('op-conflict'))?.state).toBe('CONFLICT'); await db.delete()
  })
})

describe('cursor de sincronización', () => {
  it('avanza y rechaza respuestas antiguas', async () => {
    const db = new PosDatabase(`cursor-${crypto.randomUUID()}`)
    await saveSyncCursor('empresa-1/terminal-1', '12', db)
    expect(await getSyncCursor('empresa-1/terminal-1', db)).toBe('12')
    await expect(saveSyncCursor('empresa-1/terminal-1', '11', db)).rejects.toThrow('SYNC_CURSOR_REGRESSION')
    await db.delete()
  })
  it('rechaza cursores que no son enteros no negativos', async () => {
    const db = new PosDatabase(`cursor-invalid-${crypto.randomUUID()}`)
    await expect(saveSyncCursor('empresa-1/terminal-1', '-1', db)).rejects.toThrow('INVALID_SYNC_CURSOR')
    await expect(saveSyncCursor('empresa-1/terminal-1', 'abc', db)).rejects.toThrow('INVALID_SYNC_CURSOR')
    await db.delete()
  })
})
