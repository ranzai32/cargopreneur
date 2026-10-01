import type { Order, Product, Supplier } from '../types'
import { daysUntil, formatDaysLeft, formatNumber, formatQuantity } from '../lib/format'
import { Badge } from './Badge'
import { ProgressBar } from './ProgressBar'

const statusLabel = {
  open: 'Open',
  confirmed: 'Confirmed',
  cancelled: 'Cancelled',
} as const

const statusTone = {
  open: 'accent',
  confirmed: 'success',
  cancelled: 'neutral',
} as const

type OrderPreviewProps = {
  supplier: Supplier
  product: Product
  order: Order
  currentQty: number
  showScaleNote?: boolean
}

export function OrderPreview({
  supplier,
  product,
  order,
  currentQty,
  showScaleNote = false,
}: OrderPreviewProps) {
  const stillNeeded = Math.max(0, product.moq - currentQty)
  const days = formatDaysLeft(daysUntil(order.deadline))

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2">
        <Badge>Example</Badge>
        <Badge tone={statusTone[order.status]}>{statusLabel[order.status]}</Badge>
        {supplier.moderated && <Badge tone="success">Moderated supplier</Badge>}
      </div>
      <h3 className="mt-4 font-serif text-2xl leading-tight text-ink-950">{product.name}</h3>
      <p className="mt-1 text-sm text-ink-700">
        {supplier.name} · {supplier.location}
      </p>
      <p className="mt-5 font-serif text-4xl tabular-nums leading-none text-ink-950">
        {formatNumber(currentQty)}
        <span className="text-xl text-ink-600"> / {formatNumber(product.moq)}</span>
      </p>
      <p className="mt-2 text-sm text-ink-700">
        {stillNeeded > 0
          ? `${formatQuantity(stillNeeded, product.unit)} still needed to reach the minimum`
          : 'Minimum reached at this quantity'}
      </p>
      <div className="mt-4">
        <ProgressBar
          current={currentQty}
          target={product.moq}
          tiers={product.tiers}
          unit={product.unit}
          showScaleNote={showScaleNote}
        />
      </div>
      {currentQty !== order.collectedQty && (
        <p className="mt-3 text-sm text-ink-700">
          The total above includes your quantity. The business count does not.
        </p>
      )}
      <dl className="mt-5 grid gap-4 border-t border-ink-200 pt-4 sm:grid-cols-2">
        <div>
          <dt className="text-sm text-ink-600">Businesses joined</dt>
          <dd className="mt-1 font-semibold tabular-nums text-ink-950">
            {formatNumber(order.participantCount)}
          </dd>
        </div>
        <div>
          <dt className="text-sm text-ink-600">Time left</dt>
          <dd className="mt-1 font-semibold text-ink-950">{days}</dd>
        </div>
      </dl>
    </div>
  )
}
