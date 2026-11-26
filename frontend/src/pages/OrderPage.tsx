import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { ProgressBar } from '../components/ProgressBar'
import { AddressField } from '../components/AddressField'
import { StatusBadge } from './OrdersPage'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'
import { quoteQuantity } from '../lib/pricing'
import { daysUntil, formatDaysLeft, formatDeliveryStatus, formatMoney, formatNumber, formatQuantity } from '../lib/format'
import { mediaUrl } from '../lib/media'
import type { ApiOrder } from '../types/catalog'

export function OrderPage() {
  const { orderId = '' } = useParams()
  const navigate = useNavigate()
  const { user, loadOrder, joinOrder, pay, openDispute } = useAppState()
  const [order, setOrder] = useState<ApiOrder | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [address, setAddress] = useState('')
  const [addressError, setAddressError] = useState('')
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [pending, setPending] = useState(false)
  const [disputeReason, setDisputeReason] = useState('')
  const [disputeError, setDisputeError] = useState('')

  useEffect(() => {
    let cancel = false
    loadOrder(orderId)
      .then((item) => {
        if (cancel) return
        setError('')
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
  const canDispute = Boolean(order.yourQuantity) && (order.status === 'confirmed' || order.status === 'paid' || order.status === 'completed')
  const hasOpenDispute = (order.disputes ?? []).some((dispute) => dispute.status === 'open' && dispute.openedBy === user?.id)

  async function handleDispute() {
    setPending(true)
    setMessage('')
    setDisputeError('')
    try {
      const updated = await openDispute(orderId, disputeReason)
      setOrder(updated)
      setDisputeReason('')
      setMessage('The dispute was opened. An admin will review it.')
    } catch (reason) {
      const fields = fieldErrorsOf(reason)
      setDisputeError(fields.reason ?? '')
      setMessage(isApiError(reason) ? reason.message : 'Could not open the dispute')
    } finally {
      setPending(false)
    }
  }

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

  async function handlePay() {
    setPending(true)
    setAddressError('')
    setMessage('')
    try {
      const updated = await pay(orderId, address)
      setOrder(updated)
    } catch (reason) {
      const fields = fieldErrorsOf(reason)
      if (fields.shippingAddress) {
        setAddressError(fields.shippingAddress)
      } else {
        setMessage(isApiError(reason) ? reason.message : 'Could not record payment')
      }
    } finally {
      setPending(false)
    }
  }

  const joined = Boolean(order.yourQuantity)
  const paid = order.yourPaymentStatus === 'paid'
  const canPay = joined && order.status === 'confirmed' && !paid
  const delivered = order.yourDeliveryStatus === 'delivered' || order.status === 'completed'
  const steps = orderSteps({ joined, open, canPay, paid, delivered, closed: !open && !joined })

  return (
    <PageLayout>
      <Container className="py-section">
        <Link to="/orders" className="text-sm font-semibold text-accent-800">Back to orders</Link>
        <ol className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {steps.map((step, index) => (
            <li key={step.title} className={`rounded-xl border px-4 py-3 ${step.tone}`}>
              <p className="text-xs font-semibold uppercase tracking-wide text-ink-600">Step {index + 1}</p>
              <p className="mt-1 font-semibold text-ink-950">{step.title}</p>
              <p className="mt-1 text-sm text-ink-700">{step.detail}</p>
            </li>
          ))}
        </ol>
        <div className="mt-8 grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={order.status} />
              <span className="text-sm text-ink-700">{formatDaysLeft(daysUntil(order.deadline.slice(0, 10)))}</span>
            </div>
            {order.product.imageUrl && (
              <img src={mediaUrl(order.product.imageUrl)} alt="" className="mt-4 aspect-[4/3] max-w-sm rounded-xl object-cover" />
            )}
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
            {canPay ? (
              <>
                <h2 className="font-serif text-2xl text-ink-950">Delivery and payment</h2>
                <p className="mt-3 text-ink-700">You joined with {formatQuantity(order.yourQuantity ?? 0, order.product.unit)}. The pool reached the minimum, so the address field is open.</p>
                <div className="mt-4">
                  <AddressField
                    id="shipping-address"
                    label="Shipping address"
                    value={address}
                    error={addressError}
                    onChange={setAddress}
                  />
                  <p className="mt-3 text-sm text-ink-700">Pay records a simulated payment to this address. No card is charged.</p>
                  <Button className="mt-4 w-full" disabled={pending || address.trim().length < 8} onClick={handlePay}>
                    {pending ? 'Paying…' : 'Pay and send this address'}
                  </Button>
                </div>
              </>
            ) : paid ? (
              <>
                <h2 className="font-serif text-2xl text-ink-950">Delivery</h2>
                <p className="mt-3 text-ink-700">You joined with {formatQuantity(order.yourQuantity ?? 0, order.product.unit)}. Payment is recorded.</p>
                <p className="mt-3 text-ink-800">The supplier ships to the address you entered. Status: {formatDeliveryStatus(order.yourDeliveryStatus)}.</p>
              </>
            ) : joined && open ? (
              <>
                <h2 className="font-serif text-2xl text-ink-950">Waiting for the pool</h2>
                <p className="mt-3 text-ink-700">You joined with {formatQuantity(order.yourQuantity ?? 0, order.product.unit)}.</p>
                <p className="mt-3 text-sm text-ink-700">The address field stays closed until the collected quantity reaches the minimum. OpenStreetMap suggestions appear on that step, not while the order is still open.</p>
              </>
            ) : open ? (
              <>
                <h2 className="font-serif text-2xl text-ink-950">Add your quantity</h2>
                <div className="mt-4">
                  {!user ? (
                    <Button onClick={() => navigate('/login', { state: { from: `/orders/${order.id}` } })}>Log in to join</Button>
                  ) : (
                    <>
                      <label className="block text-sm font-semibold text-ink-900" htmlFor="join-qty">Quantity</label>
                      <input id="join-qty" type="number" min={1} value={quantity} onChange={(event) => setQuantity(Math.max(1, Number(event.target.value) || 1))} className="mt-1.5 h-11 w-full rounded-md border border-ink-300 px-3" />
                      <p className="mt-3 text-sm text-ink-700">
                        {quote.unitPrice === null
                          ? `${formatQuantity(quote.shortfallToMinimum, order.product.unit)} still needed before anyone pays.`
                          : `${formatMoney(quote.unitPrice)} per ${order.product.unit}. Estimated savings ${formatMoney(quote.savings ?? 0)} versus retail.`}
                      </p>
                      <Button className="mt-4 w-full" disabled={pending} onClick={handleJoin}>
                        {pending ? 'Joining…' : 'Join this order'}
                      </Button>
                    </>
                  )}
                </div>
              </>
            ) : (
              <>
                <h2 className="font-serif text-2xl text-ink-950">This order is closed</h2>
                <p className="mt-3 text-ink-700">{closedCopy(order.status)}</p>
              </>
            )}
            {message && <p role="status" className="mt-3 text-sm text-ink-800">{message}</p>}
          </Card>
        </div>
        {(canDispute || (order.disputes ?? []).length > 0) && <Card className="mt-8 p-5 sm:p-6">
          <h2 className="font-serif text-2xl text-ink-950">Disputes</h2>
          {(order.disputes ?? []).length === 0 && (
            <p className="mt-3 text-sm text-ink-700">No disputes on this order.</p>
          )}
          <ul className="mt-4 grid gap-3">
            {(order.disputes ?? []).map((dispute) => (
              <li key={dispute.id} className="rounded-md border border-ink-200 p-3">
                <p className="text-sm font-semibold text-ink-900">{dispute.businessName || 'A business'} · {dispute.status === 'open' ? 'Open' : 'Resolved'}</p>
                <p className="mt-1 text-sm text-ink-800">{dispute.reason}</p>
                {dispute.outcome && <p className="mt-2 text-sm text-ink-700">Outcome: {dispute.outcome}</p>}
              </li>
            ))}
          </ul>
          {canDispute && !hasOpenDispute && (
            <div className="mt-4">
              <label className="block text-sm font-semibold text-ink-900" htmlFor="dispute-reason">What went wrong</label>
              <textarea
                id="dispute-reason"
                value={disputeReason}
                rows={3}
                className={`mt-1.5 w-full rounded-md border bg-white px-3 py-2 text-base text-ink-900 ${disputeError ? 'border-danger' : 'border-ink-300'}`}
                onChange={(event) => setDisputeReason(event.target.value)}
              />
              {disputeError && <p className="mt-1 text-sm text-danger">{disputeError}</p>}
              <Button className="mt-3" disabled={pending} onClick={() => void handleDispute()}>
                {pending ? 'Opening…' : 'Open a dispute'}
              </Button>
            </div>
          )}
        </Card>}
      </Container>
    </PageLayout>
  )
}

function orderSteps(input: {
  joined: boolean
  open: boolean
  canPay: boolean
  paid: boolean
  delivered: boolean
  closed: boolean
}) {
  const quantity = input.joined ? 'done' : input.open ? 'current' : 'muted'
  const address = input.paid || input.canPay ? (input.paid ? 'done' : 'current') : 'muted'
  const payment = input.paid ? 'done' : input.canPay ? 'current' : 'muted'
  const delivery = input.delivered ? 'done' : input.paid ? 'current' : 'muted'
  return [
    { title: 'Quantity', detail: input.joined ? 'Your quantity is in the pool.' : input.closed ? 'Joining is closed.' : 'Add only what you need.', tone: stepTone(quantity) },
    { title: 'Address', detail: input.canPay ? 'OpenStreetMap suggests the street here.' : input.paid ? 'Address sent with payment.' : 'Opens after the minimum is reached.', tone: stepTone(address) },
    { title: 'Payment', detail: input.paid ? 'Simulated payment recorded.' : 'No card is charged.', tone: stepTone(payment) },
    { title: 'Delivery', detail: input.delivered ? 'Delivered.' : input.paid ? 'Waiting on the supplier.' : 'The supplier updates this.', tone: stepTone(delivery) },
  ]
}

function stepTone(state: 'done' | 'current' | 'muted') {
  if (state === 'done') return 'border-accent-300 bg-success-soft'
  if (state === 'current') return 'border-accent-700 bg-white'
  return 'border-ink-200 bg-ink-50'
}

function closedCopy(status: ApiOrder['status']) {
  if (status === 'completed') return 'Every delivery on this order is finished.'
  if (status === 'paid') return 'Everyone has paid. The supplier is shipping.'
  if (status === 'confirmed') return 'The minimum was reached. Participants can pay from their own accounts.'
  return 'The deadline passed before the minimum. Nobody has to buy.'
}
