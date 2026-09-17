import 'fake-indexeddb/auto'
import { describe, expect, it, vi } from 'vitest'
import { hashOperation, PosDatabase } from './adapters/dexie'
import { syncPending } from './sync'

describe('sincronización POS', () => {
  it('envía pendientes, confirma y conserva cursor', async () => {
    const db = new PosDatabase(`sync-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-1', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    const operation = { operationId: 'op-1', number: '1', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: '2026-09-16T00:00:00Z' }
    const hash = await hashOperation(operation)
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ operation_id: 'op-1', content_hash: hash, cursor: '4' }), { status: 200 }))
    const result = await syncPending('http://central', '1/2', db, fetcher)
    expect(result).toEqual({ sent: 1, cursor: '4' })
    expect((await db.operations.get('op-1'))?.state).toBe('APPLIED')
    expect(fetcher).toHaveBeenCalledOnce(); await db.delete()
  })

  it('propaga la sesión al endpoint central', async () => {
    const db = new PosDatabase(`session-${crypto.randomUUID()}`)
    const operation = { operationId: 'op-session', number: '1', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: '2026-09-16T00:00:00Z' }
    await db.operations.add(operation)
    const hash = await hashOperation(operation)
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ operation_id: operation.operationId, content_hash: hash }), { status: 200 }))
    await syncPending('http://central', '1/2', db, fetcher, { sessionId: 'session-1' })
    expect(fetcher.mock.calls[0][1]).toMatchObject({ headers: { 'X-Session-ID': 'session-1' } })
    await db.delete()
  })

  it('marca conflicto ante hash de ACK incorrecto', async () => {
    const db = new PosDatabase(`hash-${crypto.randomUUID()}`)
    const operation = { operationId: 'op-hash', number: '1', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: '2026-09-16T00:00:00Z' }
    await db.operations.add(operation)
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ operation_id: 'op-hash', content_hash: 'bad' }), { status: 200 }))
    await expect(syncPending('http://central', '1/2', db, fetcher)).rejects.toThrow('ACK_CONTENT_MISMATCH')
    expect((await db.operations.get('op-hash'))?.state).toBe('CONFLICT')
    expect(await hashOperation(operation)).toMatch(/^[a-f0-9]{64}$/)
    await db.delete()
  })

  it('recupera SENDING a PENDING cuando falla la red', async () => {
    const db = new PosDatabase(`network-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-network', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    const fetcher = vi.fn().mockRejectedValue(new Error('offline'))
    await expect(syncPending('http://central', '1/2', db, fetcher)).rejects.toThrow('offline')
    expect((await db.operations.get('op-network'))?.state).toBe('PENDING')
    await db.delete()
  })

  it('reintenta errores temporales con backoff inyectable', async () => {
    const db = new PosDatabase(`retry-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-retry', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    const operation = { operationId: 'op-retry', number: '1', total: '10', currency: 'PYG', state: 'PENDING' as const, createdAt: '2026-09-16T00:00:00Z' }
    const hash = await hashOperation(operation)
    const fetcher = vi.fn()
      .mockResolvedValueOnce(new Response('{}', { status: 503 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ operation_id: 'op-retry', content_hash: hash }), { status: 200 }))
    const sleep = vi.fn().mockResolvedValue(undefined)
    const result = await syncPending('http://central', '1/2', db, fetcher, { maxAttempts: 2, sleep })
    expect(result.sent).toBe(1)
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(sleep).toHaveBeenCalledWith(100)
    await db.delete()
  })

  it('marca conflicto si el ACK exitoso no contiene JSON válido', async () => {
    const db = new PosDatabase(`ack-invalid-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-invalid-ack', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    const fetcher = vi.fn().mockResolvedValue(new Response('not-json', { status: 200 }))
    await expect(syncPending('http://central', '1/2', db, fetcher)).rejects.toThrow('ACK_INVALID_JSON')
    expect((await db.operations.get('op-invalid-ack'))?.state).toBe('CONFLICT')
    await db.delete()
  })

  it('rechaza un alcance de sincronización inválido antes de enviar', async () => {
    const db = new PosDatabase(`scope-invalid-${crypto.randomUUID()}`)
    const fetcher = vi.fn()
    await expect(syncPending('http://central', 'empresa/terminal', db, fetcher)).rejects.toThrow('INVALID_SYNC_SCOPE')
    expect(fetcher).not.toHaveBeenCalled()
    await db.delete()
  })

  it('limita el backoff para muchos reintentos', async () => {
    const db = new PosDatabase(`retry-cap-${crypto.randomUUID()}`)
    await db.operations.add({ operationId: 'op-cap', number: '1', total: '10', currency: 'PYG', state: 'PENDING', createdAt: '2026-09-16T00:00:00Z' })
    const fetcher = vi.fn().mockResolvedValue(new Response('{}', { status: 503 }))
    const sleep = vi.fn().mockResolvedValue(undefined)
    await syncPending('http://central', '1/2', db, fetcher, { maxAttempts: 12, sleep })
    expect(Math.max(...sleep.mock.calls.map(call => call[0] as number))).toBe(30_000)
    await db.delete()
  })
})
