import type { PriceTier } from '../types'

export type QuantityQuote = {
  yourQty: number
  pooledQty: number
  tier: PriceTier | null
  nextTier: PriceTier | null
  unitPrice: number | null
  yourTotal: number | null
  retailTotal: number
  savings: number | null
  shortfallToMinimum: number
}

export function resolveTier(qty: number, tiers: PriceTier[]): PriceTier | null {
  let best: PriceTier | null = null

  for (const tier of tiers) {
    if (tier.minQty > qty) continue
    if (
      !best ||
      tier.minQty > best.minQty ||
      (tier.minQty === best.minQty && tier.price < best.price)
    ) {
      best = tier
    }
  }

  return best
}

export function upcomingTier(qty: number, tiers: PriceTier[]): PriceTier | null {
  let next: PriceTier | null = null

  for (const tier of tiers) {
    if (tier.minQty <= qty) continue
    if (
      !next ||
      tier.minQty < next.minQty ||
      (tier.minQty === next.minQty && tier.price < next.price)
    ) {
      next = tier
    }
  }

  return next
}

export function quoteQuantity(input: {
  collectedQty: number
  yourQty: number
  minimumQty: number
  retailPrice: number
  tiers: PriceTier[]
}): QuantityQuote {
  const yourQty = Math.max(0, Math.round(input.yourQty))
  const collectedQty = Math.max(0, input.collectedQty)
  const pooledQty = collectedQty + yourQty
  const tier = resolveTier(pooledQty, input.tiers)
  const unitPrice = tier ? tier.price : null
  const yourTotal = unitPrice === null ? null : roundMoney(unitPrice * yourQty)
  const retailTotal = roundMoney(input.retailPrice * yourQty)
  const savings = yourTotal === null ? null : roundMoney(retailTotal - yourTotal)

  return {
    yourQty,
    pooledQty,
    tier,
    nextTier: upcomingTier(pooledQty, input.tiers),
    unitPrice,
    yourTotal,
    retailTotal,
    savings,
    shortfallToMinimum: Math.max(0, input.minimumQty - pooledQty),
  }
}

function roundMoney(value: number): number {
  return Math.round(value * 100) / 100
}
