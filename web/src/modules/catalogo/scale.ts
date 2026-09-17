import { PosDatabase, getScaleRules, saveScaleRules } from '../../pos/adapters/dexie'

export type ScaleRule = { code: string; prefix: string; weightDigits: number; priceDigits: number; priority: number }
export type ScaleResult = { itemCode: string; quantity: string; price: string }
export type V4ScaleRule = { prefix: string; length: number; product_start: number; product_length: number; value_start: number; value_length: number; decimals: number; content: 'PESO' | 'PRECIO' }

function validV4ScaleRule(rule: V4ScaleRule): boolean {
  const integers = [rule.length, rule.product_start, rule.product_length, rule.value_start, rule.value_length, rule.decimals]
  return typeof rule.prefix === 'string' && !!rule.prefix && integers.every(Number.isInteger) && rule.length > 0 && rule.prefix.length <= rule.length && rule.product_start > 0 && rule.product_length > 0 && rule.value_start > 0 && rule.value_length > 0 && rule.decimals >= 0 && rule.decimals < rule.value_length && rule.decimals <= 6 && (rule.content === 'PESO' || rule.content === 'PRECIO') && rule.product_start + rule.product_length - 1 <= rule.length && rule.value_start + rule.value_length - 1 <= rule.length
}

export async function fetchScaleRules(baseUrl: string, empresaId: string, fetcher: typeof fetch = fetch): Promise<V4ScaleRule[]> {
  if (!/^\d+$/.test(empresaId) || BigInt(empresaId) <= 0n) throw new Error('INVALID_SCOPE')
  const response = await fetcher(`${baseUrl}/api/v1/catalog/scale-rules?empresa_id=${encodeURIComponent(empresaId)}`)
  if (!response.ok) throw new Error('SCALE_RULES_READ_FAILED')
  const body = await response.json() as { rules?: V4ScaleRule[] }
  if (!Array.isArray(body.rules) || body.rules.some(rule => !validV4ScaleRule(rule))) throw new Error('INVALID_SCALE_RULES')
  return body.rules
}

export async function bootstrapScaleRules(baseUrl: string, empresaId: string, db = new PosDatabase(), fetcher: typeof fetch = fetch) {
  try {
    const rules = await fetchScaleRules(baseUrl, empresaId, fetcher)
    await saveScaleRules(empresaId, rules, db)
    return rules
  } catch (error) {
    const cached = await getScaleRules(empresaId, db)
    if (cached.length > 0) return cached as V4ScaleRule[]
    throw error
  }
}

export function parseScaleCode(value: string, rules: ScaleRule[]): ScaleResult {
  const valid = rules.filter(rule => rule.code && rule.weightDigits > 0 && rule.priceDigits > 0)
  const exact = valid.find(rule => rule.code === value)
  const candidates = exact ? [exact] : valid.filter(rule => rule.prefix && value.startsWith(rule.prefix)).sort((a, b) => b.priority - a.priority)
  if (!candidates.length || (!exact && candidates.length > 1 && candidates[0].priority === candidates[1].priority)) throw new Error('AMBIGUOUS_OR_INVALID_SCALE_CODE')
  const rule = candidates[0]; const start = rule.prefix.length; const end = start + rule.weightDigits + rule.priceDigits
  if (value.length < end) throw new Error('AMBIGUOUS_OR_INVALID_SCALE_CODE')
  const quantity = value.slice(start, start + rule.weightDigits); const price = value.slice(start + rule.weightDigits, end)
  if (!/^[0-9]+$/.test(quantity) || !/^[0-9]+$/.test(price) || BigInt(quantity) <= 0n || BigInt(price) <= 0n) throw new Error('AMBIGUOUS_OR_INVALID_SCALE_CODE')
  return { itemCode: rule.code, quantity, price }
}
