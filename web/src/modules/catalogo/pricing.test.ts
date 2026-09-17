import { describe, expect, it } from 'vitest'
import { resolvePrice } from './catalog'

describe('precios versionados', () => {
  it('preserva el precio histórico según fecha', () => {
    const prices = [{ itemId: '1', amount: '100', validFrom: '2026-01-01', validTo: '2026-09-01', active: true, version: 1 }, { itemId: '1', amount: '120', validFrom: '2026-09-01', active: true, version: 2 }]
    expect(resolvePrice(prices, '1', '2026-08-31')?.amount).toBe('100'); expect(resolvePrice(prices, '1', '2026-09-16')?.amount).toBe('120')
  })
  it('ignora precios inválidos o ventanas invertidas', () => {
    const prices = [
      { itemId: '1', amount: '-1', validFrom: '2026-01-01', active: true, version: 9 },
      { itemId: '1', amount: '100', validFrom: '2026-09-01', validTo: '2026-01-01', active: true, version: 8 },
    ]
    expect(resolvePrice(prices, '1', '2026-09-16')).toBeUndefined()
  })
})
