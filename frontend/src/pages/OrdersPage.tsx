import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageHeader } from '../components/PageHeader'
import { PageLayout } from '../components/PageLayout'
import { ProgressBar } from '../components/ProgressBar'
import { isApiError, useAppState } from '../lib/app-state'
import { formatDaysLeft, daysUntil, formatNumber } from '../lib/format'
import type { ApiOrder } from '../types/catalog'
import type { PriceTier } from '../types'

const statuses = [
  { value: 'open', label: 'Open' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'cancelled', label: 'Cancelled' },
  { value: 'all', label: 'All' },
]

export function OrdersPage() {
  const { loadOrders } = useAppState()
  const [params, setParams] = useSearchParams()
  const status = params.get('status') ?? 'open'
  const [orders, setOrders] = useState<ApiOrder[]>([])
  const [loaded, setLoaded] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    document.title = 'Open orders · CargoPrener'
  }, [])

  useEffect(() => {
    let cancel = false
    setLoaded(false)
    loadOrders(status)
      .then((items) => {
        if (!cancel) {
          setOrders(items)
          setLoaded(true)
        }
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Orders could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [loadOrders, status])

  return (
    <PageLayout>
      <Container className="py-section">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
          <PageHeader
            title="Open orders"
            lede="Join with only the quantity you need. If the minimum is reached before the deadline, everyone pays the wholesale price."
          />
          <Button to="/orders/new">Create an order</Button>
        </div>
        <div className="mt-6 flex flex-wrap gap-2" role="tablist" aria-label="Order status">
          {statuses.map((item) => (
            <button
              key={item.value}
              type="button"
              aria-pressed={status === item.value}
              className={status === item.value ? 'rounded-md bg-accent-700 px-3 py-2 text-sm font-semibold text-white' : 'rounded-md border border-ink-300 bg-white px-3 py-2 text-sm font-semibold text-ink-800'}
              onClick={() => setParams(item.value === 'open' ? {} : { status: item.value })}
            >
              {item.label}
            </button>
          ))}
        </div>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {!error && !loaded && <p className="mt-8 text-ink-700">Loading orders…</p>}
        {!error && loaded && orders.length === 0 && <p className="mt-8 text-ink-700">No orders in this list.</p>}
        <ul className="mt-8 grid gap-4">
          {loaded && orders.map((order) => (
            <li key={order.id}>
              <Link to={`/orders/${order.id}`} className="block rounded-xl">
                <Card className="p-5 sm:p-6">
                  <div className="flex flex-wrap items-center gap-2">
                    <StatusBadge status={order.status} />
                    <span className="text-sm text-ink-700">{formatDaysLeft(daysUntil(order.deadline.slice(0, 10)))}</span>
                  </div>
                  <h2 className="mt-3 font-serif text-2xl text-ink-950">{order.product.name}</h2>
                  <p className="mt-1 text-sm text-ink-700">
                    {order.product.supplier.name} · {formatNumber(order.participantCount)} businesses
                  </p>
                  <p className="mt-4 font-serif text-3xl tabular-nums text-ink-950">
                    {formatNumber(order.collectedQty)}
                    <span className="text-lg text-ink-600"> / {formatNumber(order.moq)}</span>
                  </p>
                  <div className="mt-3">
                    <ProgressBar current={order.collectedQty} target={order.moq} tiers={barTiers(order)} unit={order.product.unit} />
                  </div>
                </Card>
              </Link>
            </li>
          ))}
        </ul>
      </Container>
    </PageLayout>
  )
}

export function StatusBadge({ status }: { status: ApiOrder['status'] }) {
  const tone = status === 'confirmed' ? 'success' : status === 'cancelled' ? 'neutral' : 'accent'
  const label = status === 'open' ? 'Open' : status === 'confirmed' ? 'Confirmed' : 'Cancelled'
  return <Badge tone={tone}>{label}</Badge>
}

function barTiers(order: ApiOrder): PriceTier[] {
  return order.product.priceTiers.map((tier) => ({ minQty: tier.minQty, price: tier.unitPrice }))
}
