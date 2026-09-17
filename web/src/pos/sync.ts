import { getSyncCursor, hashOperation, markApplied, pendingOperations, PosDatabase, saveSyncCursor, stableStringify } from './adapters/dexie'

type Ack = { operation_id?: string; cursor?: string; content_hash?: string }
export type SyncOptions = { maxAttempts?: number; sleep?: (ms: number) => Promise<void>; sessionId?: string }

function parseScope(scope: string): [string, string] {
  const parts = scope.split('/')
  if (parts.length !== 2 || !/^\d+$/.test(parts[0]) || !/^\d+$/.test(parts[1]) || BigInt(parts[0]) <= 0n || BigInt(parts[1]) <= 0n) {
    throw new Error('INVALID_SYNC_SCOPE')
  }
  return [parts[0], parts[1]]
}

export async function syncPending(baseUrl: string, scope: string, db = new PosDatabase(), fetcher: typeof fetch = fetch, options: SyncOptions = {}) {
  const [empresaId, terminalId] = parseScope(scope)
  const pending = await pendingOperations(db)
  const maxAttempts = Math.max(1, options.maxAttempts ?? 1)
  const sleep = options.sleep ?? ((ms: number) => new Promise<void>(resolve => setTimeout(resolve, ms)))
  let sent = 0
  for (const operation of pending) {
    await db.operations.update(operation.operationId, { state: 'SENDING' })
    let response: Response | undefined
    let lastError: unknown
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        const headers: Record<string, string> = { 'Content-Type': 'application/json' }
        if (options.sessionId?.trim()) headers['X-Session-ID'] = options.sessionId.trim()
        response = await fetcher(`${baseUrl}/api/v1/pos/operations`, {
          method: 'POST', headers,
          body: JSON.stringify({ id: operation.operationId, empresa_id: empresaId, terminal_id: terminalId, payload: stableStringify(operation) }),
        })
      } catch (error) {
        lastError = error
        if (attempt < maxAttempts) await sleep(Math.min(30_000, 100 * 2 ** (attempt - 1)))
        continue
      }
      if (response.ok || ![408, 429, 500, 502, 503, 504].includes(response.status)) break
      if (attempt < maxAttempts) await sleep(Math.min(30_000, 100 * 2 ** (attempt - 1)))
    }
    if (lastError && !response) {
      await db.operations.update(operation.operationId, { state: 'PENDING' })
      throw lastError
    }
    if (!response || !response.ok) {
      await db.operations.update(operation.operationId, { state: 'PENDING' })
      break
    }
    let ack: Ack
    try {
      ack = await response.json() as Ack
    } catch (error) {
      await db.operations.update(operation.operationId, { state: 'CONFLICT' })
      throw new Error('ACK_INVALID_JSON', { cause: error })
    }
    const expectedHash = await hashOperation(operation)
    if (ack.operation_id !== operation.operationId) {
      await db.operations.update(operation.operationId, { state: 'CONFLICT' })
      throw new Error('ACK_OPERATION_MISMATCH')
    }
    if (ack.content_hash !== expectedHash) {
      await db.operations.update(operation.operationId, { state: 'CONFLICT' })
      throw new Error('ACK_CONTENT_MISMATCH')
    }
    await markApplied(operation.operationId, db)
    if (ack.cursor) await saveSyncCursor(scope, ack.cursor, db)
    sent++
  }
  return { sent, cursor: await getSyncCursor(scope, db) }
}
