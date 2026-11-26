import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { CheckoutFrame } from '../components/CheckoutFrame'
import { Container } from '../components/Container'
import { isApiError, useAppState } from '../lib/app-state'
import { deliveryChoice } from '../lib/delivery'
import { formatMoney, formatQuantity } from '../lib/format'
import { quoteQuantity } from '../lib/pricing'
import type { ApiOrder } from '../types/catalog'

export function PaymentProcessingPage() {
  const { orderId = '' } = useParams()
  const navigate = useNavigate()
  const { loadOrder } = useAppState()
  const [order, setOrder] = useState<ApiOrder | null>(null)
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    document.title = 'Placing your order · CargoPrener'
    let cancel = false
    loadOrder(orderId)
      .then((item) => {
        if (cancel) return
        if (item.yourPaymentStatus !== 'paid') {
          const checkout = item.status === 'confirmed' && item.yourQuantity
          navigate(checkout ? `/orders/${orderId}/checkout` : `/orders/${orderId}`, { replace: true })
          return
        }
        setOrder(item)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Order not found')
      })
    return () => {
      cancel = true
    }
  }, [loadOrder, navigate, orderId])

  useEffect(() => {
    if (!order) return
    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    const timer = window.setTimeout(() => setDone(true), reduce ? 400 : 2200)
    return () => window.clearTimeout(timer)
  }, [order])

  const receipt = order ? receiptOf(order) : null

  return (
    <CheckoutFrame backTo="/account#orders" backLabel="Your orders">
      <Container>
        <div className="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center text-center">
          {error && <p role="alert" className="text-danger">{error}</p>}
          {!error && !done && (
            <>
              <div className="checkout-spin size-24 rounded-full border-4 border-ink-200 border-t-accent-700" role="status" aria-label="Processing your order" />
              <h1 className="mt-6 font-serif text-3xl text-ink-950">Processing your order</h1>
              <p className="mt-2 text-ink-700">Recording the simulated payment and the delivery you chose.</p>
            </>
          )}
          {!error && done && receipt && (
            <>
              <div className="flex size-24 items-center justify-center rounded-full bg-success text-white" role="status" aria-label="Payment recorded">
                <svg viewBox="0 0 24 24" className="size-12" aria-hidden="true">
                  <path d="M5 12.5 9.5 17 19 7.5" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </div>
              <h1 className="mt-6 font-serif text-3xl text-ink-950">Payment recorded</h1>
              <p className="mt-2 text-ink-700">{receipt.summary}</p>
              <p className="mt-1 font-serif text-3xl tabular-nums text-ink-950">{formatMoney(receipt.total)}</p>
              <Link to="/account#orders" className="mt-6 text-base font-semibold text-accent-800">Check details</Link>
            </>
          )}
        </div>
      </Container>
    </CheckoutFrame>
  )
}

function receiptOf(order: ApiOrder) {
  const quantity = order.yourQuantity ?? 0
  const quote = quoteQuantity({
    collectedQty: Math.max(0, order.collectedQty - quantity),
    yourQty: quantity,
    minimumQty: order.moq,
    retailPrice: order.product.retailPrice,
    tiers: order.product.priceTiers.map((tier) => ({ minQty: tier.minQty, price: tier.unitPrice })),
  })
  const delivery = deliveryChoice(order.yourDeliveryMethod, quantity || 1)
  const shipping = order.yourShippingFee ?? delivery?.price ?? 0
  const goods = quote.yourTotal ?? 0
  const name = delivery?.name ?? 'Delivery'
  return {
    total: goods + shipping,
    summary: `${order.product.name}, ${formatQuantity(quantity, order.product.unit)}, ${name}.`,
  }
}
