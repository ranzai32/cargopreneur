import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { ProgressBar } from '../components/ProgressBar'
import { StatusBadge } from './OrdersPage'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'
import { quoteQuantity } from '../lib/pricing'
import { daysUntil, formatDaysLeft, formatMoney, formatNumber, formatQuantity } from '../lib/format'
import type { ApiOrder } from '../types/catalog'

export function OrderPage() {
  const { orderId = '' } = useParams()
  const navigate = useNavigate()
  const { user, loadOrder, joinOrder } = useAppState()
  const [order, setOrder] = useState<ApiOrder | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [pending, setPending] = useState(false)

  useEffect(() => {
    let cancel = false
    loadOrder(orderId)
      .then((item) => {
        if (cancel) return
        setOrder(item)
        document.title = `${item.product.name} order · CargoPrener`
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Order not found')
      })
    return () => {
      cancel = true
    }
  }, [loadOrder, orderId])

  if (error || !order) {
    return (
      <PageLayout>
        <Container className="py-section">
          <p role="alert" className="text-danger">{error || 'Loading the order…'}</p>
        </Container>
      </PageLayout>
    )
  }

  const tiers = order.product.priceTiers.map((tier) => ({ minQty: tier.minQty, price: tier.unitPrice }))
  const quote = quoteQuantity({
    collectedQty: order.collectedQty,
    yourQty: quantity,
    minimumQty: order.moq,
    retailPrice: order.product.retailPrice,
    tiers,
  })
  const open = order.status === 'open' && daysUntil(order.deadline.slice(0, 10)) > 0

  async function handleJoin() {
    setPending(true)
    setMessage('')
    try {
      const updated = await joinOrder(orderId, quantity)
      setOrder(updated)
      setMessage(updated.status === 'confirmed' ? 'The minimum is reached. This order is confirmed.' : 'You joined this order.')
    } catch (reason) {
      const fields = fieldErrorsOf(reason)
      setMessage(fields.quantity || (isApiError(reason) ? reason.message : 'Could not join'))
    } finally {
      setPending(false)
    }
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <Link to="/orders" className="text-sm font-semibold text-accent-800">Back to orders</Link>
        <div className="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={order.status} />
              <span className="text-sm text-ink-700">{formatDaysLeft(daysUntil(order.deadline.slice(0, 10)))}</span>
            </div>
            <h1 className="mt-3 font-serif text-4xl leading-tight text-ink-950">{order.product.name}</h1>
            <p className="mt-3 text-ink-700">
              <Link to={`/catalog/${order.product.id}`} className="font-semibold text-accent-800">{order.product.supplier.name}</Link>
              {' · '}{order.product.supplier.region}
            </p>
            <p className="mt-4 text-ink-800">
              {formatNumber(order.participantCount)} businesses have added {formatQuantity(order.collectedQty, order.product.unit)}. The minimum is {formatQuantity(order.moq, order.product.unit)}.
            </p>
            <Card className="mt-6 p-5">
              <ProgressBar current={order.collectedQty} target={order.moq} tiers={tiers} unit={order.product.unit} showScaleNote />
            </Card>
          </div>
          <Card className="h-fit p-5 sm:p-6">
            <h2 className="font-serif text-2xl text-ink-950">Your quantity</h2>
            {order.yourQuantity ? (
              <p className="mt-3 text-ink-700">You joined with {formatQuantity(order.yourQuantity, order.product.unit)}.</p>
            ) : open ? (
              <div className="mt-4">
                {!user ? (
                  <Button onClick={() => navigate('/login', { state: { from: `/orders/${order.id}` } })}>Log in to join</Button>
                ) : (
                  <>
                    <label className="block text-sm font-semibold text-ink-900" htmlFor="join-qty">Quantity</label>
                    <input id="join-qty" type="number" min={1} value={quantity} onChange={(event) => setQuantity(Math.max(1, Number(event.target.value) || 1))} className="mt-1.5 h-11 w-full rounded-md border border-ink-300 px-3" />
                    <p className="mt-3 text-sm text-ink-700">
                      {quote.unitPrice === null
                        ? `${formatQuantity(quote.shortfallToMinimum, order.product.unit)} short of the minimum.`
                        : `${formatMoney(quote.unitPrice)} per ${order.product.unit}. Estimated savings ${formatMoney(quote.savings ?? 0)} versus retail.`}
                    </p>
                    <Button className="mt-4 w-full" disabled={pending} onClick={handleJoin}>
                      {pending ? 'Joining…' : 'Join this order'}
                    </Button>
                  </>
                )}
              </div>
            ) : (
              <p className="mt-3 text-ink-700">
                {order.status === 'confirmed'
                  ? 'This order reached the minimum and is confirmed at the wholesale price.'
                  : 'This order is closed. Nobody has to buy.'}
              </p>
            )}
            {message && <p role="status" className="mt-3 text-sm text-ink-800">{message}</p>}
          </Card>
        </div>
      </Container>
    </PageLayout>
  )
}
