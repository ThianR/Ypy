import { describe, expect, it } from 'vitest'
import { resolveBarcode, searchCatalog } from './catalog'

describe('catálogo POS', () => {
  it('resuelve barras activas', () => expect(resolveBarcode([{ itemId: '1', barcode: '779', presentationId: 'u', unit: 'UN', active: true }], '779')?.itemId).toBe('1'))
  it('no ofrece artículos inactivos', () => expect(resolveBarcode([{ itemId: '1', barcode: '779', presentationId: 'u', unit: 'UN', active: false }], '779')).toBeUndefined())
  it('busca por código o nombre y excluye inactivos', () => {
    const items = [
      { itemId: '1', barcode: '779', presentationId: 'u', unit: 'UN', name: 'Café', active: true },
      { itemId: '2', barcode: '780', presentationId: 'u', unit: 'UN', name: 'Café viejo', active: false },
    ]
    expect(searchCatalog(items, ' café ')).toHaveLength(1)
    expect(searchCatalog(items, '779')[0]?.itemId).toBe('1')
  })
})
