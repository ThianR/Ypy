import { describe, expect, it } from 'vitest'
import { addDecimal } from './money'

describe('decimal exacto', () => {
  it('no pierde precisión', () => expect(addDecimal('0.1', '0.2')).toBe('0.300000'))
  it('conserva signo', () => expect(addDecimal('-1.25', '0.5')).toBe('-0.750000'))
  it('rechaza más de seis decimales', () => expect(() => addDecimal('1.0000001')).toThrow())
})
