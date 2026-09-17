/** Suma decimal exacta para importes transportados como cadenas. */
export function addDecimal(...values: string[]): string {
  const scale = 6
  const factor = 10n ** BigInt(scale)
  const total = values.reduce((sum, value) => {
    const [whole, fraction = ''] = value.split('.')
    if (!/^-?[0-9]+(?:\.[0-9]+)?$/.test(value) || fraction.length > scale) throw new Error('invalid decimal amount')
    const sign = whole.startsWith('-') ? -1n : 1n
    const digits = whole.replace('-', '') + fraction.padEnd(scale, '0')
    return sum + sign * BigInt(digits || '0')
  }, 0n)
  const negative = total < 0n ? '-' : ''
  const absolute = total < 0n ? -total : total
  const raw = absolute.toString().padStart(scale + 1, '0')
  return `${negative}${raw.slice(0, -scale)}.${raw.slice(-scale)}`
}
