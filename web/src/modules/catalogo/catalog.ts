export type CatalogItem = { itemId: string; barcode: string; presentationId: string; unit: string; name?: string; active: boolean }
export type PriceVersion = { itemId: string; amount: string; validFrom: string; validTo?: string; active: boolean; version: number }

export function resolveBarcode(items: CatalogItem[], barcode: string): CatalogItem | undefined {
  const normalized = barcode.trim()
  if (!normalized) return undefined
  const item = items.find(candidate => candidate.barcode === normalized)
  return item?.active ? item : undefined
}

/** Busca solo artículos ofrecibles; evita filtrar inactivos al circuito de venta. */
export function searchCatalog(items: CatalogItem[], query: string): CatalogItem[] {
  const normalized = query.trim().toLocaleLowerCase()
  if (!normalized) return items.filter(item => item.active)
  return items.filter(item => item.active && (
    item.barcode.toLocaleLowerCase().includes(normalized) ||
    (item.name ?? '').toLocaleLowerCase().includes(normalized)
  ))
}

export function resolvePrice(prices: PriceVersion[], itemId: string, at: string): PriceVersion | undefined {
  return prices.filter(p => {
    if (p.itemId !== itemId || !p.active || !/^\d+(?:\.\d+)?$/.test(p.amount) || !/^\d{4}-\d{2}-\d{2}/.test(p.validFrom)) return false
    if (p.validTo && (!/^\d{4}-\d{2}-\d{2}/.test(p.validTo) || p.validTo <= p.validFrom)) return false
    return p.validFrom <= at && (!p.validTo || at < p.validTo)
  }).sort((a, b) => b.version - a.version)[0]
}
