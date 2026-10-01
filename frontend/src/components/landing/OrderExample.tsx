import { useState } from 'react'
import { exampleOrder, exampleProduct, exampleQuantity, exampleSupplier } from '../../mocks/exampleOrder'
import { formatMoney, formatQuantity } from '../../lib/format'
import { quoteQuantity } from '../../lib/pricing'
import type { QuantityQuote } from '../../lib/pricing'
import { Card } from '../Card'
import { Container } from '../Container'
import { OrderPreview } from '../OrderPreview'
import { Section } from '../Section'

export function OrderExample() {
  const [qty, setQty] = useState(exampleQuantity.initial)
  const quote = quoteQuantity({
    collectedQty: exampleOrder.collectedQty,
    yourQty: qty,
    minimumQty: exampleProduct.moq,
    retailPrice: exampleProduct.retailPrice,
    tiers: exampleProduct.tiers,
  })
  const summary = priceSummary(quote, exampleProduct.unit)
  const shortOfMinimum = Math.max(0, exampleProduct.moq - exampleOrder.collectedQty)

  function setClamped(next: number) {
    const clamped = Math.min(exampleQuantity.max, Math.max(exampleQuantity.min, next))
    setQty(clamped)
  }

  return (
    <Section id="example" labelledBy="example-heading">
      <Container>
        <div className="grid items-start gap-10 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)] lg:gap-16">
          <div>
            <h2
              id="example-heading"
              className="font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
            >
              See the price change before you join
            </h2>
            <p className="mt-4 text-ink-700">
              This card uses example data. It is not a live order. Change the quantity
              to see which wholesale tier the pool would reach, and what that quantity
              would save against the example retail price.
            </p>
            <p className="mt-4 text-ink-700">
              A quantity under {formatQuantity(shortOfMinimum, exampleProduct.unit)} leaves
              this pool short of the minimum, so no wholesale tier applies. A higher
              quantity can reach a lower unit price.
            </p>
          </div>
          <Card className="p-5 sm:p-6">
            <OrderPreview
              supplier={exampleSupplier}
              product={exampleProduct}
              order={exampleOrder}
              currentQty={quote.pooledQty}
              showScaleNote
            />
            <div className="mt-6 border-t border-ink-200 pt-5">
              <div className="flex items-baseline justify-between gap-4">
                <label htmlFor="example-quantity" className="font-semibold text-ink-950">
                  Your quantity
                </label>
                <span className="tabular-nums text-sm text-ink-700">
                  {formatQuantity(qty, exampleProduct.unit)}
                </span>
              </div>
              <div className="mt-3 flex items-center gap-3">
                <button
                  type="button"
                  className="inline-flex size-11 shrink-0 items-center justify-center rounded-md border border-ink-300 text-lg text-ink-900 hover:bg-ink-50 disabled:opacity-40"
                  aria-label="Decrease your quantity"
                  disabled={qty <= exampleQuantity.min}
                  onClick={() => setClamped(qty - 1)}
                >
                  −
                </button>
                <input
                  id="example-quantity"
                  className="w-full min-w-0"
                  type="range"
                  min={exampleQuantity.min}
                  max={exampleQuantity.max}
                  step={1}
                  value={qty}
                  aria-valuemin={exampleQuantity.min}
                  aria-valuemax={exampleQuantity.max}
                  aria-valuenow={qty}
                  aria-valuetext={formatQuantity(qty, exampleProduct.unit)}
                  onChange={(event) => setClamped(Number(event.target.value))}
                />
                <button
                  type="button"
                  className="inline-flex size-11 shrink-0 items-center justify-center rounded-md border border-ink-300 text-lg text-ink-900 hover:bg-ink-50 disabled:opacity-40"
                  aria-label="Increase your quantity"
                  disabled={qty >= exampleQuantity.max}
                  onClick={() => setClamped(qty + 1)}
                >
                  +
                </button>
              </div>
              <div className="mt-5 grid gap-3 sm:grid-cols-3">
                <QuoteStat
                  label="Unit price"
                  value={quote.unitPrice === null ? 'Not yet' : formatMoney(quote.unitPrice)}
                />
                <QuoteStat
                  label="Your total"
                  value={quote.yourTotal === null ? 'Not yet' : formatMoney(quote.yourTotal)}
                />
                <QuoteStat
                  label="Estimated savings"
                  value={savingsLabel(quote)}
                  emphasis={quote.savings !== null && quote.savings > 0}
                />
              </div>
              <p className="mt-4 text-sm leading-relaxed text-ink-700" aria-live="polite">
                {summary}
              </p>
              <p className="mt-2 text-sm text-ink-700">
                Example retail price: {formatMoney(exampleProduct.retailPrice)} per{' '}
                {exampleProduct.unit}
                {quote.yourQty > 0
                  ? ` (${formatMoney(quote.retailTotal)} for your quantity).`
                  : '.'}
              </p>
              {quote.nextTier && (
                <p className="mt-2 text-sm text-ink-700">
                  {formatQuantity(quote.nextTier.minQty - quote.pooledQty, exampleProduct.unit)}{' '}
                  more in the pool would reach {formatMoney(quote.nextTier.price)} per{' '}
                  {exampleProduct.unit}.
                </p>
              )}
            </div>
          </Card>
        </div>
      </Container>
    </Section>
  )
}

function QuoteStat({
  label,
  value,
  emphasis = false,
}: {
  label: string
  value: string
  emphasis?: boolean
}) {
  return (
    <div className="rounded-md bg-ink-50 px-3 py-3">
      <p className="text-sm text-ink-600">{label}</p>
      <p
        className={
          emphasis
            ? 'mt-1 font-semibold tabular-nums text-accent-800'
            : 'mt-1 font-semibold tabular-nums text-ink-950'
        }
      >
        {value}
      </p>
    </div>
  )
}

function priceSummary(quote: QuantityQuote, unit: string): string {
  if (quote.unitPrice === null || quote.savings === null) {
    return `No wholesale tier yet. ${formatQuantity(quote.shortfallToMinimum, unit)} short of the minimum.`
  }
  if (quote.savings > 0) {
    return `Wholesale price ${formatMoney(quote.unitPrice)} per ${unit}. Estimated savings ${formatMoney(quote.savings)} versus retail.`
  }
  if (quote.savings < 0) {
    return `Wholesale price ${formatMoney(quote.unitPrice)} per ${unit}, above the example retail price.`
  }
  return `Wholesale price ${formatMoney(quote.unitPrice)} per ${unit}, the same as the example retail price.`
}

function savingsLabel(quote: QuantityQuote): string {
  if (quote.savings === null) return 'Not yet'
  if (quote.savings > 0) return formatMoney(quote.savings)
  if (quote.savings < 0) return `${formatMoney(Math.abs(quote.savings))} more`
  return formatMoney(0)
}
