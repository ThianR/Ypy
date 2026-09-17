export type PosPayment = { method: string; currency: string; amount: string }

export type PosEnvelope = {
  operationId: string
  empresaId: string
  sucursalId: string
  terminalId: string
  cashSessionId: string
  sequence: string
  contractVersion: number
  commercialDate: string
  currency: string
  total: string
  payments: PosPayment[]
}

export type PrinterResult = { status: 'PRINTED' | 'UNKNOWN' | 'FAILED'; receiptId: string }

/** La impresión es posterior a la transacción comercial y nunca cambia su resultado. */
export function simulatePrint(operationId: string, available = true): PrinterResult {
  return { status: available ? 'PRINTED' : 'UNKNOWN', receiptId: `receipt-${operationId}` }
}

