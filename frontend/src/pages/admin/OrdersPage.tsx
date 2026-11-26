import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../components/Button'
import { Card } from '../../components/Card'
import { isApiError, useAppState } from '../../lib/app-state'
import { formatDeliveryStatus, formatNumber } from '../../lib/format'
import type { ApiOrder } from '../../types/catalog'
import { StatusBadge } from '../OrdersPage'
import { AdminFrame } from './AdminFrame'

export function AdminOrdersPage() {
  return (
    <AdminFrame
      title="Orders"
      lede="Moderate pooled orders. Cancel one while it is still open or confirmed. The supplier ships and marks delivery."
    >
      <OrderManager />
    </AdminFrame>
  )
}

function OrderManager() {
  const { loadAdminOrders, cancelOrder } = useAppState()
  const [orders, setOrders] = useState<ApiOrder[]>([])
  const [loaded, setLoaded] = useState(false)
  const [message, setMessage] = useState('')
  const [cancelling, setCancelling] = useState('')
  const [blocked, setBlocked] = useState('')

  useEffect(() => {
    let cancel = false
    loadAdminOrders()
      .then((items) => {
        if (!cancel) {
          setOrders(items)
          setLoaded(true)
        }
      })
      .catch((error) => {
        if (cancel) return
        if (isApiError(error) && (error.status === 401 || error.status === 403)) {
          setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
          return
        }
        setMessage(isApiError(error) ? error.message : 'Orders could not be loaded')
        setLoaded(true)
      })
    return () => {
      cancel = true
    }
  }, [loadAdminOrders])

  async function handleCancel(id: string) {
    setCancelling(id)
    setMessage('')
    try {
      const cancelled = await cancelOrder(id)
      setOrders((current) => current.map((item) => (item.id === id ? cancelled : item)))
    } catch (error) {
      if (isApiError(error) && (error.status === 401 || error.status === 403)) {
        setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
        return
      }
      setMessage(isApiError(error) ? error.message : 'Could not cancel the order')
    } finally {
      setCancelling('')
    }
  }

  if (blocked) {
    return <p role="alert" className="text-danger">{blocked}</p>
  }

  return (
    <div>
      {message && <p role="alert" className="mb-4 text-danger">{message}</p>}
      {!loaded && <p className="text-ink-700">Loading orders…</p>}
      {loaded && orders.length === 0 && <p className="text-ink-700">No orders yet.</p>}
      <ul className="grid gap-4">
        {orders.map((order) => {
          const canCancel = order.status === 'open' || order.status === 'confirmed'
          return (
            <li key={order.id}>
              <Card className="p-5 sm:p-6">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <StatusBadge status={order.status} />
                    <h2 className="mt-3 font-serif text-2xl text-ink-950">
                      <Link to={`/orders/${order.id}`} className="rounded-sm hover:text-accent-800">{order.product.name}</Link>
                    </h2>
                    <p className="mt-1 text-sm text-ink-700">
                      {order.product.supplier.name} · {formatNumber(order.collectedQty)} / {formatNumber(order.moq)} {order.product.unit}
                    </p>
                    {order.yourQuantity != null && (
                      <p className="mt-1 text-sm text-ink-700">Your quantity {formatNumber(order.yourQuantity)}</p>
                    )}
                  </div>
                  {canCancel && (
                    <Button variant="secondary" size="sm" disabled={cancelling === order.id} onClick={() => void handleCancel(order.id)}>
                      {cancelling === order.id ? 'Cancelling…' : 'Cancel'}
                    </Button>
                  )}
                </div>
                {(order.status === 'paid' || order.status === 'completed') && (
                  <ul className="mt-4 grid gap-3">
                    {(order.participants ?? []).map((participant) => {
                      return (
                        <li key={participant.id} className="rounded-md border border-ink-200 p-3">
                          <p className="text-sm text-ink-800">
                            {formatNumber(participant.quantity)} {order.product.unit}
                            {participant.shippingAddress ? ` · ${participant.shippingAddress}` : ''}
                            {` · ${participant.paymentStatus === 'paid' ? 'Paid' : 'Unpaid'}`}
                            {` · ${formatDeliveryStatus(participant.deliveryStatus)}`}
                          </p>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </Card>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
